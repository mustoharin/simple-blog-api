package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequirePermission returns a Gin handler that aborts with 403 if the
// authenticated user does not hold the given permission.
// The JWT middleware must run before this handler.
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get(ContextKeyPermissions)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "FORBIDDEN",
			})
			return
		}

		perms, ok := raw.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "FORBIDDEN",
			})
			return
		}

		for _, p := range perms {
			if p == perm {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
			"code":  "FORBIDDEN",
		})
	}
}
