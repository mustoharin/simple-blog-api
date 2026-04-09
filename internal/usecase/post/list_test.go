package post_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/post"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListPosts_Success(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)

	posts := []*domain.Post{
		{ID: "p1"},
		{ID: "p2"},
	}
	filter := domain.PostFilter{Page: 1, Limit: 10}
	postRepo.On("List", mock.Anything, filter, false).Return(posts, 2, nil)
	postTagRepo.On("GetTagsForPost", mock.Anything, mock.Anything).Return([]domain.Tag{}, nil)

	uc := post.NewListPostsUsecase(postRepo, postTagRepo)
	out, err := uc.Execute(context.Background(), filter, false)

	assert.NoError(t, err)
	assert.Equal(t, 2, out.Total)
	assert.Len(t, out.Posts, 2)
	postRepo.AssertExpectations(t)
	postTagRepo.AssertExpectations(t)
}

func TestListPosts_DefaultPage(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)

	expectedFilter := domain.PostFilter{Page: 1, Limit: 5}
	postRepo.On("List", mock.Anything, expectedFilter, false).Return([]*domain.Post{}, 0, nil)

	uc := post.NewListPostsUsecase(postRepo, postTagRepo)
	out, err := uc.Execute(context.Background(), domain.PostFilter{Page: 0, Limit: 5}, false)

	assert.NoError(t, err)
	assert.Equal(t, 1, out.Page)
	postRepo.AssertExpectations(t)
	postTagRepo.AssertExpectations(t)
}

func TestListPosts_LimitClamped(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)

	expectedFilter := domain.PostFilter{Page: 1, Limit: 20}
	postRepo.On("List", mock.Anything, expectedFilter, false).Return([]*domain.Post{}, 0, nil)

	uc := post.NewListPostsUsecase(postRepo, postTagRepo)
	out, err := uc.Execute(context.Background(), domain.PostFilter{Page: 1, Limit: 200}, false)

	assert.NoError(t, err)
	assert.Equal(t, 20, out.Limit)
	postRepo.AssertExpectations(t)
	postTagRepo.AssertExpectations(t)
}
