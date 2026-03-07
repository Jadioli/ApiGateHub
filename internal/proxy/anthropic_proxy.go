package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"apihub/internal/models"
	"apihub/internal/repository"
)

type AnthropicProxy struct {
	resolver   *ModelResolver
	httpClient *http.Client
	logRepo    *repository.LogRepo
}

func NewAnthropicProxy(resolver *ModelResolver, httpClient *http.Client, logRepo *repository.LogRepo) *AnthropicProxy {
	return &AnthropicProxy{
		resolver:   resolver,
		httpClient: httpClient,
		logRepo:    logRepo,
	}
}

type anthropicRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

func (p *AnthropicProxy) HandleMessages(w http.ResponseWriter, r *http.Request, apiKey *models.APIKey) {
	start := time.Now()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var req anthropicRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Model == "" {
		writeJSONError(w, http.StatusBadRequest, "model field is required")
		return
	}

	allResults, err := p.resolver.ResolveAll(apiKey.ID, req.Model, models.ProtocolAnthropic)
	if err != nil {
		status := http.StatusNotFound
		if err == ErrNoPermission {
			status = http.StatusForbidden
		}
		writeJSONError(w, status, err.Error())
		return
	}

	key := LBKey(models.ProtocolAnthropic, req.Model)
	startIdx := p.resolver.lb.NextIndex(key, len(allResults))

	var lastErr error
	for i := 0; i < len(allResults); i++ {
		idx := (startIdx + i) % len(allResults)
		result := allResults[idx]

		modifiedBody := replaceModelInBody(body, result.ActualModel)
		statusCode, respErr := p.forwardRequest(w, r, &result, modifiedBody, req.Stream)

		go p.logRequest(apiKey.ID, result, req.Model, statusCode, time.Since(start), respErr)

		if respErr == nil {
			return
		}

		lastErr = respErr
		log.Printf("[Anthropic Proxy] Provider %s failed for model %s: %v, trying next...",
			result.Provider.Name, result.ActualModel, respErr)
	}

	writeJSONError(w, http.StatusBadGateway, fmt.Sprintf("all providers failed: %v", lastErr))
}

func (p *AnthropicProxy) forwardRequest(w http.ResponseWriter, r *http.Request, result *ResolveResult, body []byte, stream bool) (int, error) {
	url := result.Provider.BaseURL + "/v1/messages"

	proxyReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("x-api-key", result.Provider.APIKey)
	proxyReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(proxyReq)
	if err != nil {
		return 0, fmt.Errorf("request to provider failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		respBody, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, fmt.Errorf("provider returned %d: %s", resp.StatusCode, string(respBody))
	}

	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	if stream && resp.StatusCode == http.StatusOK {
		p.streamResponse(w, resp.Body)
	} else {
		io.Copy(w, resp.Body)
	}

	return resp.StatusCode, nil
}

func (p *AnthropicProxy) streamResponse(w http.ResponseWriter, body io.Reader) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		io.Copy(w, body)
		return
	}

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)
		flusher.Flush()
	}
}

func (p *AnthropicProxy) HandleListModels(w http.ResponseWriter, r *http.Request, apiKey *models.APIKey) {
	modelNames, err := p.resolver.ListAvailableModels(apiKey.ID, models.ProtocolAnthropic)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list models")
		return
	}

	type modelObj struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		DisplayName string `json:"display_name"`
	}

	var data []modelObj
	for _, m := range modelNames {
		data = append(data, modelObj{
			ID:          m,
			Type:        "model",
			DisplayName: m,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":     data,
		"has_more": false,
	})
}

func (p *AnthropicProxy) logRequest(apiKeyID uint, result ResolveResult, requestModel string, statusCode int, duration time.Duration, err error) {
	logEntry := &models.RequestLog{
		APIKeyID:      apiKeyID,
		ProviderID:    result.Provider.ID,
		RequestModel:  requestModel,
		ProviderModel: result.ActualModel,
		StatusCode:    statusCode,
		ResponseTime:  int(duration.Milliseconds()),
	}
	if err != nil {
		logEntry.Error = err.Error()
	}
	p.logRepo.Create(logEntry)
}
