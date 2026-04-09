package profile_test

import (
	"context"
	"errors"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/profile"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMe_Success(t *testing.T) {
	repo := new(mockUserRepo)

	u := &domain.User{ID: "u1", Email: "alice@example.com"}
	repo.On("GetByID", mock.Anything, "u1").Return(u, nil)
	repo.On("GetRoles", mock.Anything, "u1").Return([]domain.Role{{ID: "r1", Name: "admin"}}, nil)

	uc := profile.NewGetMeUsecase(repo)
	result, err := uc.Execute(context.Background(), "u1")

	assert.NoError(t, err)
	assert.Equal(t, u, result)
	assert.Equal(t, []domain.Role{{ID: "r1", Name: "admin"}}, result.Roles)
	repo.AssertExpectations(t)
}

func TestGetMe_NotFound(t *testing.T) {
	repo := new(mockUserRepo)

	repo.On("GetByID", mock.Anything, "u1").Return(nil, domain.ErrNotFound)

	uc := profile.NewGetMeUsecase(repo)
	result, err := uc.Execute(context.Background(), "u1")

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domain.ErrNotFound))
	repo.AssertExpectations(t)
}
