package domain

import "time"

type DashboardOverview struct {
	TotalPosts    int `json:"total_posts"`
	TotalComments int `json:"total_comments"`
	TotalUsers    int `json:"total_users"`
	TotalImages   int `json:"total_images"`
}

type PostAnalyticItem struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	ViewCount    int        `json:"view_count"`
	CommentCount int        `json:"comment_count"`
	PublishedAt  *time.Time `json:"published_at"`
}

type RecentActivityItem struct {
	ID         string    `json:"id"`
	Action     string    `json:"action"`
	ActorEmail string    `json:"actor_email"`
	CreatedAt  time.Time `json:"created_at"`
}

type ActiveUserItem struct {
	ID          string     `json:"id"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

type PostAnalytics struct {
	RangeDays     int                `json:"range_days"`
	TopByViews    []PostAnalyticItem `json:"top_by_views"`
	TopByComments []PostAnalyticItem `json:"top_by_comments"`
}

type ActiveUsers struct {
	Count int              `json:"count"`
	Users []ActiveUserItem `json:"users"`
}

type DashboardResponse struct {
	Overview       DashboardOverview    `json:"overview"`
	PostAnalytics  PostAnalytics        `json:"post_analytics"`
	RecentActivity []RecentActivityItem `json:"recent_activity"`
	ActiveUsers    ActiveUsers          `json:"active_users"`
}
