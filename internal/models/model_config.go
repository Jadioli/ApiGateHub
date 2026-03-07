package models

import "time"

type ModelConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:512" json:"description"`
	Enabled     bool      `gorm:"default:true;index" json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Items   []ModelConfigItem `gorm:"foreignKey:ModelConfigID" json:"items,omitempty"`
	APIKeys []APIKey          `gorm:"foreignKey:ModelConfigID" json:"api_keys,omitempty"`
}
