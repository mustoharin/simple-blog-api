package auth_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/auth"

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

type mockAuditLogger struct{ mock.Mock }

func (m *mockAuditLogger) Log(ctx context.Context, entry *domain.AuditLog) error {
	return m.Called(ctx, entry).Error(0)
}

type mockPasswordValidator struct{ mock.Mock }

func (m *mockPasswordValidator) Validate(ctx context.Context, plain string) error {
	return m.Called(ctx, plain).Error(0)
}

func TestRegister_Success(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)
	pwv := new(mockPasswordValidator)

	repo.On("GetByEmail", mock.Anything, "alice@example.com").Return(nil, domain.ErrNotFound)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "alice@example.com" &&
			u.DisplayName == "Alice" &&
			u.Status == domain.UserStatusActive
	})).Return(nil)
	pwv.On("Validate", mock.Anything, "Str0ng&Pass#99").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := auth.NewRegisterUsecase(repo, pwv, audit)
	err := uc.Execute(context.Background(), auth.RegisterInput{
		Email:       "alice@example.com",
		Password:    "Str0ng&Pass#99",
		DisplayName: "Alice",
	})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)
	pwv := new(mockPasswordValidator)

	existing := &domain.User{ID: "u1", Email: "alice@example.com", Status: domain.UserStatusActive}
	repo.On("GetByEmail", mock.Anything, "alice@example.com").Return(existing, nil)

	uc := auth.NewRegisterUsecase(repo, pwv, audit)
	err := uc.Execute(context.Background(), auth.RegisterInput{
		Email:       "alice@example.com",
		Password:    "Str0ng&Pass#99",
		DisplayName: "Alice",
	})
	assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
}

func TestRegister_WeakPassword(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)
	pwv := new(mockPasswordValidator)

	repo.On("GetByEmail", mock.Anything, "alice@example.com").Return(nil, domain.ErrNotFound)
	pwv.On("Validate", mock.Anything, "weak").Return(domain.ErrPasswordTooWeak)

	uc := auth.NewRegisterUsecase(repo, pwv, audit)
	err := uc.Execute(context.Background(), auth.RegisterInput{
		Email:    "alice@example.com",
		Password: "weak",
	})
	assert.ErrorIs(t, err, domain.ErrPasswordTooWeak)
}
