package post_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/post"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeletePost_Success(t *testing.T) {
	postRepo := new(mockPostRepo)
	commentRepo := new(mockCommentRepo)
	audit := new(mockAuditLogger)

	existing := &domain.Post{ID: "p1"}
	postRepo.On("GetByID", mock.Anything, "p1").Return(existing, nil)
	commentRepo.On("SoftDeleteByPostID", mock.Anything, "p1").Return(nil)
	postRepo.On("SoftDelete", mock.Anything, "p1").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := post.NewDeletePostUsecase(postRepo, commentRepo, audit)
	err := uc.Execute(context.Background(), "p1", "u1", "actor@example.com")

	assert.NoError(t, err)
	mock.AssertExpectations(t, postRepo, commentRepo, audit)
}

func TestDeletePost_NotFound(t *testing.T) {
	postRepo := new(mockPostRepo)
	commentRepo := new(mockCommentRepo)
	audit := new(mockAuditLogger)

	postRepo.On("GetByID", mock.Anything, "missing").Return(nil, domain.ErrNotFound)

	uc := post.NewDeletePostUsecase(postRepo, commentRepo, audit)
	err := uc.Execute(context.Background(), "missing", "u1", "actor@example.com")

	assert.ErrorIs(t, err, domain.ErrNotFound)
	mock.AssertExpectations(t, postRepo, commentRepo, audit)
}
