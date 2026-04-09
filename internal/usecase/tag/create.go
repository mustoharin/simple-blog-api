package tag

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/sanitize"
)

var nonAlpha = regexp.MustCompile(`[^a-z0-9]+`)

type CreateTagUsecase struct {
	tags  domain.TagRepository
	audit domain.AuditLogger
}

func NewCreateTagUsecase(tags domain.TagRepository, audit domain.AuditLogger) *CreateTagUsecase {
	return &CreateTagUsecase{tags: tags, audit: audit}
}

func (uc *CreateTagUsecase) Execute(ctx context.Context, name, actorID, actorEmail string) (*domain.Tag, error) {
	name = sanitize.SanitizeStrict(sanitize.Trim(name))
	slug := strings.Trim(nonAlpha.ReplaceAllString(strings.ToLower(name), "-"), "-")

	t := &domain.Tag{
		ID:   uuid.NewString(),
		Name: name,
		Slug: slug,
	}

	if err := uc.tags.Create(ctx, t); err != nil {
		return nil, err
	}

	go func() {
		_ = uc.audit.Log(context.WithoutCancel(ctx), &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &actorID,
			ActorEmail:   actorEmail,
			Action:       domain.AuditTagCreated,
			ResourceType: "tag",
			ResourceID:   &t.ID,
		})
	}()

	return t, nil
}
