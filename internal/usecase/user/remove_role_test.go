package user_test

import (
	"context"
	"testing"
	"time"

	"simple-blog-api/internal/usecase/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRemoveRole_Success(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	repo.On("RemoveRole", mock.Anything, "u1", "role1").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewRemoveRoleUsecase(repo, audit)
	err := uc.Execute(context.Background(), "u1", "role1", "admin1", "admin@example.com")
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}
