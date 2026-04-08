package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"simple-blog-api/internal/domain"
)

// --- RefreshTokenRepository ---

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, t *domain.RefreshToken) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)`,
		t.ID, t.UserID, t.TokenHash, t.ExpiresAt)
	return err
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	t := &domain.RefreshToken{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens WHERE token_hash = $1`, hash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return t, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at=NOW() WHERE user_id=$1 AND revoked_at IS NULL`, userID)
	return err
}

// --- PasswordResetTokenRepository ---

type PasswordResetTokenRepository struct {
	db *pgxpool.Pool
}

func NewPasswordResetTokenRepository(db *pgxpool.Pool) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{db: db}
}

func (r *PasswordResetTokenRepository) Create(ctx context.Context, t *domain.PasswordResetToken) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)`,
		t.ID, t.UserID, t.TokenHash, t.ExpiresAt)
	return err
}

func (r *PasswordResetTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.PasswordResetToken, error) {
	t := &domain.PasswordResetToken{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at
		FROM password_reset_tokens WHERE token_hash = $1`, hash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("get password reset token: %w", err)
	}
	return t, nil
}

func (r *PasswordResetTokenRepository) MarkUsed(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE password_reset_tokens SET used_at=NOW() WHERE id=$1`, id)
	return err
}

// --- InvitationTokenRepository ---

type InvitationTokenRepository struct {
	db *pgxpool.Pool
}

func NewInvitationTokenRepository(db *pgxpool.Pool) *InvitationTokenRepository {
	return &InvitationTokenRepository{db: db}
}

func (r *InvitationTokenRepository) Create(ctx context.Context, t *domain.InvitationToken) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO invitation_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)`,
		t.ID, t.UserID, t.TokenHash, t.ExpiresAt)
	return err
}

func (r *InvitationTokenRepository) GetByUserID(ctx context.Context, userID string) (*domain.InvitationToken, error) {
	t := &domain.InvitationToken{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM invitation_tokens WHERE user_id = $1 AND used_at IS NULL
		ORDER BY created_at DESC LIMIT 1`, userID,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("get invitation token by user: %w", err)
	}
	return t, nil
}

func (r *InvitationTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.InvitationToken, error) {
	t := &domain.InvitationToken{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM invitation_tokens WHERE token_hash = $1`, hash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("get invitation token by hash: %w", err)
	}
	return t, nil
}

func (r *InvitationTokenRepository) MarkUsed(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE invitation_tokens SET used_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *InvitationTokenRepository) InvalidatePrevious(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE invitation_tokens SET used_at=NOW() WHERE user_id=$1 AND used_at IS NULL`, userID)
	return err
}
