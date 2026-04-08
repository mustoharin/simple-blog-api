package user_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListUsers_Success(t *testing.T) {
	repo := new(mockUserRepo)

	users := []*domain.User{{ID: "u1"}, {ID: "u2"}}
	repo.On("List", mock.Anything, 1, 20).Return(users, 2, nil)

	uc := user.NewListUsersUsecase(repo)
	out, err := uc.Execute(context.Background(), 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 2, out.Total)
	assert.Len(t, out.Users, 2)
	repo.AssertExpectations(t)
}

func TestListUsers_DefaultPagination(t *testing.T) {
	repo := new(mockUserRepo)

	repo.On("List", mock.Anything, 1, 20).Return([]*domain.User{}, 0, nil)

	uc := user.NewListUsersUsecase(repo)
	out, err := uc.Execute(context.Background(), 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, out.Page)
	assert.Equal(t, 20, out.Limit)
	repo.AssertExpectations(t)
}

func TestListUsers_LimitClamped(t *testing.T) {
	repo := new(mockUserRepo)

	repo.On("List", mock.Anything, 1, 100).Return([]*domain.User{}, 0, nil)

	uc := user.NewListUsersUsecase(repo)
	out, err := uc.Execute(context.Background(), 1, 200)
	assert.NoError(t, err)
	assert.Equal(t, 100, out.Limit)
	repo.AssertExpectations(t)
}
