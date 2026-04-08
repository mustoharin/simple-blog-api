package post_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/post"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTogglePublish_Publish(t *testing.T) {
	postRepo := new(mockPostRepo)
	audit := new(mockAuditLogger)

	existing := &domain.Post{ID: "p1"}
	postRepo.On("GetByID", mock.Anything, "p1").Return(existing, nil)
	postRepo.On("SetPublished", mock.Anything, "p1", true).Return(nil)
	audit.On("Log", mock.Anything, mock.MatchedBy(func(e *domain.AuditLog) bool {
		return e.Action == domain.AuditPostPublished
	})).Return(nil)

	uc := post.NewTogglePublishUsecase(postRepo, audit)
	err := uc.Execute(context.Background(), post.TogglePublishInput{
		PostID:     "p1",
		Published:  true,
		ActorID:    "u1",
		ActorEmail: "actor@example.com",
	})

	assert.NoError(t, err)
	mock.AssertExpectations(t, postRepo, audit)
}

func TestTogglePublish_Unpublish(t *testing.T) {
	postRepo := new(mockPostRepo)
	audit := new(mockAuditLogger)

	existing := &domain.Post{ID: "p1"}
	postRepo.On("GetByID", mock.Anything, "p1").Return(existing, nil)
	postRepo.On("SetPublished", mock.Anything, "p1", false).Return(nil)
	audit.On("Log", mock.Anything, mock.MatchedBy(func(e *domain.AuditLog) bool {
		return e.Action == domain.AuditPostUnpublished
	})).Return(nil)

	uc := post.NewTogglePublishUsecase(postRepo, audit)
	err := uc.Execute(context.Background(), post.TogglePublishInput{
		PostID:     "p1",
		Published:  false,
		ActorID:    "u1",
		ActorEmail: "actor@example.com",
	})

	assert.NoError(t, err)
	mock.AssertExpectations(t, postRepo, audit)
}

func TestTogglePublish_NotFound(t *testing.T) {
	postRepo := new(mockPostRepo)
	audit := new(mockAuditLogger)

	postRepo.On("GetByID", mock.Anything, "missing").Return(nil, domain.ErrNotFound)

	uc := post.NewTogglePublishUsecase(postRepo, audit)
	err := uc.Execute(context.Background(), post.TogglePublishInput{
		PostID:  "missing",
		ActorID: "u1",
	})

	assert.ErrorIs(t, err, domain.ErrNotFound)
	mock.AssertExpectations(t, postRepo, audit)
}
