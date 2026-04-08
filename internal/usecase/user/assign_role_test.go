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
	repo.On("AssignRole", mock.Anything, "u1", "role1").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewAssignRoleUsecase(repo, audit)
	err := uc.Execute(context.Background(), "u1", "role1", "admin1", "admin@example.com")
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
