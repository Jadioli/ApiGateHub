package models

import "time"

type RequestLog struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	APIKeyID         uint      `gorm:"index" json:"api_key_id"`
	ProviderID       uint      `gorm:"index" json:"provider_id"`
	RequestModel     string    `gorm:"size:128;index" json:"request_model"`
	ProviderModel    string    `gorm:"size:128" json:"provider_model"`
	StatusCode       int       `gorm:"index" json:"status_code"`
	ResponseTime     int       `json:"response_time"` // ms
	TokensPrompt     int       `json:"tokens_prompt"`
	TokensCompletion int       `json:"tokens_completion"`
	Error            string    `gorm:"size:2048" json:"error"`
	CreatedAt        time.Time `gorm:"index" json:"created_at"`
}
