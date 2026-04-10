package password

import (
	"context"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"simple-blog-api/internal/domain"
)

const bcryptCost = 12

var bannedPatterns = []string{
	"123456", "234567", "345678", "456789", "567890",
	"abcdef", "bcdefg", "cdefgh", "defghi", "efghij",
	"qwerty", "wertyu", "ertyui", "rtyuio", "tyuiop",
	"password", "letmein", "iloveyou", "admin123",
}

var allowedSpecial = "!@#$%^&*()_+-=[]{}';:\"\\|,.<>/?"

// Validator validates passwords against NIST SP 800-63B and HIBP.
type Validator struct {
	hibp *hibpChecker
}

// NewValidator creates a Validator. cfg may be nil for production HIBP endpoint.
func NewValidator(cfg *HIBPConfig) *Validator {
	return &Validator{hibp: newHIBPChecker(cfg)}
}

// Validate checks password complexity and HIBP breach status.
// The caller must have already called strings.TrimSpace on the input.
func (v *Validator) Validate(ctx context.Context, plain string) error {
	if len(plain) < 12 {
		return domain.ErrPasswordTooWeak
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range plain {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune(allowedSpecial, r):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return domain.ErrPasswordTooWeak
	}

	lower := strings.ToLower(plain)
	for _, pattern := range bannedPatterns {
		if strings.Contains(lower, pattern) {
			return domain.ErrPasswordTooWeak
		}
	}

	if v.hibp.isPwned(ctx, plain) {
		return domain.ErrPasswordPwned
	}

	return nil
}

// Hash hashes the plain password with bcrypt cost 12.
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify reports whether plain matches the bcrypt hash.
func Verify(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
