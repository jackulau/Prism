package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Organization represents an organization in the database
type Organization struct {
	ID                         string
	WorkOSOrganizationID       *string
	Name                       string
	StripeCustomerID           *string
	StripeSubscriptionID       *string
	SubscriptionTier           string
	SubscriptionStatus         string
	CancelAtPeriodEnd          bool
	TokenCostUsedMicrodollars  int64
	TokenCostLimitMicrodollars int64
	SandboxTimeUsedSeconds     int64
	SandboxTimeLimitSeconds    int64
	BillingPeriodStart         *time.Time
	BillingPeriodEnd           *time.Time
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
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
func (r *OrganizationRepository) Create(name string) (*Organization, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO organizations (id, name, subscription_tier, subscription_status, created_at, updated_at)
		 VALUES (?, ?, 'FREE', 'ACTIVE', ?, ?)`,
		id, name, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return &Organization{
		ID:                 id,
		Name:               name,
		SubscriptionTier:   "FREE",
		SubscriptionStatus: "ACTIVE",
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(id string) (*Organization, error) {
	row := r.db.QueryRow(
		`SELECT id, workos_organization_id, name, stripe_customer_id, stripe_subscription_id,
		        subscription_tier, subscription_status, cancel_at_period_end,
		        token_cost_used_microdollars, token_cost_limit_microdollars,
		        sandbox_time_used_seconds, sandbox_time_limit_seconds,
		        billing_period_start, billing_period_end, created_at, updated_at
		 FROM organizations WHERE id = ?`,
		id,
	)
	return r.scanOrganization(row)
}

// GetByWorkOSID retrieves an organization by WorkOS organization ID
func (r *OrganizationRepository) GetByWorkOSID(workosOrgID string) (*Organization, error) {
	row := r.db.QueryRow(
		`SELECT id, workos_organization_id, name, stripe_customer_id, stripe_subscription_id,
		        subscription_tier, subscription_status, cancel_at_period_end,
		        token_cost_used_microdollars, token_cost_limit_microdollars,
		        sandbox_time_used_seconds, sandbox_time_limit_seconds,
		        billing_period_start, billing_period_end, created_at, updated_at
		 FROM organizations WHERE workos_organization_id = ?`,
		workosOrgID,
	)
	return r.scanOrganization(row)
}

// GetByStripeCustomerID retrieves an organization by Stripe customer ID
func (r *OrganizationRepository) GetByStripeCustomerID(customerID string) (*Organization, error) {
	row := r.db.QueryRow(
		`SELECT id, workos_organization_id, name, stripe_customer_id, stripe_subscription_id,
		        subscription_tier, subscription_status, cancel_at_period_end,
		        token_cost_used_microdollars, token_cost_limit_microdollars,
		        sandbox_time_used_seconds, sandbox_time_limit_seconds,
		        billing_period_start, billing_period_end, created_at, updated_at
		 FROM organizations WHERE stripe_customer_id = ?`,
		customerID,
	)
	return r.scanOrganization(row)
}

// Update updates an organization
func (r *OrganizationRepository) Update(org *Organization) error {
	now := time.Now()
	org.UpdatedAt = now

	_, err := r.db.Exec(
		`UPDATE organizations SET
			workos_organization_id = ?,
			name = ?,
			stripe_customer_id = ?,
			stripe_subscription_id = ?,
			subscription_tier = ?,
			subscription_status = ?,
			cancel_at_period_end = ?,
			token_cost_used_microdollars = ?,
			token_cost_limit_microdollars = ?,
			sandbox_time_used_seconds = ?,
			sandbox_time_limit_seconds = ?,
			billing_period_start = ?,
			billing_period_end = ?,
			updated_at = ?
		 WHERE id = ?`,
		org.WorkOSOrganizationID,
		org.Name,
		org.StripeCustomerID,
		org.StripeSubscriptionID,
		org.SubscriptionTier,
		org.SubscriptionStatus,
		org.CancelAtPeriodEnd,
		org.TokenCostUsedMicrodollars,
		org.TokenCostLimitMicrodollars,
		org.SandboxTimeUsedSeconds,
		org.SandboxTimeLimitSeconds,
		org.BillingPeriodStart,
		org.BillingPeriodEnd,
		now,
		org.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

// Delete deletes an organization by ID
func (r *OrganizationRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM organizations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}

// List lists organizations with pagination
func (r *OrganizationRepository) List(limit, offset int) ([]*Organization, error) {
	rows, err := r.db.Query(
		`SELECT id, workos_organization_id, name, stripe_customer_id, stripe_subscription_id,
		        subscription_tier, subscription_status, cancel_at_period_end,
		        token_cost_used_microdollars, token_cost_limit_microdollars,
		        sandbox_time_used_seconds, sandbox_time_limit_seconds,
		        billing_period_start, billing_period_end, created_at, updated_at
		 FROM organizations
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer rows.Close()

	return r.scanOrganizations(rows)
}

// UpdateSubscription updates subscription-related fields
func (r *OrganizationRepository) UpdateSubscription(id, tier, status string, cancelAtPeriodEnd bool) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET
			subscription_tier = ?,
			subscription_status = ?,
			cancel_at_period_end = ?,
			updated_at = ?
		 WHERE id = ?`,
		tier, status, cancelAtPeriodEnd, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	return nil
}

// SetStripeIDs sets the Stripe customer and subscription IDs
func (r *OrganizationRepository) SetStripeIDs(id, customerID, subscriptionID string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET
			stripe_customer_id = ?,
			stripe_subscription_id = ?,
			updated_at = ?
		 WHERE id = ?`,
		customerID, subscriptionID, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to set stripe IDs: %w", err)
	}
	return nil
}

