package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GitHubInstallation represents a GitHub App installation in the database
type GitHubInstallation struct {
	ID                  string    `json:"id"`
	InstallationID      int64     `json:"installation_id"`
	AccountID           int64     `json:"account_id"`
	AccountLogin        string    `json:"account_login"`
	AccountType         string    `json:"account_type"` // "User" or "Organization"
	AccountAvatarURL    string    `json:"account_avatar_url"`
	AppID               int64     `json:"app_id"`
	TargetType          string    `json:"target_type"`
	Permissions         string    `json:"permissions"` // JSON-encoded permissions
	Events              string    `json:"events"`      // JSON-encoded event list
	RepositorySelection string    `json:"repository_selection"`
	SuspendedAt         *time.Time `json:"suspended_at,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// GitHubInstallationRepository represents a repository accessible via an installation
type GitHubInstallationRepository struct {
	ID             string    `json:"id"`
	InstallationID int64     `json:"installation_id"`
	RepositoryID   int64     `json:"repository_id"`
	FullName       string    `json:"full_name"`
	Name           string    `json:"name"`
	Private        bool      `json:"private"`
	HTMLURL        string    `json:"html_url"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
}

// GitHubInstallationRepo handles GitHub installation database operations
type GitHubInstallationRepo struct {
	db *sql.DB
}

// NewGitHubInstallationRepo creates a new GitHub installation repository
func NewGitHubInstallationRepo(db *sql.DB) *GitHubInstallationRepo {
	return &GitHubInstallationRepo{db: db}
}

// Create creates a new GitHub App installation record
func (r *GitHubInstallationRepo) Create(installation *GitHubInstallation) error {
	installation.ID = uuid.New().String()
	now := time.Now()
	installation.CreatedAt = now
	installation.UpdatedAt = now

	_, err := r.db.Exec(`
		INSERT INTO github_app_installations (
			id, installation_id, account_id, account_login, account_type,
			account_avatar_url, app_id, target_type, permissions, events,
			repository_selection, suspended_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		installation.ID, installation.InstallationID, installation.AccountID,
		installation.AccountLogin, installation.AccountType, installation.AccountAvatarURL,
		installation.AppID, installation.TargetType, installation.Permissions,
		installation.Events, installation.RepositorySelection, installation.SuspendedAt,
		installation.CreatedAt, installation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create installation: %w", err)
	}

	return nil
}

// GetByInstallationID retrieves an installation by its GitHub installation ID
func (r *GitHubInstallationRepo) GetByInstallationID(installationID int64) (*GitHubInstallation, error) {
	installation := &GitHubInstallation{}
	var suspendedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, installation_id, account_id, account_login, account_type,
			account_avatar_url, app_id, target_type, permissions, events,
			repository_selection, suspended_at, created_at, updated_at
		FROM github_app_installations
		WHERE installation_id = ?
	`, installationID).Scan(
		&installation.ID, &installation.InstallationID, &installation.AccountID,
		&installation.AccountLogin, &installation.AccountType, &installation.AccountAvatarURL,
		&installation.AppID, &installation.TargetType, &installation.Permissions,
		&installation.Events, &installation.RepositorySelection, &suspendedAt,
		&installation.CreatedAt, &installation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get installation: %w", err)
	}

	if suspendedAt.Valid {
		installation.SuspendedAt = &suspendedAt.Time
	}

	return installation, nil
}

