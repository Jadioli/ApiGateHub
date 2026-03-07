package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort    string
	DatabasePath  string
	AdminPassword string
	SyncInterval  int // minutes
	LogLevel      string
}

func Load() *Config {
	loadEnvFile(".env")

	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "9011"),
		DatabasePath:  getEnv("DATABASE_PATH", "./data/apihub.db"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		SyncInterval:  getEnvInt("SYNC_INTERVAL", 60),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}
