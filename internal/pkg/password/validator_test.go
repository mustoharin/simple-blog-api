package password_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/password"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_TooShort(t *testing.T) {
	v := password.NewValidator(nil)
	err := v.Validate(context.Background(), "Short1!")
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrPasswordTooWeak)
}

func TestValidate_NoUppercase(t *testing.T) {
	v := password.NewValidator(nil)
	err := v.Validate(context.Background(), "nouppercase1!")
	assert.ErrorIs(t, err, domain.ErrPasswordTooWeak)
}

func TestValidate_NoLowercase(t *testing.T) {
	v := password.NewValidator(nil)
	err := v.Validate(context.Background(), "NOLOWERCASE1!")
	assert.ErrorIs(t, err, domain.ErrPasswordTooWeak)
}

func TestValidate_NoDigit(t *testing.T) {
	v := password.NewValidator(nil)
	err := v.Validate(context.Background(), "NoDigitHere!")
	assert.ErrorIs(t, err, domain.ErrPasswordTooWeak)
}

func TestValidate_NoSpecialChar(t *testing.T) {
	v := password.NewValidator(nil)
	err := v.Validate(context.Background(), "NoSpecialChar1")
	assert.ErrorIs(t, err, domain.ErrPasswordTooWeak)
}

func TestValidate_CommonSequential(t *testing.T) {
	v := password.NewValidator(nil)
	sequences := []string{
		"Password123!",  // contains "password"
		"Qwerty12345!",  // contains "qwerty"
		"Abcdef1234!",   // contains "abcdef"
	}
	for _, pw := range sequences {
		err := v.Validate(context.Background(), pw)
		assert.ErrorIs(t, err, domain.ErrPasswordTooWeak, "expected weak for: %s", pw)
	}
}

func TestValidate_ValidPassword(t *testing.T) {
	v := password.NewValidator(nil)
	err := v.Validate(context.Background(), "Str0ng&Secure#2024")
	assert.NoError(t, err)
}

func TestValidate_HIBPTimeout_FailOpen(t *testing.T) {
	// Mock HIBP server that is already closed (connection refused)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // Close immediately to cause connection refused

	v := password.NewValidator(&password.HIBPConfig{BaseURL: srv.URL + "/range/"})
	// Should fail open (not return ErrPasswordPwned)
	err := v.Validate(context.Background(), "Str0ng&Secure#XYZ9")
	assert.NotErrorIs(t, err, domain.ErrPasswordPwned)
}

func TestHash(t *testing.T) {
	hash, err := password.Hash("Str0ng&Secure#2024")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "Str0ng&Secure#2024", hash)
}

func TestVerify(t *testing.T) {
	hash, err := password.Hash("Str0ng&Secure#2024")
	require.NoError(t, err)

	assert.True(t, password.Verify("Str0ng&Secure#2024", hash))
	assert.False(t, password.Verify("WrongPassword1!", hash))
}
