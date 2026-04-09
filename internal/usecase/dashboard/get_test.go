package dashboard_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/dashboard"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDashboardRepo struct{ mock.Mock }

func (m *mockDashboardRepo) GetOverview(ctx context.Context) (domain.DashboardOverview, error) {
	args := m.Called(ctx)
	return args.Get(0).(domain.DashboardOverview), args.Error(1)
}
func (m *mockDashboardRepo) GetTopPostsByViews(ctx context.Context, days, limit int) ([]domain.PostAnalyticItem, error) {
	args := m.Called(ctx, days, limit)
	return args.Get(0).([]domain.PostAnalyticItem), args.Error(1)
}
func (m *mockDashboardRepo) GetTopPostsByComments(ctx context.Context, days, limit int) ([]domain.PostAnalyticItem, error) {
	args := m.Called(ctx, days, limit)
	return args.Get(0).([]domain.PostAnalyticItem), args.Error(1)
}
func (m *mockDashboardRepo) GetRecentActivity(ctx context.Context, limit int) ([]domain.RecentActivityItem, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]domain.RecentActivityItem), args.Error(1)
}
func (m *mockDashboardRepo) GetActiveUsers(ctx context.Context, minutesWindow int) ([]domain.ActiveUserItem, error) {
	args := m.Called(ctx, minutesWindow)
	return args.Get(0).([]domain.ActiveUserItem), args.Error(1)
}

func TestGetDashboard_InvalidRange(t *testing.T) {
	repo := new(mockDashboardRepo)
	uc := dashboard.NewGetDashboardUsecase(repo)
	_, err := uc.Execute(context.Background(), 60) // 60 is not in [7,30,90]
	assert.ErrorIs(t, err, domain.ErrInvalidRange)
}

func TestGetDashboard_ValidRange(t *testing.T) {
	repo := new(mockDashboardRepo)

	overview := domain.DashboardOverview{TotalPosts: 10, TotalUsers: 5}
	repo.On("GetOverview", mock.Anything).Return(overview, nil)
	repo.On("GetTopPostsByViews", mock.Anything, 30, 5).Return([]domain.PostAnalyticItem{}, nil)
	repo.On("GetTopPostsByComments", mock.Anything, 30, 5).Return([]domain.PostAnalyticItem{}, nil)
	repo.On("GetRecentActivity", mock.Anything, 20).Return([]domain.RecentActivityItem{}, nil)
	repo.On("GetActiveUsers", mock.Anything, 15).Return([]domain.ActiveUserItem{}, nil)

	uc := dashboard.NewGetDashboardUsecase(repo)
	out, err := uc.Execute(context.Background(), 30)
	assert.NoError(t, err)
	assert.Equal(t, 30, out.PostAnalytics.RangeDays)
	assert.Equal(t, 10, out.Overview.TotalPosts)
}
