package middleware

import (
	"net/http"
	"strings"

	"apihub/internal/models"
	"apihub/pkg"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")

		var admin models.Admin
		if err := db.Order("id asc").First(&admin).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "admin not initialized"})
			c.Abort()
			return
		}

		claims, err := pkg.ParseToken(admin.Password, tokenString)
		if err != nil || claims.AdminID != admin.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Next()
	}
}
