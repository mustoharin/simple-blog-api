package post

import (
	"context"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type TogglePublishUsecase struct {
	posts domain.PostRepository
	audit domain.AuditLogger
}

func NewTogglePublishUsecase(posts domain.PostRepository, audit domain.AuditLogger) *TogglePublishUsecase {
	return &TogglePublishUsecase{posts: posts, audit: audit}
}

type TogglePublishInput struct {
	PostID     string
	Published  bool
	ActorID    string
	ActorEmail string
}

func (uc *TogglePublishUsecase) Execute(ctx context.Context, in TogglePublishInput) error {
	if _, err := uc.posts.GetByID(ctx, in.PostID); err != nil {
		return err
	}

	if err := uc.posts.SetPublished(ctx, in.PostID, in.Published); err != nil {
		return err
	}

	action := domain.AuditPostPublished
	if !in.Published {
		action = domain.AuditPostUnpublished
	}

	actorID := in.ActorID
	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &actorID,
		ActorEmail:   in.ActorEmail,
		Action:       action,
		ResourceType: "post",
		ResourceID:   &in.PostID,
	})

	return nil
}
