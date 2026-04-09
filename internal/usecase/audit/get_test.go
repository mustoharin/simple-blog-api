package audit_test

import (
	"context"
	"errors"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/audit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAuditLog_Success(t *testing.T) {
	repo := new(mockAuditRepo)
	repo.On("GetByID", mock.Anything, "log1").Return(&domain.AuditLog{ID: "log1", Action: "user.login"}, nil)

	uc := audit.NewGetAuditLogUsecase(repo)
	log, err := uc.Execute(context.Background(), "log1")

	assert.NoError(t, err)
	assert.Equal(t, "log1", log.ID)
}

func TestGetAuditLog_NotFound(t *testing.T) {
	repo := new(mockAuditRepo)
	repo.On("GetByID", mock.Anything, "ghost").Return(nil, domain.ErrNotFound)

	uc := audit.NewGetAuditLogUsecase(repo)
	_, err := uc.Execute(context.Background(), "ghost")

	assert.True(t, errors.Is(err, domain.ErrNotFound))
}
