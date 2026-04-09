package comment_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/comment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListComments_Success(t *testing.T) {
	commentRepo := new(mockCommentRepo)

	comments := []*domain.Comment{
		{ID: "c1", PostID: "p1", Body: "First"},
		{ID: "c2", PostID: "p1", Body: "Second"},
	}
	commentRepo.On("List", mock.Anything, "p1", 1, 10).Return(comments, 2, nil)

	uc := comment.NewListCommentsUsecase(commentRepo)
	out, err := uc.Execute(context.Background(), "p1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, out.Total)
	assert.Len(t, out.Comments, 2)
}

func TestListComments_DefaultPage(t *testing.T) {
	commentRepo := new(mockCommentRepo)

	commentRepo.On("List", mock.Anything, "p1", 1, 20).Return([]*domain.Comment{}, 0, nil)

	uc := comment.NewListCommentsUsecase(commentRepo)
	out, err := uc.Execute(context.Background(), "p1", 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, out.Page)
	assert.Equal(t, 20, out.Limit)
	commentRepo.AssertExpectations(t)
}

func TestListComments_LimitClamped(t *testing.T) {
	commentRepo := new(mockCommentRepo)

	commentRepo.On("List", mock.Anything, "p1", 1, 100).Return([]*domain.Comment{}, 0, nil)

	uc := comment.NewListCommentsUsecase(commentRepo)
	_, err := uc.Execute(context.Background(), "p1", 1, 200)
	assert.NoError(t, err)
	commentRepo.AssertExpectations(t)
}
