package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"simple-blog-api/internal/usecase/auth"
)

const (
	ContextKeyUserID      = "userID"
	ContextKeyEmail       = "email"
	ContextKeyPermissions = "permissions"
)

// JWT extracts and validates the Bearer JWT from the Authorization header.
// On success, it stores userID, email, and permissions in the Gin context.
func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header must be 'Bearer <token>'",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		claims, err := auth.ParseAccessToken(parts[1], secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyPermissions, claims.Permissions)
		c.Next()
	}
}
