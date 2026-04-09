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

func TestAssignRole_Success(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	u := &domain.User{ID: "u1"}
	repo.On("GetByID", mock.Anything, "u1").Return(u, nil)
	// "role1" != "superadmin-role-id", so guard does not block.
	repo.On("GetRoleByName", mock.Anything, "superadmin").Return(&domain.Role{ID: "superadmin-role-id", Name: "superadmin"}, nil)
	repo.On("AssignRole", mock.Anything, "u1", "role1").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewAssignRoleUsecase(repo, audit)
	err := uc.Execute(context.Background(), "u1", "role1", "admin1", "admin@example.com")
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestAssignRole_SuperadminForbidden(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	superadminRoleID := "superadmin-role-id"
	repo.On("GetByID", mock.Anything, "u1").Return(&domain.User{ID: "u1"}, nil)
	repo.On("GetRoleByName", mock.Anything, "superadmin").Return(&domain.Role{ID: superadminRoleID, Name: "superadmin"}, nil)
	// actor does not hold the superadmin role
	repo.On("GetRoles", mock.Anything, "admin1").Return([]domain.Role{{ID: "admin-role", Name: "admin"}}, nil)

	uc := user.NewAssignRoleUsecase(repo, audit)
	err := uc.Execute(context.Background(), "u1", superadminRoleID, "admin1", "admin@example.com")
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertExpectations(t)
}

func TestAssignRole_SuperadminAllowedForSuperadmin(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	superadminRoleID := "superadmin-role-id"
	repo.On("GetByID", mock.Anything, "u1").Return(&domain.User{ID: "u1"}, nil)
	repo.On("GetRoleByName", mock.Anything, "superadmin").Return(&domain.Role{ID: superadminRoleID, Name: "superadmin"}, nil)
	// actor holds superadmin
	repo.On("GetRoles", mock.Anything, "actor1").Return([]domain.Role{{ID: superadminRoleID, Name: "superadmin"}}, nil)
	repo.On("AssignRole", mock.Anything, "u1", superadminRoleID).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewAssignRoleUsecase(repo, audit)
	err := uc.Execute(context.Background(), "u1", superadminRoleID, "actor1", "actor@example.com")
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestAssignRole_UserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	repo.On("GetByID", mock.Anything, "u1").Return(nil, domain.ErrNotFound)

	uc := user.NewAssignRoleUsecase(repo, audit)
	err := uc.Execute(context.Background(), "u1", "role1", "admin1", "admin@example.com")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
