package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"simple-blog-api/internal/delivery/http/handler"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestAuthHandler(enabled bool, provider, siteKey string) *handler.AuthHandler {
	return handler.NewAuthHandler(nil, nil, nil, nil, nil, nil, nil, enabled, provider, siteKey)
}

func TestGetCaptchaConfig_Enabled(t *testing.T) {
	h := newTestAuthHandler(true, "hcaptcha", "test-site-key-123")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/captcha-config", nil)

	h.GetCaptchaConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, true, body["enabled"])
	assert.Equal(t, "hcaptcha", body["provider"])
	assert.Equal(t, "test-site-key-123", body["site_key"])
}

func TestGetCaptchaConfig_Disabled(t *testing.T) {
	h := newTestAuthHandler(false, "hcaptcha", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/captcha-config", nil)

	h.GetCaptchaConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, false, body["enabled"])
	assert.Nil(t, body["provider"])
	assert.Nil(t, body["site_key"])
}
