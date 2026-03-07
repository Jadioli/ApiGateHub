package services

import (
	"apihub/internal/models"
	"apihub/internal/repository"
	"apihub/pkg"
)

type APIKeyService struct {
	repo       *repository.APIKeyRepo
	configRepo *repository.ModelConfigRepo
}

func NewAPIKeyService(repo *repository.APIKeyRepo, configRepo *repository.ModelConfigRepo) *APIKeyService {
	return &APIKeyService{repo: repo, configRepo: configRepo}
}

type CreateAPIKeyResult struct {
	APIKey *models.APIKey `json:"api_key"`
	RawKey string         `json:"raw_key"`
}

func (s *APIKeyService) Create(name string, template *models.APIKey) (*CreateAPIKeyResult, error) {
	if template.ModelConfigID != nil {
		if _, err := s.configRepo.FindByID(*template.ModelConfigID); err != nil {
			return nil, err
		}
	}

	rawKey, err := pkg.GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	apiKey := &models.APIKey{
		Name:          name,
		Key:           rawKey,
		KeyHint:       pkg.MaskAPIKey(rawKey),
		ModelConfigID: template.ModelConfigID,
		Enabled:       true,
		ExpiresAt:     template.ExpiresAt,
	}
	if err := s.repo.Create(apiKey); err != nil {
		return nil, err
	}

	return &CreateAPIKeyResult{APIKey: apiKey, RawKey: rawKey}, nil
}

func (s *APIKeyService) FindAll() ([]models.APIKey, error) {
	return s.repo.FindAll()
}

func (s *APIKeyService) FindByID(id uint) (*models.APIKey, error) {
	return s.repo.FindByID(id)
}

func (s *APIKeyService) Update(apiKey *models.APIKey) error {
	return s.repo.Update(apiKey)
}

func (s *APIKeyService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *APIKeyService) Toggle(id uint) (*models.APIKey, error) {
	key, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	key.Enabled = !key.Enabled
	if err := s.repo.Update(key); err != nil {
		return nil, err
	}
	return key, nil
}

// --- Model mapping operations ---

func (s *APIKeyService) GetModels(apiKeyID uint) ([]models.APIKeyModel, error) {
	return s.repo.FindModelsByKeyID(apiKeyID)
}

func (s *APIKeyService) AddModel(m *models.APIKeyModel) error {
	return s.repo.CreateModel(m)
}

func (s *APIKeyService) UpdateModel(m *models.APIKeyModel) error {
	return s.repo.UpdateModel(m)
}

func (s *APIKeyService) DeleteModel(id uint) error {
	return s.repo.DeleteModel(id)
}

func (s *APIKeyService) GetModelsGrouped(apiKeyID uint) (map[string][]models.APIKeyModel, error) {
	return s.repo.FindModelsGrouped(apiKeyID)
}

func (s *APIKeyService) BatchReplace(apiKeyID uint, ms []models.APIKeyModel) error {
	for i := range ms {
		ms[i].APIKeyID = apiKeyID
		ms[i].Enabled = true
	}
	return s.repo.ReplaceModels(apiKeyID, ms)
}

// --- ModelConfig operations ---

func (s *APIKeyService) SetModelConfig(apiKeyID uint, configID *uint) error {
	key, err := s.repo.FindByID(apiKeyID)
	if err != nil {
		return err
	}
	if configID != nil {
		if _, err := s.configRepo.FindByID(*configID); err != nil {
			return err
		}
	}
	key.ModelConfigID = configID
	return s.repo.Update(key)
}

func (s *APIKeyService) GetModelConfig(apiKeyID uint) (*models.ModelConfig, error) {
	key, err := s.repo.FindByID(apiKeyID)
	if err != nil {
		return nil, err
	}
	if key.ModelConfigID == nil {
		return nil, nil
	}
	return s.configRepo.FindByIDWithItems(*key.ModelConfigID)
}
