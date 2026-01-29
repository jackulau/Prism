---
id: org-workos-integration
name: WorkOS Organization Integration
wave: 4
priority: 3
dependencies:
- org-repository
- org-api-handlers
estimated_hours: 4
tags:
- backend
- integration
- auth
- workos
---

## Objective

Integrate WorkOS organization management for external authentication and SSO support.

## Context

WorkOS provides enterprise features like SSO, directory sync, and organization management. This task connects the Organization entity with WorkOS's organization API to enable external identity management. The workos_organization_id field links our internal organizations to WorkOS.

## Implementation

1. Create WorkOS configuration in `backend/internal/config/config.go`:
   ```go
   // Add to Config struct
   WorkOSAPIKey      string `env:"WORKOS_API_KEY"`
   WorkOSClientID    string `env:"WORKOS_CLIENT_ID"`
   WorkOSWebhookSecret string `env:"WORKOS_WEBHOOK_SECRET"`
   ```

2. Create `backend/internal/integrations/workos/client.go`:
   ```go
   type WorkOSClient struct {
       apiKey   string
       clientID string
       baseURL  string
   }

   func NewWorkOSClient(apiKey, clientID string) *WorkOSClient
   ```

3. Implement organization sync methods:
   ```go
   // Create organization in WorkOS
   func (c *WorkOSClient) CreateOrganization(name string) (*WorkOSOrg, error)

   // Get organization from WorkOS
   func (c *WorkOSClient) GetOrganization(workosOrgID string) (*WorkOSOrg, error)

   // Update organization in WorkOS
   func (c *WorkOSClient) UpdateOrganization(workosOrgID, name string) error

   // Delete organization from WorkOS
   func (c *WorkOSClient) DeleteOrganization(workosOrgID string) error

   // List organizations from WorkOS (for sync)
   func (c *WorkOSClient) ListOrganizations(limit int, after string) ([]*WorkOSOrg, string, error)
   ```

4. Define WorkOS organization type:
   ```go
   type WorkOSOrg struct {
       ID        string `json:"id"`
       Name      string `json:"name"`
       CreatedAt string `json:"created_at"`
       UpdatedAt string `json:"updated_at"`
   }
   ```

5. Update OrganizationHandler to optionally sync with WorkOS:
   ```go
   type OrganizationHandler struct {
       orgRepo     *repository.OrganizationRepository
       workosClient *workos.WorkOSClient  // Optional, can be nil
   }
   ```

6. Add handler method for WorkOS webhook:
   ```go
   func (h *OrganizationHandler) HandleWorkOSWebhook(c *fiber.Ctx) error
   ```
   - Handle `organization.created` event
   - Handle `organization.updated` event
   - Handle `organization.deleted` event
   - Verify webhook signature

7. Update routes to include WorkOS webhook endpoint:
   ```go
   // Public endpoint with webhook signature verification
   api.Post("/webhooks/workos", orgHandler.HandleWorkOSWebhook)
   ```

8. Create organization sync service (optional background job):
   ```go
   type OrgSyncService struct {
       orgRepo      *repository.OrganizationRepository
       workosClient *workos.WorkOSClient
   }

   func (s *OrgSyncService) SyncFromWorkOS() error
   func (s *OrgSyncService) SyncToWorkOS(orgID string) error
   ```

## Acceptance Criteria

- [ ] WorkOS configuration added to config struct
- [ ] WorkOSClient implements organization CRUD operations
- [ ] HTTP client properly configured with API key header
- [ ] Organization creation optionally creates in WorkOS
- [ ] WorkOS webhook endpoint receives and processes events
- [ ] Webhook signature verification implemented
- [ ] Error handling for WorkOS API failures (graceful degradation)
- [ ] WorkOS organization ID stored/updated in local database
- [ ] Logging for WorkOS API calls and responses

## Files to Create/Modify

- `backend/internal/config/config.go` - Add WorkOS config fields
- `backend/internal/integrations/workos/client.go` - Create WorkOS client
- `backend/internal/integrations/workos/types.go` - Define WorkOS types
- `backend/internal/api/handlers/organization.go` - Add webhook handler
- `backend/internal/api/routes/router.go` - Add webhook route
- `backend/cmd/server/main.go` - Initialize WorkOS client

## Integration Points

- **Provides**: WorkOS integration for external organization management
- **Consumes**: OrganizationRepository, OrganizationHandler
- **Conflicts**: Minor updates to organization.go handler

## Technical Notes

- WorkOS API docs: https://workos.com/docs/reference/organization
- Use HMAC-SHA256 for webhook signature verification
- API Key header: `Authorization: Bearer {api_key}`
- Base URL: https://api.workos.com
- Handle rate limiting (429 responses)
- WorkOS client should be optional (nil check for self-hosted without WorkOS)
- Consider retry logic for transient failures

## Environment Variables

```
WORKOS_API_KEY=sk_test_...
WORKOS_CLIENT_ID=client_...
WORKOS_WEBHOOK_SECRET=whsec_...
```
