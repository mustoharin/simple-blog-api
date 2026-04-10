package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// TestLoginHandler_NoCaptchaToken_BindingDoesNotReject verifies that omitting
// captcha_token from the login request body does NOT cause a 400 validation error.
// If binding:"required" is ever re-added to CaptchaToken, this test will catch it.
// (We get 500 from nil usecase pointer — that's fine; it proves binding accepted the request.)
func TestLoginHandler_NoCaptchaToken_BindingDoesNotReject(t *testing.T) {
	h := handler.NewAuthHandler(nil, nil, nil, nil, nil, nil, nil, false, "", "")

	router := gin.New()
	router.Use(gin.RecoveryWithWriter(io.Discard)) // catch nil-pointer panic silently; we only care about the status code
	router.POST("/login", h.Login)

	body := `{"email":"alice@example.com","password":"Str0ng&Pass#99"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusBadRequest, w.Code,
		"captcha_token must not be required; binding must not reject login without it")
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
