package post_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/post"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBySlug_Success_IncrementView(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)

	p := &domain.Post{ID: "p1", Slug: "my-post", Status: domain.PostStatusPublished}
	postRepo.On("GetBySlug", mock.Anything, "my-post").Return(p, nil)
	postTagRepo.On("GetTagsForPost", mock.Anything, "p1").Return([]domain.Tag{}, nil)
	postRepo.On("IncrementViewCount", mock.Anything, "p1").Return(nil)

	uc := post.NewGetBySlugUsecase(postRepo, postTagRepo)
	result, err := uc.Execute(context.Background(), "my-post", false)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	postRepo.AssertExpectations(t)
	postTagRepo.AssertExpectations(t)
}

func TestGetBySlug_Success_SkipIncrement(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)

	p := &domain.Post{ID: "p1", Slug: "my-post", Status: domain.PostStatusPublished}
	postRepo.On("GetBySlug", mock.Anything, "my-post").Return(p, nil)
	postTagRepo.On("GetTagsForPost", mock.Anything, "p1").Return([]domain.Tag{}, nil)

	uc := post.NewGetBySlugUsecase(postRepo, postTagRepo)
	result, err := uc.Execute(context.Background(), "my-post", true)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	postRepo.AssertNotCalled(t, "IncrementViewCount", mock.Anything, mock.Anything)
	postRepo.AssertExpectations(t)
	postTagRepo.AssertExpectations(t)
}

func TestGetBySlug_NotFound(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)

	postRepo.On("GetBySlug", mock.Anything, "no-such-post").Return(nil, domain.ErrNotFound)

	uc := post.NewGetBySlugUsecase(postRepo, postTagRepo)
	result, err := uc.Execute(context.Background(), "no-such-post", false)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	postRepo.AssertExpectations(t)
	postTagRepo.AssertExpectations(t)
}
