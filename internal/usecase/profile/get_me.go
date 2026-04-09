package profile

import (
	"context"

	"simple-blog-api/internal/domain"
)

type GetMeUsecase struct {
	users domain.UserRepository
}

func NewGetMeUsecase(users domain.UserRepository) *GetMeUsecase {
	return &GetMeUsecase{users: users}
}

func (uc *GetMeUsecase) Execute(ctx context.Context, userID string) (*domain.User, error) {
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	roles, _ := uc.users.GetRoles(ctx, userID)
	user.Roles = roles
	return user, nil
}
