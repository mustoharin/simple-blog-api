package comment

import (
	"context"

	"simple-blog-api/internal/domain"
)

type ListCommentsUsecase struct {
	comments domain.CommentRepository
}

func NewListCommentsUsecase(comments domain.CommentRepository) *ListCommentsUsecase {
	return &ListCommentsUsecase{comments: comments}
}

type ListCommentsOutput struct {
	Comments []*domain.Comment
	Total    int
	Page     int
	Limit    int
}

func (uc *ListCommentsUsecase) Execute(ctx context.Context, postID string, page, limit int) (ListCommentsOutput, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	comments, total, err := uc.comments.List(ctx, postID, page, limit)
	if err != nil {
		return ListCommentsOutput{}, err
	}
	return ListCommentsOutput{Comments: comments, Total: total, Page: page, Limit: limit}, nil
}
