package domain

import "time"

type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"
	CommentStatusApproved CommentStatus = "approved"
	CommentStatusRejected CommentStatus = "rejected"
)

type Comment struct {
	ID        string        `json:"id"`
	PostID    string        `json:"post_id"`
	AuthorID  string        `json:"author_id"`
	Body      string        `json:"body"`
	Status    CommentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	DeletedAt *time.Time    `json:"-"`
	Author    *User         `json:"author,omitempty"`
}
