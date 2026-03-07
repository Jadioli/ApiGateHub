package middleware

import (
	"net/http"
	"strings"
	"time"

	"apihub/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := extractAPIKey(c)
		if key == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			c.Abort()
			return
		}

		var apiKey models.APIKey
		if err := db.Where("key = ?", key).First(&apiKey).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		if !apiKey.Enabled {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key is disabled"})
			c.Abort()
			return
		}

		if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key has expired"})
			c.Abort()
			return
		}

		// Update last used time asynchronously
		go func() {
			now := time.Now()
			db.Model(&models.APIKey{}).Where("id = ?", apiKey.ID).Update("last_used_at", &now)
		}()

		c.Set("api_key", &apiKey)
		c.Next()
	}
}

func extractAPIKey(c *gin.Context) string {
	// OpenAI style: Authorization: Bearer sk-xxx
	if auth := c.GetHeader("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	// Anthropic style: x-api-key: sk-xxx
	if key := c.GetHeader("X-Api-Key"); key != "" {
		return key
	}
	return ""
}
