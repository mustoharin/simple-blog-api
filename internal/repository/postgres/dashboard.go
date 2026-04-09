package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"simple-blog-api/internal/domain"
)

type DashboardRepository struct {
	db *pgxpool.Pool
}

func NewDashboardRepository(db *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetOverview(ctx context.Context) (domain.DashboardOverview, error) {
	var o domain.DashboardOverview
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM posts    WHERE deleted_at IS NULL) AS total_posts,
			(SELECT COUNT(*) FROM comments WHERE deleted_at IS NULL) AS total_comments,
			(SELECT COUNT(*) FROM users    WHERE deleted_at IS NULL) AS total_users,
			(SELECT COUNT(*) FROM images) AS total_images
	`).Scan(&o.TotalPosts, &o.TotalComments, &o.TotalUsers, &o.TotalImages)
	return o, err
}

func (r *DashboardRepository) GetTopPostsByViews(ctx context.Context, days, limit int) ([]domain.PostAnalyticItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, slug, view_count, published_at
		FROM posts
		WHERE deleted_at IS NULL AND status = 'published'
		  AND published_at >= NOW() - ($1 || ' days')::INTERVAL
		ORDER BY view_count DESC LIMIT $2`, days, limit)
	if err != nil {
		return nil, fmt.Errorf("top posts by views: %w", err)
	}
	defer rows.Close()
	var items []domain.PostAnalyticItem
	for rows.Next() {
		var item domain.PostAnalyticItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Slug, &item.ViewCount, &item.PublishedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("top posts by views scan: %w", err)
	}
	return items, nil
}

func (r *DashboardRepository) GetTopPostsByComments(ctx context.Context, days, limit int) ([]domain.PostAnalyticItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, slug, comment_count, published_at
		FROM posts
		WHERE deleted_at IS NULL AND status = 'published'
		  AND published_at >= NOW() - ($1 || ' days')::INTERVAL
		ORDER BY comment_count DESC LIMIT $2`, days, limit)
	if err != nil {
		return nil, fmt.Errorf("top posts by comments: %w", err)
	}
	defer rows.Close()
	var items []domain.PostAnalyticItem
	for rows.Next() {
		var item domain.PostAnalyticItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Slug, &item.CommentCount, &item.PublishedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("top posts by comments scan: %w", err)
	}
	return items, nil
}

func (r *DashboardRepository) GetRecentActivity(ctx context.Context, limit int) ([]domain.RecentActivityItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, action, actor_email, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("recent activity: %w", err)
	}
	defer rows.Close()
	var items []domain.RecentActivityItem
	for rows.Next() {
		var item domain.RecentActivityItem
		if err := rows.Scan(&item.ID, &item.Action, &item.ActorEmail, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recent activity scan: %w", err)
	}
	return items, nil
}

func (r *DashboardRepository) GetActiveUsers(ctx context.Context, minutesWindow int) ([]domain.ActiveUserItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, display_name, email, last_login_at
		FROM users
		WHERE last_login_at >= NOW() - ($1 || ' minutes')::INTERVAL
		  AND deleted_at IS NULL
		ORDER BY last_login_at DESC`, minutesWindow)
	if err != nil {
		return nil, fmt.Errorf("active users: %w", err)
	}
	defer rows.Close()
	var users []domain.ActiveUserItem
	for rows.Next() {
		var u domain.ActiveUserItem
		if err := rows.Scan(&u.ID, &u.DisplayName, &u.Email, &u.LastLoginAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("active users scan: %w", err)
	}
	return users, nil
}
