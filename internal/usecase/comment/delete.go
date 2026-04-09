package comment

import (
	"context"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type DeleteCommentUsecase struct {
	comments domain.CommentRepository
	audit    domain.AuditLogger
}

func NewDeleteCommentUsecase(comments domain.CommentRepository, audit domain.AuditLogger) *DeleteCommentUsecase {
	return &DeleteCommentUsecase{comments: comments, audit: audit}
}

func (uc *DeleteCommentUsecase) Execute(ctx context.Context, commentID, actorID, actorEmail string) error {
	c, err := uc.comments.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	if err := uc.comments.SoftDelete(ctx, commentID); err != nil {
		return err
	}

	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &actorID,
		ActorEmail:   actorEmail,
		Action:       domain.AuditCommentDeleted,
		ResourceType: "comment",
		ResourceID:   &c.ID,
	})

	return nil
}