// GetByAccountLogin retrieves an installation by account login
func (r *GitHubInstallationRepo) GetByAccountLogin(accountLogin string) (*GitHubInstallation, error) {
	installation := &GitHubInstallation{}
	var suspendedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, installation_id, account_id, account_login, account_type,
			account_avatar_url, app_id, target_type, permissions, events,
			repository_selection, suspended_at, created_at, updated_at
		FROM github_app_installations
		WHERE account_login = ?
	`, accountLogin).Scan(
		&installation.ID, &installation.InstallationID, &installation.AccountID,
		&installation.AccountLogin, &installation.AccountType, &installation.AccountAvatarURL,
		&installation.AppID, &installation.TargetType, &installation.Permissions,
		&installation.Events, &installation.RepositorySelection, &suspendedAt,
		&installation.CreatedAt, &installation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get installation: %w", err)
	}

	if suspendedAt.Valid {
		installation.SuspendedAt = &suspendedAt.Time
	}

	return installation, nil
}

// List retrieves all installations
func (r *GitHubInstallationRepo) List() ([]*GitHubInstallation, error) {
	rows, err := r.db.Query(`
		SELECT id, installation_id, account_id, account_login, account_type,
			account_avatar_url, app_id, target_type, permissions, events,
			repository_selection, suspended_at, created_at, updated_at
		FROM github_app_installations
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list installations: %w", err)
	}
	defer rows.Close()

	var installations []*GitHubInstallation
	for rows.Next() {
		installation := &GitHubInstallation{}
		var suspendedAt sql.NullTime

		err := rows.Scan(
			&installation.ID, &installation.InstallationID, &installation.AccountID,
			&installation.AccountLogin, &installation.AccountType, &installation.AccountAvatarURL,
			&installation.AppID, &installation.TargetType, &installation.Permissions,
			&installation.Events, &installation.RepositorySelection, &suspendedAt,
			&installation.CreatedAt, &installation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan installation: %w", err)
		}

		if suspendedAt.Valid {
			installation.SuspendedAt = &suspendedAt.Time
		}

		installations = append(installations, installation)
	}

	return installations, nil
}

// Update updates an existing installation
func (r *GitHubInstallationRepo) Update(installation *GitHubInstallation) error {
	installation.UpdatedAt = time.Now()

	_, err := r.db.Exec(`
		UPDATE github_app_installations
		SET account_login = ?, account_type = ?, account_avatar_url = ?,
			permissions = ?, events = ?, repository_selection = ?,
			suspended_at = ?, updated_at = ?
		WHERE installation_id = ?
	`,
		installation.AccountLogin, installation.AccountType, installation.AccountAvatarURL,
		installation.Permissions, installation.Events, installation.RepositorySelection,
		installation.SuspendedAt, installation.UpdatedAt, installation.InstallationID,
	)
	if err != nil {
		return fmt.Errorf("failed to update installation: %w", err)
	}

	return nil
}

// Delete removes an installation by its GitHub installation ID
func (r *GitHubInstallationRepo) Delete(installationID int64) error {
	// Delete associated repositories first
	_, err := r.db.Exec(`DELETE FROM github_installation_repositories WHERE installation_id = ?`, installationID)
	if err != nil {
		return fmt.Errorf("failed to delete installation repositories: %w", err)
	}

	// Delete the installation
	_, err = r.db.Exec(`DELETE FROM github_app_installations WHERE installation_id = ?`, installationID)
	if err != nil {
		return fmt.Errorf("failed to delete installation: %w", err)
	}

	return nil
}

// Suspend marks an installation as suspended
func (r *GitHubInstallationRepo) Suspend(installationID int64) error {
	now := time.Now()
	_, err := r.db.Exec(`
		UPDATE github_app_installations
		SET suspended_at = ?, updated_at = ?
		WHERE installation_id = ?
	`, now, now, installationID)
	if err != nil {
		return fmt.Errorf("failed to suspend installation: %w", err)
	}

	return nil
}

// Unsuspend removes suspension from an installation
func (r *GitHubInstallationRepo) Unsuspend(installationID int64) error {
	now := time.Now()
	_, err := r.db.Exec(`
		UPDATE github_app_installations
		SET suspended_at = NULL, updated_at = ?
		WHERE installation_id = ?
	`, now, installationID)
	if err != nil {
		return fmt.Errorf("failed to unsuspend installation: %w", err)
	}

	return nil
}

