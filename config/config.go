package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                    string
	Env                     string
	DatabaseURL             string
	JWTSecret               string
	JWTAccessExpiry         time.Duration
	JWTRefreshExpiry        time.Duration
	S3Endpoint              string
	S3Bucket                string
	S3Region                string
	S3AccessKey             string
	S3SecretKey             string
	SMTPHost                string
	SMTPPort                int
	SMTPUser                string
	SMTPPass                string
	EmailFrom               string
	CaptchaProvider         string
	CaptchaSecret           string
	CaptchaSiteKey          string
	AllowedOrigins          []string
	FrontendURL             string
	AuditLogRetentionDays   int
	SoftDeleteRetentionDays int
	SeedAdminEmail          string
	SeedAdminPassword       string
}

func Load() *Config {
	return &Config{
		Port:                    getEnv("PORT", "8080"),
		Env:                     getEnv("ENV", "development"),
		DatabaseURL:             getEnv("DATABASE_URL", ""),
		JWTSecret:               getEnv("JWT_SECRET", ""),
		JWTAccessExpiry:         getDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry:        getDuration("JWT_REFRESH_EXPIRY", 30*24*time.Hour),
		S3Endpoint:              getEnv("S3_ENDPOINT", ""),
		S3Bucket:                getEnv("S3_BUCKET", ""),
		S3Region:                getEnv("S3_REGION", "us-east-1"),
		S3AccessKey:             getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:             getEnv("S3_SECRET_KEY", ""),
		SMTPHost:                getEnv("SMTP_HOST", ""),
		SMTPPort:                getInt("SMTP_PORT", 587),
		SMTPUser:                getEnv("SMTP_USER", ""),
		SMTPPass:                getEnv("SMTP_PASS", ""),
		EmailFrom:               getEnv("EMAIL_FROM", ""),
		CaptchaProvider:         getEnv("CAPTCHA_PROVIDER", "hcaptcha"),
		CaptchaSecret:           getEnv("CAPTCHA_SECRET", ""),
		CaptchaSiteKey:          getEnv("CAPTCHA_SITE_KEY", ""),
		AllowedOrigins:          getSlice("ALLOWED_ORIGINS", []string{"*"}),
		FrontendURL:             getEnv("FRONTEND_URL", "http://localhost:3000"),
		AuditLogRetentionDays:   getInt("AUDIT_LOG_RETENTION_DAYS", 365),
		SoftDeleteRetentionDays: getInt("SOFT_DELETE_RETENTION_DAYS", 90),
		SeedAdminEmail:          getEnv("SEED_ADMIN_EMAIL", ""),
		SeedAdminPassword:       getEnv("SEED_ADMIN_PASSWORD", ""),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getSlice(key string, def []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				result = append(result, s)
			}
		}
		return result
	}
	return def
}
