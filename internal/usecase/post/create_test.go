package post_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/post"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPostRepo struct{ mock.Mock }

func (m *mockPostRepo) Create(ctx context.Context, p *domain.Post) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockPostRepo) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
}
func (m *mockPostRepo) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
}
func (m *mockPostRepo) List(ctx context.Context, filter domain.PostFilter, publicOnly bool) ([]*domain.Post, int, error) {
	args := m.Called(ctx, filter, publicOnly)
	return args.Get(0).([]*domain.Post), args.Int(1), args.Error(2)
}
func (m *mockPostRepo) Update(ctx context.Context, p *domain.Post) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockPostRepo) SoftDelete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockPostRepo) IncrementViewCount(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockPostRepo) SetPublished(ctx context.Context, id string, published bool) error {
	return m.Called(ctx, id, published).Error(0)
}

type mockPostTagRepo struct{ mock.Mock }

func (m *mockPostTagRepo) SetPostTags(ctx context.Context, postID string, tagIDs []string) error {
	return m.Called(ctx, postID, tagIDs).Error(0)
}
func (m *mockPostTagRepo) GetTagsForPost(ctx context.Context, postID string) ([]domain.Tag, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).([]domain.Tag), args.Error(1)
}

type mockCommentRepo struct{ mock.Mock }

func (m *mockCommentRepo) Create(ctx context.Context, c *domain.Comment) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockCommentRepo) List(ctx context.Context, postID string, page, limit int) ([]*domain.Comment, int, error) {
	args := m.Called(ctx, postID, page, limit)
	return args.Get(0).([]*domain.Comment), args.Int(1), args.Error(2)
}
func (m *mockCommentRepo) GetByID(ctx context.Context, id string) (*domain.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Comment), args.Error(1)
}
func (m *mockCommentRepo) UpdateStatus(ctx context.Context, id string, status domain.CommentStatus) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *mockCommentRepo) SoftDelete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockCommentRepo) SoftDeleteByPostID(ctx context.Context, postID string) error {
	return m.Called(ctx, postID).Error(0)
}

type mockAuditLogger struct{ mock.Mock }

func (m *mockAuditLogger) Log(ctx context.Context, entry *domain.AuditLog) error {
	return m.Called(ctx, entry).Error(0)
}

func TestCreatePost_Success_SlugGenerated(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)
	audit := new(mockAuditLogger)

	postRepo.On("GetBySlug", mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)
	postRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *domain.Post) bool {
		return p.Title == "Hello World" && p.Slug == "hello-world" && p.AuthorID == "u1"
	})).Return(nil)
	postTagRepo.On("SetPostTags", mock.Anything, mock.Anything, []string{}).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := post.NewCreatePostUsecase(postRepo, postTagRepo, audit)
	p, err := uc.Execute(context.Background(), post.CreatePostInput{
		Title:       "Hello World",
		Content:     "Some content here",
		AuthorID:    "u1",
		AuthorEmail: "author@example.com",
		TagIDs:      []string{},
	})
	assert.NoError(t, err)
	assert.Equal(t, "hello-world", p.Slug)
}

func TestCreatePost_SlugConflict_Suffixed(t *testing.T) {
	postRepo := new(mockPostRepo)
	postTagRepo := new(mockPostTagRepo)
	audit := new(mockAuditLogger)

	existing := &domain.Post{ID: "other", Slug: "hello-world"}
	postRepo.On("GetBySlug", mock.Anything, "hello-world").Return(existing, nil)
	postRepo.On("GetBySlug", mock.Anything, mock.MatchedBy(func(s string) bool {
		return s != "hello-world"
	})).Return(nil, domain.ErrNotFound)
	postRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	postTagRepo.On("SetPostTags", mock.Anything, mock.Anything, []string{}).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := post.NewCreatePostUsecase(postRepo, postTagRepo, audit)
	p, err := uc.Execute(context.Background(), post.CreatePostInput{
		Title:    "Hello World",
		Content:  "Some content",
		AuthorID: "u1",
		TagIDs:   []string{},
	})
	assert.NoError(t, err)
	assert.NotEqual(t, "hello-world", p.Slug)
}
