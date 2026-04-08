package post

import (
	"context"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type DeletePostUsecase struct {
	posts    domain.PostRepository
	comments domain.CommentRepository
	audit    domain.AuditLogger
}

func NewDeletePostUsecase(
	posts domain.PostRepository,
	comments domain.CommentRepository,
	audit domain.AuditLogger,
) *DeletePostUsecase {
	return &DeletePostUsecase{posts: posts, comments: comments, audit: audit}
}

func (uc *DeletePostUsecase) Execute(ctx context.Context, postID, actorID, actorEmail string) error {
	if _, err := uc.posts.GetByID(ctx, postID); err != nil {
		return err
	}

	_ = uc.comments.SoftDeleteByPostID(ctx, postID)

	if err := uc.posts.SoftDelete(ctx, postID); err != nil {
		return err
	}

	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &actorID,
		ActorEmail:   actorEmail,
		Action:       domain.AuditPostDeleted,
		ResourceType: "post",
		ResourceID:   &postID,
	})

	return nil
}
