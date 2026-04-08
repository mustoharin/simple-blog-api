package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type ResendInvitationUsecase struct {
	users       domain.UserRepository
	invTokens   domain.InvitationTokenRepository
	email       EmailSender
	audit       domain.AuditLogger
	frontendURL string
}

func NewResendInvitationUsecase(
	users domain.UserRepository,
	invTokens domain.InvitationTokenRepository,
	email EmailSender,
	audit domain.AuditLogger,
	frontendURL string,
) *ResendInvitationUsecase {
	return &ResendInvitationUsecase{
		users:       users,
		invTokens:   invTokens,
		email:       email,
		audit:       audit,
		frontendURL: frontendURL,
	}
}

func (uc *ResendInvitationUsecase) Execute(ctx context.Context, userID, actorID, actorEmail string) error {
	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if u.Status == domain.UserStatusActive {
		return domain.ErrUserAlreadyActive
	}

	if u.Status == domain.UserStatusExpiredInvitation {
		if err := uc.users.UpdateStatus(ctx, userID, domain.UserStatusPendingInvitation); err != nil {
			return err
		}
	}

	if err := uc.invTokens.InvalidatePrevious(ctx, userID); err != nil {
		return err
	}

	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return err
	}
	rawTokenStr := hex.EncodeToString(rawToken)
	hash := sha256.Sum256([]byte(rawTokenStr))
	tokenHash := hex.EncodeToString(hash[:])

	inv := &domain.InvitationToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := uc.invTokens.Create(ctx, inv); err != nil {
		return err
	}

	inviteURL := fmt.Sprintf("%s/accept-invitation?token=%s", uc.frontendURL, rawTokenStr)
	go func() {
		_ = uc.email.Send(u.Email, "Your invitation has been resent",
			fmt.Sprintf(`<p>Click <a href="%s">here</a> to accept your invitation. Expires in 7 days.</p>`, inviteURL))
	}()

	go func() {
		_ = uc.audit.Log(ctx, &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &actorID,
			ActorEmail:   actorEmail,
			Action:       domain.AuditUserInvitationResent,
			ResourceType: "user",
			ResourceID:   &userID,
		})
	}()

	return nil
}
