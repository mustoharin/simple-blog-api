package comment_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/comment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateCommentStatus_Reject(t *testing.T) {
	commentRepo := new(mockCommentRepo)
	audit := new(mockAuditLogger)

	c := &domain.Comment{ID: "c1", PostID: "p1", AuthorID: "u2", Status: domain.CommentStatusPending}
	commentRepo.On("GetByID", mock.Anything, "c1").Return(c, nil)
	commentRepo.On("UpdateStatus", mock.Anything, "c1", domain.CommentStatusRejected).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := comment.NewUpdateStatusUsecase(commentRepo, audit)
	err := uc.Execute(context.Background(), "c1", "rejected", "mod1", "mod@example.com")
	assert.NoError(t, err)
	commentRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestUpdateCommentStatus_InvalidStatus(t *testing.T) {
	commentRepo := new(mockCommentRepo)
	audit := new(mockAuditLogger)

	c := &domain.Comment{ID: "c1", PostID: "p1", AuthorID: "u2", Status: domain.CommentStatusPending}
	commentRepo.On("GetByID", mock.Anything, "c1").Return(c, nil)

	uc := comment.NewUpdateStatusUsecase(commentRepo, audit)
	err := uc.Execute(context.Background(), "c1", "unknown", "mod1", "mod@example.com")
	assert.Error(t, err)
}
