package auth_test

import (
	"context"
	"testing"
	"time"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/password"
	"simple-blog-api/internal/usecase/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRefreshTokenRepo struct{ mock.Mock }

func (m *mockRefreshTokenRepo) Create(ctx context.Context, t *domain.RefreshToken) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockRefreshTokenRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}
func (m *mockRefreshTokenRepo) Revoke(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

type mockCaptchaVerifier struct{ mock.Mock }

func (m *mockCaptchaVerifier) Verify(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

func makeActiveUser(t *testing.T) *domain.User {
	t.Helper()
	hash, _ := password.Hash("Str0ng&Pass#99")
	return &domain.User{
		ID:           "u1",
		Email:        "alice@example.com",
		PasswordHash: &hash,
		Status:       domain.UserStatusActive,
	}
}

func TestLogin_Success(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	audit := new(mockAuditLogger)
	captcha := new(mockCaptchaVerifier)

	user := makeActiveUser(t)
	repo.On("GetByEmail", mock.Anything, "alice@example.com").Return(user, nil)
	repo.On("GetPermissions", mock.Anything, "u1").Return([]string{"posts:read"}, nil)
	repo.On("UpdateLastLogin", mock.Anything, "u1").Return(nil)
	rtRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	captcha.On("Verify", mock.Anything, "valid-token").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := auth.NewLoginUsecase(repo, rtRepo, captcha, audit, "secret", 15*time.Minute, 30*24*time.Hour)
	out, err := uc.Execute(context.Background(), auth.LoginInput{
		Email:        "alice@example.com",
		Password:     "Str0ng&Pass#99",
		CaptchaToken: "valid-token",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, out.AccessToken)
	assert.NotEmpty(t, out.RefreshToken)
	repo.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	audit := new(mockAuditLogger)
	captcha := new(mockCaptchaVerifier)

	user := makeActiveUser(t)
	repo.On("GetByEmail", mock.Anything, "alice@example.com").Return(user, nil)
	captcha.On("Verify", mock.Anything, "valid-token").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := auth.NewLoginUsecase(repo, rtRepo, captcha, audit, "secret", 15*time.Minute, 30*24*time.Hour)
	_, err := uc.Execute(context.Background(), auth.LoginInput{
		Email:        "alice@example.com",
		Password:     "WrongPass1!",
		CaptchaToken: "valid-token",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogin_AccountNotActivated(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	audit := new(mockAuditLogger)
	captcha := new(mockCaptchaVerifier)

	hash, _ := password.Hash("Str0ng&Pass#99")
	user := &domain.User{
		ID:           "u1",
		Email:        "bob@example.com",
		PasswordHash: &hash,
		Status:       domain.UserStatusPendingInvitation,
	}
	repo.On("GetByEmail", mock.Anything, "bob@example.com").Return(user, nil)
	captcha.On("Verify", mock.Anything, "valid-token").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := auth.NewLoginUsecase(repo, rtRepo, captcha, audit, "secret", 15*time.Minute, 30*24*time.Hour)
	_, err := uc.Execute(context.Background(), auth.LoginInput{
		Email:        "bob@example.com",
		Password:     "Str0ng&Pass#99",
		CaptchaToken: "valid-token",
	})
	assert.ErrorIs(t, err, domain.ErrAccountNotActivated)
}

func TestLogin_InvalidCaptcha(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	audit := new(mockAuditLogger)
	captcha := new(mockCaptchaVerifier)

	captcha.On("Verify", mock.Anything, "bad-token").Return(domain.ErrInvalidCaptcha)

	uc := auth.NewLoginUsecase(repo, rtRepo, captcha, audit, "secret", 15*time.Minute, 30*24*time.Hour)
	_, err := uc.Execute(context.Background(), auth.LoginInput{
		Email:        "alice@example.com",
		Password:     "Str0ng&Pass#99",
		CaptchaToken: "bad-token",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidCaptcha)
}
