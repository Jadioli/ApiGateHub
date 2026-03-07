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

type OpenAIProxy struct {
	resolver   *ModelResolver
	httpClient *http.Client
	logRepo    *repository.LogRepo
}

func NewOpenAIProxy(resolver *ModelResolver, httpClient *http.Client, logRepo *repository.LogRepo) *OpenAIProxy {
	return &OpenAIProxy{
		resolver:   resolver,
		httpClient: httpClient,
		logRepo:    logRepo,
	}
}

type openAIRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

func (p *OpenAIProxy) HandleChatCompletions(w http.ResponseWriter, r *http.Request, apiKey *models.APIKey) {
	p.handleRequest(w, r, apiKey, "/v1/chat/completions")
}

func (p *OpenAIProxy) HandleCompletions(w http.ResponseWriter, r *http.Request, apiKey *models.APIKey) {
	p.handleRequest(w, r, apiKey, "/v1/completions")
}

func (p *OpenAIProxy) handleRequest(w http.ResponseWriter, r *http.Request, apiKey *models.APIKey, upstreamPath string) {
	start := time.Now()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var req openAIRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Model == "" {
		writeJSONError(w, http.StatusBadRequest, "model field is required")
		return
	}

	allResults, err := p.resolver.ResolveAll(apiKey.ID, req.Model, models.ProtocolOpenAI)
	if err != nil {
		status := http.StatusNotFound
		if err == ErrNoPermission {
			status = http.StatusForbidden
		}
		writeJSONError(w, status, err.Error())
		return
	}

	key := LBKey(models.ProtocolOpenAI, req.Model)
	startIdx := p.resolver.lb.NextIndex(key, len(allResults))

	var lastErr error
	for i := 0; i < len(allResults); i++ {
		idx := (startIdx + i) % len(allResults)
		result := allResults[idx]

		modifiedBody := replaceModelInBody(body, result.ActualModel)
		statusCode, respErr := p.forwardRequest(w, r, &result, modifiedBody, req.Stream, upstreamPath)

		go p.logRequest(apiKey.ID, result, req.Model, statusCode, time.Since(start), respErr)

		if respErr == nil {
			return
		}

		lastErr = respErr
		log.Printf("[OpenAI Proxy] Provider %s failed for model %s: %v, trying next...",
			result.Provider.Name, result.ActualModel, respErr)
	}

	writeJSONError(w, http.StatusBadGateway, fmt.Sprintf("all providers failed: %v", lastErr))
}

func (p *OpenAIProxy) forwardRequest(w http.ResponseWriter, r *http.Request, result *ResolveResult, body []byte, stream bool, upstreamPath string) (int, error) {
	url := result.Provider.BaseURL + upstreamPath

	proxyReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Authorization", "Bearer "+result.Provider.APIKey)

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

func (p *OpenAIProxy) streamResponse(w http.ResponseWriter, body io.Reader) {
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

func (p *OpenAIProxy) HandleListModels(w http.ResponseWriter, r *http.Request, apiKey *models.APIKey) {
	modelNames, err := p.resolver.ListAvailableModels(apiKey.ID, models.ProtocolOpenAI)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list models")
		return
	}

	type modelObj struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	var data []modelObj
	for _, m := range modelNames {
		data = append(data, modelObj{
			ID:      m,
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "apihub",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   data,
	})
}

func (p *OpenAIProxy) logRequest(apiKeyID uint, result ResolveResult, requestModel string, statusCode int, duration time.Duration, err error) {
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

func replaceModelInBody(body []byte, newModel string) []byte {
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return body
	}
	m["model"] = newModel
	modified, err := json.Marshal(m)
	if err != nil {
		return body
	}
	return modified
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    "apihub_error",
		},
	})
}
