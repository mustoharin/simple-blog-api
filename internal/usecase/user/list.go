package user

import (
	"context"

	"simple-blog-api/internal/domain"
)

type ListUsersUsecase struct {
	users domain.UserRepository
}

func NewListUsersUsecase(users domain.UserRepository) *ListUsersUsecase {
	return &ListUsersUsecase{users: users}
}

type ListUsersOutput struct {
	Users []*domain.User
	Total int
	Page  int
	Limit int
}

func (uc *ListUsersUsecase) Execute(ctx context.Context, page, limit int) (ListUsersOutput, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	users, total, err := uc.users.List(ctx, page, limit)
	if err != nil {
		return ListUsersOutput{}, err
	}
	return ListUsersOutput{Users: users, Total: total, Page: page, Limit: limit}, nil
}
