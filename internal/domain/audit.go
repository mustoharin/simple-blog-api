package domain

import (
	"context"
	"time"
)

type AuditLog struct {
	ID           string
	ActorID      *string
	ActorEmail   string
	Action       string
	ResourceType string
	ResourceID   *string
	IPAddress    string
	UserAgent    string
	CreatedAt    time.Time
}

type AuditFilter struct {
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	From         *time.Time
	To           *time.Time
	Page         int
	Limit        int
}

const (
	AuditUserRegistered              = "user.registered"
	AuditUserInvited                 = "user.invited"
	AuditUserInvitationAccepted      = "user.invitation_accepted"
	AuditUserInvitationResent        = "user.invitation_resent"
	AuditUserInvitationExpired       = "user.invitation_expired"
	AuditUserLogin                   = "user.login"
	AuditUserLoginFailed             = "user.login_failed"
	AuditUserLogout                  = "user.logout"
	AuditUserPasswordResetRequested  = "user.password_reset_requested"
	AuditUserPasswordResetCompleted  = "user.password_reset_completed"
	AuditUserPasswordChanged         = "user.password_changed"
	AuditUserCreated                 = "user.created"
	AuditUserUpdated                 = "user.updated"
	AuditUserDeleted                 = "user.deleted"
	AuditUserRoleAssigned            = "user.role_assigned"
	AuditUserRoleRemoved             = "user.role_removed"
	AuditPostCreated                 = "post.created"
	AuditPostUpdated                 = "post.updated"
	AuditPostPublished               = "post.published"
	AuditPostUnpublished             = "post.unpublished"
	AuditPostDeleted                 = "post.deleted"
	AuditCommentCreated              = "comment.created"
	AuditCommentApproved             = "comment.approved"
	AuditCommentRejected             = "comment.rejected"
	AuditCommentDeleted              = "comment.deleted"
	AuditTagCreated                  = "tag.created"
	AuditTagDeleted                  = "tag.deleted"
	AuditImageUploaded               = "image.uploaded"
	AuditImageDeleted                = "image.deleted"
)

type AuditLogger interface {
	Log(ctx context.Context, entry *AuditLog) error
}
