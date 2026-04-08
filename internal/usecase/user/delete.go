package user

import (
	"context"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type DeleteUserUsecase struct {
	users         domain.UserRepository
	refreshTokens domain.RefreshTokenRepository
	audit         domain.AuditLogger
}

func NewDeleteUserUsecase(
	users domain.UserRepository,
	refreshTokens domain.RefreshTokenRepository,
	audit domain.AuditLogger,
) *DeleteUserUsecase {
	return &DeleteUserUsecase{users: users, refreshTokens: refreshTokens, audit: audit}
}

func (uc *DeleteUserUsecase) Execute(ctx context.Context, userID, actorID, actorEmail string) error {
	if _, err := uc.users.GetByID(ctx, userID); err != nil {
		return err
	}

	if err := uc.users.SoftDelete(ctx, userID); err != nil {
		return err
	}

	go func() { _ = uc.refreshTokens.RevokeAllForUser(ctx, userID) }()

	go func() {
		_ = uc.audit.Log(ctx, &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &actorID,
			ActorEmail:   actorEmail,
			Action:       domain.AuditUserDeleted,
			ResourceType: "user",
			ResourceID:   &userID,
		})
	}()

	return nil
}
