package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OrgWorkspace represents an organization-scoped workspace for agent sessions
type OrgWorkspace struct {
	ID                   string
	Name                 string
	OrganizationID       string
	GitHubRepositoryName string
	WorkerID             string
	CurrentBranch        string
	SlackChannelID       string
	SlackMessageTs       string
	CreatedAt            time.Time
}

// OrgWorkspaceRepository handles org workspace database operations
type OrgWorkspaceRepository struct {
	db *sql.DB
}

// NewOrgWorkspaceRepository creates a new org workspace repository
func NewOrgWorkspaceRepository(db *sql.DB) *OrgWorkspaceRepository {
	return &OrgWorkspaceRepository{db: db}
}

// Create creates a new organization workspace
func (r *OrgWorkspaceRepository) Create(ws *OrgWorkspace) (*OrgWorkspace, error) {
	ws.ID = uuid.New().String()
	ws.CreatedAt = time.Now()

	_, err := r.db.Exec(
		`INSERT INTO org_workspaces (id, name, organization_id, github_repository_name, worker_id, current_branch, slack_channel_id, slack_message_ts, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ws.ID, ws.Name, ws.OrganizationID, nullString(ws.GitHubRepositoryName), nullString(ws.WorkerID),
		nullString(ws.CurrentBranch), nullString(ws.SlackChannelID), nullString(ws.SlackMessageTs), ws.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create org workspace: %w", err)
	}

	return ws, nil
}

// GetByID retrieves an organization workspace by ID
func (r *OrgWorkspaceRepository) GetByID(id string) (*OrgWorkspace, error) {
	ws := &OrgWorkspace{}
	var githubRepo, workerID, currentBranch sql.NullString
	var slackChannelID, slackMessageTs sql.NullString

	err := r.db.QueryRow(
		`SELECT id, name, organization_id, github_repository_name, worker_id, current_branch, slack_channel_id, slack_message_ts, created_at
		 FROM org_workspaces WHERE id = ?`,
		id,
	).Scan(&ws.ID, &ws.Name, &ws.OrganizationID, &githubRepo, &workerID, &currentBranch, &slackChannelID, &slackMessageTs, &ws.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get org workspace: %w", err)
	}

	if githubRepo.Valid {
		ws.GitHubRepositoryName = githubRepo.String
	}
	if workerID.Valid {
		ws.WorkerID = workerID.String
	}
	if currentBranch.Valid {
		ws.CurrentBranch = currentBranch.String
	}
	if slackChannelID.Valid {
		ws.SlackChannelID = slackChannelID.String
	}
	if slackMessageTs.Valid {
		ws.SlackMessageTs = slackMessageTs.String
	}

	return ws, nil
}

// ListByOrganizationID retrieves all workspaces for an organization
func (r *OrgWorkspaceRepository) ListByOrganizationID(orgID string) ([]*OrgWorkspace, error) {
	rows, err := r.db.Query(
		`SELECT id, name, organization_id, github_repository_name, worker_id, current_branch, slack_channel_id, slack_message_ts, created_at
		 FROM org_workspaces WHERE organization_id = ? ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list org workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*OrgWorkspace
	for rows.Next() {
		ws := &OrgWorkspace{}
		var githubRepo, workerID, currentBranch sql.NullString
		var slackChannelID, slackMessageTs sql.NullString

		err := rows.Scan(&ws.ID, &ws.Name, &ws.OrganizationID, &githubRepo, &workerID, &currentBranch, &slackChannelID, &slackMessageTs, &ws.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan org workspace: %w", err)
		}

		if githubRepo.Valid {
			ws.GitHubRepositoryName = githubRepo.String
		}
		if workerID.Valid {
			ws.WorkerID = workerID.String
		}
		if currentBranch.Valid {
			ws.CurrentBranch = currentBranch.String
		}
		if slackChannelID.Valid {
			ws.SlackChannelID = slackChannelID.String
		}
		if slackMessageTs.Valid {
			ws.SlackMessageTs = slackMessageTs.String
		}

		workspaces = append(workspaces, ws)
	}

	return workspaces, nil
}

// Update updates an organization workspace
func (r *OrgWorkspaceRepository) Update(ws *OrgWorkspace) error {
	_, err := r.db.Exec(
		`UPDATE org_workspaces SET name = ?, github_repository_name = ?, worker_id = ?, current_branch = ?, slack_channel_id = ?, slack_message_ts = ?
		 WHERE id = ?`,
		ws.Name, nullString(ws.GitHubRepositoryName), nullString(ws.WorkerID),
		nullString(ws.CurrentBranch), nullString(ws.SlackChannelID), nullString(ws.SlackMessageTs), ws.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update org workspace: %w", err)
	}
	return nil
}

// UpdateBranch updates only the current_branch field
func (r *OrgWorkspaceRepository) UpdateBranch(id, branch string) error {
	_, err := r.db.Exec(
		`UPDATE org_workspaces SET current_branch = ? WHERE id = ?`,
		nullString(branch), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update org workspace branch: %w", err)
	}
	return nil
}

// Delete removes an organization workspace by ID
func (r *OrgWorkspaceRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM org_workspaces WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete org workspace: %w", err)
	}
	return nil
}

// nullString converts an empty string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
