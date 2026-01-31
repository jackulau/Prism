package security

import (
	"testing"
)

func TestHasPermission(t *testing.T) {
	permissions := []Permission{
		{Resource: ResourceAgent, Action: ActionRead, Scope: ScopeTeam},
		{Resource: ResourceAgent, Action: ActionWrite, Scope: ScopeTeam},
		{Resource: ResourceWorkflow, Action: ActionAdmin, Scope: ScopeOrganization},
	}

	tests := []struct {
		name     string
		resource ResourceType
		action   Action
		want     bool
	}{
		{"has read agent", ResourceAgent, ActionRead, true},
		{"has write agent", ResourceAgent, ActionWrite, true},
		{"no delete agent", ResourceAgent, ActionDelete, false},
		{"admin includes read workflow", ResourceWorkflow, ActionRead, true},
		{"admin includes write workflow", ResourceWorkflow, ActionWrite, true},
		{"admin includes delete workflow", ResourceWorkflow, ActionDelete, true},
		{"no conversation permission", ResourceConversation, ActionRead, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPermission(permissions, tt.resource, tt.action)
			if got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPermissionWithScope(t *testing.T) {
	permissions := []Permission{
		{Resource: ResourceAgent, Action: ActionRead, Scope: ScopeTeam},
		{Resource: ResourceWorkflow, Action: ActionRead, Scope: ScopeOrganization},
		{Resource: ResourceConversation, Action: ActionRead, Scope: ScopeOwn},
	}

	tests := []struct {
		name     string
		resource ResourceType
		action   Action
		minScope Scope
		want     bool
	}{
		{"team scope satisfies team", ResourceAgent, ActionRead, ScopeTeam, true},
		{"team scope satisfies own", ResourceAgent, ActionRead, ScopeOwn, true},
		{"team scope doesn't satisfy org", ResourceAgent, ActionRead, ScopeOrganization, false},
		{"org scope satisfies org", ResourceWorkflow, ActionRead, ScopeOrganization, true},
		{"org scope satisfies team", ResourceWorkflow, ActionRead, ScopeTeam, true},
		{"own scope satisfies own", ResourceConversation, ActionRead, ScopeOwn, true},
		{"own scope doesn't satisfy team", ResourceConversation, ActionRead, ScopeTeam, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPermissionWithScope(permissions, tt.resource, tt.action, tt.minScope)
			if got != tt.want {
				t.Errorf("HasPermissionWithScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergePermissions(t *testing.T) {
	set1 := []Permission{
		{Resource: ResourceAgent, Action: ActionRead, Scope: ScopeOwn},
		{Resource: ResourceWorkflow, Action: ActionRead, Scope: ScopeTeam},
	}

	set2 := []Permission{
		{Resource: ResourceAgent, Action: ActionRead, Scope: ScopeTeam},
		{Resource: ResourceConversation, Action: ActionWrite, Scope: ScopeOwn},
	}

	merged := MergePermissions(set1, set2)

	// Should have 3 unique resource:action combinations
	if len(merged) != 3 {
		t.Errorf("MergePermissions() length = %d, want 3", len(merged))
	}

	// Agent:Read should have Team scope (higher)
	found := false
	for _, p := range merged {
		if p.Resource == ResourceAgent && p.Action == ActionRead {
			found = true
			if p.Scope != ScopeTeam {
				t.Errorf("Agent:Read scope = %v, want %v", p.Scope, ScopeTeam)
			}
		}
	}
	if !found {
		t.Error("Agent:Read permission not found in merged result")
	}
}

func TestPredefinedRoles(t *testing.T) {
	roles := PredefinedRoles()

	// Verify all expected roles exist
	expectedRoles := []RoleType{RoleOrgOwner, RoleOrgAdmin, RoleTeamAdmin, RoleTeamMember, RoleViewer}
	for _, role := range expectedRoles {
		if _, ok := roles[role]; !ok {
			t.Errorf("PredefinedRoles() missing role %v", role)
		}
	}

	// Verify org_owner has admin on all resources
	orgOwnerPerms := roles[RoleOrgOwner]
	if !HasPermission(orgOwnerPerms, ResourceOrganization, ActionAdmin) {
		t.Error("OrgOwner should have admin on organization")
	}
	if !HasPermission(orgOwnerPerms, ResourceTeam, ActionAdmin) {
		t.Error("OrgOwner should have admin on team")
	}

	// Verify viewer only has read permissions
	viewerPerms := roles[RoleViewer]
	if HasPermission(viewerPerms, ResourceAgent, ActionWrite) {
		t.Error("Viewer should not have write on agent")
	}
	if !HasPermission(viewerPerms, ResourceAgent, ActionRead) {
		t.Error("Viewer should have read on agent")
	}
}
