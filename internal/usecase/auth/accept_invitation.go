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

type AcceptInvitationUsecase struct {
	users     domain.UserRepository
	invTokens domain.InvitationTokenRepository
	pwv       PasswordValidator
	audit     domain.AuditLogger
}

func NewAcceptInvitationUsecase(
	users domain.UserRepository,
	invTokens domain.InvitationTokenRepository,
	pwv PasswordValidator,
	audit domain.AuditLogger,
) *AcceptInvitationUsecase {
	return &AcceptInvitationUsecase{
		users:     users,
		invTokens: invTokens,
		pwv:       pwv,
		audit:     audit,
	}
}

func (uc *AcceptInvitationUsecase) Execute(ctx context.Context, rawToken, newPassword string) error {
	newPassword = sanitize.Trim(newPassword)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	inv, err := uc.invTokens.GetByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrTokenNotFound
		}
		return err
	}

	if inv.UsedAt != nil || inv.ExpiresAt.Before(time.Now()) {
		return domain.ErrTokenExpired
	}

	if err := uc.pwv.Validate(ctx, newPassword); err != nil {
		return err
	}

	user, err := uc.users.GetByID(ctx, inv.UserID)
	if err != nil {
		return err
	}

	pwHash, err := password.Hash(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = &pwHash
	user.Status = domain.UserStatusActive
	if err := uc.users.Update(ctx, user); err != nil {
		return err
	}

	if err := uc.invTokens.MarkUsed(ctx, inv.ID); err != nil {
		return err
	}

	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &user.ID,
		ActorEmail:   user.Email,
		Action:       domain.AuditUserInvitationAccepted,
		ResourceType: "user",
		ResourceID:   &user.ID,
	})

	return nil
}
