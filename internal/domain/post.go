package domain

import "time"

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

type Post struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Content       string     `json:"content"`
	Excerpt       string     `json:"excerpt"`
	CoverImageURL string     `json:"cover_image_url"`
	Status        PostStatus `json:"status"`
	AuthorID      string     `json:"author_id"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"-"`
	ViewCount     int64      `json:"view_count"`
	CommentCount  int        `json:"comment_count"`
	Tags          []Tag      `json:"tags,omitempty"`
}

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type PostFilter struct {
	Query    string
	Tag      string
	AuthorID string
	Page     int
	Limit    int
	Sort     string
}
