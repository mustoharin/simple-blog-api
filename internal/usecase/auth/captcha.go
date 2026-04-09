package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"simple-blog-api/internal/domain"
)

// CaptchaVerifier abstracts CAPTCHA verification.
type CaptchaVerifier interface {
	Verify(ctx context.Context, token string) error
}

type hCaptchaVerifier struct {
	secret  string
	client  *http.Client
	siteURL string
}

// NewHCaptchaVerifier creates a verifier for hCaptcha.
func NewHCaptchaVerifier(secret string) CaptchaVerifier {
	return &hCaptchaVerifier{
		secret:  secret,
		client:  &http.Client{Timeout: 5 * time.Second},
		siteURL: "https://hcaptcha.com/siteverify",
	}
}

func (v *hCaptchaVerifier) Verify(ctx context.Context, token string) error {
	body := strings.NewReader(url.Values{
		"secret":   {v.secret},
		"response": {token},
	}.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.siteURL, body)
	if err != nil {
		return domain.ErrInvalidCaptcha
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return domain.ErrInvalidCaptcha
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return domain.ErrInvalidCaptcha
	}
	if !result.Success {
		return domain.ErrInvalidCaptcha
	}
	return nil
}

type reCaptchaVerifier struct {
	secret  string
	client  *http.Client
	siteURL string
}

// NewReCaptchaVerifier creates a verifier for Google reCAPTCHA v2/v3.
func NewReCaptchaVerifier(secret string) CaptchaVerifier {
	return &reCaptchaVerifier{
		secret:  secret,
		client:  &http.Client{Timeout: 5 * time.Second},
		siteURL: "https://www.google.com/recaptcha/api/siteverify",
	}
}

func (v *reCaptchaVerifier) Verify(ctx context.Context, token string) error {
	body := strings.NewReader(url.Values{
		"secret":   {v.secret},
		"response": {token},
	}.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.siteURL, body)
	if err != nil {
		return domain.ErrInvalidCaptcha
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return domain.ErrInvalidCaptcha
	}
	defer resp.Body.Close()

	var result struct {
		Success bool    `json:"success"`
		Score   float64 `json:"score"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return domain.ErrInvalidCaptcha
	}
	if !result.Success || result.Score < 0.5 {
		return domain.ErrInvalidCaptcha
	}
	return nil
}

// NewCaptchaVerifier returns the appropriate verifier based on provider name.
func NewCaptchaVerifier(provider, secret string) (CaptchaVerifier, error) {
	switch strings.ToLower(provider) {
	case "hcaptcha":
		return NewHCaptchaVerifier(secret), nil
	case "recaptcha":
		return NewReCaptchaVerifier(secret), nil
	default:
		return nil, fmt.Errorf("unknown captcha provider: %s", provider)
	}
}
