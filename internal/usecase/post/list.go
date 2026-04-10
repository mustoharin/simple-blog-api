package post

import (
	"context"
	"fmt"

	"simple-blog-api/internal/domain"
)

type ListPostsUsecase struct {
	posts    domain.PostRepository
	postTags domain.PostTagRepository
}

func NewListPostsUsecase(posts domain.PostRepository, postTags domain.PostTagRepository) *ListPostsUsecase {
	return &ListPostsUsecase{posts: posts, postTags: postTags}
}

type ListPostsOutput struct {
	Posts []*domain.Post
	Total int
	Page  int
	Limit int
}

func (uc *ListPostsUsecase) Execute(ctx context.Context, filter domain.PostFilter, publicOnly bool) (ListPostsOutput, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	posts, total, err := uc.posts.List(ctx, filter, publicOnly)
	if err != nil {
		return ListPostsOutput{}, err
	}

	if len(posts) > 0 {
		postIDs := make([]string, len(posts))
		for i, p := range posts {
			postIDs[i] = p.ID
		}
		tagMap, err := uc.postTags.GetTagsForPosts(ctx, postIDs)
		if err != nil {
			return ListPostsOutput{}, fmt.Errorf("fetch post tags: %w", err)
		}
		for _, p := range posts {
			if tags, ok := tagMap[p.ID]; ok {
				p.Tags = tags
			} else {
				p.Tags = []domain.Tag{}
			}
		}
	}

	return ListPostsOutput{
		Posts: posts,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	}, nil
}
