package services

import (
	"errors"

	"apihub/internal/repository"
	"apihub/pkg"
)

type AuthService struct {
	adminRepo *repository.AdminRepo
}

func NewAuthService(adminRepo *repository.AdminRepo) *AuthService {
	return &AuthService{adminRepo: adminRepo}
}

func (s *AuthService) Login(password string) (string, error) {
	admin, err := s.adminRepo.FindFirst()
	if err != nil {
		return "", errors.New("invalid password")
	}

	if !pkg.CheckPassword(password, admin.Password) {
		return "", errors.New("invalid password")
	}

	token, err := pkg.GenerateToken(admin.Password, admin.ID)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}
