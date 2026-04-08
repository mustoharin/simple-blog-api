package domain

import "time"

type UserStatus string

const (
	UserStatusPendingInvitation UserStatus = "pending_invitation"
	UserStatusActive            UserStatus = "active"
	UserStatusExpiredInvitation UserStatus = "expired_invitation"
)

type User struct {
	ID           string
	Email        string
	PasswordHash *string
	DisplayName  string
	Bio          string
	AvatarURL    string
	LastLoginAt  *time.Time
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	Roles        []Role
	Permissions  []string
}