// IncrementTokenUsage atomically increments token usage
func (r *OrganizationRepository) IncrementTokenUsage(id string, microdollars int64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET
			token_cost_used_microdollars = token_cost_used_microdollars + ?,
			updated_at = ?
		 WHERE id = ?`,
		microdollars, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to increment token usage: %w", err)
	}
	return nil
}

// IncrementSandboxUsage atomically increments sandbox time usage
func (r *OrganizationRepository) IncrementSandboxUsage(id string, seconds int64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET
			sandbox_time_used_seconds = sandbox_time_used_seconds + ?,
			updated_at = ?
		 WHERE id = ?`,
		seconds, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to increment sandbox usage: %w", err)
	}
	return nil
}

// ResetUsageCounters resets all usage counters to zero
func (r *OrganizationRepository) ResetUsageCounters(id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET
			token_cost_used_microdollars = 0,
			sandbox_time_used_seconds = 0,
			updated_at = ?
		 WHERE id = ?`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to reset usage counters: %w", err)
	}
	return nil
}

// SetUsageLimits sets the usage limits for the organization
func (r *OrganizationRepository) SetUsageLimits(id string, tokenLimit, sandboxLimit int64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET
			token_cost_limit_microdollars = ?,
			sandbox_time_limit_seconds = ?,
			updated_at = ?
		 WHERE id = ?`,
		tokenLimit, sandboxLimit, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to set usage limits: %w", err)
	}
	return nil
}

// SetBillingPeriod sets the billing period dates
func (r *OrganizationRepository) SetBillingPeriod(id string, start, end time.Time) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET
			billing_period_start = ?,
			billing_period_end = ?,
			updated_at = ?
		 WHERE id = ?`,
		start, end, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to set billing period: %w", err)
	}
	return nil
}

// scanOrganization scans a single row into an Organization struct
func (r *OrganizationRepository) scanOrganization(row *sql.Row) (*Organization, error) {
	org := &Organization{}
	var workosOrgID, stripeCustomerID, stripeSubscriptionID sql.NullString
	var billingStart, billingEnd sql.NullTime
	var cancelAtPeriodEnd int

	err := row.Scan(
		&org.ID,
		&workosOrgID,
		&org.Name,
		&stripeCustomerID,
		&stripeSubscriptionID,
		&org.SubscriptionTier,
		&org.SubscriptionStatus,
		&cancelAtPeriodEnd,
		&org.TokenCostUsedMicrodollars,
		&org.TokenCostLimitMicrodollars,
		&org.SandboxTimeUsedSeconds,
		&org.SandboxTimeLimitSeconds,
		&billingStart,
		&billingEnd,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan organization: %w", err)
	}

	if workosOrgID.Valid {
		org.WorkOSOrganizationID = &workosOrgID.String
	}
	if stripeCustomerID.Valid {
		org.StripeCustomerID = &stripeCustomerID.String
	}
	if stripeSubscriptionID.Valid {
		org.StripeSubscriptionID = &stripeSubscriptionID.String
	}
	if billingStart.Valid {
		org.BillingPeriodStart = &billingStart.Time
	}
	if billingEnd.Valid {
		org.BillingPeriodEnd = &billingEnd.Time
	}
	org.CancelAtPeriodEnd = cancelAtPeriodEnd != 0

	return org, nil
}

// scanOrganizations scans multiple rows into Organization structs
func (r *OrganizationRepository) scanOrganizations(rows *sql.Rows) ([]*Organization, error) {
	var organizations []*Organization

	for rows.Next() {
		org := &Organization{}
		var workosOrgID, stripeCustomerID, stripeSubscriptionID sql.NullString
		var billingStart, billingEnd sql.NullTime
		var cancelAtPeriodEnd int

		err := rows.Scan(
			&org.ID,
			&workosOrgID,
			&org.Name,
			&stripeCustomerID,
			&stripeSubscriptionID,
			&org.SubscriptionTier,
			&org.SubscriptionStatus,
			&cancelAtPeriodEnd,
			&org.TokenCostUsedMicrodollars,
			&org.TokenCostLimitMicrodollars,
			&org.SandboxTimeUsedSeconds,
			&org.SandboxTimeLimitSeconds,
			&billingStart,
			&billingEnd,
			&org.CreatedAt,
			&org.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}

		if workosOrgID.Valid {
			org.WorkOSOrganizationID = &workosOrgID.String
		}
		if stripeCustomerID.Valid {
			org.StripeCustomerID = &stripeCustomerID.String
		}
		if stripeSubscriptionID.Valid {
			org.StripeSubscriptionID = &stripeSubscriptionID.String
		}
		if billingStart.Valid {
			org.BillingPeriodStart = &billingStart.Time
		}
		if billingEnd.Valid {
			org.BillingPeriodEnd = &billingEnd.Time
		}
		org.CancelAtPeriodEnd = cancelAtPeriodEnd != 0

		organizations = append(organizations, org)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating organizations: %w", err)
	}

	return organizations, nil
}
