package model

const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// Permission 权限项。
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// Role 角色定义。
type Role struct {
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}
