---
id: org-schema-migration
name: Organization Database Schema Migration
wave: 1
priority: 1
dependencies: []
estimated_hours: 2
tags:
- backend
- database
- schema
---

## Objective

Add the Organization entity table and schema migration to the SQLite database.

## Context

The Organization entity is the foundation for multi-tenant organization management. This task creates the database schema following existing patterns in `backend/internal/database/sqlite.go`. All other organization-related tasks depend on this schema being in place.

## Implementation

1. Add the organizations table to `backend/internal/database/sqlite.go` in the `Migrate()` function

2. Schema definition based on the feature spec:
   ```sql
   CREATE TABLE IF NOT EXISTS organizations (
       id TEXT PRIMARY KEY,
       workos_organization_id TEXT UNIQUE,
       name TEXT NOT NULL,
       stripe_customer_id TEXT,
       stripe_subscription_id TEXT,
       subscription_tier TEXT NOT NULL DEFAULT 'FREE' CHECK(subscription_tier IN ('FREE', 'PAID', 'ENTERPRISE')),
       subscription_status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(subscription_status IN ('ACTIVE', 'CANCELED', 'PAST_DUE', 'INCOMPLETE')),
       cancel_at_period_end INTEGER NOT NULL DEFAULT 0,
       token_cost_used_microdollars INTEGER NOT NULL DEFAULT 0,
       token_cost_limit_microdollars INTEGER NOT NULL DEFAULT 0,
       sandbox_time_used_seconds INTEGER NOT NULL DEFAULT 0,
       sandbox_time_limit_seconds INTEGER NOT NULL DEFAULT 0,
       billing_period_start DATETIME,
       billing_period_end DATETIME,
       created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
       updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
   );
   ```

3. Add indexes for frequently queried fields:
   ```sql
   CREATE INDEX IF NOT EXISTS idx_organizations_workos_org_id ON organizations(workos_organization_id);
   CREATE INDEX IF NOT EXISTS idx_organizations_stripe_customer_id ON organizations(stripe_customer_id);
   ```

4. Follow the existing migration pattern - append to the `Migrate()` function after existing table creations

## Acceptance Criteria

- [ ] Organizations table is created with all specified fields
- [ ] Field constraints are properly defined (CHECK constraints for enums)
- [ ] Indexes are created for workos_organization_id and stripe_customer_id
- [ ] Default values are set correctly (FREE tier, ACTIVE status, 0 for counters)
- [ ] cancel_at_period_end uses INTEGER (SQLite boolean pattern)
- [ ] Migration is idempotent (IF NOT EXISTS)
- [ ] Application starts successfully after migration

## Files to Create/Modify

- `backend/internal/database/sqlite.go` - Add organizations table and indexes to Migrate()

## Integration Points

- **Provides**: Organizations table schema for repository layer
- **Consumes**: None (foundational task)
- **Conflicts**: None - isolated schema addition

## Technical Notes

- Follow existing patterns in sqlite.go (lines 50-386)
- Use TEXT for UUIDs (consistent with other tables)
- Use INTEGER for boolean flags (SQLite convention)
- Use INTEGER for monetary values (microdollars = millionths of a dollar)
- Use DATETIME with DEFAULT CURRENT_TIMESTAMP for timestamps
