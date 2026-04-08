package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
}

type RefreshUsecase struct {
	users         domain.UserRepository
	refreshTokens domain.RefreshTokenRepository
	audit         domain.AuditLogger
	jwtSecret     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewRefreshUsecase(
	users domain.UserRepository,
	refreshTokens domain.RefreshTokenRepository,
	audit domain.AuditLogger,
	jwtSecret string,
	accessExpiry, refreshExpiry time.Duration,
) *RefreshUsecase {
	return &RefreshUsecase{
		users:         users,
		refreshTokens: refreshTokens,
		audit:         audit,
		jwtSecret:     jwtSecret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

func (uc *RefreshUsecase) Execute(ctx context.Context, rawToken string) (RefreshOutput, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	rt, err := uc.refreshTokens.GetByHash(ctx, tokenHash)
	if err != nil {
		return RefreshOutput{}, domain.ErrTokenNotFound
	}

	if rt.RevokedAt != nil || rt.ExpiresAt.Before(time.Now()) {
		return RefreshOutput{}, domain.ErrTokenExpired
	}

	// Revoke old token (rotation)
	if err := uc.refreshTokens.Revoke(ctx, rt.ID); err != nil {
		return RefreshOutput{}, err
	}

	user, err := uc.users.GetByID(ctx, rt.UserID)
	if err != nil {
		return RefreshOutput{}, err
	}

	perms, err := uc.users.GetPermissions(ctx, user.ID)
	if err != nil {
		return RefreshOutput{}, err
	}

	accessToken, err := IssueAccessToken(user.ID, user.Email, perms, uc.jwtSecret, uc.accessExpiry)
	if err != nil {
		return RefreshOutput{}, err
	}

	rawNew := make([]byte, 32)
	if _, err := rand.Read(rawNew); err != nil {
		return RefreshOutput{}, err
	}
	rawNewStr := hex.EncodeToString(rawNew)
	newHash := sha256.Sum256([]byte(rawNewStr))
	newTokenHash := hex.EncodeToString(newHash[:])

	newRT := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: newTokenHash,
		ExpiresAt: time.Now().Add(uc.refreshExpiry),
	}
	if err := uc.refreshTokens.Create(ctx, newRT); err != nil {
		return RefreshOutput{}, err
	}

	return RefreshOutput{
		AccessToken:  accessToken,
		RefreshToken: rawNewStr,
	}, nil
}
