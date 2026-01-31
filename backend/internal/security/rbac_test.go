package security

import (
	"testing"
)

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"valid user role", "user", true},
		{"valid admin role", "admin", true},
		{"invalid role", "superuser", false},
		{"empty role", "", false},
		{"uppercase role", "ADMIN", false},
		{"mixed case role", "Admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidRole(tt.role)
			if result != tt.expected {
				t.Errorf("IsValidRole(%q) = %v, expected %v", tt.role, result, tt.expected)
			}
		})
	}
}

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       Role
		permission Permission
		expected   bool
	}{
		// User permissions
		{"user can view conversations", RoleUser, PermViewConversations, true},
		{"user can create conversations", RoleUser, PermCreateConversations, true},
		{"user can manage own data", RoleUser, PermManageOwnData, true},
		{"user cannot manage users", RoleUser, PermManageUsers, false},
		{"user cannot manage organization", RoleUser, PermManageOrganization, false},
		{"user cannot view all conversations", RoleUser, PermViewAllConversations, false},
		{"user cannot manage providers", RoleUser, PermManageProviders, false},
		{"user cannot manage settings", RoleUser, PermManageSettings, false},

		// Admin permissions
		{"admin can view conversations", RoleAdmin, PermViewConversations, true},
		{"admin can create conversations", RoleAdmin, PermCreateConversations, true},
		{"admin can manage own data", RoleAdmin, PermManageOwnData, true},
		{"admin can manage users", RoleAdmin, PermManageUsers, true},
		{"admin can manage organization", RoleAdmin, PermManageOrganization, true},
		{"admin can view all conversations", RoleAdmin, PermViewAllConversations, true},
		{"admin can manage providers", RoleAdmin, PermManageProviders, true},
		{"admin can manage settings", RoleAdmin, PermManageSettings, true},

		// Invalid role
		{"invalid role has no permissions", Role("invalid"), PermViewConversations, false},
		{"empty role has no permissions", Role(""), PermViewConversations, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasPermission(tt.role, tt.permission)
			if result != tt.expected {
				t.Errorf("HasPermission(%q, %q) = %v, expected %v", tt.role, tt.permission, result, tt.expected)
			}
		})
	}
}

func TestGetPermissions(t *testing.T) {
	tests := []struct {
		name           string
		role           Role
		minPermissions int
	}{
		{"user has basic permissions", RoleUser, 3},
		{"admin has all permissions", RoleAdmin, 8},
		{"invalid role has no permissions", Role("invalid"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := GetPermissions(tt.role)
			if len(permissions) < tt.minPermissions {
				t.Errorf("GetPermissions(%q) returned %d permissions, expected at least %d", tt.role, len(permissions), tt.minPermissions)
			}
		})
	}
}

func TestGetPermissionStrings(t *testing.T) {
	permissions := GetPermissionStrings(RoleUser)

	if len(permissions) == 0 {
		t.Error("GetPermissionStrings(RoleUser) returned empty slice")
	}

	// Check that permissions are strings, not Permission types
	for _, p := range permissions {
		if p == "" {
			t.Error("GetPermissionStrings returned empty permission string")
		}
	}
}

func TestGetPermissionsReturnsCopy(t *testing.T) {
	// Get permissions twice
	perms1 := GetPermissions(RoleUser)
	perms2 := GetPermissions(RoleUser)

	// Modify the first slice
	if len(perms1) > 0 {
		perms1[0] = "modified"
	}

	// Check that the second slice is not affected
	if len(perms2) > 0 && perms2[0] == "modified" {
		t.Error("GetPermissions did not return a copy; modification affected other results")
	}
}

func TestRolePermissionsConsistency(t *testing.T) {
	// Admin should have all user permissions
	userPerms := GetPermissions(RoleUser)
	adminPerms := GetPermissions(RoleAdmin)

	for _, userPerm := range userPerms {
		found := false
		for _, adminPerm := range adminPerms {
			if userPerm == adminPerm {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Admin role missing user permission: %s", userPerm)
		}
	}
}

func TestValidRolesConstant(t *testing.T) {
	if len(ValidRoles) != 2 {
		t.Errorf("Expected 2 valid roles, got %d", len(ValidRoles))
	}

	hasUser := false
	hasAdmin := false
	for _, role := range ValidRoles {
		if role == RoleUser {
			hasUser = true
		}
		if role == RoleAdmin {
			hasAdmin = true
		}
	}

	if !hasUser {
		t.Error("ValidRoles missing RoleUser")
	}
	if !hasAdmin {
		t.Error("ValidRoles missing RoleAdmin")
	}
}

func TestRoleConstants(t *testing.T) {
	if RoleUser != "user" {
		t.Errorf("RoleUser = %q, expected 'user'", RoleUser)
	}
	if RoleAdmin != "admin" {
		t.Errorf("RoleAdmin = %q, expected 'admin'", RoleAdmin)
	}
}

func TestPermissionConstants(t *testing.T) {
	expectedPermissions := map[Permission]string{
		PermViewConversations:    "conversations:view",
		PermCreateConversations:  "conversations:create",
		PermManageOwnData:        "own_data:manage",
		PermManageUsers:          "users:manage",
		PermManageOrganization:   "organization:manage",
		PermViewAllConversations: "conversations:view_all",
		PermManageProviders:      "providers:manage",
		PermManageSettings:       "settings:manage",
	}

	for perm, expected := range expectedPermissions {
		if string(perm) != expected {
			t.Errorf("Permission %v = %q, expected %q", perm, string(perm), expected)
		}
	}
}
