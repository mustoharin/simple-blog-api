package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/password"
	"simple-blog-api/internal/pkg/sanitize"
)

type ResetPasswordUsecase struct {
	users         domain.UserRepository
	tokens        domain.PasswordResetTokenRepository
	refreshTokens domain.RefreshTokenRepository
	pwv           PasswordValidator
	audit         domain.AuditLogger
}

func NewResetPasswordUsecase(
	users domain.UserRepository,
	tokens domain.PasswordResetTokenRepository,
	refreshTokens domain.RefreshTokenRepository,
	pwv PasswordValidator,
	audit domain.AuditLogger,
) *ResetPasswordUsecase {
	return &ResetPasswordUsecase{
		users:         users,
		tokens:        tokens,
		refreshTokens: refreshTokens,
		pwv:           pwv,
		audit:         audit,
	}
}

func (uc *ResetPasswordUsecase) Execute(ctx context.Context, rawToken, newPassword string) error {
	newPassword = sanitize.Trim(newPassword)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	prt, err := uc.tokens.GetByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrTokenNotFound
		}
		return err
	}

	if prt.UsedAt != nil || prt.ExpiresAt.Before(time.Now()) {
		return domain.ErrTokenExpired
	}

	if err := uc.pwv.Validate(ctx, newPassword); err != nil {
		return err
	}

	user, err := uc.users.GetByID(ctx, prt.UserID)
	if err != nil {
		return err
	}

	newHash, err := password.Hash(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = &newHash
	if err := uc.users.Update(ctx, user); err != nil {
		return err
	}

	if err := uc.tokens.MarkUsed(ctx, prt.ID); err != nil {
		return err
	}

	go func() {
		_ = uc.refreshTokens.RevokeAllForUser(context.WithoutCancel(ctx), user.ID)
	}()

	go func() {
		_ = uc.audit.Log(context.WithoutCancel(ctx), &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &user.ID,
			ActorEmail:   user.Email,
			Action:       domain.AuditUserPasswordResetCompleted,
			ResourceType: "user",
			ResourceID:   &user.ID,
		})
	}()

	return nil
}
