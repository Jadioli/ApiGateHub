package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
		if err := s.SyncAllProviders(); err != nil {
			log.Printf("[Sync] Scheduled sync finished with errors: %v", err)
		}
	})
	s.cron.Start()
	log.Printf("[Sync] Scheduler started, interval: %d minutes", intervalMinutes)
}

func (s *SyncService) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

func (s *SyncService) SyncAllProviders() error {
	providers, err := s.providerRepo.FindAllEnabled()
	if err != nil {
		return fmt.Errorf("failed to fetch providers: %w", err)
	}

	var failed []string
	for _, p := range providers {
		if err := s.SyncProvider(p.ID); err != nil {
			failed = append(failed, fmt.Sprintf("%s(%d): %v", p.Name, p.ID, err))
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("sync failed for %d provider(s): %s", len(failed), strings.Join(failed, "; "))
	}
	return nil
}

func (s *SyncService) SyncProvider(providerID uint) error {
	provider, err := s.providerRepo.FindByID(providerID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	// Mark as syncing
	if err := s.updateSyncState(providerID, "syncing", "", nil); err != nil {
		log.Printf("[Sync] Failed to mark provider %s as syncing: %v", provider.Name, err)
	}

	modelNames, err := s.fetchModels(provider)
	if err != nil {
		if updateErr := s.updateSyncState(providerID, "failed", err.Error(), nil); updateErr != nil {
			log.Printf("[Sync] Failed to update failed status for provider %s: %v", provider.Name, updateErr)
		}
		log.Printf("[Sync] Failed to sync provider %s: %v", provider.Name, err)
		return err
	}

	if len(modelNames) == 0 {
		const syncError = "no models returned from provider"
		if updateErr := s.updateSyncState(providerID, "failed", syncError, nil); updateErr != nil {
			log.Printf("[Sync] Failed to update failed status for provider %s: %v", provider.Name, updateErr)
		}
		log.Printf("[Sync] Provider %s returned 0 models, marked as failed", provider.Name)
		return fmt.Errorf("no models returned from provider")
	}

	if err := s.providerService.UpsertModels(providerID, modelNames); err != nil {
		if updateErr := s.updateSyncState(providerID, "failed", err.Error(), nil); updateErr != nil {
			log.Printf("[Sync] Failed to update failed status for provider %s: %v", provider.Name, updateErr)
		}
		return err
	}

	now := time.Now()
	if err := s.updateSyncState(providerID, "success", "", &now); err != nil {
		log.Printf("[Sync] Models synced for provider %s but failed to persist success status: %v", provider.Name, err)
		return fmt.Errorf("models synced but failed to persist sync status: %w", err)
	}

	log.Printf("[Sync] Provider %s synced successfully, %d models found", provider.Name, len(modelNames))
	return nil
}

func (s *SyncService) updateSyncState(providerID uint, status string, syncError string, lastSyncAt *time.Time) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		lastErr = s.providerRepo.UpdateSyncState(providerID, status, syncError, lastSyncAt)
		if lastErr == nil {
			return nil
		}
		time.Sleep(time.Duration(attempt+1) * 120 * time.Millisecond)
	}
	return lastErr
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
