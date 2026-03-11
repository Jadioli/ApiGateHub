package services

import (
	"strings"

	"apihub/internal/models"
	"apihub/internal/repository"
)

type ProviderService struct {
	repo *repository.ProviderRepo
}

func NewProviderService(repo *repository.ProviderRepo) *ProviderService {
	return &ProviderService{repo: repo}
}

func (s *ProviderService) Create(provider *models.Provider) error {
	return s.repo.Create(provider)
}

func (s *ProviderService) FindAll() ([]models.Provider, error) {
	return s.repo.FindAll()
}

func (s *ProviderService) FindByID(id uint) (*models.Provider, error) {
	return s.repo.FindByID(id)
}

func (s *ProviderService) FindByIDWithModels(id uint) (*models.Provider, error) {
	return s.repo.FindByIDWithModels(id)
}

func (s *ProviderService) Update(provider *models.Provider) error {
	return s.repo.Update(provider)
}

func (s *ProviderService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *ProviderService) Toggle(id uint) (*models.Provider, error) {
	provider, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	provider.Enabled = !provider.Enabled
	if err := s.repo.Update(provider); err != nil {
		return nil, err
	}
	return provider, nil
}

func (s *ProviderService) GetModels(providerID uint) ([]models.ProviderModel, error) {
	return s.repo.FindModelsByProviderID(providerID)
}

func (s *ProviderService) ToggleModel(modelID uint) (*models.ProviderModel, error) {
	pm, err := s.repo.FindModelByID(modelID)
	if err != nil {
		return nil, err
	}
	pm.Enabled = !pm.Enabled
	if err := s.repo.UpdateModel(pm); err != nil {
		return nil, err
	}
	return pm, nil
}

func (s *ProviderService) UpsertModels(providerID uint, modelNames []string) error {
	return s.repo.UpsertModels(providerID, modelNames)
}

func (s *ProviderService) SetModelsEnabled(providerID uint, enabled bool) error {
	return s.repo.SetModelsEnabled(providerID, enabled)
}

func (s *ProviderService) GetAllTags() ([]string, error) {
	providers, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var tags []string
	for _, p := range providers {
		if p.Tags == "" {
			continue
		}
		for _, tag := range strings.Split(p.Tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" && !seen[tag] {
				seen[tag] = true
				tags = append(tags, tag)
			}
		}
	}
	if tags == nil {
		tags = []string{}
	}
	return tags, nil
}
