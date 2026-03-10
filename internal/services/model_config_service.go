package services

import (
	"fmt"

	"apihub/internal/models"
	"apihub/internal/repository"
)

type ModelConfigService struct {
	repo         *repository.ModelConfigRepo
	providerRepo *repository.ProviderRepo
}

func NewModelConfigService(repo *repository.ModelConfigRepo, providerRepo *repository.ProviderRepo) *ModelConfigService {
	return &ModelConfigService{
		repo:         repo,
		providerRepo: providerRepo,
	}
}

// CRUD 鎿嶄綔

func (s *ModelConfigService) Create(name, description string) (*models.ModelConfig, error) {
	config := &models.ModelConfig{
		Name:        name,
		Description: description,
		Enabled:     true,
	}
	if err := s.repo.Create(config); err != nil {
		return nil, err
	}
	return config, nil
}

func (s *ModelConfigService) FindAll() ([]models.ModelConfig, error) {
	return s.repo.FindAll()
}

func (s *ModelConfigService) FindByID(id uint) (*models.ModelConfig, error) {
	return s.repo.FindByID(id)
}

func (s *ModelConfigService) FindByIDWithItems(id uint) (*models.ModelConfig, error) {
	return s.repo.FindByIDWithItems(id)
}

func (s *ModelConfigService) Update(config *models.ModelConfig) error {
	return s.repo.Update(config)
}

func (s *ModelConfigService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *ModelConfigService) Toggle(id uint) (*models.ModelConfig, error) {
	config, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	config.Enabled = !config.Enabled
	if err := s.repo.Update(config); err != nil {
		return nil, err
	}
	return config, nil
}

// 妯″瀷閰嶇疆椤规搷浣?
func (s *ModelConfigService) GetItems(configID uint) ([]models.ModelConfigItem, error) {
	return s.repo.FindItemsByConfigID(configID)
}

func (s *ModelConfigService) AddItem(item *models.ModelConfigItem) error {
	return s.repo.CreateItem(item)
}

func (s *ModelConfigService) UpdateItem(item *models.ModelConfigItem) error {
	return s.repo.UpdateItem(item)
}

func (s *ModelConfigService) DeleteItem(id uint) error {
	return s.repo.DeleteItem(id)
}

func (s *ModelConfigService) GetItemsGrouped(configID uint) (map[string][]models.ModelConfigItem, error) {
	return s.repo.FindItemsGrouped(configID)
}

func (s *ModelConfigService) BatchReplace(configID uint, items []models.ModelConfigItem) error {
	for i := range items {
		items[i].ModelConfigID = configID
		items[i].Enabled = true
	}
	return s.repo.ReplaceItems(configID, items)
}

// 鍏嬮殕閰嶇疆鏂规

func (s *ModelConfigService) Clone(sourceID uint, newName string) (*models.ModelConfig, error) {
	source, err := s.repo.FindByIDWithItems(sourceID)
	if err != nil {
		return nil, err
	}

	newConfig := &models.ModelConfig{
		Name:        newName,
		Description: fmt.Sprintf("Cloned from %s", source.Name),
		Enabled:     true,
	}

	if err := s.repo.Create(newConfig); err != nil {
		return nil, err
	}

	for _, item := range source.Items {
		newItem := item
		newItem.ID = 0
		newItem.ModelConfigID = newConfig.ID
		if err := s.repo.CreateItem(&newItem); err != nil {
			return nil, err
		}
	}

	return newConfig, nil
}

// GetAllAvailableModels 鑾峰彇鎵€鏈?Provider 鐨勬墍鏈夋ā鍨嬶紝鐢ㄤ簬鍓嶇灞曠ず閰嶇疆鐣岄潰
func (s *ModelConfigService) GetAllAvailableModels() ([]ProviderWithModels, error) {
	providers, err := s.providerRepo.FindAllEnabled()
	if err != nil {
		return nil, err
	}

	result := make([]ProviderWithModels, 0, len(providers))
	for _, provider := range providers {
		providerModels, err := s.providerRepo.FindModelsByProviderID(provider.ID)
		if err != nil {
			continue
		}

		enabledModels := make([]models.ProviderModel, 0, len(providerModels))
		for _, model := range providerModels {
			if model.Enabled {
				enabledModels = append(enabledModels, model)
			}
		}

		result = append(result, ProviderWithModels{
			Provider: provider,
			Models:   enabledModels,
		})
	}

	return result, nil
}

type ProviderWithModels struct {
	Provider models.Provider        `json:"provider"`
	Models   []models.ProviderModel `json:"models"`
}
