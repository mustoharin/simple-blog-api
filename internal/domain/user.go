package domain

import "time"

type UserStatus string

const (
	UserStatusPendingInvitation UserStatus = "pending_invitation"
	UserStatusActive            UserStatus = "active"
	UserStatusExpiredInvitation UserStatus = "expired_invitation"
)

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash *string    `json:"-"`
	DisplayName  string     `json:"display_name"`
	Bio          string     `json:"bio"`
	AvatarURL    string     `json:"avatar_url"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	Status       UserStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
	Roles        []Role     `json:"roles,omitempty"`
	Permissions  []string   `json:"permissions,omitempty"`
}
