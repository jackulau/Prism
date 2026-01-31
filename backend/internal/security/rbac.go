package security

// ResourceType represents a type of resource that can be access-controlled
type ResourceType string

const (
	ResourceAgent        ResourceType = "agent"
	ResourceWorkflow     ResourceType = "workflow"
	ResourceConversation ResourceType = "conversation"
	ResourceWorkspace    ResourceType = "workspace"
	ResourceTeam         ResourceType = "team"
	ResourceOrganization ResourceType = "organization"
	ResourceTool         ResourceType = "tool"
	ResourceIntegration  ResourceType = "integration"
)

// Action represents an action that can be performed on a resource
type Action string

const (
	ActionRead   Action = "read"
	ActionWrite  Action = "write"
	ActionDelete Action = "delete"
	ActionAdmin  Action = "admin"
	ActionCreate Action = "create"
	ActionInvite Action = "invite"
	ActionRemove Action = "remove"
)

// Scope represents the scope of a permission
type Scope string

const (
	ScopeOwn          Scope = "own"          // Only resources the user owns
	ScopeTeam         Scope = "team"         // Resources within the user's team(s)
	ScopeOrganization Scope = "organization" // Resources within the user's organization(s)
	ScopeGlobal       Scope = "global"       // All resources (super admin)
)

// Permission represents a single permission grant
type Permission struct {
	Resource ResourceType `json:"resource"`
	Action   Action       `json:"action"`
	Scope    Scope        `json:"scope"`
}

// RoleType represents a predefined or custom role
type RoleType string

const (
	RoleOrgOwner   RoleType = "org_owner"   // Full organization control
	RoleOrgAdmin   RoleType = "org_admin"   // Organization administration
	RoleTeamAdmin  RoleType = "team_admin"  // Team administration
	RoleTeamMember RoleType = "team_member" // Standard team member
	RoleViewer     RoleType = "viewer"      // Read-only access
	RoleCustom     RoleType = "custom"      // Custom role with specific permissions
)

// Role represents a role with a set of permissions
type Role struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	Type           RoleType     `json:"type"`
	OrganizationID string       `json:"organization_id,omitempty"`
	Permissions    []Permission `json:"permissions"`
	IsSystem       bool         `json:"is_system"` // System roles cannot be modified
}

