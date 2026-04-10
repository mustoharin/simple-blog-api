package profile

import (
	"context"
	"net/url"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/sanitize"
)

type UpdateMeInput struct {
	UserID      string
	DisplayName string
	Bio         string
	AvatarURL   string
}

type UpdateMeUsecase struct {
	users domain.UserRepository
	audit domain.AuditLogger
}

func NewUpdateMeUsecase(users domain.UserRepository, audit domain.AuditLogger) *UpdateMeUsecase {
	return &UpdateMeUsecase{users: users, audit: audit}
}

func (uc *UpdateMeUsecase) Execute(ctx context.Context, in UpdateMeInput) (*domain.User, error) {
	user, err := uc.users.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	user.DisplayName = sanitize.SanitizeStrict(sanitize.Trim(in.DisplayName))
	user.Bio = sanitize.SanitizeStrict(sanitize.Trim(in.Bio))
	trimmedURL := sanitize.Trim(in.AvatarURL)
	if trimmedURL != "" {
		parsed, err := url.Parse(trimmedURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return nil, domain.ErrInvalidInput
		}
	}
	user.AvatarURL = trimmedURL

	if err := uc.users.Update(ctx, user); err != nil {
		return nil, err
	}

	go func() {
		_ = uc.audit.Log(context.WithoutCancel(ctx), &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &user.ID,
			ActorEmail:   user.Email,
			Action:       domain.AuditUserUpdated,
			ResourceType: "user",
			ResourceID:   &user.ID,
		})
	}()

	return user, nil
}
