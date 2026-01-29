---
id: org-api-handlers
name: Organization API Handlers and Routes
wave: 3
priority: 2
dependencies:
- org-repository
estimated_hours: 4
tags:
- backend
- api
- handlers
---

## Objective

Create HTTP API handlers and routes for Organization CRUD operations and management.

## Context

This task exposes the Organization entity through the REST API. It follows the handler patterns established in `backend/internal/api/handlers/` and integrates with the router in `backend/internal/api/routes/router.go`.

## Implementation

1. Create `backend/internal/api/handlers/organization.go`

2. Define request/response DTOs:
   ```go
   type CreateOrganizationRequest struct {
       Name string `json:"name" validate:"required,min=1,max=255"`
   }

   type UpdateOrganizationRequest struct {
       Name                       *string `json:"name,omitempty"`
       SubscriptionTier           *string `json:"subscription_tier,omitempty"`
       SubscriptionStatus         *string `json:"subscription_status,omitempty"`
       CancelAtPeriodEnd          *bool   `json:"cancel_at_period_end,omitempty"`
       TokenCostLimitMicrodollars *int64  `json:"token_cost_limit_microdollars,omitempty"`
       SandboxTimeLimitSeconds    *int64  `json:"sandbox_time_limit_seconds,omitempty"`
   }

   type OrganizationResponse struct {
       ID                         string     `json:"id"`
       WorkOSOrganizationID       *string    `json:"workos_organization_id,omitempty"`
       Name                       string     `json:"name"`
       StripeCustomerID           *string    `json:"stripe_customer_id,omitempty"`
       StripeSubscriptionID       *string    `json:"stripe_subscription_id,omitempty"`
       SubscriptionTier           string     `json:"subscription_tier"`
       SubscriptionStatus         string     `json:"subscription_status"`
       CancelAtPeriodEnd          bool       `json:"cancel_at_period_end"`
       TokenCostUsedMicrodollars  int64      `json:"token_cost_used_microdollars"`
       TokenCostLimitMicrodollars int64      `json:"token_cost_limit_microdollars"`
       SandboxTimeUsedSeconds     int64      `json:"sandbox_time_used_seconds"`
       SandboxTimeLimitSeconds    int64      `json:"sandbox_time_limit_seconds"`
       BillingPeriodStart         *string    `json:"billing_period_start,omitempty"`
       BillingPeriodEnd           *string    `json:"billing_period_end,omitempty"`
       CreatedAt                  string     `json:"created_at"`
       UpdatedAt                  string     `json:"updated_at"`
   }
   ```

3. Implement OrganizationHandler struct:
   ```go
   type OrganizationHandler struct {
       orgRepo *repository.OrganizationRepository
   }

   func NewOrganizationHandler(orgRepo *repository.OrganizationRepository) *OrganizationHandler {
       return &OrganizationHandler{orgRepo: orgRepo}
   }
   ```

4. Implement handler methods:
   - `Create(c *fiber.Ctx) error` - POST /organizations
   - `GetByID(c *fiber.Ctx) error` - GET /organizations/:id
   - `Update(c *fiber.Ctx) error` - PATCH /organizations/:id
   - `Delete(c *fiber.Ctx) error` - DELETE /organizations/:id
   - `List(c *fiber.Ctx) error` - GET /organizations

5. Implement helper for converting entity to response:
   ```go
   func toOrganizationResponse(org *repository.Organization) *OrganizationResponse
   ```

6. Update `backend/internal/api/routes/router.go`:
   - Add OrganizationRepository to Dependencies struct
   - Create OrganizationHandler in SetupRoutes
   - Add organization routes under `/api/v1/organizations`

7. Update `backend/cmd/server/main.go`:
   - Create OrganizationRepository instance
   - Pass to Dependencies struct

## Acceptance Criteria

- [ ] OrganizationHandler struct created with repository dependency
- [ ] Create endpoint validates name field and returns 201 on success
- [ ] GetByID endpoint returns 404 if not found
- [ ] Update endpoint allows partial updates (PATCH semantics)
- [ ] Delete endpoint returns 204 on success
- [ ] List endpoint supports pagination (limit, offset query params)
- [ ] All endpoints require authentication (use AuthMiddleware)
- [ ] Proper HTTP status codes (200, 201, 204, 400, 404, 500)
- [ ] JSON request/response with proper error messages
- [ ] Routes registered under /api/v1/organizations
- [ ] Repository initialized in main.go and passed to router

## Files to Create/Modify

- `backend/internal/api/handlers/organization.go` - Create handler file
- `backend/internal/api/routes/router.go` - Add routes and dependencies
- `backend/cmd/server/main.go` - Initialize repository

## Integration Points

- **Provides**: REST API for organization management
- **Consumes**: OrganizationRepository (org-repository)
- **Conflicts**: Avoid conflicts with router.go modifications from other tasks

## Technical Notes

- Reference `backend/internal/api/handlers/workspace.go` for pattern examples
- Use Fiber's c.Params("id"), c.Query(), c.BodyParser()
- Validate enum values for subscription_tier and subscription_status
- Return ISO 8601 format for datetime fields in responses
- Use middleware.GetUserID(c) for authenticated user context
