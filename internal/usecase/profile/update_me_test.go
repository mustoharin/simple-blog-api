package profile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/profile"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateMe_Success(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	u := &domain.User{ID: "u1", Email: "alice@example.com"}
	repo.On("GetByID", mock.Anything, "u1").Return(u, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := profile.NewUpdateMeUsecase(repo, audit)
	result, err := uc.Execute(context.Background(), profile.UpdateMeInput{
		UserID:      "u1",
		DisplayName: "Alice",
		Bio:         "Hello!",
		AvatarURL:   "https://example.com/avatar.png",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Alice", result.DisplayName)

	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestUpdateMe_NotFound(t *testing.T) {
	repo := new(mockUserRepo)
	audit := new(mockAuditLogger)

	repo.On("GetByID", mock.Anything, "u1").Return(nil, domain.ErrNotFound)

	uc := profile.NewUpdateMeUsecase(repo, audit)
	result, err := uc.Execute(context.Background(), profile.UpdateMeInput{UserID: "u1"})

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domain.ErrNotFound))
	repo.AssertExpectations(t)
}
