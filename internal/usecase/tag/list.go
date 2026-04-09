package tag

import (
	"context"

	"simple-blog-api/internal/domain"
)

type ListTagsUsecase struct {
	tags domain.TagRepository
}

func NewListTagsUsecase(tags domain.TagRepository) *ListTagsUsecase {
	return &ListTagsUsecase{tags: tags}
}

func (uc *ListTagsUsecase) Execute(ctx context.Context) ([]*domain.Tag, error) {
	return uc.tags.List(ctx)
}
