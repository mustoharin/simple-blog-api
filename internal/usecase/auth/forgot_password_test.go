package auth_test

import (
	"context"
	"testing"
	"time"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPRTRepo struct{ mock.Mock }

func (m *mockPRTRepo) Create(ctx context.Context, t *domain.PasswordResetToken) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockPRTRepo) GetByHash(ctx context.Context, hash string) (*domain.PasswordResetToken, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PasswordResetToken), args.Error(1)
}
func (m *mockPRTRepo) MarkUsed(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

type mockEmailSender struct{ mock.Mock }

func (m *mockEmailSender) Send(msg auth.EmailMessage) error {
	return m.Called(msg).Error(0)
}

func TestForgotPassword_Success(t *testing.T) {
	repo := new(mockUserRepo)
	prtRepo := new(mockPRTRepo)
	emailSender := new(mockEmailSender)
	audit := new(mockAuditLogger)

	user := &domain.User{ID: "u1", Email: "alice@example.com", Status: domain.UserStatusActive}
	repo.On("GetByEmail", mock.Anything, "alice@example.com").Return(user, nil)
	prtRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	emailSender.On("Send", mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := auth.NewForgotPasswordUsecase(repo, prtRepo, emailSender, audit, "http://localhost:3000")
	err := uc.Execute(context.Background(), "alice@example.com")
	assert.NoError(t, err)
}

func TestForgotPassword_UserNotFound_NoError(t *testing.T) {
	// Should return nil even for non-existent emails (prevent enumeration)
	repo := new(mockUserRepo)
	prtRepo := new(mockPRTRepo)
	emailSender := new(mockEmailSender)
	audit := new(mockAuditLogger)

	repo.On("GetByEmail", mock.Anything, "ghost@example.com").Return(nil, domain.ErrNotFound)

	uc := auth.NewForgotPasswordUsecase(repo, prtRepo, emailSender, audit, "http://localhost:3000")
	err := uc.Execute(context.Background(), "ghost@example.com")
	assert.NoError(t, err) // no error to prevent email enumeration
}

func TestResetPassword_Success(t *testing.T) {
	repo := new(mockUserRepo)
	prtRepo := new(mockPRTRepo)
	rtRepo := new(mockRefreshTokenRepo)
	pwv := new(mockPasswordValidator)
	audit := new(mockAuditLogger)

	rawToken := "raw-reset-token-xyz"
	hash := makeTokenHash(rawToken)
	prt := &domain.PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    "u1",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	user := &domain.User{ID: "u1", Email: "alice@example.com"}

	prtRepo.On("GetByHash", mock.Anything, hash).Return(prt, nil)
	prtRepo.On("MarkUsed", mock.Anything, prt.ID).Return(nil)
	repo.On("GetByID", mock.Anything, "u1").Return(user, nil)
	pwv.On("Validate", mock.Anything, "NewStr0ng!Pass#77").Return(nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	rtRepo.On("RevokeAllForUser", mock.Anything, "u1").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := auth.NewResetPasswordUsecase(repo, prtRepo, rtRepo, pwv, audit)
	err := uc.Execute(context.Background(), rawToken, "NewStr0ng!Pass#77")
	assert.NoError(t, err)
}

func TestResetPassword_TokenExpired(t *testing.T) {
	repo := new(mockUserRepo)
	prtRepo := new(mockPRTRepo)
	rtRepo := new(mockRefreshTokenRepo)
	pwv := new(mockPasswordValidator)
	audit := new(mockAuditLogger)

	rawToken := "raw-expired-reset"
	hash := makeTokenHash(rawToken)
	prt := &domain.PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    "u1",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(-time.Hour), // expired
	}
	prtRepo.On("GetByHash", mock.Anything, hash).Return(prt, nil)

	uc := auth.NewResetPasswordUsecase(repo, prtRepo, rtRepo, pwv, audit)
	err := uc.Execute(context.Background(), rawToken, "NewStr0ng!Pass#77")
	assert.ErrorIs(t, err, domain.ErrTokenExpired)
}
