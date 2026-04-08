package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"simple-blog-api/internal/delivery/http/middleware"
	"simple-blog-api/internal/usecase/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func makeTestToken(t *testing.T, userID, email string, perms []string) string {
	t.Helper()
	token, err := auth.IssueAccessToken(userID, email, perms, "test-secret", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	router := gin.New()
	router.Use(middleware.JWT("test-secret"))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, "u1", userID)
		c.Status(http.StatusOK)
	})

	token := makeTestToken(t, "u1", "alice@example.com", []string{"post:create"})
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	router := gin.New()
	router.Use(middleware.JWT("test-secret"))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	router := gin.New()
	router.Use(middleware.JWT("test-secret"))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRBACMiddleware_HasPermission(t *testing.T) {
	router := gin.New()
	router.Use(middleware.JWT("test-secret"))
	router.POST("/posts", middleware.RequirePermission("post:create"), func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	token := makeTestToken(t, "u1", "alice@example.com", []string{"post:create"})
	req := httptest.NewRequest(http.MethodPost, "/posts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRBACMiddleware_MissingPermission(t *testing.T) {
	router := gin.New()
	router.Use(middleware.JWT("test-secret"))
	router.POST("/posts", middleware.RequirePermission("post:create"), func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	token := makeTestToken(t, "u1", "alice@example.com", []string{"comment:create"})
	req := httptest.NewRequest(http.MethodPost, "/posts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
