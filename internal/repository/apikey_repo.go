package repository

import (
	"apihub/internal/models"

	"gorm.io/gorm"
)

type APIKeyRepo struct {
	db *gorm.DB
}

func NewAPIKeyRepo(db *gorm.DB) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

func (r *APIKeyRepo) Create(apiKey *models.APIKey) error {
	return r.db.Create(apiKey).Error
}

func (r *APIKeyRepo) FindAll() ([]models.APIKey, error) {
	var keys []models.APIKey
	err := r.db.Preload("ModelConfig").Order("id asc").Find(&keys).Error
	return keys, err
}

func (r *APIKeyRepo) FindByID(id uint) (*models.APIKey, error) {
	var key models.APIKey
	err := r.db.Preload("ModelConfig").First(&key, id).Error
	return &key, err
}

func (r *APIKeyRepo) FindByKey(key string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.Where("key = ?", key).First(&apiKey).Error
	return &apiKey, err
}

func (r *APIKeyRepo) Update(apiKey *models.APIKey) error {
	return r.db.Save(apiKey).Error
}

func (r *APIKeyRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("api_key_id = ?", id).Delete(&models.APIKeyModel{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.APIKey{}, id).Error
	})
}

// --- APIKeyModel (model mapping) operations ---

func (r *APIKeyRepo) FindModelsByKeyID(apiKeyID uint) ([]models.APIKeyModel, error) {
	var ms []models.APIKeyModel
	err := r.db.Preload("Provider").Preload("ProviderModel").
		Where("api_key_id = ?", apiKeyID).Order("mapped_name asc, id asc").Find(&ms).Error
	if err != nil {
		return nil, err
	}
	return ms, nil
}

func (r *APIKeyRepo) CreateModel(m *models.APIKeyModel) error {
	return r.db.Create(m).Error
}

func (r *APIKeyRepo) UpdateModel(m *models.APIKeyModel) error {
	return r.db.Save(m).Error
}

func (r *APIKeyRepo) DeleteModel(id uint) error {
	return r.db.Delete(&models.APIKeyModel{}, id).Error
}

func (r *APIKeyRepo) ReplaceModels(apiKeyID uint, ms []models.APIKeyModel) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("api_key_id = ?", apiKeyID).Delete(&models.APIKeyModel{}).Error; err != nil {
			return err
		}
		if len(ms) == 0 {
			return nil
		}
		for i := range ms {
			ms[i].APIKeyID = apiKeyID
		}
		return tx.Create(&ms).Error
	})
}

// --- Proxy queries ---

// Find all enabled APIKeyModels matching a mapped_name for a specific key + protocol
func (r *APIKeyRepo) FindModelsByMappedName(apiKeyID uint, mappedName string, protocol models.ProviderProtocol) ([]models.APIKeyModel, error) {
	var ms []models.APIKeyModel
	err := r.db.
		Joins("JOIN providers ON providers.id = api_key_models.provider_id").
		Joins("JOIN provider_models ON provider_models.id = api_key_models.provider_model_id").
		Where("api_key_models.api_key_id = ?", apiKeyID).
		Where("api_key_models.mapped_name = ?", mappedName).
		Where("api_key_models.enabled = ?", true).
		Where("providers.enabled = ?", true).
		Where("providers.protocol = ?", protocol).
		Where("provider_models.enabled = ?", true).
		Preload("Provider").
		Preload("ProviderModel").
		Find(&ms).Error
	if err != nil {
		return nil, err
	}
	return ms, nil
}

// Get all distinct mapped names for an API Key + protocol
func (r *APIKeyRepo) FindAllMappedNames(apiKeyID uint, protocol models.ProviderProtocol) ([]string, error) {
	var names []string
	err := r.db.Model(&models.APIKeyModel{}).
		Joins("JOIN providers ON providers.id = api_key_models.provider_id").
		Joins("JOIN provider_models ON provider_models.id = api_key_models.provider_model_id").
		Where("api_key_models.api_key_id = ?", apiKeyID).
		Where("api_key_models.enabled = ?", true).
		Where("providers.enabled = ?", true).
		Where("providers.protocol = ?", protocol).
		Where("provider_models.enabled = ?", true).
		Distinct("api_key_models.mapped_name").
		Pluck("api_key_models.mapped_name", &names).Error
	return names, err
}

// Get grouped view: mapped_name -> []APIKeyModel
func (r *APIKeyRepo) FindModelsGrouped(apiKeyID uint) (map[string][]models.APIKeyModel, error) {
	ms, err := r.FindModelsByKeyID(apiKeyID)
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]models.APIKeyModel)
	for _, m := range ms {
		grouped[m.MappedName] = append(grouped[m.MappedName], m)
	}
	return grouped, nil
}