// PredefinedRoles returns the default system roles
func PredefinedRoles() map[RoleType][]Permission {
	return map[RoleType][]Permission{
		RoleOrgOwner: {
			// Full organization control
			{Resource: ResourceOrganization, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceTeam, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceAgent, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceWorkflow, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceConversation, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceWorkspace, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceTool, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceIntegration, Action: ActionAdmin, Scope: ScopeOrganization},
		},
		RoleOrgAdmin: {
			// Organization administration (cannot delete org)
			{Resource: ResourceOrganization, Action: ActionRead, Scope: ScopeOrganization},
			{Resource: ResourceOrganization, Action: ActionWrite, Scope: ScopeOrganization},
			{Resource: ResourceTeam, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceAgent, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceWorkflow, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceConversation, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceWorkspace, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceTool, Action: ActionAdmin, Scope: ScopeOrganization},
			{Resource: ResourceIntegration, Action: ActionAdmin, Scope: ScopeOrganization},
		},
		RoleTeamAdmin: {
			// Team administration
			{Resource: ResourceTeam, Action: ActionRead, Scope: ScopeTeam},
			{Resource: ResourceTeam, Action: ActionWrite, Scope: ScopeTeam},
			{Resource: ResourceTeam, Action: ActionInvite, Scope: ScopeTeam},
			{Resource: ResourceTeam, Action: ActionRemove, Scope: ScopeTeam},
			{Resource: ResourceAgent, Action: ActionAdmin, Scope: ScopeTeam},
			{Resource: ResourceWorkflow, Action: ActionAdmin, Scope: ScopeTeam},
			{Resource: ResourceConversation, Action: ActionAdmin, Scope: ScopeTeam},
			{Resource: ResourceWorkspace, Action: ActionAdmin, Scope: ScopeTeam},
			{Resource: ResourceTool, Action: ActionRead, Scope: ScopeTeam},
			{Resource: ResourceTool, Action: ActionWrite, Scope: ScopeTeam},
		},
		RoleTeamMember: {
			// Standard team member
			{Resource: ResourceTeam, Action: ActionRead, Scope: ScopeTeam},
			{Resource: ResourceAgent, Action: ActionRead, Scope: ScopeTeam},
			{Resource: ResourceAgent, Action: ActionWrite, Scope: ScopeTeam},
			{Resource: ResourceAgent, Action: ActionCreate, Scope: ScopeTeam},
			{Resource: ResourceWorkflow, Action: ActionRead, Scope: ScopeTeam},
			{Resource: ResourceWorkflow, Action: ActionWrite, Scope: ScopeTeam},
			{Resource: ResourceWorkflow, Action: ActionCreate, Scope: ScopeTeam},
			{Resource: ResourceConversation, Action: ActionRead, Scope: ScopeOwn},
			{Resource: ResourceConversation, Action: ActionWrite, Scope: ScopeOwn},
			{Resource: ResourceConversation, Action: ActionCreate, Scope: ScopeTeam},
			{Resource: ResourceWorkspace, Action: ActionRead, Scope: ScopeTeam},
			{Resource: ResourceWorkspace, Action: ActionWrite, Scope: ScopeTeam},
			{Resource: ResourceTool, Action: ActionRead, Scope: ScopeTeam},
		},
		RoleViewer: {
			// Read-only access
			{Resource: ResourceTeam, Action: ActionRead, Scope: ScopeTeam},
			{Resource: ResourceAgent, Action: ActionRead, Scope: ScopeTeam},
			{Resource: ResourceWorkflow, Action: ActionRead, Scope: ScopeTeam},
			{Resource: ResourceConversation, Action: ActionRead, Scope: ScopeOwn},
			{Resource: ResourceWorkspace, Action: ActionRead, Scope: ScopeTeam},
			{Resource: ResourceTool, Action: ActionRead, Scope: ScopeTeam},
		},
	}
}

// HasPermission checks if a set of permissions includes a specific permission
func HasPermission(permissions []Permission, resource ResourceType, action Action) bool {
	for _, p := range permissions {
		if p.Resource == resource && (p.Action == action || p.Action == ActionAdmin) {
			return true
		}
	}
	return false
}

// HasPermissionWithScope checks if permissions include a specific permission with at least the given scope
func HasPermissionWithScope(permissions []Permission, resource ResourceType, action Action, minScope Scope) bool {
	scopePrecedence := map[Scope]int{
		ScopeOwn:          1,
		ScopeTeam:         2,
		ScopeOrganization: 3,
		ScopeGlobal:       4,
	}

	minScopeLevel := scopePrecedence[minScope]

	for _, p := range permissions {
		if p.Resource == resource && (p.Action == action || p.Action == ActionAdmin) {
			if scopePrecedence[p.Scope] >= minScopeLevel {
				return true
			}
		}
	}
	return false
}

// MergePermissions merges multiple permission sets, keeping the highest scope for each resource+action
func MergePermissions(permissionSets ...[]Permission) []Permission {
	scopePrecedence := map[Scope]int{
		ScopeOwn:          1,
		ScopeTeam:         2,
		ScopeOrganization: 3,
		ScopeGlobal:       4,
	}

	// Use a map to track the highest scope for each resource+action combination
	permMap := make(map[string]Permission)

	for _, perms := range permissionSets {
		for _, p := range perms {
			key := string(p.Resource) + ":" + string(p.Action)
			existing, exists := permMap[key]
			if !exists || scopePrecedence[p.Scope] > scopePrecedence[existing.Scope] {
				permMap[key] = p
			}
		}
	}

	// Convert map back to slice
	result := make([]Permission, 0, len(permMap))
	for _, p := range permMap {
		result = append(result, p)
	}

	return result
}
