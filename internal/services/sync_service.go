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

// intervalToSpec 将 sync_interval 转换为 cron 表达式
func intervalToSpec(interval string) string {
	switch interval {
	case "hourly":
		return "@every 1h"
	case "daily":
		return "@every 24h"
	case "weekly":
		return "@every 168h"
	default:
		return ""
	}
}

// StartPerProviderScheduler 根据每个 Provider 各自的 sync_interval 启动定时同步
func (s *SyncService) StartPerProviderScheduler() {
	s.cron = cron.New()

	providers, err := s.providerRepo.FindAllEnabled()
	if err != nil {
		log.Printf("[Sync] 获取 Provider 列表失败: %v", err)
		return
	}

	count := 0
	for _, p := range providers {
		if p.SyncInterval == "" || p.SyncInterval == "none" {
			continue
		}
		spec := intervalToSpec(p.SyncInterval)
		if spec == "" {
			continue
		}
		pid := p.ID
		pName := p.Name
		s.cron.AddFunc(spec, func() {
			log.Printf("[Sync] 定时同步触发: Provider %s (ID=%d)", pName, pid)
			s.SyncProvider(pid)
		})
		count++
	}

	s.cron.Start()
	log.Printf("[Sync] Per-provider 调度器已启动，共 %d 个 Provider 注册了定时同步", count)
}

// RefreshScheduler 刷新调度器（当 Provider 的同步设置变更时调用）
func (s *SyncService) RefreshScheduler() {
	if s.cron != nil {
		s.cron.Stop()
	}
	s.StartPerProviderScheduler()
}

func (s *SyncService) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

func (s *SyncService) SyncAllProviders() {
	providers, err := s.providerRepo.FindAllEnabled()
	if err != nil {
		log.Printf("[Sync] 获取 Provider 列表失败: %v", err)
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

	// 标记为同步中
	provider.SyncStatus = "syncing"
	provider.SyncError = ""
	s.providerRepo.Update(provider)

	modelNames, err := s.fetchModels(provider)
	if err != nil {
		provider.SyncStatus = "failed"
		provider.SyncError = err.Error()
		s.providerRepo.Update(provider)
		log.Printf("[Sync] Provider %s 同步失败: %v", provider.Name, err)
		return err
	}

	if len(modelNames) == 0 {
		provider.SyncStatus = "failed"
		provider.SyncError = "no models returned from provider"
		s.providerRepo.Update(provider)
		log.Printf("[Sync] Provider %s 返回了 0 个模型，标记为失败", provider.Name)
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

	log.Printf("[Sync] Provider %s 同步成功，共 %d 个模型", provider.Name, len(modelNames))
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

// parseModelIDs 解析模型列表响应
// 同时支持 { "data": [{ "id": "..." }] } 和 { "models": [...] } 格式
func parseModelIDs(body []byte) ([]string, error) {
	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
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

	log.Printf("[Sync] 解析出 %d 个模型 (body preview: %.200s)", len(names), string(body))
	return names, nil
}
