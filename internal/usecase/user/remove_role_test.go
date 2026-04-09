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

func TestRemoveRole_Success(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	// "role1" != "superadmin-role-id", so guard does not block.
	repo.On("GetRoleByName", mock.Anything, "superadmin").Return(&domain.Role{ID: "superadmin-role-id", Name: "superadmin"}, nil)
	repo.On("RemoveRole", mock.Anything, "u1", "role1").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewRemoveRoleUsecase(repo, audit)
	err := uc.Execute(context.Background(), "u1", "role1", "admin1", "admin@example.com")
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestRemoveRole_SuperadminForbidden(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	superadminRoleID := "superadmin-role-id"
	repo.On("GetRoleByName", mock.Anything, "superadmin").Return(&domain.Role{ID: superadminRoleID, Name: "superadmin"}, nil)
	// actor does not hold the superadmin role
	repo.On("GetRoles", mock.Anything, "admin1").Return([]domain.Role{{ID: "admin-role", Name: "admin"}}, nil)

	uc := user.NewRemoveRoleUsecase(repo, audit)
	err := uc.Execute(context.Background(), "u1", superadminRoleID, "admin1", "admin@example.com")
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertExpectations(t)
}

func TestRemoveRole_SuperadminAllowedForSuperadmin(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	superadminRoleID := "superadmin-role-id"
	repo.On("GetRoleByName", mock.Anything, "superadmin").Return(&domain.Role{ID: superadminRoleID, Name: "superadmin"}, nil)
	// actor holds superadmin
	repo.On("GetRoles", mock.Anything, "actor1").Return([]domain.Role{{ID: superadminRoleID, Name: "superadmin"}}, nil)
	repo.On("RemoveRole", mock.Anything, "u1", superadminRoleID).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewRemoveRoleUsecase(repo, audit)
	err := uc.Execute(context.Background(), "u1", superadminRoleID, "actor1", "actor@example.com")
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}
