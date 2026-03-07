package models

import "time"

type ModelConfigItem struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ModelConfigID   uint      `gorm:"not null;index:idx_config_mapped" json:"model_config_id"`
	ProviderID      uint      `gorm:"not null;index" json:"provider_id"`
	ProviderModelID uint      `gorm:"not null;index" json:"provider_model_id"`
	MappedName      string    `gorm:"size:128;not null;index:idx_config_mapped" json:"mapped_name"`
	Priority        int       `gorm:"default:0" json:"priority"`
	Enabled         bool      `gorm:"default:true;index" json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	ModelConfig   ModelConfig   `gorm:"foreignKey:ModelConfigID" json:"model_config,omitempty"`
	Provider      Provider      `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	ProviderModel ProviderModel `gorm:"foreignKey:ProviderModelID" json:"provider_model,omitempty"`
}
