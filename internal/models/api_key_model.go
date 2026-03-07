package models

import "time"

// APIKeyModel combines permission + model mapping in one table.
// Each record means: this API Key can access this provider model, exposed as mapped_name.
type APIKeyModel struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	APIKeyID        uint      `gorm:"not null;index:idx_key_mapped_name" json:"api_key_id"`
	ProviderID      uint      `gorm:"not null;index" json:"provider_id"`
	ProviderModelID uint      `gorm:"not null;index" json:"provider_model_id"`
	MappedName      string    `gorm:"size:128;not null;index:idx_key_mapped_name" json:"mapped_name"`
	Priority        int       `gorm:"default:0" json:"priority"`
	Enabled         bool      `gorm:"default:true;index" json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	APIKey        APIKey        `gorm:"foreignKey:APIKeyID" json:"api_key,omitempty"`
	Provider      Provider      `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	ProviderModel ProviderModel `gorm:"foreignKey:ProviderModelID" json:"provider_model,omitempty"`
}
