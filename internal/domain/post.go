package domain

import "time"

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

type Post struct {
	ID            string
	Title         string
	Slug          string
	Content       string
	Excerpt       string
	CoverImageURL string
	Status        PostStatus
	AuthorID      string
	PublishedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	ViewCount     int64
	CommentCount  int
	Tags          []Tag
}

type Tag struct {
	ID   string
	Name string
	Slug string
}

type PostFilter struct {
	Query    string
	Tag      string
	AuthorID string
	Page     int
	Limit    int
	Sort     string
}
