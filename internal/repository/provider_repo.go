package repository

import (
	"apihub/internal/models"
	"time"

	"gorm.io/gorm"
)

type ProviderRepo struct {
	db *gorm.DB
}

func NewProviderRepo(db *gorm.DB) *ProviderRepo {
	return &ProviderRepo{db: db}
}

func (r *ProviderRepo) Create(provider *models.Provider) error {
	return r.db.Create(provider).Error
}

func (r *ProviderRepo) FindAll() ([]models.Provider, error) {
	var providers []models.Provider
	err := r.db.Order("id asc").Find(&providers).Error
	return providers, err
}

func (r *ProviderRepo) FindByID(id uint) (*models.Provider, error) {
	var provider models.Provider
	err := r.db.First(&provider, id).Error
	return &provider, err
}

func (r *ProviderRepo) FindByIDWithModels(id uint) (*models.Provider, error) {
	var provider models.Provider
	err := r.db.Preload("Models", func(db *gorm.DB) *gorm.DB {
		return db.Order("model_name asc")
	}).First(&provider, id).Error
	return &provider, err
}

func (r *ProviderRepo) FindEnabledByProtocol(protocol models.ProviderProtocol) ([]models.Provider, error) {
	var providers []models.Provider
	err := r.db.Where("enabled = ? AND protocol = ?", true, protocol).Find(&providers).Error
	return providers, err
}

func (r *ProviderRepo) FindAllEnabled() ([]models.Provider, error) {
	var providers []models.Provider
	err := r.db.Where("enabled = ?", true).Find(&providers).Error
	return providers, err
}

func (r *ProviderRepo) Update(provider *models.Provider) error {
	return r.db.Save(provider).Error
}

func (r *ProviderRepo) UpdateSyncState(providerID uint, status string, syncError string, lastSyncAt *time.Time) error {
	updates := map[string]interface{}{
		"sync_status": status,
		"sync_error":  syncError,
	}
	if lastSyncAt != nil {
		updates["last_sync_at"] = *lastSyncAt
	}
	return r.db.Model(&models.Provider{}).Where("id = ?", providerID).Updates(updates).Error
}

func (r *ProviderRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete key-level mappings that point at this provider.
		if err := tx.Where("provider_id = ?", id).Delete(&models.APIKeyModel{}).Error; err != nil {
			return err
		}
		// Delete config-level mappings before removing provider models.
		if err := tx.Where("provider_id = ?", id).Delete(&models.ModelConfigItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("provider_id = ?", id).Delete(&models.ProviderModel{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Provider{}, id).Error
	})
}

// ProviderModel operations

func (r *ProviderRepo) FindModelsByProviderID(providerID uint) ([]models.ProviderModel, error) {
	var pm []models.ProviderModel
	err := r.db.Where("provider_id = ?", providerID).Order("model_name asc").Find(&pm).Error
	return pm, err
}

func (r *ProviderRepo) FindModelByID(id uint) (*models.ProviderModel, error) {
	var pm models.ProviderModel
	err := r.db.First(&pm, id).Error
	return &pm, err
}

func (r *ProviderRepo) UpsertModels(providerID uint, modelNames []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, name := range modelNames {
			var existing models.ProviderModel
			err := tx.Where("provider_id = ? AND model_name = ?", providerID, name).First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				pm := models.ProviderModel{
					ProviderID: providerID,
					ModelName:  name,
					Enabled:    true,
				}
				if err := tx.Create(&pm).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *ProviderRepo) UpdateModel(pm *models.ProviderModel) error {
	return r.db.Save(pm).Error
}

func (r *ProviderRepo) SetModelsEnabled(providerID uint, enabled bool) error {
	return r.db.Model(&models.ProviderModel{}).
		Where("provider_id = ?", providerID).
		Update("enabled", enabled).Error
}
