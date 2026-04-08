package comment_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/comment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteComment_Success(t *testing.T) {
	commentRepo := new(mockCommentRepo)
	audit := new(mockAuditLogger)

	c := &domain.Comment{ID: "c1", PostID: "p1", AuthorID: "u1"}
	commentRepo.On("GetByID", mock.Anything, "c1").Return(c, nil)
	commentRepo.On("SoftDelete", mock.Anything, "c1").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := comment.NewDeleteCommentUsecase(commentRepo, audit)
	err := uc.Execute(context.Background(), "c1", "mod1", "mod@example.com")
	assert.NoError(t, err)
	commentRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestDeleteComment_NotFound(t *testing.T) {
	commentRepo := new(mockCommentRepo)
	audit := new(mockAuditLogger)

	commentRepo.On("GetByID", mock.Anything, "ghost").Return(nil, domain.ErrNotFound)

	uc := comment.NewDeleteCommentUsecase(commentRepo, audit)
	err := uc.Execute(context.Background(), "ghost", "mod1", "mod@example.com")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
