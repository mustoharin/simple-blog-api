package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/password"
	"simple-blog-api/internal/pkg/sanitize"
)

// PasswordValidator abstracts password validation for dependency injection.
type PasswordValidator interface {
	Validate(ctx context.Context, plain string) error
}

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type RegisterUsecase struct {
	users domain.UserRepository
	pwv   PasswordValidator
	audit domain.AuditLogger
}

func NewRegisterUsecase(users domain.UserRepository, pwv PasswordValidator, audit domain.AuditLogger) *RegisterUsecase {
	return &RegisterUsecase{users: users, pwv: pwv, audit: audit}
}

func (uc *RegisterUsecase) Execute(ctx context.Context, in RegisterInput) error {
	// Trim → Sanitize → Validate
	in.Email = strings.ToLower(sanitize.Trim(in.Email))
	in.Password = sanitize.Trim(in.Password)
	in.DisplayName = sanitize.SanitizeStrict(sanitize.Trim(in.DisplayName))

	if in.Email == "" {
		return domain.ErrInvalidCredentials
	}

	// Check for existing user
	existing, err := uc.users.GetByEmail(ctx, in.Email)
	if err != nil && err != domain.ErrNotFound {
		return err
	}
	if existing != nil {
		return domain.ErrEmailAlreadyExists
	}

	// Validate password
	if err := uc.pwv.Validate(ctx, in.Password); err != nil {
		return err
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return err
	}

	user := &domain.User{
		ID:           uuid.NewString(),
		Email:        in.Email,
		PasswordHash: &hash,
		DisplayName:  in.DisplayName,
		Status:       domain.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return err
	}

	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &user.ID,
		ActorEmail:   user.Email,
		Action:       domain.AuditUserRegistered,
		ResourceType: "user",
		ResourceID:   &user.ID,
	})

	return nil
}
