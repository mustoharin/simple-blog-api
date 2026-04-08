package user

import (
	"context"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/sanitize"
)

type UpdateUserInput struct {
	ID          string
	DisplayName string
	Bio         string
	AvatarURL   string
	ActorID     string
	ActorEmail  string
}

type UpdateUserUsecase struct {
	users domain.UserRepository
	audit domain.AuditLogger
}

func NewUpdateUserUsecase(users domain.UserRepository, audit domain.AuditLogger) *UpdateUserUsecase {
	return &UpdateUserUsecase{users: users, audit: audit}
}

func (uc *UpdateUserUsecase) Execute(ctx context.Context, in UpdateUserInput) (*domain.User, error) {
	u, err := uc.users.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	u.DisplayName = sanitize.SanitizeStrict(sanitize.Trim(in.DisplayName))
	u.Bio = sanitize.SanitizeStrict(sanitize.Trim(in.Bio))
	u.AvatarURL = sanitize.Trim(in.AvatarURL)

	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}

	actorID := in.ActorID
	go func() {
		_ = uc.audit.Log(context.WithoutCancel(ctx), &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &actorID,
			ActorEmail:   in.ActorEmail,
			Action:       domain.AuditUserUpdated,
			ResourceType: "user",
			ResourceID:   &u.ID,
		})
	}()

	return u, nil
}
