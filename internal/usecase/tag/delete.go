package tag

import (
	"context"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type DeleteTagUsecase struct {
	tags  domain.TagRepository
	audit domain.AuditLogger
}

func NewDeleteTagUsecase(tags domain.TagRepository, audit domain.AuditLogger) *DeleteTagUsecase {
	return &DeleteTagUsecase{tags: tags, audit: audit}
}

func (uc *DeleteTagUsecase) Execute(ctx context.Context, tagID, actorID, actorEmail string) error {
	if err := uc.tags.HardDelete(ctx, tagID); err != nil {
		return err
	}

	go func() {
		_ = uc.audit.Log(ctx, &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &actorID,
			ActorEmail:   actorEmail,
			Action:       domain.AuditTagDeleted,
			ResourceType: "tag",
			ResourceID:   &tagID,
		})
	}()

	return nil
}
