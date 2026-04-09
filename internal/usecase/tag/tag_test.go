package tag_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/tag"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type mockTagRepo struct{ mock.Mock }

func (m *mockTagRepo) Create(ctx context.Context, t *domain.Tag) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockTagRepo) List(ctx context.Context) ([]*domain.Tag, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Tag), args.Error(1)
}
func (m *mockTagRepo) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tag), args.Error(1)
}
func (m *mockTagRepo) HardDelete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockTagRepo) GetOrCreateByName(ctx context.Context, name string) (*domain.Tag, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tag), args.Error(1)
}

type mockAuditLogger struct{ mock.Mock }

func (m *mockAuditLogger) Log(ctx context.Context, entry *domain.AuditLog) error {
	return m.Called(ctx, entry).Error(0)
}

// --- CreateTag tests ---

func TestCreateTag_Success(t *testing.T) {
	repo := &mockTagRepo{}
	audit := &mockAuditLogger{}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(tg *domain.Tag) bool {
		return tg.Name == "Hello World" && tg.Slug == "hello-world" && tg.ID != ""
	})).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := tag.NewCreateTagUsecase(repo, audit)
	result, err := uc.Execute(context.Background(), "Hello World", "actor-1", "actor@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Hello World", result.Name)
	assert.Equal(t, "hello-world", result.Slug)
	assert.NotEmpty(t, result.ID)

	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestCreateTag_SlugGeneration(t *testing.T) {
	cases := []struct {
		input    string
		wantSlug string
	}{
		{"Hello World", "hello-world"},
		{"Go Testing", "go-testing"},
		{"  spaces  ", "spaces"},
		{"Multiple   Spaces", "multiple-spaces"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			repo := &mockTagRepo{}
			audit := &mockAuditLogger{}

			repo.On("Create", mock.Anything, mock.MatchedBy(func(tg *domain.Tag) bool {
				return tg.Slug == tc.wantSlug
			})).Return(nil)
			audit.On("Log", mock.Anything, mock.Anything).Return(nil)

			uc := tag.NewCreateTagUsecase(repo, audit)
			result, err := uc.Execute(context.Background(), tc.input, "actor-1", "actor@example.com")

			assert.NoError(t, err)
			assert.Equal(t, tc.wantSlug, result.Slug)

			time.Sleep(50 * time.Millisecond)
			repo.AssertExpectations(t)
			audit.AssertExpectations(t)
		})
	}
}

func TestCreateTag_RepoError(t *testing.T) {
	repo := &mockTagRepo{}
	audit := &mockAuditLogger{}

	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	uc := tag.NewCreateTagUsecase(repo, audit)
	result, err := uc.Execute(context.Background(), "Hello World", "actor-1", "actor@example.com")

	assert.Error(t, err)
	assert.Nil(t, result)
	audit.AssertNotCalled(t, "Log")
	repo.AssertExpectations(t)
}

// --- DeleteTag tests ---

func TestDeleteTag_Success(t *testing.T) {
	repo := &mockTagRepo{}
	audit := &mockAuditLogger{}

	repo.On("HardDelete", mock.Anything, "tag-123").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := tag.NewDeleteTagUsecase(repo, audit)
	err := uc.Execute(context.Background(), "tag-123", "actor-1", "actor@example.com")

	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestDeleteTag_RepoError(t *testing.T) {
	repo := &mockTagRepo{}
	audit := &mockAuditLogger{}

	repo.On("HardDelete", mock.Anything, "tag-999").Return(errors.New("not found"))

	uc := tag.NewDeleteTagUsecase(repo, audit)
	err := uc.Execute(context.Background(), "tag-999", "actor-1", "actor@example.com")

	assert.Error(t, err)
	audit.AssertNotCalled(t, "Log")
	repo.AssertExpectations(t)
}

// --- ListTags tests ---

func TestListTags_Success(t *testing.T) {
	repo := &mockTagRepo{}

	expected := []*domain.Tag{
		{ID: "1", Name: "Go", Slug: "go"},
		{ID: "2", Name: "Testing", Slug: "testing"},
	}
	repo.On("List", mock.Anything).Return(expected, nil)

	uc := tag.NewListTagsUsecase(repo)
	result, err := uc.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestListTags_Empty(t *testing.T) {
	repo := &mockTagRepo{}

	repo.On("List", mock.Anything).Return([]*domain.Tag{}, nil)

	uc := tag.NewListTagsUsecase(repo)
	result, err := uc.Execute(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

func TestListTags_RepoError(t *testing.T) {
	repo := &mockTagRepo{}

	repo.On("List", mock.Anything).Return([]*domain.Tag{}, errors.New("db error"))

	uc := tag.NewListTagsUsecase(repo)
	result, err := uc.Execute(context.Background())

	assert.Error(t, err)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}