// AddRepository adds a repository to an installation
func (r *GitHubInstallationRepo) AddRepository(repo *GitHubInstallationRepository) error {
	repo.ID = uuid.New().String()
	repo.CreatedAt = time.Now()

	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO github_installation_repositories (
			id, installation_id, repository_id, full_name, name,
			private, html_url, description, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		repo.ID, repo.InstallationID, repo.RepositoryID, repo.FullName,
		repo.Name, repo.Private, repo.HTMLURL, repo.Description, repo.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add repository: %w", err)
	}

	return nil
}

// RemoveRepository removes a repository from an installation
func (r *GitHubInstallationRepo) RemoveRepository(installationID int64, repositoryID int64) error {
	_, err := r.db.Exec(`
		DELETE FROM github_installation_repositories
		WHERE installation_id = ? AND repository_id = ?
	`, installationID, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}

	return nil
}

// ListRepositories lists all repositories for an installation
func (r *GitHubInstallationRepo) ListRepositories(installationID int64) ([]*GitHubInstallationRepository, error) {
	rows, err := r.db.Query(`
		SELECT id, installation_id, repository_id, full_name, name,
			private, html_url, description, created_at
		FROM github_installation_repositories
		WHERE installation_id = ?
		ORDER BY full_name ASC
	`, installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}
	defer rows.Close()

	var repos []*GitHubInstallationRepository
	for rows.Next() {
		repo := &GitHubInstallationRepository{}
		err := rows.Scan(
			&repo.ID, &repo.InstallationID, &repo.RepositoryID, &repo.FullName,
			&repo.Name, &repo.Private, &repo.HTMLURL, &repo.Description, &repo.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan repository: %w", err)
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

// GetRepositoryByFullName retrieves a repository by its full name
func (r *GitHubInstallationRepo) GetRepositoryByFullName(fullName string) (*GitHubInstallationRepository, error) {
	repo := &GitHubInstallationRepository{}

	err := r.db.QueryRow(`
		SELECT id, installation_id, repository_id, full_name, name,
			private, html_url, description, created_at
		FROM github_installation_repositories
		WHERE full_name = ?
	`, fullName).Scan(
		&repo.ID, &repo.InstallationID, &repo.RepositoryID, &repo.FullName,
		&repo.Name, &repo.Private, &repo.HTMLURL, &repo.Description, &repo.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return repo, nil
}

// SetRepositories replaces all repositories for an installation
func (r *GitHubInstallationRepo) SetRepositories(installationID int64, repos []*GitHubInstallationRepository) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing repositories
	_, err = tx.Exec(`DELETE FROM github_installation_repositories WHERE installation_id = ?`, installationID)
	if err != nil {
		return fmt.Errorf("failed to delete existing repositories: %w", err)
	}

	// Insert new repositories
	for _, repo := range repos {
		repo.ID = uuid.New().String()
		repo.InstallationID = installationID
		repo.CreatedAt = time.Now()

		_, err := tx.Exec(`
			INSERT INTO github_installation_repositories (
				id, installation_id, repository_id, full_name, name,
				private, html_url, description, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			repo.ID, repo.InstallationID, repo.RepositoryID, repo.FullName,
			repo.Name, repo.Private, repo.HTMLURL, repo.Description, repo.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert repository: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Helper function to encode permissions to JSON
func EncodePermissions(perms map[string]string) string {
	data, _ := json.Marshal(perms)
	return string(data)
}

// Helper function to decode permissions from JSON
func DecodePermissions(encoded string) map[string]string {
	var perms map[string]string
	json.Unmarshal([]byte(encoded), &perms)
	return perms
}

// Helper function to encode events to JSON
func EncodeEvents(events []string) string {
	data, _ := json.Marshal(events)
	return string(data)
}

// Helper function to decode events from JSON
func DecodeEvents(encoded string) []string {
	var events []string
	json.Unmarshal([]byte(encoded), &events)
	return events
}
