package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type LogoutUsecase struct {
	refreshTokens domain.RefreshTokenRepository
	audit         domain.AuditLogger
}

func NewLogoutUsecase(refreshTokens domain.RefreshTokenRepository, audit domain.AuditLogger) *LogoutUsecase {
	return &LogoutUsecase{refreshTokens: refreshTokens, audit: audit}
}

func (uc *LogoutUsecase) Execute(ctx context.Context, rawToken, userID, userEmail string) error {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	rt, err := uc.refreshTokens.GetByHash(ctx, tokenHash)
	if err != nil {
		// Token not found is OK for logout (idempotent)
		return nil
	}

	if err := uc.refreshTokens.Revoke(ctx, rt.ID); err != nil {
		return err
	}

	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &userID,
		ActorEmail:   userEmail,
		Action:       domain.AuditUserLogout,
		ResourceType: "user",
		ResourceID:   &userID,
	})

	return nil
}
