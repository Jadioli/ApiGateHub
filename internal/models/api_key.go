package models

import "time"

type APIKey struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Name          string     `gorm:"size:128;not null" json:"name"`
	Key           string     `gorm:"uniqueIndex;size:128;not null" json:"key"` // full key, always visible
	KeyHint       string     `gorm:"size:16" json:"key_hint"`
	ModelConfigID *uint      `gorm:"index" json:"model_config_id"`
	Enabled       bool       `gorm:"default:true;index" json:"enabled"`
	ExpiresAt     *time.Time `json:"expires_at"`
	LastUsedAt    *time.Time `json:"last_used_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	ModelConfig *ModelConfig  `gorm:"foreignKey:ModelConfigID" json:"model_config,omitempty"`
	Models      []APIKeyModel `gorm:"foreignKey:APIKeyID" json:"models,omitempty"`
}
