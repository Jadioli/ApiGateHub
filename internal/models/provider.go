package models

import "time"

type ProviderProtocol string

const (
	ProtocolOpenAI    ProviderProtocol = "openai"
	ProtocolAnthropic ProviderProtocol = "anthropic"
)

type Provider struct {
	ID           uint             `gorm:"primaryKey" json:"id"`
	Name         string           `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Protocol     ProviderProtocol `gorm:"size:16;not null;index" json:"protocol"`
	BaseURL      string           `gorm:"size:512;not null" json:"base_url"`
	APIKey       string           `gorm:"size:2048;not null" json:"-"`
	Enabled      bool             `gorm:"default:true;index" json:"enabled"`
	LastSyncAt   *time.Time       `json:"last_sync_at"`
	SyncStatus   string           `gorm:"size:16;default:'pending'" json:"sync_status"`
	SyncError    string           `gorm:"size:1024" json:"sync_error"`
	SyncInterval string           `gorm:"size:16;default:'none'" json:"sync_interval"` // none / hourly / daily / weekly
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`

	Models []ProviderModel `gorm:"foreignKey:ProviderID" json:"models,omitempty"`
}
