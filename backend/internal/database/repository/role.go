package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/security"
)

// RoleRecord represents a role record in the database
type RoleRecord struct {
	ID              string
	Name            string
	Description     string
	Type            string
	OrganizationID  string
	PermissionsJSON string
	IsSystem        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ToRole converts a RoleRecord to a security.Role
func (r *RoleRecord) ToRole() (*security.Role, error) {
	var permissions []security.Permission
	if r.PermissionsJSON != "" {
		if err := json.Unmarshal([]byte(r.PermissionsJSON), &permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}
	}

	return &security.Role{
		ID:             r.ID,
		Name:           r.Name,
		Description:    r.Description,
		Type:           security.RoleType(r.Type),
		OrganizationID: r.OrganizationID,
		Permissions:    permissions,
		IsSystem:       r.IsSystem,
	}, nil
}

// RoleRepository handles role database operations
type RoleRepository struct {
	db *sql.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// InitializeSystemRoles creates the predefined system roles if they don't exist
func (r *RoleRepository) InitializeSystemRoles() error {
	predefinedRoles := security.PredefinedRoles()
	now := time.Now()

	roleDescriptions := map[security.RoleType]string{
		security.RoleOrgOwner:   "Full organization control including deletion",
		security.RoleOrgAdmin:   "Organization administration without deletion rights",
		security.RoleTeamAdmin:  "Team administration and management",
		security.RoleTeamMember: "Standard team member with read/write access",
		security.RoleViewer:     "Read-only access to team resources",
	}

	for roleType, permissions := range predefinedRoles {
		permissionsJSON, err := json.Marshal(permissions)
		if err != nil {
			return fmt.Errorf("failed to marshal permissions for %s: %w", roleType, err)
		}

		// Use UPSERT to create or update system roles
		_, err = r.db.Exec(
			`INSERT INTO roles (id, name, description, type, organization_id, permissions_json, is_system, created_at, updated_at)
			 VALUES (?, ?, ?, ?, '', ?, TRUE, ?, ?)
			 ON CONFLICT(id) DO UPDATE SET
			   permissions_json = excluded.permissions_json,
			   updated_at = excluded.updated_at`,
			string(roleType), // Use role type as ID for system roles
			string(roleType),
			roleDescriptions[roleType],
			string(roleType),
			string(permissionsJSON),
			now,
			now,
		)
		if err != nil {
			return fmt.Errorf("failed to initialize system role %s: %w", roleType, err)
		}
	}

	return nil
}

// Create creates a new custom role for an organization
func (r *RoleRepository) Create(orgID, name, description string, permissions []security.Permission) (*security.Role, error) {
	id := uuid.New().String()
	now := time.Now()

	permissionsJSON, err := json.Marshal(permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}

	_, err = r.db.Exec(
		`INSERT INTO roles (id, name, description, type, organization_id, permissions_json, is_system, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, FALSE, ?, ?)`,
		id, name, description, string(security.RoleCustom), orgID, string(permissionsJSON), now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &security.Role{
		ID:             id,
		Name:           name,
		Description:    description,
		Type:           security.RoleCustom,
		OrganizationID: orgID,
		Permissions:    permissions,
		IsSystem:       false,
	}, nil
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(id string) (*security.Role, error) {
	record := &RoleRecord{}
	var orgID, description sql.NullString

	err := r.db.QueryRow(
		`SELECT id, name, description, type, organization_id, permissions_json, is_system, created_at, updated_at
		 FROM roles WHERE id = ?`,
		id,
	).Scan(&record.ID, &record.Name, &description, &record.Type, &orgID,
		&record.PermissionsJSON, &record.IsSystem, &record.CreatedAt, &record.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	record.Description = description.String
	record.OrganizationID = orgID.String

	return record.ToRole()
}

// GetByType retrieves a system role by type
func (r *RoleRepository) GetByType(roleType security.RoleType) (*security.Role, error) {
	return r.GetByID(string(roleType))
}

// Update updates a role's name, description, and permissions
func (r *RoleRepository) Update(id, name, description string, permissions []security.Permission) error {
	now := time.Now()

	permissionsJSON, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	result, err := r.db.Exec(
		`UPDATE roles SET name = ?, description = ?, permissions_json = ?, updated_at = ?
		 WHERE id = ? AND is_system = FALSE`,
		name, description, string(permissionsJSON), now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("role not found or is a system role")
	}

	return nil
}

// Delete deletes a custom role
func (r *RoleRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM roles WHERE id = ? AND is_system = FALSE`, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("role not found or is a system role")
	}

	return nil
}

// ListSystemRoles returns all system roles
func (r *RoleRepository) ListSystemRoles() ([]*security.Role, error) {
	rows, err := r.db.Query(
		`SELECT id, name, description, type, organization_id, permissions_json, is_system, created_at, updated_at
		 FROM roles WHERE is_system = TRUE ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list system roles: %w", err)
	}
	defer rows.Close()

	return r.scanRoles(rows)
}

// ListByOrganization returns all roles for an organization (system + custom)
func (r *RoleRepository) ListByOrganization(orgID string) ([]*security.Role, error) {
	rows, err := r.db.Query(
		`SELECT id, name, description, type, organization_id, permissions_json, is_system, created_at, updated_at
		 FROM roles
		 WHERE is_system = TRUE OR organization_id = ?
		 ORDER BY is_system DESC, name`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	return r.scanRoles(rows)
}

// ListCustomByOrganization returns only custom roles for an organization
func (r *RoleRepository) ListCustomByOrganization(orgID string) ([]*security.Role, error) {
	rows, err := r.db.Query(
		`SELECT id, name, description, type, organization_id, permissions_json, is_system, created_at, updated_at
		 FROM roles
		 WHERE is_system = FALSE AND organization_id = ?
		 ORDER BY name`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list custom roles: %w", err)
	}
	defer rows.Close()

	return r.scanRoles(rows)
}

// scanRoles scans role rows into a slice
func (r *RoleRepository) scanRoles(rows *sql.Rows) ([]*security.Role, error) {
	var roles []*security.Role

	for rows.Next() {
		record := &RoleRecord{}
		var orgID, description sql.NullString

		err := rows.Scan(&record.ID, &record.Name, &description, &record.Type, &orgID,
			&record.PermissionsJSON, &record.IsSystem, &record.CreatedAt, &record.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		record.Description = description.String
		record.OrganizationID = orgID.String

		role, err := record.ToRole()
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetPermissionsForRole returns the permissions for a role by ID
func (r *RoleRepository) GetPermissionsForRole(roleID string) ([]security.Permission, error) {
	role, err := r.GetByID(roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role not found: %s", roleID)
	}
	return role.Permissions, nil
}

// GetEffectivePermissions returns the combined permissions for a user based on their roles
// in teams and organizations
func (r *RoleRepository) GetEffectivePermissions(userID string) ([]security.Permission, error) {
	// Get permissions from organization memberships
	orgRows, err := r.db.Query(
		`SELECT r.permissions_json
		 FROM organization_members om
		 INNER JOIN roles r ON om.role = r.id OR om.role = r.type
		 WHERE om.user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get org permissions: %w", err)
	}
	defer orgRows.Close()

	var allPermissions [][]security.Permission

	for orgRows.Next() {
		var permJSON string
		if err := orgRows.Scan(&permJSON); err != nil {
			return nil, fmt.Errorf("failed to scan org permissions: %w", err)
		}
		if permJSON != "" {
			var perms []security.Permission
			if err := json.Unmarshal([]byte(permJSON), &perms); err != nil {
				return nil, fmt.Errorf("failed to unmarshal org permissions: %w", err)
			}
			allPermissions = append(allPermissions, perms)
		}
	}

	// Get permissions from team memberships
	teamRows, err := r.db.Query(
		`SELECT r.permissions_json
		 FROM team_members tm
		 INNER JOIN roles r ON tm.role_id = r.id OR tm.role = r.type
		 WHERE tm.user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get team permissions: %w", err)
	}
	defer teamRows.Close()

	for teamRows.Next() {
		var permJSON sql.NullString
		if err := teamRows.Scan(&permJSON); err != nil {
			return nil, fmt.Errorf("failed to scan team permissions: %w", err)
		}
		if permJSON.Valid && permJSON.String != "" {
			var perms []security.Permission
			if err := json.Unmarshal([]byte(permJSON.String), &perms); err != nil {
				return nil, fmt.Errorf("failed to unmarshal team permissions: %w", err)
			}
			allPermissions = append(allPermissions, perms)
		}
	}

	// Merge all permissions
	return security.MergePermissions(allPermissions...), nil
}

// GetPermissionsForTeam returns permissions for a user in a specific team
func (r *RoleRepository) GetPermissionsForTeam(userID, teamID string) ([]security.Permission, error) {
	var permJSON sql.NullString
	err := r.db.QueryRow(
		`SELECT r.permissions_json
		 FROM team_members tm
		 INNER JOIN roles r ON tm.role_id = r.id OR tm.role = r.type
		 WHERE tm.user_id = ? AND tm.team_id = ?`,
		userID, teamID,
	).Scan(&permJSON)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team permissions: %w", err)
	}

	if !permJSON.Valid || permJSON.String == "" {
		return nil, nil
	}

	var perms []security.Permission
	if err := json.Unmarshal([]byte(permJSON.String), &perms); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	return perms, nil
}

// GetPermissionsForOrganization returns permissions for a user in a specific organization
func (r *RoleRepository) GetPermissionsForOrganization(userID, orgID string) ([]security.Permission, error) {
	var permJSON sql.NullString
	err := r.db.QueryRow(
		`SELECT r.permissions_json
		 FROM organization_members om
		 INNER JOIN roles r ON om.role = r.id OR om.role = r.type
		 WHERE om.user_id = ? AND om.organization_id = ?`,
		userID, orgID,
	).Scan(&permJSON)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get org permissions: %w", err)
	}

	if !permJSON.Valid || permJSON.String == "" {
		return nil, nil
	}

	var perms []security.Permission
	if err := json.Unmarshal([]byte(permJSON.String), &perms); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	return perms, nil
}
