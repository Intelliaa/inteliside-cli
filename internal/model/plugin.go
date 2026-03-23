package model

// Plugin represents a marketplace plugin that can be installed.
type Plugin struct {
	ID          string
	Name        string
	Description string
	Role        Role
	Deps        []string // dependency IDs
}

// Role represents the user role a plugin is designed for.
type Role string

const (
	RolePM         Role = "pm"
	RoleDesigner   Role = "designer"
	RoleDev        Role = "dev"
	RoleAutomation Role = "automation"
)
