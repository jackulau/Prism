package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/audit"
)

// RetentionRepository handles retention policy database operations
type RetentionRepository struct {
	db *sql.DB
}

// NewRetentionRepository creates a new retention repository
func NewRetentionRepository(db *sql.DB) *RetentionRepository {
	return &RetentionRepository{db: db}
}

// CreatePolicy stores a new retention policy
func (r *RetentionRepository) CreatePolicy(policy *audit.RetentionPolicy) error {
	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}

	resourceTypesJSON := "[]"
	if len(policy.ResourceTypes) > 0 {
		if data, err := json.Marshal(policy.ResourceTypes); err == nil {
			resourceTypesJSON = string(data)
		}
	}

	actionTypesJSON := "[]"
	if len(policy.ActionTypes) > 0 {
		if data, err := json.Marshal(policy.ActionTypes); err == nil {
			actionTypesJSON = string(data)
		}
	}

	_, err := r.db.Exec(`
		INSERT INTO data_retention_policies (
			id, organization_id, name, retention_days, resource_types,
			action_types, enabled, created_at, updated_at, last_executed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		policy.ID, policy.OrgID, policy.Name, policy.RetentionDays,
		resourceTypesJSON, actionTypesJSON, policy.Enabled,
		policy.CreatedAt, policy.UpdatedAt, policy.LastExecutedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create retention policy: %w", err)
	}

	return nil
}

// UpdatePolicy updates an existing retention policy
func (r *RetentionRepository) UpdatePolicy(policy *audit.RetentionPolicy) error {
	resourceTypesJSON := "[]"
	if len(policy.ResourceTypes) > 0 {
		if data, err := json.Marshal(policy.ResourceTypes); err == nil {
			resourceTypesJSON = string(data)
		}
	}

	actionTypesJSON := "[]"
	if len(policy.ActionTypes) > 0 {
		if data, err := json.Marshal(policy.ActionTypes); err == nil {
			actionTypesJSON = string(data)
		}
	}

	_, err := r.db.Exec(`
		UPDATE data_retention_policies SET
			name = ?, retention_days = ?, resource_types = ?,
			action_types = ?, enabled = ?, updated_at = ?, last_executed_at = ?
		WHERE id = ?`,
		policy.Name, policy.RetentionDays, resourceTypesJSON,
		actionTypesJSON, policy.Enabled, policy.UpdatedAt, policy.LastExecutedAt,
		policy.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update retention policy: %w", err)
	}

	return nil
}

// GetPolicy retrieves a retention policy by ID
func (r *RetentionRepository) GetPolicy(id string) (*audit.RetentionPolicy, error) {
	var policy audit.RetentionPolicy
	var orgID, resourceTypesJSON, actionTypesJSON sql.NullString
	var lastExecutedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, organization_id, name, retention_days, resource_types,
			   action_types, enabled, created_at, updated_at, last_executed_at
		FROM data_retention_policies WHERE id = ?`, id,
	).Scan(
		&policy.ID, &orgID, &policy.Name, &policy.RetentionDays,
		&resourceTypesJSON, &actionTypesJSON, &policy.Enabled,
		&policy.CreatedAt, &policy.UpdatedAt, &lastExecutedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get retention policy: %w", err)
	}

	policy.OrgID = orgID.String

	if resourceTypesJSON.Valid && resourceTypesJSON.String != "" {
		var types []string
		if err := json.Unmarshal([]byte(resourceTypesJSON.String), &types); err == nil {
			policy.ResourceTypes = types
		}
	}

	if actionTypesJSON.Valid && actionTypesJSON.String != "" {
		var types []string
		if err := json.Unmarshal([]byte(actionTypesJSON.String), &types); err == nil {
			policy.ActionTypes = types
		}
	}

	if lastExecutedAt.Valid {
		policy.LastExecutedAt = &lastExecutedAt.Time
	}

	return &policy, nil
}

