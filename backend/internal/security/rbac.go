package security

// Role represents a user's role in the system
type Role string

const (
	// RoleUser is the default role for regular users
	RoleUser Role = "user"
	// RoleAdmin is the role for administrators with elevated privileges
	RoleAdmin Role = "admin"
)

// ValidRoles contains all valid role values
var ValidRoles = []Role{RoleUser, RoleAdmin}

// IsValidRole checks if a role string is a valid role
func IsValidRole(role string) bool {
	for _, r := range ValidRoles {
		if string(r) == role {
			return true
		}
	}
	return false
}

// Permission represents a specific action that can be performed
type Permission string

const (
	// User permissions - available to all authenticated users
	PermViewConversations   Permission = "conversations:view"
	PermCreateConversations Permission = "conversations:create"
	PermManageOwnData       Permission = "own_data:manage"

	// Admin permissions - available only to administrators
	PermManageUsers          Permission = "users:manage"
	PermManageOrganization   Permission = "organization:manage"
	PermViewAllConversations Permission = "conversations:view_all"
	PermManageProviders      Permission = "providers:manage"
	PermManageSettings       Permission = "settings:manage"
)

// ResourceType represents a type of resource in the system
type ResourceType string

const (
	ResourceConversation ResourceType = "conversation"
	ResourceAgent        ResourceType = "agent"
	ResourceWorkflow     ResourceType = "workflow"
	ResourceTool         ResourceType = "tool"
	ResourceUser         ResourceType = "user"
	ResourceOrganization ResourceType = "organization"
	ResourceTeam         ResourceType = "team"
	ResourceSettings     ResourceType = "settings"
	ResourceProvider     ResourceType = "provider"
)

// Action represents an action that can be performed on a resource
type Action string

const (
	ActionView   Action = "view"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionExecute Action = "execute"
)

// RolePermissions maps roles to their allowed permissions
var RolePermissions = map[Role][]Permission{
	RoleUser: {
		PermViewConversations,
		PermCreateConversations,
		PermManageOwnData,
	},
	RoleAdmin: {
		// Admins have all user permissions
		PermViewConversations,
		PermCreateConversations,
		PermManageOwnData,
		// Plus admin-specific permissions
		PermManageUsers,
		PermManageOrganization,
		PermViewAllConversations,
		PermManageProviders,
		PermManageSettings,
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role Role, perm Permission) bool {
	permissions, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// GetPermissions returns all permissions for a role
func GetPermissions(role Role) []Permission {
	permissions, ok := RolePermissions[role]
	if !ok {
		return []Permission{}
	}
	// Return a copy to prevent modification
	result := make([]Permission, len(permissions))
	copy(result, permissions)
	return result
}

// GetPermissionStrings returns all permissions for a role as strings
func GetPermissionStrings(role Role) []string {
	permissions := GetPermissions(role)
	result := make([]string, len(permissions))
	for i, p := range permissions {
		result[i] = string(p)
	}
	return result
}
