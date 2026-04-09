package auth_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func makeTokenHash(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func TestRefresh_Success(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	audit := new(mockAuditLogger)

	raw := "raw-refresh-token-value-abcdef"
	hash := makeTokenHash(raw)

	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    "u1",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	user := &domain.User{ID: "u1", Email: "alice@example.com", Status: domain.UserStatusActive}

	rtRepo.On("GetByHash", mock.Anything, hash).Return(rt, nil)
	rtRepo.On("Revoke", mock.Anything, rt.ID).Return(nil)
	rtRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetByID", mock.Anything, "u1").Return(user, nil)
	repo.On("GetPermissions", mock.Anything, "u1").Return([]string{"post:create"}, nil)

	uc := auth.NewRefreshUsecase(repo, rtRepo, audit, "secret", 15*time.Minute, 30*24*time.Hour)
	out, err := uc.Execute(context.Background(), raw)
	assert.NoError(t, err)
	assert.NotEmpty(t, out.AccessToken)
	assert.NotEmpty(t, out.RefreshToken)
	assert.NotEqual(t, raw, out.RefreshToken) // token was rotated
}

func TestRefresh_TokenExpired(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	audit := new(mockAuditLogger)

	raw := "raw-expired-token"
	hash := makeTokenHash(raw)

	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    "u1",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(-time.Hour), // expired
	}
	rtRepo.On("GetByHash", mock.Anything, hash).Return(rt, nil)

	uc := auth.NewRefreshUsecase(repo, rtRepo, audit, "secret", 15*time.Minute, 30*24*time.Hour)
	_, err := uc.Execute(context.Background(), raw)
	assert.ErrorIs(t, err, domain.ErrTokenExpired)
}

func TestRefresh_TokenRevoked(t *testing.T) {
	repo := new(mockUserRepo)
	rtRepo := new(mockRefreshTokenRepo)
	audit := new(mockAuditLogger)

	raw := "raw-revoked-token"
	hash := makeTokenHash(raw)
	now := time.Now()

	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    "u1",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(time.Hour),
		RevokedAt: &now,
	}
	rtRepo.On("GetByHash", mock.Anything, hash).Return(rt, nil)

	uc := auth.NewRefreshUsecase(repo, rtRepo, audit, "secret", 15*time.Minute, 30*24*time.Hour)
	_, err := uc.Execute(context.Background(), raw)
	assert.ErrorIs(t, err, domain.ErrTokenExpired)
}

func TestLogout_Success(t *testing.T) {
	rtRepo := new(mockRefreshTokenRepo)
	audit := new(mockAuditLogger)

	raw := "raw-logout-token"
	hash := makeTokenHash(raw)
	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    "u1",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	rtRepo.On("GetByHash", mock.Anything, hash).Return(rt, nil)
	rtRepo.On("Revoke", mock.Anything, rt.ID).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := auth.NewLogoutUsecase(rtRepo, audit)
	err := uc.Execute(context.Background(), raw, "u1", "alice@example.com")
	assert.NoError(t, err)
}
