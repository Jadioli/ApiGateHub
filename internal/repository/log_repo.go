package repository

import (
	"apihub/internal/models"

	"gorm.io/gorm"
)

type LogRepo struct {
	db *gorm.DB
}

func NewLogRepo(db *gorm.DB) *LogRepo {
	return &LogRepo{db: db}
}

func (r *LogRepo) Create(log *models.RequestLog) error {
	return r.db.Create(log).Error
}

type LogQuery struct {
	Page     int
	PageSize int
	APIKeyID *uint
	Model    string
	Status   *int
}

type LogResult struct {
	Logs  []models.RequestLog `json:"logs"`
	Total int64               `json:"total"`
}

func (r *LogRepo) Query(q LogQuery) (*LogResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}

	query := r.db.Model(&models.RequestLog{})

	if q.APIKeyID != nil {
		query = query.Where("api_key_id = ?", *q.APIKeyID)
	}
	if q.Model != "" {
		query = query.Where("request_model = ?", q.Model)
	}
	if q.Status != nil {
		query = query.Where("status_code = ?", *q.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var logs []models.RequestLog
	err := query.Order("id desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return &LogResult{Logs: logs, Total: total}, nil
}

type DashboardStats struct {
	ProviderCount int64 `json:"provider_count"`
	APIKeyCount   int64 `json:"api_key_count"`
	TodayRequests int64 `json:"today_requests"`
}

func (r *LogRepo) GetDashboardStats() (*DashboardStats, error) {
	var stats DashboardStats

	r.db.Model(&models.Provider{}).Count(&stats.ProviderCount)
	r.db.Model(&models.APIKey{}).Count(&stats.APIKeyCount)
	r.db.Model(&models.RequestLog{}).
		Where("created_at >= date('now')").
		Count(&stats.TodayRequests)

	return &stats, nil
}
