package post

import (
	"context"

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

	for _, p := range posts {
		tags, _ := uc.postTags.GetTagsForPost(ctx, p.ID)
		p.Tags = tags
	}

	return ListPostsOutput{
		Posts: posts,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	}, nil
}
