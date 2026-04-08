package audit

import (
	"context"

	"simple-blog-api/internal/domain"
)

type GetAuditLogUsecase struct {
	repo AuditRepository
}

func NewGetAuditLogUsecase(repo AuditRepository) *GetAuditLogUsecase {
	return &GetAuditLogUsecase{repo: repo}
}

func (uc *GetAuditLogUsecase) Execute(ctx context.Context, id string) (*domain.AuditLog, error) {
	return uc.repo.GetByID(ctx, id)
}
