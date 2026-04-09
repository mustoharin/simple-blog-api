package profile_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/password"
	"simple-blog-api/internal/usecase/profile"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) List(ctx context.Context, page, limit int) ([]*domain.User, int, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]*domain.User), args.Int(1), args.Error(2)
}
func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepo) SoftDelete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) UpdateStatus(ctx context.Context, id string, s domain.UserStatus) error {
	return m.Called(ctx, id, s).Error(0)
}
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) UpdatePasswordHash(ctx context.Context, id, passwordHash string) error {
	return m.Called(ctx, id, passwordHash).Error(0)
}
func (m *mockUserRepo) AssignRole(ctx context.Context, userID, roleID string) error {
	return m.Called(ctx, userID, roleID).Error(0)
}
func (m *mockUserRepo) RemoveRole(ctx context.Context, userID, roleID string) error {
	return m.Called(ctx, userID, roleID).Error(0)
}
func (m *mockUserRepo) GetRoles(ctx context.Context, userID string) ([]domain.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Role), args.Error(1)
}
func (m *mockUserRepo) GetPermissions(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockUserRepo) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

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

type mockAuditLogger struct{ mock.Mock }

func (m *mockAuditLogger) Log(ctx context.Context, entry *domain.AuditLog) error {
	return m.Called(ctx, entry).Error(0)
}

type mockPasswordValidator struct{ mock.Mock }

func (m *mockPasswordValidator) Validate(ctx context.Context, plain string) error {
	return m.Called(ctx, plain).Error(0)
}

func TestChangePassword_Success(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	pwv := new(mockPasswordValidator)
	audit := new(mockAuditLogger)

	currentHash, _ := password.Hash("OldStr0ng!Pass")
	u := &domain.User{ID: "u1", Email: "alice@example.com", PasswordHash: &currentHash}

	repo.On("GetByID", mock.Anything, "u1").Return(u, nil)
	pwv.On("Validate", mock.Anything, "NewStr0ng!Pass#99").Return(nil)
	repo.On("UpdatePasswordHash", mock.Anything, "u1", mock.AnythingOfType("string")).Return(nil)
	rtRepo.On("RevokeAllForUser", mock.Anything, "u1").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := profile.NewChangePasswordUsecase(repo, rtRepo, pwv, audit)
	err := uc.Execute(context.Background(), "u1", "OldStr0ng!Pass", "NewStr0ng!Pass#99")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	rtRepo.AssertExpectations(t)
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	pwv := new(mockPasswordValidator)
	audit := new(mockAuditLogger)

	currentHash, _ := password.Hash("OldStr0ng!Pass")
	u := &domain.User{ID: "u1", Email: "alice@example.com", PasswordHash: &currentHash}
	repo.On("GetByID", mock.Anything, "u1").Return(u, nil)

	uc := profile.NewChangePasswordUsecase(repo, rtRepo, pwv, audit)
	err := uc.Execute(context.Background(), "u1", "WrongPassword!", "NewStr0ng!Pass#99")
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}
