package user_test

import (
	"context"
	"testing"
	"time"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRefreshTokenRepo struct{ mock.Mock }

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	return m.Called(ctx, token).Error(0)
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

func TestDeleteUser_Success(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	audit := new(mockAuditLogger)

	u := &domain.User{ID: "u1"}
	repo.On("GetByID", mock.Anything, "u1").Return(u, nil)
	repo.On("SoftDelete", mock.Anything, "u1").Return(nil)
	rtRepo.On("RevokeAllForUser", mock.Anything, "u1").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewDeleteUserUsecase(repo, rtRepo, audit)
	err := uc.Execute(context.Background(), "u1", "admin1", "admin@example.com")
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	rtRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestDeleteUser_NotFound(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	audit := new(mockAuditLogger)

	repo.On("GetByID", mock.Anything, "u1").Return(nil, domain.ErrNotFound)

	uc := user.NewDeleteUserUsecase(repo, rtRepo, audit)
	err := uc.Execute(context.Background(), "u1", "admin1", "admin@example.com")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
