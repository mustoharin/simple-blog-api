package user

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

type EmailSender interface {
	Send(to, subject, body string) error
}

type CreateUserInput struct {
	Email       string
	DisplayName string
	ActorID     string
	ActorEmail  string
}

type CreateUserUsecase struct {
	users       domain.UserRepository
	invTokens   domain.InvitationTokenRepository
	email       EmailSender
	audit       domain.AuditLogger
	frontendURL string
}

func NewCreateUserUsecase(
	users domain.UserRepository,
	invTokens domain.InvitationTokenRepository,
	email EmailSender,
	audit domain.AuditLogger,
	frontendURL string,
) *CreateUserUsecase {
	return &CreateUserUsecase{users: users, invTokens: invTokens, email: email, audit: audit, frontendURL: frontendURL}
}

func (uc *CreateUserUsecase) Execute(ctx context.Context, in CreateUserInput) (*domain.User, error) {
	in.Email = strings.ToLower(sanitize.Trim(in.Email))
	in.DisplayName = sanitize.SanitizeStrict(sanitize.Trim(in.DisplayName))

	existing, err := uc.users.GetByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	newUser := &domain.User{
		ID:          uuid.NewString(),
		Email:       in.Email,
		DisplayName: in.DisplayName,
		Status:      domain.UserStatusPendingInvitation,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.users.Create(ctx, newUser); err != nil {
		return nil, err
	}

	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return nil, err
	}
	rawTokenStr := hex.EncodeToString(rawToken)
	hash := sha256.Sum256([]byte(rawTokenStr))
	tokenHash := hex.EncodeToString(hash[:])

	inv := &domain.InvitationToken{
		ID:        uuid.NewString(),
		UserID:    newUser.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := uc.invTokens.Create(ctx, inv); err != nil {
		return nil, err
	}

	inviteURL := fmt.Sprintf("%s/accept-invitation?token=%s", uc.frontendURL, rawTokenStr)
	go func() {
		_ = uc.email.Send(newUser.Email, "You've been invited",
			fmt.Sprintf(`<p>Click <a href="%s">here</a> to accept your invitation. Expires in 24 hours.</p>`, inviteURL))
	}()

	actorID := in.ActorID
	go func() {
		_ = uc.audit.Log(context.WithoutCancel(ctx), &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &actorID,
			ActorEmail:   in.ActorEmail,
			Action:       domain.AuditUserInvited,
			ResourceType: "user",
			ResourceID:   &newUser.ID,
		})
	}()

	return newUser, nil
}
