package profile

import (
	"context"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/auth"
	"simple-blog-api/internal/pkg/password"
	"simple-blog-api/internal/pkg/sanitize"
)

type ChangePasswordUsecase struct {
	users         domain.UserRepository
	refreshTokens domain.RefreshTokenRepository
	pwv           auth.PasswordValidator
	audit         domain.AuditLogger
}

func NewChangePasswordUsecase(
	users domain.UserRepository,
	refreshTokens domain.RefreshTokenRepository,
	pwv auth.PasswordValidator,
	audit domain.AuditLogger,
) *ChangePasswordUsecase {
	return &ChangePasswordUsecase{
		users:         users,
		refreshTokens: refreshTokens,
		pwv:           pwv,
		audit:         audit,
	}
}

func (uc *ChangePasswordUsecase) Execute(ctx context.Context, userID, currentPassword, newPassword string) error {
	currentPassword = sanitize.Trim(currentPassword)
	newPassword = sanitize.Trim(newPassword)

	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if u.PasswordHash == nil || !password.Verify(currentPassword, *u.PasswordHash) {
		return domain.ErrInvalidCredentials
	}

	if err := uc.pwv.Validate(ctx, newPassword); err != nil {
		return err
	}

	newHash, err := password.Hash(newPassword)
	if err != nil {
		return err
	}

	u.PasswordHash = &newHash
	if err := uc.users.UpdatePasswordHash(ctx, u.ID, newHash); err != nil {
		return err
	}

	if err := uc.refreshTokens.RevokeAllForUser(ctx, userID); err != nil {
		return err
	}

	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &userID,
		ActorEmail:   u.Email,
		Action:       domain.AuditUserPasswordChanged,
		ResourceType: "user",
		ResourceID:   &userID,
	})

	return nil
}
