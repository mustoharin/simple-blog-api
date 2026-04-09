package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/password"
	"simple-blog-api/internal/pkg/sanitize"
)

type LoginInput struct {
	Email        string
	Password     string
	CaptchaToken string
	IPAddress    string
	UserAgent    string
}

type LoginOutput struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
}

type LoginUsecase struct {
	users         domain.UserRepository
	refreshTokens domain.RefreshTokenRepository
	captcha       CaptchaVerifier
	audit         domain.AuditLogger
	jwtSecret     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewLoginUsecase(
	users domain.UserRepository,
	refreshTokens domain.RefreshTokenRepository,
	captcha CaptchaVerifier,
	audit domain.AuditLogger,
	jwtSecret string,
	accessExpiry, refreshExpiry time.Duration,
) *LoginUsecase {
	return &LoginUsecase{
		users:         users,
		refreshTokens: refreshTokens,
		captcha:       captcha,
		audit:         audit,
		jwtSecret:     jwtSecret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

func (uc *LoginUsecase) Execute(ctx context.Context, in LoginInput) (LoginOutput, error) {
	in.Email = strings.ToLower(sanitize.Trim(in.Email))
	in.Password = sanitize.Trim(in.Password)

	// Verify CAPTCHA first
	if err := uc.captcha.Verify(ctx, in.CaptchaToken); err != nil {
		return LoginOutput{}, domain.ErrInvalidCaptcha
	}

	user, err := uc.users.GetByEmail(ctx, in.Email)
	if err != nil {
		if err == domain.ErrNotFound {
			_ = uc.audit.Log(ctx, &domain.AuditLog{
				ID:         uuid.NewString(),
				ActorEmail: in.Email,
				Action:     domain.AuditUserLoginFailed,
				IPAddress:  in.IPAddress,
				UserAgent:  in.UserAgent,
			})
			return LoginOutput{}, domain.ErrInvalidCredentials
		}
		return LoginOutput{}, err
	}

	if user.Status != domain.UserStatusActive {
		return LoginOutput{}, domain.ErrAccountNotActivated
	}

	if user.PasswordHash == nil || !password.Verify(in.Password, *user.PasswordHash) {
		_ = uc.audit.Log(ctx, &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &user.ID,
			ActorEmail:   user.Email,
			Action:       domain.AuditUserLoginFailed,
			ResourceType: "user",
			ResourceID:   &user.ID,
			IPAddress:    in.IPAddress,
			UserAgent:    in.UserAgent,
		})
		return LoginOutput{}, domain.ErrInvalidCredentials
	}

	perms, err := uc.users.GetPermissions(ctx, user.ID)
	if err != nil {
		return LoginOutput{}, err
	}

	accessToken, err := IssueAccessToken(user.ID, user.Email, perms, uc.jwtSecret, uc.accessExpiry)
	if err != nil {
		return LoginOutput{}, err
	}

	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return LoginOutput{}, err
	}
	rawTokenStr := hex.EncodeToString(rawToken)
	hash := sha256.Sum256([]byte(rawTokenStr))
	tokenHash := hex.EncodeToString(hash[:])

	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(uc.refreshExpiry),
	}
	if err := uc.refreshTokens.Create(ctx, rt); err != nil {
		return LoginOutput{}, err
	}

	_ = uc.users.UpdateLastLogin(ctx, user.ID)

	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &user.ID,
		ActorEmail:   user.Email,
		Action:       domain.AuditUserLogin,
		ResourceType: "user",
		ResourceID:   &user.ID,
		IPAddress:    in.IPAddress,
		UserAgent:    in.UserAgent,
	})

	return LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: rawTokenStr,
		User:         user,
	}, nil
}
