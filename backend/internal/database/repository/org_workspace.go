package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OrgWorkspace represents an organization-scoped workspace for agent sessions.
//
// Workspace has One-to-Many relation with Agent:
// When Agent entity is created, add:
//   - ListAgentsByWorkspaceID(workspaceID string) ([]*Agent, error)
//   - GetWorkspaceByAgentID(agentID string) (*OrgWorkspace, error)
//
// Workspace has One-to-Many relation with WorkspaceTool:
// When WorkspaceTool entity is created, add:
//   - ListToolsByWorkspaceID(workspaceID string) ([]*WorkspaceTool, error)
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

// OrgWorkspaceRepository handles organization workspace database operations
type OrgWorkspaceRepository struct {
	db *sql.DB
}

// NewOrgWorkspaceRepository creates a new organization workspace repository
func NewOrgWorkspaceRepository(db *sql.DB) *OrgWorkspaceRepository {
	return &OrgWorkspaceRepository{db: db}
}

// nullString converts an empty string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// scanWorkspace is a helper to scan workspace rows with nullable fields
func (r *OrgWorkspaceRepository) scanWorkspace(scanner interface {
	Scan(dest ...interface{}) error
}) (*OrgWorkspace, error) {
	ws := &OrgWorkspace{}
	var githubRepo, workerID, currentBranch sql.NullString
	var slackChannelID, slackMessageTs sql.NullString

	err := scanner.Scan(
		&ws.ID, &ws.Name, &ws.OrganizationID,
		&githubRepo, &workerID, &currentBranch,
		&slackChannelID, &slackMessageTs, &ws.CreatedAt,
	)
	if err != nil {
		return nil, err
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

// Create creates a new organization workspace
func (r *OrgWorkspaceRepository) Create(workspace *OrgWorkspace) (*OrgWorkspace, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO org_workspaces (id, name, organization_id, github_repository_name, worker_id, current_branch, slack_channel_id, slack_message_ts, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, workspace.Name, workspace.OrganizationID,
		nullString(workspace.GitHubRepositoryName), nullString(workspace.WorkerID), nullString(workspace.CurrentBranch),
		nullString(workspace.SlackChannelID), nullString(workspace.SlackMessageTs), now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	return &OrgWorkspace{
		ID:                   id,
		Name:                 workspace.Name,
		OrganizationID:       workspace.OrganizationID,
		GitHubRepositoryName: workspace.GitHubRepositoryName,
		WorkerID:             workspace.WorkerID,
		CurrentBranch:        workspace.CurrentBranch,
		SlackChannelID:       workspace.SlackChannelID,
		SlackMessageTs:       workspace.SlackMessageTs,
		CreatedAt:            now,
	}, nil
}

// GetByID retrieves an organization workspace by ID
func (r *OrgWorkspaceRepository) GetByID(id string) (*OrgWorkspace, error) {
	row := r.db.QueryRow(
		`SELECT id, name, organization_id, github_repository_name, worker_id, current_branch, slack_channel_id, slack_message_ts, created_at
		 FROM org_workspaces WHERE id = ?`,
		id,
	)

	ws, err := r.scanWorkspace(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	return ws, nil
}

// Update updates an organization workspace
func (r *OrgWorkspaceRepository) Update(workspace *OrgWorkspace) error {
	_, err := r.db.Exec(
		`UPDATE org_workspaces SET name = ?, github_repository_name = ?, worker_id = ?, current_branch = ?, slack_channel_id = ?, slack_message_ts = ?
		 WHERE id = ?`,
		workspace.Name, nullString(workspace.GitHubRepositoryName), nullString(workspace.WorkerID),
		nullString(workspace.CurrentBranch), nullString(workspace.SlackChannelID), nullString(workspace.SlackMessageTs),
		workspace.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	return nil
}

// Delete deletes an organization workspace by ID
func (r *OrgWorkspaceRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM org_workspaces WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}

// ListByOrganizationID retrieves all workspaces for an organization
func (r *OrgWorkspaceRepository) ListByOrganizationID(orgID string) ([]*OrgWorkspace, error) {
	rows, err := r.db.Query(
		`SELECT id, name, organization_id, github_repository_name, worker_id, current_branch, slack_channel_id, slack_message_ts, created_at
		 FROM org_workspaces WHERE organization_id = ? ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*OrgWorkspace
	for rows.Next() {
		ws, err := r.scanWorkspace(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, ws)
	}

	return workspaces, nil
}

// GetByGitHubRepo retrieves a workspace by GitHub repository name within an organization
func (r *OrgWorkspaceRepository) GetByGitHubRepo(orgID, repoName string) (*OrgWorkspace, error) {
	row := r.db.QueryRow(
		`SELECT id, name, organization_id, github_repository_name, worker_id, current_branch, slack_channel_id, slack_message_ts, created_at
		 FROM org_workspaces WHERE organization_id = ? AND github_repository_name = ?`,
		orgID, repoName,
	)

	ws, err := r.scanWorkspace(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace by GitHub repo: %w", err)
	}

	return ws, nil
}

// ListByBranch retrieves all workspaces on a specific branch
func (r *OrgWorkspaceRepository) ListByBranch(branch string) ([]*OrgWorkspace, error) {
	rows, err := r.db.Query(
		`SELECT id, name, organization_id, github_repository_name, worker_id, current_branch, slack_channel_id, slack_message_ts, created_at
		 FROM org_workspaces WHERE current_branch = ? ORDER BY created_at DESC`,
		branch,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces by branch: %w", err)
	}
	defer rows.Close()

	var workspaces []*OrgWorkspace
	for rows.Next() {
		ws, err := r.scanWorkspace(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, ws)
	}

	return workspaces, nil
}

// UpdateBranch updates only the current_branch field for a workspace
func (r *OrgWorkspaceRepository) UpdateBranch(id, branch string) error {
	_, err := r.db.Exec(
		`UPDATE org_workspaces SET current_branch = ? WHERE id = ?`,
		nullString(branch), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update workspace branch: %w", err)
	}
	return nil
}

// UpdateSlackInfo updates the Slack integration fields for a workspace
func (r *OrgWorkspaceRepository) UpdateSlackInfo(id, channelID, messageTs string) error {
	_, err := r.db.Exec(
		`UPDATE org_workspaces SET slack_channel_id = ?, slack_message_ts = ? WHERE id = ?`,
		nullString(channelID), nullString(messageTs), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update workspace Slack info: %w", err)
	}
	return nil
}
