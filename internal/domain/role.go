package domain

type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions,omitempty"`
}

type Permission struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
