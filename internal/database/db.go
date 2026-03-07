package database

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"apihub/internal/config"
	"apihub/internal/models"
	"apihub/pkg"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

func Init(cfg *config.Config) *gorm.DB {
	dir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	logLevel := logger.Warn
	if cfg.LogLevel == "debug" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        cfg.DatabasePath,
	}, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	sqlDB.Exec("PRAGMA journal_mode=WAL")
	sqlDB.Exec("PRAGMA foreign_keys=ON")

	err = db.AutoMigrate(
		&models.Admin{},
		&models.Provider{},
		&models.ProviderModel{},
		&models.ModelConfig{},
		&models.ModelConfigItem{},
		&models.APIKey{},
		&models.APIKeyModel{},
		&models.RequestLog{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	migrateLegacyProviderAPIKeys(db)
	seedAdmin(db, cfg)

	return db
}

func migrateLegacyProviderAPIKeys(db *gorm.DB) {
	legacyKey := os.Getenv("DATA_ENCRYPTION_KEY")
	if legacyKey == "" {
		return
	}

	cipher, err := pkg.NewTextCipher(legacyKey)
	if err != nil {
		log.Printf("Skipping legacy provider key migration: %v", err)
		return
	}

	var providers []models.Provider
	if err := db.Find(&providers).Error; err != nil {
		log.Printf("Failed to scan providers for legacy key migration: %v", err)
		return
	}

	migrated := 0
	for _, provider := range providers {
		if !strings.HasPrefix(provider.APIKey, "enc:") {
			continue
		}

		plaintext, err := cipher.Decrypt(provider.APIKey)
		if err != nil {
			log.Printf("Failed to migrate provider %d API key: %v", provider.ID, err)
			continue
		}

		if err := db.Model(&models.Provider{}).Where("id = ?", provider.ID).Update("api_key", plaintext).Error; err != nil {
			log.Printf("Failed to persist migrated API key for provider %d: %v", provider.ID, err)
			continue
		}

		migrated++
	}

	if migrated > 0 {
		log.Printf("Migrated %d legacy provider API keys to plaintext storage", migrated)
	}
}

func seedAdmin(db *gorm.DB, cfg *config.Config) {
	var count int64
	db.Model(&models.Admin{}).Count(&count)
	if count > 0 {
		return
	}

	hash, err := pkg.HashPassword(cfg.AdminPassword)
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}

	admin := models.Admin{
		Username: "admin",
		Password: hash,
	}
	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("Failed to create initial admin: %v", err)
	}
	log.Printf("Initial admin created")
}
