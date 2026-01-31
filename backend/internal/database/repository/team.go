package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Team represents a team within an organization
type Team struct {
	ID             string
	OrganizationID string
	Name           string
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TeamMember represents a member of a team
type TeamMember struct {
	ID        string
	TeamID    string
	UserID    string
	RoleID    string    // References the role table
	Role      string    // Denormalized role name for convenience
	CreatedAt time.Time
}

// TeamWithMemberCount includes member count for listing
type TeamWithMemberCount struct {
	Team
	MemberCount int
}

// TeamRepository handles team database operations
type TeamRepository struct {
	db *sql.DB
}

// NewTeamRepository creates a new team repository
func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create creates a new team
func (r *TeamRepository) Create(orgID, name, description string) (*Team, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO teams (id, organization_id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, orgID, name, description, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return &Team{
		ID:             id,
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// GetByID retrieves a team by ID
func (r *TeamRepository) GetByID(id string) (*Team, error) {
	team := &Team{}
	var description sql.NullString

	err := r.db.QueryRow(
		`SELECT id, organization_id, name, description, created_at, updated_at FROM teams WHERE id = ?`,
		id,
	).Scan(&team.ID, &team.OrganizationID, &team.Name, &description, &team.CreatedAt, &team.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	team.Description = description.String
	return team, nil
}

// Update updates a team's name and description
func (r *TeamRepository) Update(id, name, description string) error {
	now := time.Now()

	result, err := r.db.Exec(
		`UPDATE teams SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		name, description, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("team not found")
	}

	return nil
}

// Delete deletes a team by ID
func (r *TeamRepository) Delete(id string) error {
	// First delete all team members
	_, err := r.db.Exec(`DELETE FROM team_members WHERE team_id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete team members: %w", err)
	}

	// Then delete the team
	_, err = r.db.Exec(`DELETE FROM teams WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}

// ListByOrganization retrieves all teams for an organization with pagination
func (r *TeamRepository) ListByOrganization(orgID string, limit, offset int) ([]*TeamWithMemberCount, error) {
	rows, err := r.db.Query(
		`SELECT t.id, t.organization_id, t.name, t.description, t.created_at, t.updated_at,
		        (SELECT COUNT(*) FROM team_members WHERE team_id = t.id) as member_count
		 FROM teams t
		 WHERE t.organization_id = ?
		 ORDER BY t.created_at DESC
		 LIMIT ? OFFSET ?`,
		orgID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	defer rows.Close()

	var teams []*TeamWithMemberCount
	for rows.Next() {
		team := &TeamWithMemberCount{}
		var description sql.NullString

		err := rows.Scan(&team.ID, &team.OrganizationID, &team.Name, &description,
			&team.CreatedAt, &team.UpdatedAt, &team.MemberCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}

		team.Description = description.String
		teams = append(teams, team)
	}

	return teams, nil
}

// CountByOrganization returns the total count of teams for an organization
func (r *TeamRepository) CountByOrganization(orgID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM teams WHERE organization_id = ?`, orgID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count teams: %w", err)
	}
	return count, nil
}

// AddMember adds a user to a team with a role
func (r *TeamRepository) AddMember(teamID, userID, roleID, roleName string) (*TeamMember, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO team_members (id, team_id, user_id, role_id, role, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, teamID, userID, roleID, roleName, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add team member: %w", err)
	}

	return &TeamMember{
		ID:        id,
		TeamID:    teamID,
		UserID:    userID,
		RoleID:    roleID,
		Role:      roleName,
		CreatedAt: now,
	}, nil
}

// RemoveMember removes a user from a team
func (r *TeamRepository) RemoveMember(teamID, userID string) error {
	_, err := r.db.Exec(`DELETE FROM team_members WHERE team_id = ? AND user_id = ?`, teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}
	return nil
}

// UpdateMemberRole updates a team member's role
func (r *TeamRepository) UpdateMemberRole(teamID, userID, roleID, roleName string) error {
	result, err := r.db.Exec(
		`UPDATE team_members SET role_id = ?, role = ? WHERE team_id = ? AND user_id = ?`,
		roleID, roleName, teamID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

// GetMembers retrieves all members of a team
func (r *TeamRepository) GetMembers(teamID string) ([]*TeamMember, error) {
	rows, err := r.db.Query(
		`SELECT id, team_id, user_id, role_id, role, created_at FROM team_members WHERE team_id = ? ORDER BY created_at`,
		teamID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	var members []*TeamMember
	for rows.Next() {
		member := &TeamMember{}
		var roleID, role sql.NullString

		err := rows.Scan(&member.ID, &member.TeamID, &member.UserID, &roleID, &role, &member.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}

		member.RoleID = roleID.String
		member.Role = role.String
		members = append(members, member)
	}

	return members, nil
}

// GetUserTeams retrieves all teams a user belongs to
func (r *TeamRepository) GetUserTeams(userID string) ([]*Team, error) {
	rows, err := r.db.Query(
		`SELECT t.id, t.organization_id, t.name, t.description, t.created_at, t.updated_at
		 FROM teams t
		 INNER JOIN team_members m ON t.id = m.team_id
		 WHERE m.user_id = ?
		 ORDER BY t.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user teams: %w", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		var description sql.NullString

		err := rows.Scan(&team.ID, &team.OrganizationID, &team.Name, &description, &team.CreatedAt, &team.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}

		team.Description = description.String
		teams = append(teams, team)
	}

	return teams, nil
}

// GetUserTeamsInOrganization retrieves teams a user belongs to within an organization
func (r *TeamRepository) GetUserTeamsInOrganization(userID, orgID string) ([]*Team, error) {
	rows, err := r.db.Query(
		`SELECT t.id, t.organization_id, t.name, t.description, t.created_at, t.updated_at
		 FROM teams t
		 INNER JOIN team_members m ON t.id = m.team_id
		 WHERE m.user_id = ? AND t.organization_id = ?
		 ORDER BY t.created_at DESC`,
		userID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user teams in organization: %w", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		var description sql.NullString

		err := rows.Scan(&team.ID, &team.OrganizationID, &team.Name, &description, &team.CreatedAt, &team.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}

		team.Description = description.String
		teams = append(teams, team)
	}

	return teams, nil
}

// IsMember checks if a user is a member of a team
func (r *TeamRepository) IsMember(teamID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM team_members WHERE team_id = ? AND user_id = ?`,
		teamID, userID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}
	return count > 0, nil
}

// GetMemberRole gets the role of a user in a team
func (r *TeamRepository) GetMemberRole(teamID, userID string) (string, string, error) {
	var roleID, role sql.NullString
	err := r.db.QueryRow(
		`SELECT role_id, role FROM team_members WHERE team_id = ? AND user_id = ?`,
		teamID, userID,
	).Scan(&roleID, &role)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("failed to get member role: %w", err)
	}
	return roleID.String, role.String, nil
}

// GetTeamIDsForUser returns all team IDs a user belongs to
func (r *TeamRepository) GetTeamIDsForUser(userID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT team_id FROM team_members WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get team IDs: %w", err)
	}
	defer rows.Close()

	var teamIDs []string
	for rows.Next() {
		var teamID string
		if err := rows.Scan(&teamID); err != nil {
			return nil, fmt.Errorf("failed to scan team ID: %w", err)
		}
		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, nil
}

// GetTeamIDsForUserInOrganization returns team IDs a user belongs to within an organization
func (r *TeamRepository) GetTeamIDsForUserInOrganization(userID, orgID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT m.team_id
		 FROM team_members m
		 INNER JOIN teams t ON m.team_id = t.id
		 WHERE m.user_id = ? AND t.organization_id = ?`,
		userID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get team IDs in organization: %w", err)
	}
	defer rows.Close()

	var teamIDs []string
	for rows.Next() {
		var teamID string
		if err := rows.Scan(&teamID); err != nil {
			return nil, fmt.Errorf("failed to scan team ID: %w", err)
		}
		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, nil
}
