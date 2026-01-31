package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Organization represents an organization in the database
type Organization struct {
	ID                   string
	Name                 string
	WorkOSOrganizationID sql.NullString
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	ID             string
	OrganizationID string
	UserID         string
	Role           string
	CreatedAt      time.Time
}

// OrganizationRepository handles organization database operations
type OrganizationRepository struct {
	db *sql.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create creates a new organization
func (r *OrganizationRepository) Create(name string, workosOrgID string) (*Organization, error) {
	id := uuid.New().String()
	now := time.Now()

	var workosID sql.NullString
	if workosOrgID != "" {
		workosID = sql.NullString{String: workosOrgID, Valid: true}
	}

	_, err := r.db.Exec(
		`INSERT INTO organizations (id, name, workos_organization_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, name, workosID, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return &Organization{
		ID:                   id,
		Name:                 name,
		WorkOSOrganizationID: workosID,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(id string) (*Organization, error) {
	org := &Organization{}
	err := r.db.QueryRow(
		`SELECT id, name, workos_organization_id, created_at, updated_at FROM organizations WHERE id = ?`,
		id,
	).Scan(&org.ID, &org.Name, &org.WorkOSOrganizationID, &org.CreatedAt, &org.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}

// GetByWorkOSID retrieves an organization by WorkOS organization ID
func (r *OrganizationRepository) GetByWorkOSID(workosOrgID string) (*Organization, error) {
	org := &Organization{}
	err := r.db.QueryRow(
		`SELECT id, name, workos_organization_id, created_at, updated_at FROM organizations WHERE workos_organization_id = ?`,
		workosOrgID,
	).Scan(&org.ID, &org.Name, &org.WorkOSOrganizationID, &org.CreatedAt, &org.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}

// Update updates an organization's name and WorkOS ID
func (r *OrganizationRepository) Update(id, name string, workosOrgID string) error {
	now := time.Now()

	var workosID sql.NullString
	if workosOrgID != "" {
		workosID = sql.NullString{String: workosOrgID, Valid: true}
	}

	result, err := r.db.Exec(
		`UPDATE organizations SET name = ?, workos_organization_id = ?, updated_at = ? WHERE id = ?`,
		name, workosID, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}

// UpdateWorkOSID updates only the WorkOS organization ID
func (r *OrganizationRepository) UpdateWorkOSID(id, workosOrgID string) error {
	now := time.Now()

	var workosID sql.NullString
	if workosOrgID != "" {
		workosID = sql.NullString{String: workosOrgID, Valid: true}
	}

	_, err := r.db.Exec(
		`UPDATE organizations SET workos_organization_id = ?, updated_at = ? WHERE id = ?`,
		workosID, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update organization workos id: %w", err)
	}

	return nil
}

// Delete deletes an organization by ID
func (r *OrganizationRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM organizations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

// DeleteByWorkOSID deletes an organization by WorkOS organization ID
func (r *OrganizationRepository) DeleteByWorkOSID(workosOrgID string) error {
	_, err := r.db.Exec(`DELETE FROM organizations WHERE workos_organization_id = ?`, workosOrgID)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

// List retrieves all organizations with pagination
func (r *OrganizationRepository) List(limit, offset int) ([]*Organization, error) {
	rows, err := r.db.Query(
		`SELECT id, name, workos_organization_id, created_at, updated_at FROM organizations ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*Organization
	for rows.Next() {
		org := &Organization{}
		if err := rows.Scan(&org.ID, &org.Name, &org.WorkOSOrganizationID, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// AddMember adds a user to an organization
func (r *OrganizationRepository) AddMember(orgID, userID, role string) (*OrganizationMember, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO organization_members (id, organization_id, user_id, role, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, orgID, userID, role, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add organization member: %w", err)
	}

	return &OrganizationMember{
		ID:             id,
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
		CreatedAt:      now,
	}, nil
}

// RemoveMember removes a user from an organization
func (r *OrganizationRepository) RemoveMember(orgID, userID string) error {
	_, err := r.db.Exec(`DELETE FROM organization_members WHERE organization_id = ? AND user_id = ?`, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove organization member: %w", err)
	}
	return nil
}

// GetMembers retrieves all members of an organization
func (r *OrganizationRepository) GetMembers(orgID string) ([]*OrganizationMember, error) {
	rows, err := r.db.Query(
		`SELECT id, organization_id, user_id, role, created_at FROM organization_members WHERE organization_id = ? ORDER BY created_at`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization members: %w", err)
	}
	defer rows.Close()

	var members []*OrganizationMember
	for rows.Next() {
		member := &OrganizationMember{}
		if err := rows.Scan(&member.ID, &member.OrganizationID, &member.UserID, &member.Role, &member.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan organization member: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

// GetUserOrganizations retrieves all organizations a user belongs to
func (r *OrganizationRepository) GetUserOrganizations(userID string) ([]*Organization, error) {
	rows, err := r.db.Query(
		`SELECT o.id, o.name, o.workos_organization_id, o.created_at, o.updated_at
		 FROM organizations o
		 INNER JOIN organization_members m ON o.id = m.organization_id
		 WHERE m.user_id = ?
		 ORDER BY o.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*Organization
	for rows.Next() {
		org := &Organization{}
		if err := rows.Scan(&org.ID, &org.Name, &org.WorkOSOrganizationID, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// IsMember checks if a user is a member of an organization
func (r *OrganizationRepository) IsMember(orgID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ?`,
		orgID, userID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}
	return count > 0, nil
}

// GetMemberRole gets the role of a user in an organization
func (r *OrganizationRepository) GetMemberRole(orgID, userID string) (string, error) {
	var role string
	err := r.db.QueryRow(
		`SELECT role FROM organization_members WHERE organization_id = ? AND user_id = ?`,
		orgID, userID,
	).Scan(&role)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get member role: %w", err)
	}
	return role, nil
}

// GetMember retrieves a specific member of an organization
func (r *OrganizationRepository) GetMember(orgID, userID string) (*OrganizationMember, error) {
	member := &OrganizationMember{}
	err := r.db.QueryRow(
		`SELECT id, organization_id, user_id, role, created_at FROM organization_members WHERE organization_id = ? AND user_id = ?`,
		orgID, userID,
	).Scan(&member.ID, &member.OrganizationID, &member.UserID, &member.Role, &member.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization member: %w", err)
	}
	return member, nil
}
