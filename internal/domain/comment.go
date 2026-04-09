package domain

import "time"

type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"
	CommentStatusApproved CommentStatus = "approved"
	CommentStatusRejected CommentStatus = "rejected"
)

type Comment struct {
	ID        string
	PostID    string
	AuthorID  string
	Body      string
	Status    CommentStatus
	CreatedAt time.Time
	DeletedAt *time.Time
	Author    *User
}
