package repository

import (
	"apihub/internal/models"

	"gorm.io/gorm"
)

type ModelConfigRepo struct {
	db *gorm.DB
}

func NewModelConfigRepo(db *gorm.DB) *ModelConfigRepo {
	return &ModelConfigRepo{db: db}
}

// ModelConfig CRUD

func (r *ModelConfigRepo) Create(config *models.ModelConfig) error {
	return r.db.Create(config).Error
}

func (r *ModelConfigRepo) FindAll() ([]models.ModelConfig, error) {
	var configs []models.ModelConfig
	err := r.db.Order("id asc").Find(&configs).Error
	return configs, err
}

func (r *ModelConfigRepo) FindByID(id uint) (*models.ModelConfig, error) {
	var config models.ModelConfig
	err := r.db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *ModelConfigRepo) FindByIDWithItems(id uint) (*models.ModelConfig, error) {
	var config models.ModelConfig
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("mapped_name asc, priority desc")
	}).Preload("Items.Provider").Preload("Items.ProviderModel").First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *ModelConfigRepo) Update(config *models.ModelConfig) error {
	return r.db.Save(config).Error
}

func (r *ModelConfigRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.APIKey{}).Where("model_config_id = ?", id).Update("model_config_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Where("model_config_id = ?", id).Delete(&models.ModelConfigItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.ModelConfig{}, id).Error
	})
}

// ModelConfigItem operations

func (r *ModelConfigRepo) FindItemsByConfigID(configID uint) ([]models.ModelConfigItem, error) {
	var items []models.ModelConfigItem
	err := r.db.Where("model_config_id = ?", configID).
		Preload("Provider").
		Preload("ProviderModel").
		Order("mapped_name asc, priority desc").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ModelConfigRepo) CreateItem(item *models.ModelConfigItem) error {
	return r.db.Create(item).Error
}

func (r *ModelConfigRepo) UpdateItem(item *models.ModelConfigItem) error {
	return r.db.Save(item).Error
}

func (r *ModelConfigRepo) DeleteItem(id uint) error {
	return r.db.Delete(&models.ModelConfigItem{}, id).Error
}

func (r *ModelConfigRepo) ReplaceItems(configID uint, items []models.ModelConfigItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("model_config_id = ?", configID).Delete(&models.ModelConfigItem{}).Error; err != nil {
			return err
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Proxy queries

func (r *ModelConfigRepo) FindItemsByMappedName(
	configID uint,
	mappedName string,
	protocol models.ProviderProtocol,
) ([]models.ModelConfigItem, error) {
	var items []models.ModelConfigItem
	err := r.db.
		Joins("JOIN providers ON providers.id = model_config_items.provider_id").
		Joins("JOIN provider_models ON provider_models.id = model_config_items.provider_model_id").
		Joins("JOIN model_configs ON model_configs.id = model_config_items.model_config_id").
		Where("model_config_items.model_config_id = ?", configID).
		Where("model_config_items.mapped_name = ?", mappedName).
		Where("model_config_items.enabled = ?", true).
		Where("model_configs.enabled = ?", true).
		Where("providers.enabled = ?", true).
		Where("providers.protocol = ?", protocol).
		Where("provider_models.enabled = ?", true).
		Preload("Provider").
		Preload("ProviderModel").
		Order("model_config_items.priority desc, model_config_items.id asc").
		Find(&items).Error

	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ModelConfigRepo) FindAllMappedNames(
	configID uint,
	protocol models.ProviderProtocol,
) ([]string, error) {
	var names []string
	err := r.db.Model(&models.ModelConfigItem{}).
		Joins("JOIN providers ON providers.id = model_config_items.provider_id").
		Joins("JOIN provider_models ON provider_models.id = model_config_items.provider_model_id").
		Joins("JOIN model_configs ON model_configs.id = model_config_items.model_config_id").
		Where("model_config_items.model_config_id = ?", configID).
		Where("model_config_items.enabled = ?", true).
		Where("model_configs.enabled = ?", true).
		Where("providers.enabled = ?", true).
		Where("providers.protocol = ?", protocol).
		Where("provider_models.enabled = ?", true).
		Distinct("model_config_items.mapped_name").
		Pluck("model_config_items.mapped_name", &names).Error

	return names, err
}

func (r *ModelConfigRepo) FindItemsGrouped(configID uint) (map[string][]models.ModelConfigItem, error) {
	items, err := r.FindItemsByConfigID(configID)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]models.ModelConfigItem)
	for _, item := range items {
		grouped[item.MappedName] = append(grouped[item.MappedName], item)
	}
	return grouped, nil
}
