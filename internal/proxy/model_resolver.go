package proxy

import (
	"errors"
	"fmt"

	"apihub/internal/models"
	"apihub/internal/repository"
)

var (
	ErrModelNotFound   = errors.New("model not found")
	ErrNoAvailProvider = errors.New("no available provider for this model")
	ErrNoPermission    = errors.New("no permission to access this model")
)

type ResolveResult struct {
	APIKeyModel   models.APIKeyModel
	Provider      models.Provider
	ProviderModel models.ProviderModel
	ActualModel   string
}

type ModelResolver struct {
	apiKeyRepo      *repository.APIKeyRepo
	modelConfigRepo *repository.ModelConfigRepo
	lb              *LoadBalancer
}

func NewModelResolver(
	apiKeyRepo *repository.APIKeyRepo,
	modelConfigRepo *repository.ModelConfigRepo,
	lb *LoadBalancer,
) *ModelResolver {
	return &ModelResolver{
		apiKeyRepo:      apiKeyRepo,
		modelConfigRepo: modelConfigRepo,
		lb:              lb,
	}
}

// ResolveAll returns all candidates for load balancing + failover
func (r *ModelResolver) ResolveAll(apiKeyID uint, modelName string, protocol models.ProviderProtocol) ([]ResolveResult, error) {
	// 获取 APIKey
	apiKey, err := r.apiKeyRepo.FindByID(apiKeyID)
	if err != nil {
		return nil, err
	}

	// 优先使用 ModelConfig
	if apiKey.ModelConfigID != nil {
		items, err := r.modelConfigRepo.FindItemsByMappedName(*apiKey.ModelConfigID, modelName, protocol)
		if err == nil && len(items) > 0 {
			return r.convertConfigItemsToResults(items), nil
		}
	}

	// 回退到旧的 APIKeyModel（向后兼容）
	candidates, err := r.apiKeyRepo.FindModelsByMappedName(apiKeyID, modelName, protocol)
	if err != nil || len(candidates) == 0 {
		return nil, ErrModelNotFound
	}

	var results []ResolveResult
	for _, c := range candidates {
		results = append(results, ResolveResult{
			APIKeyModel:   c,
			Provider:      c.Provider,
			ProviderModel: c.ProviderModel,
			ActualModel:   c.ProviderModel.ModelName,
		})
	}

	return results, nil
}

func (r *ModelResolver) convertConfigItemsToResults(items []models.ModelConfigItem) []ResolveResult {
	var results []ResolveResult
	for _, item := range items {
		results = append(results, ResolveResult{
			Provider:      item.Provider,
			ProviderModel: item.ProviderModel,
			ActualModel:   item.ProviderModel.ModelName,
		})
	}
	return results
}

// ListAvailableModels returns all mapped model names for an API Key + protocol
func (r *ModelResolver) ListAvailableModels(apiKeyID uint, protocol models.ProviderProtocol) ([]string, error) {
	// 获取 APIKey
	apiKey, err := r.apiKeyRepo.FindByID(apiKeyID)
	if err != nil {
		return nil, err
	}

	// 优先使用 ModelConfig
	if apiKey.ModelConfigID != nil {
		names, err := r.modelConfigRepo.FindAllMappedNames(*apiKey.ModelConfigID, protocol)
		if err == nil && len(names) > 0 {
			return names, nil
		}
	}

	// 回退到旧方式
	return r.apiKeyRepo.FindAllMappedNames(apiKeyID, protocol)
}

// LBKey builds the load balancer key
func LBKey(protocol models.ProviderProtocol, modelName string) string {
	return fmt.Sprintf("%s:%s", protocol, modelName)
}
