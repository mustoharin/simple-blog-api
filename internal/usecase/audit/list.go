package audit

import (
	"context"
	"time"

	"simple-blog-api/internal/domain"
)

type AuditRepository interface {
	Create(ctx context.Context, l *domain.AuditLog) error
	List(ctx context.Context, f domain.AuditFilter) ([]*domain.AuditLog, int, error)
	GetByID(ctx context.Context, id string) (*domain.AuditLog, error)
}

type ListInput struct {
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	From         *time.Time
	To           *time.Time
	Page         int
	Limit        int
}

type ListOutput struct {
	Logs  []*domain.AuditLog
	Total int
	Page  int
	Limit int
}

type ListAuditLogsUsecase struct {
	repo AuditRepository
}

func NewListAuditLogsUsecase(repo AuditRepository) *ListAuditLogsUsecase {
	return &ListAuditLogsUsecase{repo: repo}
}

func (uc *ListAuditLogsUsecase) Execute(ctx context.Context, in ListInput) (ListOutput, error) {
	if in.Page < 1 {
		in.Page = 1
	}
	if in.Limit < 1 || in.Limit > 100 {
		in.Limit = 20
	}

	filter := domain.AuditFilter{
		ActorID:      in.ActorID,
		Action:       in.Action,
		ResourceType: in.ResourceType,
		ResourceID:   in.ResourceID,
		From:         in.From,
		To:           in.To,
		Page:         in.Page,
		Limit:        in.Limit,
	}

	logs, total, err := uc.repo.List(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}

	return ListOutput{
		Logs:  logs,
		Total: total,
		Page:  in.Page,
		Limit: in.Limit,
	}, nil
}
