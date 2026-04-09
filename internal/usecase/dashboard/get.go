package dashboard

import (
	"context"

	"simple-blog-api/internal/domain"
)

type DashboardRepository interface {
	GetOverview(ctx context.Context) (domain.DashboardOverview, error)
	GetTopPostsByViews(ctx context.Context, days, limit int) ([]domain.PostAnalyticItem, error)
	GetTopPostsByComments(ctx context.Context, days, limit int) ([]domain.PostAnalyticItem, error)
	GetRecentActivity(ctx context.Context, limit int) ([]domain.RecentActivityItem, error)
	GetActiveUsers(ctx context.Context, minutesWindow int) ([]domain.ActiveUserItem, error)
}

var validRanges = map[int]bool{7: true, 30: true, 90: true}

type GetDashboardUsecase struct {
	repo DashboardRepository
}

func NewGetDashboardUsecase(repo DashboardRepository) *GetDashboardUsecase {
	return &GetDashboardUsecase{repo: repo}
}

func (uc *GetDashboardUsecase) Execute(ctx context.Context, rangeDays int) (domain.DashboardResponse, error) {
	if !validRanges[rangeDays] {
		return domain.DashboardResponse{}, domain.ErrInvalidRange
	}

	overview, err := uc.repo.GetOverview(ctx)
	if err != nil {
		return domain.DashboardResponse{}, err
	}

	topByViews, err := uc.repo.GetTopPostsByViews(ctx, rangeDays, 5)
	if err != nil {
		return domain.DashboardResponse{}, err
	}

	topByComments, err := uc.repo.GetTopPostsByComments(ctx, rangeDays, 5)
	if err != nil {
		return domain.DashboardResponse{}, err
	}

	recentActivity, err := uc.repo.GetRecentActivity(ctx, 20)
	if err != nil {
		return domain.DashboardResponse{}, err
	}

	activeUsers, err := uc.repo.GetActiveUsers(ctx, 15)
	if err != nil {
		return domain.DashboardResponse{}, err
	}

	return domain.DashboardResponse{
		Overview: overview,
		PostAnalytics: domain.PostAnalytics{
			RangeDays:     rangeDays,
			TopByViews:    topByViews,
			TopByComments: topByComments,
		},
		RecentActivity: recentActivity,
		ActiveUsers: domain.ActiveUsers{
			Count: len(activeUsers),
			Users: activeUsers,
		},
	}, nil
}