// ListPolicies retrieves all retention policies for an organization
func (r *RetentionRepository) ListPolicies(orgID string) ([]*audit.RetentionPolicy, error) {
	var rows *sql.Rows
	var err error

	if orgID != "" {
		rows, err = r.db.Query(`
			SELECT id, organization_id, name, retention_days, resource_types,
				   action_types, enabled, created_at, updated_at, last_executed_at
			FROM data_retention_policies
			WHERE organization_id = ? OR organization_id IS NULL
			ORDER BY created_at DESC`, orgID,
		)
	} else {
		rows, err = r.db.Query(`
			SELECT id, organization_id, name, retention_days, resource_types,
				   action_types, enabled, created_at, updated_at, last_executed_at
			FROM data_retention_policies
			ORDER BY created_at DESC`,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query retention policies: %w", err)
	}
	defer rows.Close()

	var policies []*audit.RetentionPolicy
	for rows.Next() {
		var policy audit.RetentionPolicy
		var orgIDVal, resourceTypesJSON, actionTypesJSON sql.NullString
		var lastExecutedAt sql.NullTime

		err := rows.Scan(
			&policy.ID, &orgIDVal, &policy.Name, &policy.RetentionDays,
			&resourceTypesJSON, &actionTypesJSON, &policy.Enabled,
			&policy.CreatedAt, &policy.UpdatedAt, &lastExecutedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retention policy: %w", err)
		}

		policy.OrgID = orgIDVal.String

		if resourceTypesJSON.Valid && resourceTypesJSON.String != "" {
			var types []string
			if err := json.Unmarshal([]byte(resourceTypesJSON.String), &types); err == nil {
				policy.ResourceTypes = types
			}
		}

		if actionTypesJSON.Valid && actionTypesJSON.String != "" {
			var types []string
			if err := json.Unmarshal([]byte(actionTypesJSON.String), &types); err == nil {
				policy.ActionTypes = types
			}
		}

		if lastExecutedAt.Valid {
			policy.LastExecutedAt = &lastExecutedAt.Time
		}

		policies = append(policies, &policy)
	}

	return policies, nil
}

// DeletePolicy removes a retention policy
func (r *RetentionRepository) DeletePolicy(id string) error {
	_, err := r.db.Exec(`DELETE FROM data_retention_policies WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete retention policy: %w", err)
	}
	return nil
}

// CreateLegalHold stores a new legal hold
func (r *RetentionRepository) CreateLegalHold(hold *audit.LegalHold) error {
	if hold.ID == "" {
		hold.ID = uuid.New().String()
	}

	_, err := r.db.Exec(`
		INSERT INTO legal_holds (
			id, organization_id, name, description, start_date, end_date,
			created_by, created_at, active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		hold.ID, hold.OrgID, hold.Name, hold.Description, hold.StartDate,
		hold.EndDate, hold.CreatedBy, hold.CreatedAt, hold.Active,
	)
	if err != nil {
		return fmt.Errorf("failed to create legal hold: %w", err)
	}

	return nil
}

// UpdateLegalHold updates an existing legal hold
func (r *RetentionRepository) UpdateLegalHold(hold *audit.LegalHold) error {
	_, err := r.db.Exec(`
		UPDATE legal_holds SET
			name = ?, description = ?, start_date = ?, end_date = ?, active = ?
		WHERE id = ?`,
		hold.Name, hold.Description, hold.StartDate, hold.EndDate, hold.Active,
		hold.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update legal hold: %w", err)
	}

	return nil
}

// GetLegalHold retrieves a legal hold by ID
func (r *RetentionRepository) GetLegalHold(id string) (*audit.LegalHold, error) {
	var hold audit.LegalHold
	var orgID, description sql.NullString

	err := r.db.QueryRow(`
		SELECT id, organization_id, name, description, start_date, end_date,
			   created_by, created_at, active
		FROM legal_holds WHERE id = ?`, id,
	).Scan(
		&hold.ID, &orgID, &hold.Name, &description, &hold.StartDate,
		&hold.EndDate, &hold.CreatedBy, &hold.CreatedAt, &hold.Active,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get legal hold: %w", err)
	}

	hold.OrgID = orgID.String
	hold.Description = description.String

	return &hold, nil
}

// ListLegalHolds retrieves all legal holds for an organization
func (r *RetentionRepository) ListLegalHolds(orgID string, activeOnly bool) ([]*audit.LegalHold, error) {
	var rows *sql.Rows
	var err error

	query := `
		SELECT id, organization_id, name, description, start_date, end_date,
			   created_by, created_at, active
		FROM legal_holds`

	var conditions []string
	var args []interface{}

	if orgID != "" {
		conditions = append(conditions, "organization_id = ?")
		args = append(args, orgID)
	}

	if activeOnly {
		conditions = append(conditions, "active = 1")
	}

	if len(conditions) > 0 {
		query += " WHERE "
		for i, cond := range conditions {
			if i > 0 {
				query += " AND "
			}
			query += cond
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err = r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query legal holds: %w", err)
	}
	defer rows.Close()

	var holds []*audit.LegalHold
	for rows.Next() {
		var hold audit.LegalHold
		var orgIDVal, description sql.NullString

		err := rows.Scan(
			&hold.ID, &orgIDVal, &hold.Name, &description, &hold.StartDate,
			&hold.EndDate, &hold.CreatedBy, &hold.CreatedAt, &hold.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan legal hold: %w", err)
		}

		hold.OrgID = orgIDVal.String
		hold.Description = description.String

		holds = append(holds, &hold)
	}

	return holds, nil
}

// DeleteLegalHold removes a legal hold
func (r *RetentionRepository) DeleteLegalHold(id string) error {
	_, err := r.db.Exec(`DELETE FROM legal_holds WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete legal hold: %w", err)
	}
	return nil
}
