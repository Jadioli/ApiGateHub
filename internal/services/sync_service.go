package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"apihub/internal/models"
	"apihub/internal/repository"

	"github.com/robfig/cron/v3"
)

type SyncService struct {
	providerRepo    *repository.ProviderRepo
	providerService *ProviderService
	httpClient      *http.Client
	cron            *cron.Cron
}

func NewSyncService(providerRepo *repository.ProviderRepo, providerService *ProviderService, httpClient *http.Client) *SyncService {
	return &SyncService{
		providerRepo:    providerRepo,
		providerService: providerService,
		httpClient:      httpClient,
	}
}

func (s *SyncService) StartScheduler(intervalMinutes int) {
	s.cron = cron.New()
	spec := fmt.Sprintf("@every %dm", intervalMinutes)
	s.cron.AddFunc(spec, func() {
		log.Println("[Sync] Starting scheduled sync for all providers")
		s.SyncAllProviders()
	})
	s.cron.Start()
	log.Printf("[Sync] Scheduler started, interval: %d minutes", intervalMinutes)
}

func (s *SyncService) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

func (s *SyncService) SyncAllProviders() {
	providers, err := s.providerRepo.FindAllEnabled()
	if err != nil {
		log.Printf("[Sync] Failed to fetch providers: %v", err)
		return
	}
	for _, p := range providers {
		s.SyncProvider(p.ID)
	}
}

func (s *SyncService) SyncProvider(providerID uint) error {
	provider, err := s.providerRepo.FindByID(providerID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	// Mark as syncing
	provider.SyncStatus = "syncing"
	provider.SyncError = ""
	s.providerRepo.Update(provider)

	modelNames, err := s.fetchModels(provider)
	if err != nil {
		provider.SyncStatus = "failed"
		provider.SyncError = err.Error()
		s.providerRepo.Update(provider)
		log.Printf("[Sync] Failed to sync provider %s: %v", provider.Name, err)
		return err
	}

	if len(modelNames) == 0 {
		provider.SyncStatus = "failed"
		provider.SyncError = "no models returned from provider"
		s.providerRepo.Update(provider)
		log.Printf("[Sync] Provider %s returned 0 models, marked as failed", provider.Name)
		return fmt.Errorf("no models returned from provider")
	}

	if err := s.providerService.UpsertModels(providerID, modelNames); err != nil {
		provider.SyncStatus = "failed"
		provider.SyncError = err.Error()
		s.providerRepo.Update(provider)
		return err
	}

	now := time.Now()
	provider.LastSyncAt = &now
	provider.SyncStatus = "success"
	provider.SyncError = ""
	s.providerRepo.Update(provider)

	log.Printf("[Sync] Provider %s synced successfully, %d models found", provider.Name, len(modelNames))
	return nil
}

func (s *SyncService) fetchModels(provider *models.Provider) ([]string, error) {
	url := provider.BaseURL + "/v1/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	switch provider.Protocol {
	case models.ProtocolOpenAI:
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	case models.ProtocolAnthropic:
		req.Header.Set("x-api-key", provider.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return parseModelIDs(body)
}

// Both OpenAI and Anthropic return { "data": [{ "id": "model-name", ... }] }
// Some providers use { "models": [...] } or { "data": [{ "name": "..." }] }
func parseModelIDs(body []byte) ([]string, error) {
	// Try the most common format first: { "data": [{ "id": "..." }] }
	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
		// Some providers use "models" instead of "data"
		Models []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	seen := make(map[string]struct{})
	var names []string

	addName := func(id, name string) {
		v := id
		if v == "" {
			v = name
		}
		if v != "" {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				names = append(names, v)
			}
		}
	}

	for _, m := range result.Data {
		addName(m.ID, m.Name)
	}
	for _, m := range result.Models {
		addName(m.ID, m.Name)
	}

	log.Printf("[Sync] Parsed %d models from response (body preview: %.200s)", len(names), string(body))
	return names, nil
}
