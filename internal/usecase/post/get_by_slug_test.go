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
	result, err := uc.Execute(context.Background(), "my-post", false, false)

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
	result, err := uc.Execute(context.Background(), "my-post", true, false)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	postRepo.AssertNotCalled(t, "IncrementViewCount", mock.Anything, mock.Anything)
	postRepo.AssertExpectations(t)
	postTagRepo.AssertExpectations(t)
}

func TestGetBySlug_PublicOnly_HidesDraft(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)

	p := &domain.Post{ID: "p1", Slug: "draft-post", Status: domain.PostStatusDraft}
	postRepo.On("GetBySlug", mock.Anything, "draft-post").Return(p, nil)
	postTagRepo.On("GetTagsForPost", mock.Anything, "p1").Return([]domain.Tag{}, nil)

	uc := post.NewGetBySlugUsecase(postRepo, postTagRepo)
	result, err := uc.Execute(context.Background(), "draft-post", false, true)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrNotFound, "public caller must not see draft posts")
}

func TestGetBySlug_PublicOnly_ShowsPublished(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)

	p := &domain.Post{ID: "p1", Slug: "pub-post", Status: domain.PostStatusPublished}
	postRepo.On("GetBySlug", mock.Anything, "pub-post").Return(p, nil)
	postTagRepo.On("GetTagsForPost", mock.Anything, "p1").Return([]domain.Tag{}, nil)
	postRepo.On("IncrementViewCount", mock.Anything, "p1").Return(nil)

	uc := post.NewGetBySlugUsecase(postRepo, postTagRepo)
	result, err := uc.Execute(context.Background(), "pub-post", false, true)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGetBySlug_AdminAccess_SeesDraft(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)

	p := &domain.Post{ID: "p1", Slug: "draft-post", Status: domain.PostStatusDraft}
	postRepo.On("GetBySlug", mock.Anything, "draft-post").Return(p, nil)
	postTagRepo.On("GetTagsForPost", mock.Anything, "p1").Return([]domain.Tag{}, nil)

	uc := post.NewGetBySlugUsecase(postRepo, postTagRepo)
	// publicOnly=false → admin/authenticated caller, drafts allowed
	result, err := uc.Execute(context.Background(), "draft-post", true, false)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGetBySlug_NotFound(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)

	postRepo.On("GetBySlug", mock.Anything, "no-such-post").Return(nil, domain.ErrNotFound)

	uc := post.NewGetBySlugUsecase(postRepo, postTagRepo)
	result, err := uc.Execute(context.Background(), "no-such-post", false, false)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	postRepo.AssertExpectations(t)
	postTagRepo.AssertExpectations(t)
}
