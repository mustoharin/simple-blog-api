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

type mockInvitationTokenRepo struct{ mock.Mock }

func (m *mockInvitationTokenRepo) Create(ctx context.Context, t *domain.InvitationToken) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockInvitationTokenRepo) GetByUserID(ctx context.Context, userID string) (*domain.InvitationToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InvitationToken), args.Error(1)
}
func (m *mockInvitationTokenRepo) GetByHash(ctx context.Context, hash string) (*domain.InvitationToken, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InvitationToken), args.Error(1)
}
func (m *mockInvitationTokenRepo) MarkUsed(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockInvitationTokenRepo) InvalidatePrevious(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func TestAcceptInvitation_Success(t *testing.T) {
	repo := new(mockUserRepo)
	invRepo := new(mockInvitationTokenRepo)
	pwv := new(mockPasswordValidator)
	audit := new(mockAuditLogger)

	rawToken := "raw-invite-token-abc123"
	hash := makeTokenHash(rawToken)
	inv := &domain.InvitationToken{
		ID:        uuid.NewString(),
		UserID:    "u1",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	user := &domain.User{ID: "u1", Email: "bob@example.com", Status: domain.UserStatusPendingInvitation}

	invRepo.On("GetByHash", mock.Anything, hash).Return(inv, nil)
	invRepo.On("MarkUsed", mock.Anything, inv.ID).Return(nil)
	repo.On("GetByID", mock.Anything, "u1").Return(user, nil)
	pwv.On("Validate", mock.Anything, "Str0ng&Pass#99").Return(nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.PasswordHash != nil && u.Status == domain.UserStatusActive
	})).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := auth.NewAcceptInvitationUsecase(repo, invRepo, pwv, audit)
	err := uc.Execute(context.Background(), rawToken, "Str0ng&Pass#99")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAcceptInvitation_TokenExpired(t *testing.T) {
	repo := new(mockUserRepo)
	invRepo := new(mockInvitationTokenRepo)
	pwv := new(mockPasswordValidator)
	audit := new(mockAuditLogger)

	rawToken := "raw-expired-invite"
	hash := makeTokenHash(rawToken)
	inv := &domain.InvitationToken{
		ID:        uuid.NewString(),
		UserID:    "u1",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	invRepo.On("GetByHash", mock.Anything, hash).Return(inv, nil)

	uc := auth.NewAcceptInvitationUsecase(repo, invRepo, pwv, audit)
	err := uc.Execute(context.Background(), rawToken, "Str0ng&Pass#99")
	assert.ErrorIs(t, err, domain.ErrTokenExpired)
}
