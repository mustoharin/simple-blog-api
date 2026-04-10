package config_test

import (
	"os"
	"testing"
	"time"

	"simple-blog-api/config"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := config.Load()
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "development", cfg.Env)
	assert.Equal(t, 15*time.Minute, cfg.JWTAccessExpiry)
	assert.Equal(t, 30*24*time.Hour, cfg.JWTRefreshExpiry)
	assert.Equal(t, 365, cfg.AuditLogRetentionDays)
	assert.Equal(t, 90, cfg.SoftDeleteRetentionDays)
	assert.Equal(t, 587, cfg.SMTPPort)
	assert.Equal(t, "hcaptcha", cfg.CaptchaProvider)
	assert.Equal(t, "http://localhost:3000", cfg.FrontendURL)
	assert.Equal(t, "us-east-1", cfg.S3Region)
	assert.Equal(t, "", cfg.CaptchaSiteKey)
}

func TestLoad_EnvOverride(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "production")
	os.Setenv("AUDIT_LOG_RETENTION_DAYS", "180")
	os.Setenv("SOFT_DELETE_RETENTION_DAYS", "60")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
		os.Unsetenv("AUDIT_LOG_RETENTION_DAYS")
		os.Unsetenv("SOFT_DELETE_RETENTION_DAYS")
	}()

	cfg := config.Load()
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "production", cfg.Env)
	assert.Equal(t, 180, cfg.AuditLogRetentionDays)
	assert.Equal(t, 60, cfg.SoftDeleteRetentionDays)
}

func TestLoad_DurationEnvOverride(t *testing.T) {
	os.Setenv("JWT_ACCESS_EXPIRY", "30m")
	os.Setenv("JWT_REFRESH_EXPIRY", "720h")
	defer func() {
		os.Unsetenv("JWT_ACCESS_EXPIRY")
		os.Unsetenv("JWT_REFRESH_EXPIRY")
	}()

	cfg := config.Load()
	assert.Equal(t, 30*time.Minute, cfg.JWTAccessExpiry)
	assert.Equal(t, 720*time.Hour, cfg.JWTRefreshExpiry)
}

func TestLoad_CaptchaSiteKeyEnvOverride(t *testing.T) {
	os.Setenv("CAPTCHA_SITE_KEY", "my-public-site-key")
	defer os.Unsetenv("CAPTCHA_SITE_KEY")

	cfg := config.Load()
	assert.Equal(t, "my-public-site-key", cfg.CaptchaSiteKey)
}

func TestLoad_SliceEnvOverride(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,https://example.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	cfg := config.Load()
	assert.Equal(t, []string{"http://localhost:3000", "https://example.com"}, cfg.AllowedOrigins)
}

func TestLoad_InvalidIntFallsToDefault(t *testing.T) {
	os.Setenv("SMTP_PORT", "notanumber")
	defer os.Unsetenv("SMTP_PORT")

	cfg := config.Load()
	assert.Equal(t, 587, cfg.SMTPPort)
}

func TestLoad_InvalidDurationFallsToDefault(t *testing.T) {
	os.Setenv("JWT_ACCESS_EXPIRY", "notaduration")
	defer os.Unsetenv("JWT_ACCESS_EXPIRY")

	cfg := config.Load()
	assert.Equal(t, 15*time.Minute, cfg.JWTAccessExpiry)
}
