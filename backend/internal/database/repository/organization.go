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

// scanOrganization scans a single row into an Organization
func (r *OrganizationRepository) scanOrganization(row *sql.Row) (*Organization, error) {
	org := &Organization{}
	var workosOrgID, stripeCustomerID, stripeSubscriptionID sql.NullString
	var billingPeriodStart, billingPeriodEnd sql.NullTime

	err := row.Scan(
		&org.ID,
		&workosOrgID,
		&org.Name,
		&stripeCustomerID,
		&stripeSubscriptionID,
		&org.SubscriptionTier,
		&org.SubscriptionStatus,
		&org.CancelAtPeriodEnd,
		&org.TokenCostUsedMicrodollars,
		&org.TokenCostLimitMicrodollars,
		&org.SandboxTimeUsedSeconds,
		&org.SandboxTimeLimitSeconds,
		&billingPeriodStart,
		&billingPeriodEnd,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
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
	if billingPeriodStart.Valid {
		org.BillingPeriodStart = &billingPeriodStart.Time
	}
	if billingPeriodEnd.Valid {
		org.BillingPeriodEnd = &billingPeriodEnd.Time
	}

	return org, nil
}

// scanOrganizations scans multiple rows into Organizations
func (r *OrganizationRepository) scanOrganizations(rows *sql.Rows) ([]*Organization, error) {
	var organizations []*Organization

	for rows.Next() {
		org := &Organization{}
		var workosOrgID, stripeCustomerID, stripeSubscriptionID sql.NullString
		var billingPeriodStart, billingPeriodEnd sql.NullTime

		err := rows.Scan(
			&org.ID,
			&workosOrgID,
			&org.Name,
			&stripeCustomerID,
			&stripeSubscriptionID,
			&org.SubscriptionTier,
			&org.SubscriptionStatus,
			&org.CancelAtPeriodEnd,
			&org.TokenCostUsedMicrodollars,
			&org.TokenCostLimitMicrodollars,
			&org.SandboxTimeUsedSeconds,
			&org.SandboxTimeLimitSeconds,
			&billingPeriodStart,
			&billingPeriodEnd,
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
		if billingPeriodStart.Valid {
			org.BillingPeriodStart = &billingPeriodStart.Time
		}
		if billingPeriodEnd.Valid {
			org.BillingPeriodEnd = &billingPeriodEnd.Time
		}

		organizations = append(organizations, org)
	}

	return organizations, nil
}

// Create creates a new organization
func (r *OrganizationRepository) Create(name string) (*Organization, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO organizations (id, name, subscription_tier, subscription_status, cancel_at_period_end,
		 token_cost_used_microdollars, token_cost_limit_microdollars, sandbox_time_used_seconds,
		 sandbox_time_limit_seconds, created_at, updated_at)
		 VALUES (?, ?, 'FREE', 'ACTIVE', 0, 0, 0, 0, 0, ?, ?)`,
		id, name, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return &Organization{
		ID:                         id,
		Name:                       name,
		SubscriptionTier:           "FREE",
		SubscriptionStatus:         "ACTIVE",
		CancelAtPeriodEnd:          false,
		TokenCostUsedMicrodollars:  0,
		TokenCostLimitMicrodollars: 0,
		SandboxTimeUsedSeconds:     0,
		SandboxTimeLimitSeconds:    0,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}, nil
}

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(id string) (*Organization, error) {
	row := r.db.QueryRow(
		`SELECT id, workos_organization_id, name, stripe_customer_id, stripe_subscription_id,
		 subscription_tier, subscription_status, cancel_at_period_end, token_cost_used_microdollars,
		 token_cost_limit_microdollars, sandbox_time_used_seconds, sandbox_time_limit_seconds,
		 billing_period_start, billing_period_end, created_at, updated_at
		 FROM organizations WHERE id = ?`,
		id,
	)

	org, err := r.scanOrganization(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return org, nil
}

// GetByWorkOSID retrieves an organization by WorkOS organization ID
func (r *OrganizationRepository) GetByWorkOSID(workosOrgID string) (*Organization, error) {
	row := r.db.QueryRow(
		`SELECT id, workos_organization_id, name, stripe_customer_id, stripe_subscription_id,
		 subscription_tier, subscription_status, cancel_at_period_end, token_cost_used_microdollars,
		 token_cost_limit_microdollars, sandbox_time_used_seconds, sandbox_time_limit_seconds,
		 billing_period_start, billing_period_end, created_at, updated_at
		 FROM organizations WHERE workos_organization_id = ?`,
		workosOrgID,
	)

	org, err := r.scanOrganization(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization by WorkOS ID: %w", err)
	}
	return org, nil
}

// GetByStripeCustomerID retrieves an organization by Stripe customer ID
func (r *OrganizationRepository) GetByStripeCustomerID(customerID string) (*Organization, error) {
	row := r.db.QueryRow(
		`SELECT id, workos_organization_id, name, stripe_customer_id, stripe_subscription_id,
		 subscription_tier, subscription_status, cancel_at_period_end, token_cost_used_microdollars,
		 token_cost_limit_microdollars, sandbox_time_used_seconds, sandbox_time_limit_seconds,
		 billing_period_start, billing_period_end, created_at, updated_at
		 FROM organizations WHERE stripe_customer_id = ?`,
		customerID,
	)

	org, err := r.scanOrganization(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization by Stripe customer ID: %w", err)
	}
	return org, nil
}

// Update updates all fields of an organization
func (r *OrganizationRepository) Update(org *Organization) error {
	now := time.Now()
	org.UpdatedAt = now

	_, err := r.db.Exec(
		`UPDATE organizations SET
		 workos_organization_id = ?, name = ?, stripe_customer_id = ?, stripe_subscription_id = ?,
		 subscription_tier = ?, subscription_status = ?, cancel_at_period_end = ?,
		 token_cost_used_microdollars = ?, token_cost_limit_microdollars = ?,
		 sandbox_time_used_seconds = ?, sandbox_time_limit_seconds = ?,
		 billing_period_start = ?, billing_period_end = ?, updated_at = ?
		 WHERE id = ?`,
		org.WorkOSOrganizationID, org.Name, org.StripeCustomerID, org.StripeSubscriptionID,
		org.SubscriptionTier, org.SubscriptionStatus, org.CancelAtPeriodEnd,
		org.TokenCostUsedMicrodollars, org.TokenCostLimitMicrodollars,
		org.SandboxTimeUsedSeconds, org.SandboxTimeLimitSeconds,
		org.BillingPeriodStart, org.BillingPeriodEnd, now, org.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
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

// List retrieves organizations with pagination
func (r *OrganizationRepository) List(limit, offset int) ([]*Organization, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.Query(
		`SELECT id, workos_organization_id, name, stripe_customer_id, stripe_subscription_id,
		 subscription_tier, subscription_status, cancel_at_period_end, token_cost_used_microdollars,
		 token_cost_limit_microdollars, sandbox_time_used_seconds, sandbox_time_limit_seconds,
		 billing_period_start, billing_period_end, created_at, updated_at
		 FROM organizations ORDER BY created_at DESC LIMIT ? OFFSET ?`,
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
		`UPDATE organizations SET subscription_tier = ?, subscription_status = ?,
		 cancel_at_period_end = ?, updated_at = ? WHERE id = ?`,
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
		`UPDATE organizations SET stripe_customer_id = ?, stripe_subscription_id = ?, updated_at = ? WHERE id = ?`,
		customerID, subscriptionID, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to set Stripe IDs: %w", err)
	}
	return nil
}

// IncrementTokenUsage atomically increments the token usage counter
func (r *OrganizationRepository) IncrementTokenUsage(id string, microdollars int64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET token_cost_used_microdollars = token_cost_used_microdollars + ?,
		 updated_at = ? WHERE id = ?`,
		microdollars, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to increment token usage: %w", err)
	}
	return nil
}

// IncrementSandboxUsage atomically increments the sandbox usage counter
func (r *OrganizationRepository) IncrementSandboxUsage(id string, seconds int64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET sandbox_time_used_seconds = sandbox_time_used_seconds + ?,
		 updated_at = ? WHERE id = ?`,
		seconds, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to increment sandbox usage: %w", err)
	}
	return nil
}

// ResetUsageCounters resets the token and sandbox usage counters to zero
func (r *OrganizationRepository) ResetUsageCounters(id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET token_cost_used_microdollars = 0, sandbox_time_used_seconds = 0,
		 updated_at = ? WHERE id = ?`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to reset usage counters: %w", err)
	}
	return nil
}

// SetUsageLimits sets the token and sandbox usage limits
func (r *OrganizationRepository) SetUsageLimits(id string, tokenLimit, sandboxLimit int64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET token_cost_limit_microdollars = ?, sandbox_time_limit_seconds = ?,
		 updated_at = ? WHERE id = ?`,
		tokenLimit, sandboxLimit, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to set usage limits: %w", err)
	}
	return nil
}

// SetBillingPeriod sets the billing period start and end times
func (r *OrganizationRepository) SetBillingPeriod(id string, start, end time.Time) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE organizations SET billing_period_start = ?, billing_period_end = ?, updated_at = ? WHERE id = ?`,
		start, end, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to set billing period: %w", err)
	}
	return nil
}
