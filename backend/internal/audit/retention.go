package audit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RetentionPolicy defines how long audit logs should be retained
type RetentionPolicy struct {
	ID              string     `json:"id"`
	OrgID           string     `json:"organization_id,omitempty"`
	Name            string     `json:"name"`
	RetentionDays   int        `json:"retention_days"`
	ResourceTypes   []string   `json:"resource_types,omitempty"` // Empty means all
	ActionTypes     []string   `json:"action_types,omitempty"`   // Empty means all
	Enabled         bool       `json:"enabled"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastExecutedAt  *time.Time `json:"last_executed_at,omitempty"`
}

// LegalHold represents a legal hold that prevents deletion
type LegalHold struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"organization_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	Active      bool      `json:"active"`
}

// RetentionRepository is the interface for managing retention policies
type RetentionRepository interface {
	// Retention policies
	CreatePolicy(policy *RetentionPolicy) error
	UpdatePolicy(policy *RetentionPolicy) error
	GetPolicy(id string) (*RetentionPolicy, error)
	ListPolicies(orgID string) ([]*RetentionPolicy, error)
	DeletePolicy(id string) error

	// Legal holds
	CreateLegalHold(hold *LegalHold) error
	UpdateLegalHold(hold *LegalHold) error
	GetLegalHold(id string) (*LegalHold, error)
	ListLegalHolds(orgID string, activeOnly bool) ([]*LegalHold, error)
	DeleteLegalHold(id string) error
}

// AuditDeleter is the interface for deleting audit logs
type AuditDeleter interface {
	DeleteBefore(timestamp time.Time, excludeLegalHold bool) (int64, error)
	SetLegalHoldByDateRange(start, end time.Time, hold bool, orgID string) (int64, error)
}

// RetentionManager handles audit log retention and legal holds
type RetentionManager struct {
	policyRepo  RetentionRepository
	auditRepo   AuditDeleter
	logger      *Logger

	mu            sync.Mutex
	stopCh        chan struct{}
	cleanupTicker *time.Ticker
}

// NewRetentionManager creates a new retention manager
func NewRetentionManager(policyRepo RetentionRepository, auditRepo AuditDeleter, logger *Logger) *RetentionManager {
	return &RetentionManager{
		policyRepo: policyRepo,
		auditRepo:  auditRepo,
		logger:     logger,
	}
}

// Start begins the background retention enforcement
func (m *RetentionManager) Start(ctx context.Context, checkInterval time.Duration) {
	m.mu.Lock()
	if m.cleanupTicker != nil {
		m.mu.Unlock()
		return
	}
	m.stopCh = make(chan struct{})
	m.cleanupTicker = time.NewTicker(checkInterval)
	m.mu.Unlock()

	go func() {
		// Run immediately on start
		m.enforceRetention(ctx)

		for {
			select {
			case <-m.cleanupTicker.C:
				m.enforceRetention(ctx)
			case <-m.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the background retention enforcement
func (m *RetentionManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
		close(m.stopCh)
		m.cleanupTicker = nil
	}
}

// enforceRetention applies all active retention policies
func (m *RetentionManager) enforceRetention(ctx context.Context) {
	// Get all active policies
	policies, err := m.policyRepo.ListPolicies("")
	if err != nil {
		return
	}

	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		m.enforcePolicy(policy)
	}
}

// enforcePolicy applies a single retention policy
func (m *RetentionManager) enforcePolicy(policy *RetentionPolicy) {
	cutoff := time.Now().AddDate(0, 0, -policy.RetentionDays)

	// Delete old audit logs, excluding those with legal holds
	deleted, err := m.auditRepo.DeleteBefore(cutoff, true)
	if err != nil {
		// Log error but continue
		if m.logger != nil {
			m.logger.LogSystemAction(ActionDelete, ResourceAuditLog, policy.ID,
				WithError(err),
				WithMetadata(map[string]interface{}{
					"policy_name": policy.Name,
					"cutoff":      cutoff,
				}),
			)
		}
		return
	}

	// Update last executed
	now := time.Now()
	policy.LastExecutedAt = &now
	_ = m.policyRepo.UpdatePolicy(policy)

	// Log the cleanup action
	if m.logger != nil && deleted > 0 {
		m.logger.LogSystemAction(ActionDelete, ResourceAuditLog, policy.ID,
			WithMetadata(map[string]interface{}{
				"policy_name":    policy.Name,
				"deleted_count":  deleted,
				"cutoff_date":    cutoff,
				"retention_days": policy.RetentionDays,
			}),
		)
	}
}

// CreatePolicy creates a new retention policy
func (m *RetentionManager) CreatePolicy(policy *RetentionPolicy) error {
	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}
	policy.CreatedAt = time.Now().UTC()
	policy.UpdatedAt = policy.CreatedAt

	if policy.RetentionDays < 1 {
		return fmt.Errorf("retention days must be at least 1")
	}

	return m.policyRepo.CreatePolicy(policy)
}

// UpdatePolicy updates an existing retention policy
func (m *RetentionManager) UpdatePolicy(policy *RetentionPolicy) error {
	policy.UpdatedAt = time.Now().UTC()
	return m.policyRepo.UpdatePolicy(policy)
}

// GetPolicy retrieves a retention policy by ID
func (m *RetentionManager) GetPolicy(id string) (*RetentionPolicy, error) {
	return m.policyRepo.GetPolicy(id)
}

// ListPolicies returns all retention policies for an organization
func (m *RetentionManager) ListPolicies(orgID string) ([]*RetentionPolicy, error) {
	return m.policyRepo.ListPolicies(orgID)
}

// DeletePolicy removes a retention policy
func (m *RetentionManager) DeletePolicy(id string) error {
	return m.policyRepo.DeletePolicy(id)
}

// CreateLegalHold creates a legal hold that prevents deletion
func (m *RetentionManager) CreateLegalHold(hold *LegalHold) error {
	if hold.ID == "" {
		hold.ID = uuid.New().String()
	}
	hold.CreatedAt = time.Now().UTC()
	hold.Active = true

	// Apply the legal hold to existing audit logs in the date range
	affected, err := m.auditRepo.SetLegalHoldByDateRange(hold.StartDate, hold.EndDate, true, hold.OrgID)
	if err != nil {
		return fmt.Errorf("failed to apply legal hold: %w", err)
	}

	if err := m.policyRepo.CreateLegalHold(hold); err != nil {
		// Rollback the legal hold flags
		_, _ = m.auditRepo.SetLegalHoldByDateRange(hold.StartDate, hold.EndDate, false, hold.OrgID)
		return err
	}

	// Log the action
	if m.logger != nil {
		m.logger.LogUserAction(hold.CreatedBy, "", ActionCreate, ResourceAuditLog, hold.ID,
			WithMetadata(map[string]interface{}{
				"hold_name":       hold.Name,
				"start_date":      hold.StartDate,
				"end_date":        hold.EndDate,
				"affected_logs":   affected,
			}),
		)
	}

	return nil
}

// ReleaseLegalHold releases a legal hold
func (m *RetentionManager) ReleaseLegalHold(id string, releasedBy string) error {
	hold, err := m.policyRepo.GetLegalHold(id)
	if err != nil {
		return err
	}
	if hold == nil {
		return fmt.Errorf("legal hold not found")
	}

	// Remove the legal hold flags
	affected, err := m.auditRepo.SetLegalHoldByDateRange(hold.StartDate, hold.EndDate, false, hold.OrgID)
	if err != nil {
		return fmt.Errorf("failed to release legal hold: %w", err)
	}

	hold.Active = false
	if err := m.policyRepo.UpdateLegalHold(hold); err != nil {
		return err
	}

	// Log the action
	if m.logger != nil {
		m.logger.LogUserAction(releasedBy, "", ActionUpdate, ResourceAuditLog, hold.ID,
			WithMetadata(map[string]interface{}{
				"hold_name":       hold.Name,
				"action":          "released",
				"affected_logs":   affected,
			}),
		)
	}

	return nil
}

// GetLegalHold retrieves a legal hold by ID
func (m *RetentionManager) GetLegalHold(id string) (*LegalHold, error) {
	return m.policyRepo.GetLegalHold(id)
}

// ListLegalHolds returns all legal holds for an organization
func (m *RetentionManager) ListLegalHolds(orgID string, activeOnly bool) ([]*LegalHold, error) {
	return m.policyRepo.ListLegalHolds(orgID, activeOnly)
}

// DeleteLegalHold removes a legal hold (only if inactive)
func (m *RetentionManager) DeleteLegalHold(id string) error {
	hold, err := m.policyRepo.GetLegalHold(id)
	if err != nil {
		return err
	}
	if hold == nil {
		return fmt.Errorf("legal hold not found")
	}
	if hold.Active {
		return fmt.Errorf("cannot delete active legal hold - release it first")
	}

	return m.policyRepo.DeleteLegalHold(id)
}

// DefaultRetentionPolicy returns a sensible default retention policy
func DefaultRetentionPolicy() *RetentionPolicy {
	return &RetentionPolicy{
		ID:            uuid.New().String(),
		Name:          "Default Retention Policy",
		RetentionDays: 365, // 1 year
		Enabled:       true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
}
