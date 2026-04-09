package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, page, limit int) ([]*User, int, error)
	Update(ctx context.Context, user *User) error
	SoftDelete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status UserStatus) error
	UpdateLastLogin(ctx context.Context, id string) error
	UpdatePasswordHash(ctx context.Context, id, passwordHash string) error
	AssignRole(ctx context.Context, userID, roleID string) error
	RemoveRole(ctx context.Context, userID, roleID string) error
	GetRoles(ctx context.Context, userID string) ([]Role, error)
	GetPermissions(ctx context.Context, userID string) ([]string, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	GetByHash(ctx context.Context, hash string) (*PasswordResetToken, error)
	MarkUsed(ctx context.Context, id string) error
}

type InvitationTokenRepository interface {
	Create(ctx context.Context, token *InvitationToken) error
	GetByUserID(ctx context.Context, userID string) (*InvitationToken, error)
	GetByHash(ctx context.Context, hash string) (*InvitationToken, error)
	MarkUsed(ctx context.Context, id string) error
	InvalidatePrevious(ctx context.Context, userID string) error
}

type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	GetByID(ctx context.Context, id string) (*Post, error)
	GetBySlug(ctx context.Context, slug string) (*Post, error)
	List(ctx context.Context, filter PostFilter, publicOnly bool) ([]*Post, int, error)
	Update(ctx context.Context, post *Post) error
	SoftDelete(ctx context.Context, id string) error
	IncrementViewCount(ctx context.Context, id string) error
	SetPublished(ctx context.Context, id string, published bool) error
}

type TagRepository interface {
	Create(ctx context.Context, tag *Tag) error
	List(ctx context.Context) ([]*Tag, error)
	GetByID(ctx context.Context, id string) (*Tag, error)
	HardDelete(ctx context.Context, id string) error
	GetOrCreateByName(ctx context.Context, name string) (*Tag, error)
}

type PostTagRepository interface {
	SetPostTags(ctx context.Context, postID string, tagIDs []string) error
	GetTagsForPost(ctx context.Context, postID string) ([]Tag, error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment *Comment) error
	List(ctx context.Context, postID string, page, limit int) ([]*Comment, int, error)
	GetByID(ctx context.Context, id string) (*Comment, error)
	UpdateStatus(ctx context.Context, id string, status CommentStatus) error
	SoftDelete(ctx context.Context, id string) error
	SoftDeleteByPostID(ctx context.Context, postID string) error
}

type ImageRepository interface {
	Create(ctx context.Context, image *Image) error
	GetByID(ctx context.Context, id string) (*Image, error)
	HardDelete(ctx context.Context, id string) error
	List(ctx context.Context, page, limit int) ([]*Image, int, error)
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, filter AuditFilter) ([]*AuditLog, int, error)
	GetByID(ctx context.Context, id string) (*AuditLog, error)
	DeleteOlderThan(ctx context.Context, days int) error
}

