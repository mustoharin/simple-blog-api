package auth_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/usecase/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCaptchaVerifier_EmptySecret_ReturnsNoop verifies that when CAPTCHA_SECRET is
// empty the factory returns a noop verifier instead of making real HTTP calls.
func TestNewCaptchaVerifier_EmptySecret_ReturnsNoop(t *testing.T) {
	for _, provider := range []string{"hcaptcha", "recaptcha"} {
		t.Run(provider, func(t *testing.T) {
			v, err := auth.NewCaptchaVerifier(provider, "")
			require.NoError(t, err)

			// Must pass any token without touching the network
			err = v.Verify(context.Background(), "any-token")
			assert.NoError(t, err, "empty secret should skip captcha verification")
		})
	}
}

// TestNewCaptchaVerifier_UnknownProvider still returns an error.
func TestNewCaptchaVerifier_UnknownProvider(t *testing.T) {
	_, err := auth.NewCaptchaVerifier("unknown", "some-secret")
	assert.Error(t, err)
}
