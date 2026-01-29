---
id: org-repository
name: Organization Repository Implementation
wave: 2
priority: 2
dependencies:
- org-schema-migration
estimated_hours: 4
tags:
- backend
- database
- repository
---

## Objective

Implement the Organization repository with CRUD operations following the existing repository pattern.

## Context

This task creates the data access layer for the Organization entity. It follows the repository pattern established in `backend/internal/database/repository/` and provides methods for creating, reading, updating, and deleting organizations.

## Implementation

1. Create `backend/internal/database/repository/organization.go`

2. Define the Organization struct matching the database schema:
   ```go
   type Organization struct {
       ID                         string
       WorkOSOrganizationID       *string    // Nullable external auth ID
       Name                       string
       StripeCustomerID           *string
       StripeSubscriptionID       *string
       SubscriptionTier           string     // FREE | PAID | ENTERPRISE
       SubscriptionStatus         string     // ACTIVE | CANCELED | PAST_DUE | INCOMPLETE
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
   ```

3. Implement OrganizationRepository struct:
   ```go
   type OrganizationRepository struct {
       db *sql.DB
   }

   func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
       return &OrganizationRepository{db: db}
   }
   ```

4. Implement CRUD methods:
   - `Create(name string) (*Organization, error)` - Create new organization
   - `GetByID(id string) (*Organization, error)` - Get by primary key
   - `GetByWorkOSID(workosOrgID string) (*Organization, error)` - Get by external ID
   - `GetByStripeCustomerID(customerID string) (*Organization, error)` - Get by Stripe ID
   - `Update(org *Organization) error` - Update all fields
   - `Delete(id string) error` - Hard delete
   - `List(limit, offset int) ([]*Organization, error)` - List with pagination

5. Implement subscription management methods:
   - `UpdateSubscription(id, tier, status string, cancelAtPeriodEnd bool) error`
   - `SetStripeIDs(id, customerID, subscriptionID string) error`

6. Implement usage tracking methods:
   - `IncrementTokenUsage(id string, microdollars int64) error`
   - `IncrementSandboxUsage(id string, seconds int64) error`
   - `ResetUsageCounters(id string) error`
   - `SetUsageLimits(id string, tokenLimit, sandboxLimit int64) error`
   - `SetBillingPeriod(id string, start, end time.Time) error`

7. Implement helper method for scanning rows:
   ```go
   func (r *OrganizationRepository) scanOrganization(row *sql.Row) (*Organization, error)
   func (r *OrganizationRepository) scanOrganizations(rows *sql.Rows) ([]*Organization, error)
   ```

## Acceptance Criteria

- [ ] Organization struct properly maps to database schema
- [ ] All CRUD operations implemented (Create, GetByID, Update, Delete, List)
- [ ] Lookup by WorkOS ID implemented
- [ ] Lookup by Stripe Customer ID implemented
- [ ] Subscription management methods implemented
- [ ] Usage tracking methods implemented with atomic updates
- [ ] Proper error handling with descriptive error messages
- [ ] sql.NullString/NullTime used for nullable fields
- [ ] UUID generation using google/uuid package
- [ ] UpdatedAt timestamp set on all updates
- [ ] Follows existing patterns (see workspace.go, user.go)

## Files to Create/Modify

- `backend/internal/database/repository/organization.go` - Create new repository file

## Integration Points

- **Provides**: Organization data access for handlers and services
- **Consumes**: Database schema (org-schema-migration)
- **Conflicts**: None - new file

## Technical Notes

- Reference `backend/internal/database/repository/workspace.go` for pattern examples
- Use `sql.NullString` for optional string fields
- Use `sql.NullTime` for optional datetime fields
- Boolean stored as INTEGER (0/1) in SQLite
- Use `uuid.New().String()` for ID generation
- Wrap errors with context: `fmt.Errorf("failed to create organization: %w", err)`
