package post

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/sanitize"
)

type UpdatePostInput struct {
	ID            string
	Title         string
	Slug          string
	Content       string
	Excerpt       string
	CoverImageURL string
	TagIDs        []string
	ActorID       string
	ActorEmail    string
}

type UpdatePostUsecase struct {
	posts    domain.PostRepository
	postTags domain.PostTagRepository
	audit    domain.AuditLogger
}

func NewUpdatePostUsecase(posts domain.PostRepository, postTags domain.PostTagRepository, audit domain.AuditLogger) *UpdatePostUsecase {
	return &UpdatePostUsecase{posts: posts, postTags: postTags, audit: audit}
}

func (uc *UpdatePostUsecase) Execute(ctx context.Context, in UpdatePostInput) (*domain.Post, error) {
	p, err := uc.posts.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	p.Title = sanitize.SanitizeStrict(sanitize.Trim(in.Title))
	p.Content = sanitize.SanitizeUGC(sanitize.Trim(in.Content))
	p.Excerpt = sanitize.SanitizeStrict(sanitize.Trim(in.Excerpt))
	p.CoverImageURL = sanitize.Trim(in.CoverImageURL)
	if in.Slug != "" {
		newSlug := slugify(sanitize.Trim(in.Slug))
		// Only check for conflicts if the slug is actually changing
		if newSlug != p.Slug {
			baseSlug := newSlug
			for i := 1; ; i++ {
				existing, err := uc.posts.GetBySlug(ctx, newSlug)
				if errors.Is(err, domain.ErrNotFound) {
					break
				}
				if err != nil {
					return nil, err
				}
				// existing post found — if it's the same post, no conflict
				if existing.ID == p.ID {
					break
				}
				newSlug = fmt.Sprintf("%s-%d", baseSlug, i)
			}
		}
		p.Slug = newSlug
	}

	if err := uc.posts.Update(ctx, p); err != nil {
		return nil, err
	}

	if in.TagIDs == nil {
		in.TagIDs = []string{}
	}
	if err := uc.postTags.SetPostTags(ctx, p.ID, in.TagIDs); err != nil {
		return nil, err
	}

	actorID := in.ActorID
	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &actorID,
		ActorEmail:   in.ActorEmail,
		Action:       domain.AuditPostUpdated,
		ResourceType: "post",
		ResourceID:   &p.ID,
	})

	return p, nil
}
