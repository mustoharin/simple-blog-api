package comment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type UpdateStatusUsecase struct {
	comments domain.CommentRepository
	audit    domain.AuditLogger
}

func NewUpdateStatusUsecase(comments domain.CommentRepository, audit domain.AuditLogger) *UpdateStatusUsecase {
	return &UpdateStatusUsecase{comments: comments, audit: audit}
}

func (uc *UpdateStatusUsecase) Execute(ctx context.Context, commentID, statusStr, actorID, actorEmail string) error {
	var status domain.CommentStatus
	switch statusStr {
	case "approved":
		status = domain.CommentStatusApproved
	case "rejected":
		status = domain.CommentStatusRejected
	default:
		return fmt.Errorf("invalid status: %s", statusStr)
	}

	c, err := uc.comments.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	if err := uc.comments.UpdateStatus(ctx, commentID, status); err != nil {
		return err
	}

	action := domain.AuditCommentApproved
	if status == domain.CommentStatusRejected {
		action = domain.AuditCommentRejected
	}

	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &actorID,
		ActorEmail:   actorEmail,
		Action:       action,
		ResourceType: "comment",
		ResourceID:   &c.ID,
	})

	return nil
}
