package comment

import (
	"context"
	"time"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/sanitize"
)

type PostGetter interface {
	GetByID(ctx context.Context, id string) (*domain.Post, error)
}

type CreateCommentInput struct {
	PostID      string
	AuthorID    string
	AuthorEmail string
	Body        string
}

type CreateCommentUsecase struct {
	posts    PostGetter
	comments domain.CommentRepository
	audit    domain.AuditLogger
}

func NewCreateCommentUsecase(posts PostGetter, comments domain.CommentRepository, audit domain.AuditLogger) *CreateCommentUsecase {
	return &CreateCommentUsecase{posts: posts, comments: comments, audit: audit}
}

func (uc *CreateCommentUsecase) Execute(ctx context.Context, in CreateCommentInput) (*domain.Comment, error) {
	in.Body = sanitize.SanitizeStrict(sanitize.Trim(in.Body))

	if _, err := uc.posts.GetByID(ctx, in.PostID); err != nil {
		return nil, err
	}

	c := &domain.Comment{
		ID:        uuid.NewString(),
		PostID:    in.PostID,
		AuthorID:  in.AuthorID,
		Body:      in.Body,
		Status:    domain.CommentStatusPending,
		CreatedAt: time.Now(),
	}

	if err := uc.comments.Create(ctx, c); err != nil {
		return nil, err
	}

	actorID := in.AuthorID
	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &actorID,
		ActorEmail:   in.AuthorEmail,
		Action:       domain.AuditCommentCreated,
		ResourceType: "comment",
		ResourceID:   &c.ID,
	})

	return c, nil
}
