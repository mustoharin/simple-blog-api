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

type mockAuditLogger struct{ mock.Mock }

func (m *mockAuditLogger) Log(ctx context.Context, entry *domain.AuditLog) error {
	return m.Called(ctx, entry).Error(0)
}

type mockEmailSender struct{ mock.Mock }

func (m *mockEmailSender) Send(to, subject, body string) error {
	return m.Called(to, subject, body).Error(0)
}

func TestCreateUser_Success_SendsInvitation(t *testing.T) {
	repo := new(mockUserRepo)
	invRepo := new(mockInvitationTokenRepo)
	email := new(mockEmailSender)
	audit := new(mockAuditLogger)

	repo.On("GetByEmail", mock.Anything, "newuser@example.com").Return(nil, domain.ErrNotFound)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "newuser@example.com" && u.Status == domain.UserStatusPendingInvitation
	})).Return(nil)
	invRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	email.On("Send", "newuser@example.com", mock.Anything, mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewCreateUserUsecase(repo, invRepo, email, audit, "http://localhost:3000")
	createdUser, err := uc.Execute(context.Background(), user.CreateUserInput{
		Email:       "newuser@example.com",
		DisplayName: "New User",
		ActorID:     "admin1",
		ActorEmail:  "admin@example.com",
	})
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, domain.UserStatusPendingInvitation, createdUser.Status)
	time.Sleep(50 * time.Millisecond) // allow fire-and-forget goroutines to complete
	repo.AssertExpectations(t)
	invRepo.AssertExpectations(t)
	email.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestCreateUser_EmailAlreadyExists(t *testing.T) {
	repo := new(mockUserRepo)
	invRepo := new(mockInvitationTokenRepo)
	email := new(mockEmailSender)
	audit := new(mockAuditLogger)

	existing := &domain.User{ID: "u1", Email: "existing@example.com"}
	repo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existing, nil)

	uc := user.NewCreateUserUsecase(repo, invRepo, email, audit, "http://localhost:3000")
	_, err := uc.Execute(context.Background(), user.CreateUserInput{
		Email: "existing@example.com",
	})
	assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
}
