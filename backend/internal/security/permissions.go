package security

import (
	"sync"
	"time"
)

// PermissionChecker provides an interface for checking permissions
type PermissionChecker interface {
	CheckPermission(userID string, resource ResourceType, action Action, resourceID string) (bool, error)
	CheckTeamPermission(userID, teamID string, resource ResourceType, action Action) (bool, error)
	CheckOrgPermission(userID, orgID string, resource ResourceType, action Action) (bool, error)
	GetUserPermissions(userID string) ([]Permission, error)
	InvalidateCache(userID string)
}

// PermissionLoader interface for loading permissions from storage
type PermissionLoader interface {
	GetEffectivePermissions(userID string) ([]Permission, error)
	GetPermissionsForTeam(userID, teamID string) ([]Permission, error)
	GetPermissionsForOrganization(userID, orgID string) ([]Permission, error)
}

// ResourceOwnerChecker interface for checking resource ownership
type ResourceOwnerChecker interface {
	GetResourceOwner(resource ResourceType, resourceID string) (ownerUserID string, teamID string, orgID string, err error)
	IsResourceInTeam(resource ResourceType, resourceID, teamID string) (bool, error)
	IsResourceInOrg(resource ResourceType, resourceID, orgID string) (bool, error)
}

// TeamMembershipChecker interface for checking team membership
type TeamMembershipChecker interface {
	IsTeamMember(userID, teamID string) (bool, error)
	IsOrgMember(userID, orgID string) (bool, error)
	GetUserTeamIDs(userID string) ([]string, error)
	GetUserOrgIDs(userID string) ([]string, error)
}

// cachedPermissions holds cached permissions with expiry
type cachedPermissions struct {
	permissions []Permission
	expiresAt   time.Time
}

// PermissionService provides permission checking with caching
type PermissionService struct {
	loader           PermissionLoader
	resourceChecker  ResourceOwnerChecker
	membershipChecker TeamMembershipChecker

	cache     map[string]*cachedPermissions
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
}

// NewPermissionService creates a new permission service
func NewPermissionService(
	loader PermissionLoader,
	resourceChecker ResourceOwnerChecker,
	membershipChecker TeamMembershipChecker,
) *PermissionService {
	return &PermissionService{
		loader:            loader,
		resourceChecker:   resourceChecker,
		membershipChecker: membershipChecker,
		cache:             make(map[string]*cachedPermissions),
		cacheTTL:          5 * time.Minute,
	}
}

// SetCacheTTL sets the cache time-to-live
func (s *PermissionService) SetCacheTTL(ttl time.Duration) {
	s.cacheTTL = ttl
}

// GetUserPermissions returns all effective permissions for a user
func (s *PermissionService) GetUserPermissions(userID string) ([]Permission, error) {
	// Check cache first
	s.cacheMu.RLock()
	cached, exists := s.cache[userID]
	s.cacheMu.RUnlock()

	if exists && time.Now().Before(cached.expiresAt) {
		return cached.permissions, nil
	}

	// Load permissions
	permissions, err := s.loader.GetEffectivePermissions(userID)
	if err != nil {
		return nil, err
	}

	// Update cache
	s.cacheMu.Lock()
	s.cache[userID] = &cachedPermissions{
		permissions: permissions,
		expiresAt:   time.Now().Add(s.cacheTTL),
	}
	s.cacheMu.Unlock()

	return permissions, nil
}

// InvalidateCache invalidates the permission cache for a user
func (s *PermissionService) InvalidateCache(userID string) {
	s.cacheMu.Lock()
	delete(s.cache, userID)
	s.cacheMu.Unlock()
}

