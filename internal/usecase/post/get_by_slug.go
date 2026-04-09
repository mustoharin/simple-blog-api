package post

import (
	"context"

	"simple-blog-api/internal/domain"
)

type GetBySlugUsecase struct {
	posts    domain.PostRepository
	postTags domain.PostTagRepository
}

func NewGetBySlugUsecase(posts domain.PostRepository, postTags domain.PostTagRepository) *GetBySlugUsecase {
	return &GetBySlugUsecase{posts: posts, postTags: postTags}
}

func (uc *GetBySlugUsecase) Execute(ctx context.Context, slug string, skipViewIncrement bool) (*domain.Post, error) {
	p, err := uc.posts.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	tags, _ := uc.postTags.GetTagsForPost(ctx, p.ID)
	p.Tags = tags

	if p.Status == domain.PostStatusPublished && !skipViewIncrement {
		_ = uc.posts.IncrementViewCount(ctx, p.ID)
		p.ViewCount++
	}

	return p, nil
}
