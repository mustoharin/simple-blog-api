package post_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/post"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdatePost_Success(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)
	audit := new(mockAuditLogger)

	existing := &domain.Post{ID: "p1", Slug: "old-slug"}
	postRepo.On("GetByID", mock.Anything, "p1").Return(existing, nil)
	postRepo.On("GetBySlug", mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)
	postRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	postTagRepo.On("SetPostTags", mock.Anything, "p1", []string{"t1"}).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := post.NewUpdatePostUsecase(postRepo, postTagRepo, audit)
	result, err := uc.Execute(context.Background(), post.UpdatePostInput{
		ID:         "p1",
		Title:      "New Title",
		Slug:       "new-slug",
		Content:    "Updated content",
		TagIDs:     []string{"t1"},
		ActorID:    "u1",
		ActorEmail: "actor@example.com",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mock.AssertExpectations(t, postRepo, postTagRepo, audit)
}

func TestUpdatePost_NotFound(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)
	audit := new(mockAuditLogger)

	postRepo.On("GetByID", mock.Anything, "missing").Return(nil, domain.ErrNotFound)

	uc := post.NewUpdatePostUsecase(postRepo, postTagRepo, audit)
	result, err := uc.Execute(context.Background(), post.UpdatePostInput{
		ID:      "missing",
		ActorID: "u1",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	mock.AssertExpectations(t, postRepo, postTagRepo, audit)
}
