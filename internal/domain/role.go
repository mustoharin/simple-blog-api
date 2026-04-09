package domain

type Role struct {
	ID          string
	Name        string
	Description string
	Permissions []Permission
}

type Permission struct {
	ID   string
	Name string
}

const (
	PermPostCreate     = "post:create"
	PermPostEdit       = "post:edit"
	PermPostDelete     = "post:delete"
	PermPostPublish    = "post:publish"
	PermImageUpload    = "image:upload"
	PermCommentCreate  = "comment:create"
	PermCommentApprove = "comment:approve"
	PermUserManage     = "user:manage"
	PermUserCreate     = "user:create"
	PermUserUpdate     = "user:update"
	PermUserDelete     = "user:delete"
	PermAuditRead      = "audit:read"
	PermDashboardRead  = "dashboard:read"
)
