package audit_test

import (
	"context"
	"testing"
	"time"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/audit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuditRepo struct{ mock.Mock }

func (m *mockAuditRepo) Create(ctx context.Context, l *domain.AuditLog) error {
	return m.Called(ctx, l).Error(0)
}
func (m *mockAuditRepo) List(ctx context.Context, f domain.AuditFilter) ([]*domain.AuditLog, int, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]*domain.AuditLog), args.Int(1), args.Error(2)
}
func (m *mockAuditRepo) GetByID(ctx context.Context, id string) (*domain.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditLog), args.Error(1)
}

func TestListAuditLogs_DefaultPagination(t *testing.T) {
	repo := new(mockAuditRepo)

	from := time.Now().Add(-24 * time.Hour)
	logs := []*domain.AuditLog{
		{ID: "a1", Action: "user.login", ActorEmail: "alice@example.com"},
	}
	repo.On("List", mock.Anything, mock.MatchedBy(func(f domain.AuditFilter) bool {
		return f.Page == 1 && f.Limit == 20
	})).Return(logs, 1, nil)

	uc := audit.NewListAuditLogsUsecase(repo)
	out, err := uc.Execute(context.Background(), audit.ListInput{
		From:  &from,
		Page:  0, // should default to 1
		Limit: 0, // should default to 20
	})
	assert.NoError(t, err)
	assert.Len(t, out.Logs, 1)
	assert.Equal(t, 1, out.Total)
}

func TestListAuditLogs_WithFilters(t *testing.T) {
	repo := new(mockAuditRepo)

	logs := []*domain.AuditLog{
		{ID: "a2", Action: "post.created", ActorEmail: "editor@example.com"},
	}
	repo.On("List", mock.Anything, mock.MatchedBy(func(f domain.AuditFilter) bool {
		return f.Action == "post.created" && f.Page == 2 && f.Limit == 10
	})).Return(logs, 1, nil)

	uc := audit.NewListAuditLogsUsecase(repo)
	out, err := uc.Execute(context.Background(), audit.ListInput{
		Action: "post.created",
		Page:   2,
		Limit:  10,
	})
	assert.NoError(t, err)
	assert.Len(t, out.Logs, 1)
}
