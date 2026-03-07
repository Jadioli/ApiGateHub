package models

import "time"

type ProviderModel struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	ProviderID uint   `gorm:"not null;uniqueIndex:idx_provider_model" json:"provider_id"`
	ModelName  string `gorm:"size:128;not null;uniqueIndex:idx_provider_model" json:"model_name"`
	Enabled    bool   `gorm:"default:true;index" json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Provider Provider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
}