// InvalidateAllCaches invalidates all permission caches
func (s *PermissionService) InvalidateAllCaches() {
	s.cacheMu.Lock()
	s.cache = make(map[string]*cachedPermissions)
	s.cacheMu.Unlock()
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (s *PermissionService) CheckPermission(userID string, resource ResourceType, action Action, resourceID string) (bool, error) {
	// Get user permissions
	permissions, err := s.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	// Check for global permission first
	if HasPermissionWithScope(permissions, resource, action, ScopeGlobal) {
		return true, nil
	}

	// Get resource ownership info
	if s.resourceChecker == nil {
		// Without resource checker, only check if user has any matching permission
		return HasPermission(permissions, resource, action), nil
	}

	ownerID, teamID, orgID, err := s.resourceChecker.GetResourceOwner(resource, resourceID)
	if err != nil {
		return false, err
	}

	// Check own scope
	if ownerID == userID && HasPermissionWithScope(permissions, resource, action, ScopeOwn) {
		return true, nil
	}

	// Check team scope
	if teamID != "" {
		isTeamMember, err := s.membershipChecker.IsTeamMember(userID, teamID)
		if err != nil {
			return false, err
		}
		if isTeamMember && HasPermissionWithScope(permissions, resource, action, ScopeTeam) {
			return true, nil
		}
	}

	// Check organization scope
	if orgID != "" {
		isOrgMember, err := s.membershipChecker.IsOrgMember(userID, orgID)
		if err != nil {
			return false, err
		}
		if isOrgMember && HasPermissionWithScope(permissions, resource, action, ScopeOrganization) {
			return true, nil
		}
	}

	return false, nil
}

// CheckTeamPermission checks if a user has permission for a resource within a team
func (s *PermissionService) CheckTeamPermission(userID, teamID string, resource ResourceType, action Action) (bool, error) {
	// First verify team membership
	if s.membershipChecker != nil {
		isMember, err := s.membershipChecker.IsTeamMember(userID, teamID)
		if err != nil {
			return false, err
		}
		if !isMember {
			return false, nil
		}
	}

	// Get team-specific permissions
	permissions, err := s.loader.GetPermissionsForTeam(userID, teamID)
	if err != nil {
		return false, err
	}

	// Check for admin or specific permission
	return HasPermission(permissions, resource, action), nil
}

// CheckOrgPermission checks if a user has permission for a resource within an organization
func (s *PermissionService) CheckOrgPermission(userID, orgID string, resource ResourceType, action Action) (bool, error) {
	// First verify org membership
	if s.membershipChecker != nil {
		isMember, err := s.membershipChecker.IsOrgMember(userID, orgID)
		if err != nil {
			return false, err
		}
		if !isMember {
			return false, nil
		}
	}

	// Get org-specific permissions
	permissions, err := s.loader.GetPermissionsForOrganization(userID, orgID)
	if err != nil {
		return false, err
	}

	// Check for admin or specific permission
	return HasPermission(permissions, resource, action), nil
}

// CanCreateInTeam checks if a user can create resources in a team
func (s *PermissionService) CanCreateInTeam(userID, teamID string, resource ResourceType) (bool, error) {
	return s.CheckTeamPermission(userID, teamID, resource, ActionCreate)
}

// CanManageTeam checks if a user can manage a team (add/remove members, update settings)
func (s *PermissionService) CanManageTeam(userID, teamID string) (bool, error) {
	return s.CheckTeamPermission(userID, teamID, ResourceTeam, ActionAdmin)
}

// CanInviteToTeam checks if a user can invite members to a team
func (s *PermissionService) CanInviteToTeam(userID, teamID string) (bool, error) {
	return s.CheckTeamPermission(userID, teamID, ResourceTeam, ActionInvite)
}

// CanManageOrg checks if a user can manage an organization
func (s *PermissionService) CanManageOrg(userID, orgID string) (bool, error) {
	return s.CheckOrgPermission(userID, orgID, ResourceOrganization, ActionAdmin)
}

// FilterResourcesByPermission filters a list of resource IDs to only those the user can access
func (s *PermissionService) FilterResourcesByPermission(
	userID string,
	resource ResourceType,
	action Action,
	resourceIDs []string,
) ([]string, error) {
	if len(resourceIDs) == 0 {
		return resourceIDs, nil
	}

	permissions, err := s.GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	// If user has global permission, return all resources
	if HasPermissionWithScope(permissions, resource, action, ScopeGlobal) {
		return resourceIDs, nil
	}

	// Otherwise check each resource
	var allowed []string
	for _, resourceID := range resourceIDs {
		hasAccess, err := s.CheckPermission(userID, resource, action, resourceID)
		if err != nil {
			return nil, err
		}
		if hasAccess {
			allowed = append(allowed, resourceID)
		}
	}

	return allowed, nil
}

// AccessContext provides context for permission checks
type AccessContext struct {
	UserID   string
	TeamID   string
	OrgID    string
	Resource ResourceType
	Action   Action
}

// RequirePermission is a helper that returns an error if permission is denied
func (s *PermissionService) RequirePermission(ctx AccessContext, resourceID string) error {
	allowed, err := s.CheckPermission(ctx.UserID, ctx.Resource, ctx.Action, resourceID)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

// RequireTeamPermission is a helper that returns an error if team permission is denied
func (s *PermissionService) RequireTeamPermission(ctx AccessContext) error {
	allowed, err := s.CheckTeamPermission(ctx.UserID, ctx.TeamID, ctx.Resource, ctx.Action)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

// RequireOrgPermission is a helper that returns an error if org permission is denied
func (s *PermissionService) RequireOrgPermission(ctx AccessContext) error {
	allowed, err := s.CheckOrgPermission(ctx.UserID, ctx.OrgID, ctx.Resource, ctx.Action)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

// ErrPermissionDenied is returned when a user lacks permission
var ErrPermissionDenied = &PermissionError{Message: "permission denied"}

// PermissionError represents a permission-related error
type PermissionError struct {
	Message  string
	Resource ResourceType
	Action   Action
}

func (e *PermissionError) Error() string {
	if e.Resource != "" && e.Action != "" {
		return e.Message + ": " + string(e.Action) + " on " + string(e.Resource)
	}
	return e.Message
}

// IsPermissionError checks if an error is a permission error
func IsPermissionError(err error) bool {
	_, ok := err.(*PermissionError)
	return ok
}
