package post

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/pkg/sanitize"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(title string) string {
	s := strings.ToLower(title)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

type CreatePostInput struct {
	Title         string
	Slug          string
	Content       string
	Excerpt       string
	CoverImageURL string
	AuthorID      string
	AuthorEmail   string
	TagIDs        []string
}

type CreatePostUsecase struct {
	posts    domain.PostRepository
	postTags domain.PostTagRepository
	audit    domain.AuditLogger
}

func NewCreatePostUsecase(
	posts domain.PostRepository,
	postTags domain.PostTagRepository,
	audit domain.AuditLogger,
) *CreatePostUsecase {
	return &CreatePostUsecase{posts: posts, postTags: postTags, audit: audit}
}

func (uc *CreatePostUsecase) Execute(ctx context.Context, in CreatePostInput) (*domain.Post, error) {
	in.Title = sanitize.SanitizeStrict(sanitize.Trim(in.Title))
	in.Content = sanitize.SanitizeUGC(sanitize.Trim(in.Content))
	in.Excerpt = sanitize.SanitizeStrict(sanitize.Trim(in.Excerpt))
	in.CoverImageURL = sanitize.Trim(in.CoverImageURL)

	if in.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	slug := in.Slug
	if slug == "" {
		slug = slugify(in.Title)
	} else {
		slug = slugify(slug)
	}

	baseSlug := slug
	for i := 1; ; i++ {
		_, err := uc.posts.GetBySlug(ctx, slug)
		if errors.Is(err, domain.ErrNotFound) {
			break
		}
		if err != nil {
			return nil, err
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
	}

	p := &domain.Post{
		ID:            uuid.NewString(),
		Title:         in.Title,
		Slug:          slug,
		Content:       in.Content,
		Excerpt:       in.Excerpt,
		CoverImageURL: in.CoverImageURL,
		Status:        domain.PostStatusDraft,
		AuthorID:      in.AuthorID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := uc.posts.Create(ctx, p); err != nil {
		return nil, err
	}

	tagIDs := in.TagIDs
	if tagIDs == nil {
		tagIDs = []string{}
	}
	if err := uc.postTags.SetPostTags(ctx, p.ID, tagIDs); err != nil {
		return nil, err
	}

	actorID := in.AuthorID
	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &actorID,
		ActorEmail:   in.AuthorEmail,
		Action:       domain.AuditPostCreated,
		ResourceType: "post",
		ResourceID:   &p.ID,
	})

	return p, nil
}
