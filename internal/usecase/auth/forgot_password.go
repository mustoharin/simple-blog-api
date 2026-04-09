package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/sanitize"
)

// EmailMessage is a thin envelope passed to the email sender to avoid a circular import.
type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

// EmailSender is a minimal interface for sending email.
type EmailSender interface {
	Send(msg EmailMessage) error
}

type ForgotPasswordUsecase struct {
	users       domain.UserRepository
	tokens      domain.PasswordResetTokenRepository
	email       EmailSender
	audit       domain.AuditLogger
	frontendURL string
}

func NewForgotPasswordUsecase(
	users domain.UserRepository,
	tokens domain.PasswordResetTokenRepository,
	email EmailSender,
	audit domain.AuditLogger,
	frontendURL string,
) *ForgotPasswordUsecase {
	return &ForgotPasswordUsecase{
		users:       users,
		tokens:      tokens,
		email:       email,
		audit:       audit,
		frontendURL: frontendURL,
	}
}

func (uc *ForgotPasswordUsecase) Execute(ctx context.Context, emailAddr string) error {
	emailAddr = strings.ToLower(sanitize.Trim(emailAddr))

	user, err := uc.users.GetByEmail(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil // prevent email enumeration
		}
		return err
	}

	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return err
	}
	rawTokenStr := hex.EncodeToString(rawToken)
	hash := sha256.Sum256([]byte(rawTokenStr))
	tokenHash := hex.EncodeToString(hash[:])

	prt := &domain.PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := uc.tokens.Create(ctx, prt); err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", uc.frontendURL, rawTokenStr)
	go func() {
		_ = uc.email.Send(EmailMessage{
			To:      user.Email,
			Subject: "Password Reset Request",
			Body:    fmt.Sprintf(`<p>Click <a href="%s">here</a> to reset your password. This link expires in 1 hour.</p>`, resetURL),
		})
	}()

	go func() {
		_ = uc.audit.Log(context.WithoutCancel(ctx), &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &user.ID,
			ActorEmail:   user.Email,
			Action:       domain.AuditUserPasswordResetRequested,
			ResourceType: "user",
			ResourceID:   &user.ID,
		})
	}()

	return nil
}
