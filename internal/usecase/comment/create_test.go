package comment_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/comment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPostRepo struct{ mock.Mock }

func (m *mockPostRepo) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
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

func TestCreateComment_Success(t *testing.T) {
	postRepo := new(mockPostRepo)
	commentRepo := new(mockCommentRepo)
	audit := new(mockAuditLogger)

	post := &domain.Post{ID: "p1", Title: "Test Post", Status: domain.PostStatusPublished}
	postRepo.On("GetByID", mock.Anything, "p1").Return(post, nil)
	commentRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Comment) bool {
		return c.PostID == "p1" &&
			c.AuthorID == "u1" &&
			c.Status == domain.CommentStatusPending &&
			c.Body == "Great post!"
	})).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := comment.NewCreateCommentUsecase(postRepo, commentRepo, audit)
	c, err := uc.Execute(context.Background(), comment.CreateCommentInput{
		PostID:      "p1",
		AuthorID:    "u1",
		AuthorEmail: "alice@example.com",
		Body:        "Great post!",
	})
	assert.NoError(t, err)
	assert.Equal(t, domain.CommentStatusPending, c.Status)
	assert.Equal(t, "Great post!", c.Body)
}

func TestCreateComment_PostNotFound(t *testing.T) {
	postRepo := new(mockPostRepo)
	commentRepo := new(mockCommentRepo)
	audit := new(mockAuditLogger)

	postRepo.On("GetByID", mock.Anything, "ghost").Return(nil, domain.ErrNotFound)

	uc := comment.NewCreateCommentUsecase(postRepo, commentRepo, audit)
	_, err := uc.Execute(context.Background(), comment.CreateCommentInput{
		PostID: "ghost",
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUpdateCommentStatus_Approve(t *testing.T) {
	commentRepo := new(mockCommentRepo)
	audit := new(mockAuditLogger)

	c := &domain.Comment{ID: "c1", PostID: "p1", AuthorID: "u2", Status: domain.CommentStatusPending}
	commentRepo.On("GetByID", mock.Anything, "c1").Return(c, nil)
	commentRepo.On("UpdateStatus", mock.Anything, "c1", domain.CommentStatusApproved).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := comment.NewUpdateStatusUsecase(commentRepo, audit)
	err := uc.Execute(context.Background(), "c1", "approved", "mod1", "mod@example.com")
	assert.NoError(t, err)
}
