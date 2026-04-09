package postgres_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository implements domain.UserRepository for unit tests.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, page, limit int) ([]*domain.User, int, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]*domain.User), args.Int(1), args.Error(2)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, id string, status domain.UserStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePasswordHash(ctx context.Context, id, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) AssignRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) RemoveRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) GetRoles(ctx context.Context, userID string) ([]domain.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Role), args.Error(1)
}

func (m *MockUserRepository) GetPermissions(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockUserRepository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

// Verify MockUserRepository satisfies the interface at compile time.
var _ domain.UserRepository = (*MockUserRepository)(nil)

func TestMockUserRepository_Create(t *testing.T) {
	repo := new(MockUserRepository)
	user := &domain.User{ID: "u1", Email: "a@b.com"}
	repo.On("Create", mock.Anything, user).Return(nil)
	err := repo.Create(context.Background(), user)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestMockUserRepository_GetByEmail_NotFound(t *testing.T) {
	repo := new(MockUserRepository)
	repo.On("GetByEmail", mock.Anything, "missing@b.com").Return(nil, domain.ErrNotFound)
	u, err := repo.GetByEmail(context.Background(), "missing@b.com")
	assert.Nil(t, u)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
