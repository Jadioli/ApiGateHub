package repository

import (
	"apihub/internal/models"

	"gorm.io/gorm"
)

type AdminRepo struct {
	db *gorm.DB
}

func NewAdminRepo(db *gorm.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) FindFirst() (*models.Admin, error) {
	var admin models.Admin
	err := r.db.Order("id asc").First(&admin).Error
	return &admin, err
}

func (r *AdminRepo) FindByID(id uint) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.First(&admin, id).Error
	return &admin, err
}
