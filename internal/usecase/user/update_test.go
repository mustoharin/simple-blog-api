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

func TestUpdateUser_Success(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	u := &domain.User{ID: "u1", Email: "user@example.com"}
	repo.On("GetByID", mock.Anything, "u1").Return(u, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewUpdateUserUsecase(repo, audit)
	result, err := uc.Execute(context.Background(), user.UpdateUserInput{
		ID:          "u1",
		DisplayName: "Updated Name",
		Bio:         "New bio",
		ActorID:     "admin1",
		ActorEmail:  "admin@example.com",
	})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "u1", result.ID)
	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestUpdateUser_NotFound(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	repo.On("GetByID", mock.Anything, "u1").Return(nil, domain.ErrNotFound)

	uc := user.NewUpdateUserUsecase(repo, audit)
	result, err := uc.Execute(context.Background(), user.UpdateUserInput{ID: "u1"})
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUpdateUser_JavascriptAvatarURL_ReturnsError(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	u := &domain.User{ID: "u1", Email: "alice@example.com"}
	repo.On("GetByID", mock.Anything, "u1").Return(u, nil)

	uc := user.NewUpdateUserUsecase(repo, audit)
	result, err := uc.Execute(context.Background(), user.UpdateUserInput{
		ID:         "u1",
		AvatarURL:  "javascript:alert(1)",
		ActorID:    "admin1",
		ActorEmail: "admin@example.com",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}
