# Prism Feature Breakdown

## Feature Categories Overview

| Category | Features | Priority |
|----------|----------|----------|
| **Core Infrastructure** | 8 features | P0 (Foundation) |
| **Agent System** | 12 features | P0 (Core) |
| **Tool System** | 14 features | P1 (Essential) |
| **Real-time** | 6 features | P1 (Essential) |
| **Workers** | 6 features | P2 (Enhancement) |
| **Integrations** | 10 features | P2 (Enhancement) |
| **Billing** | 8 features | P3 (Monetization) |
| **Frontend** | 12 features | P1 (User-facing) |

**Total: 76 distinct features**

---

## Category 1: Core Infrastructure

### F1.1 - Database Setup
**Description:** PostgreSQL database with TypeORM ORM
**Scope:**
- Database connection configuration
- Environment variable handling (PGHOST, PGPORT, etc.)
- SSL support toggle
- Auto-sync in development mode

**Entities Required:**
```
- Organization
- OrganizationMember
- Workspace
- WorkspaceTool
- Agent
- Message
- ToolCall
- WorkerDefinition
- WorkerDefinitionTool
- Tool
- IntegrationProvider
- IntegrationConnection
- SlackUserMapping
```

---

### F1.2 - Organization Entity
**Description:** Multi-tenant organization management
**Fields:**
- `id` - Primary key
- `workosOrganizationId` - External auth ID (unique, indexed)
- `name` - Organization name
- `stripeCustomerId` - Stripe customer reference
- `stripeSubscriptionId` - Subscription reference
- `subscriptionTier` - FREE | PAID | ENTERPRISE
- `subscriptionStatus` - ACTIVE | CANCELED | PAST_DUE | INCOMPLETE
- `cancelAtPeriodEnd` - Boolean flag
- `tokenCostUsedMicrodollars` - Usage counter
- `tokenCostLimitMicrodollars` - Usage limit
- `sandboxTimeUsedSeconds` - Sandbox usage
- `sandboxTimeLimitSeconds` - Sandbox limit
- `billingPeriodStart` - Date
- `billingPeriodEnd` - Date
- `createdAt`, `updatedAt` - Timestamps

---

### F1.3 - Workspace Entity
**Description:** Container for agent sessions linked to repositories
**Fields:**
- `id` - Primary key
- `name` - Workspace name
- `organizationId` - Foreign key
- `githubRepositoryName` - Repository reference
- `workerId` - Optional worker template link
- `currentBranch` - Git branch (indexed)
- `slackChannelId` - Optional Slack integration
- `slackMessageTs` - Slack message timestamp
- `createdAt` - Timestamp

**Relations:**
- Many-to-One: Organization
- One-to-Many: Agent, WorkspaceTool

---

### F1.4 - Authentication System
**Description:** WorkOS-based SSO/OAuth authentication
**Components:**
- WorkOS SDK integration
- OAuth callback handler
- Session cookie management (`wos-session`)
- Organization context loading
- Protected route middleware

**Environment Variables:**
- `WORKOS_API_KEY`
- `WORKOS_CLIENT_ID`
- `WORKOS_COOKIE_PASSWORD`
- `WORKOS_REDIRECT_URI`

---

### F1.5 - tRPC API Setup
**Description:** Type-safe RPC API layer
**Components:**
- tRPC router configuration
- Context creation with auth
- Protected procedure middleware
- Error handling (TRPCError)
- Input validation with Zod

**Routers:**
- `workspace` - Workspace operations
- `workers` - Worker management
- `integrations` - Integration connections
- `organization` - Org settings
- `payment` - Billing operations

---

### F1.6 - Database Seed Data
**Description:** Default data population on startup
**Seed Items:**
- Default integration providers (GitHub, PostHog, Cursor, Jules, Slack)
- Default tools catalog
- System configuration

---

### F1.7 - Environment Configuration
**Description:** Centralized environment variable management
**Categories:**
- Database (PGHOST, PGPORT, PGUSER, PGPASSWORD, PGDATABASE, PGSSL)
- Auth (WORKOS_*)
- LLM (OPENAI_API_KEY)
- Analytics (POSTHOG_API_KEY)
- Payments (STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET)
- GitHub (GITHUB_APP_ID, GITHUB_PRIVATE_KEY, GITHUB_WEBHOOK_SECRET)

---

### F1.8 - Data Encryption
**Description:** Sensitive data encryption for integration credentials
**Implementation:**
- `setDataConfig()` - Encrypt and store
- `getDataConfig()` - Decrypt and retrieve
- JSONB storage format

---

## Category 2: Agent System

### F2.1 - Agent Entity
**Description:** AI execution instance data model
**Fields:**
- `id` - Primary key
- `workspaceId` - Foreign key
- `status` - PENDING | RUNNING | COMPLETED | FAILED
- `providerType` - | CURSOR | JULES
- `conversationId` - Provider conversation ID
- `url` - Agent URL
- `githubBranchName` - Created branch
- `name` - Agent name
- `model` - LLM model used
- `sandboxId` - Sandbox instance ID
- `isOrchestratorAgent` - Boolean flag
- `createdAt`, `updatedAt` - Timestamps

---

### F2.2 - Agent Status Lifecycle
**Description:** Agent state machine management
**States:**
```
PENDING → RUNNING → COMPLETED
              ↓
           FAILED
```
**Functions:**
- `markAgentComplete(agentId)`
- `markAgentFailed(agentId, error)`
- `getAgentStatus(agentId)`

---

### F2.3 - Message Entity
**Description:** Conversation history storage
**Fields:**
- `id` - Primary key
- `agentId` - Foreign key
- `content` - Message text
- `sender` - USER | AGENT
- `images` - Array of {data, mimeType}
- `promptTokens`, `completionTokens`, `totalTokens` - Token tracking
- `costMicrodollars` - Cost tracking
- `error` - Optional error message
- `model` - LLM model
- `sandboxDurationMs` - Execution time
- `createdAt` - Timestamp

---

### F2.4 - ToolCall Entity
**Description:** Tool execution tracking
**Fields:**
- `id` - Primary key
- `agentId` - Foreign key
- `messageId` - Foreign key
- `toolName` - Tool identifier (indexed)
- `arguments` - JSON input
- `result` - Execution result
- `status` - success | error
- `createdAt` - Timestamp

---

### F2.5 - CloudProvider Interface
**Description:** Abstract interface for agent backends
**Methods:**
```typescript
interface CloudProvider {
  createAgent(params): Promise<Agent>
  getMessages(agent): Promise<ProviderMessage[]>
  sendMessage(agent, message, images): Promise<boolean>
}
```

---

### F2.6 - Prism Provider (Native)
**Description:** Built-in agent provider with full sandbox control
**Features:**
- Internal workflow execution
- Token tracking and cost calculation
- SSE streaming support
- Database persistence
- Sandbox management

---

### F2.7 - Cursor Provider
**Description:** External Cursor AI integration
**Endpoints:**
- `POST https://api.cursor.com/v0/agents` - Create agent
- `GET /messages` - Get history
- `POST /followup` - Send message

**Auth:** Basic auth with API key (Base64 encoded)

---

### F2.8 - Jules Provider
**Description:** External Jules AI integration
**Similar to Cursor provider with Jules-specific API**

---

### F2.9 - Standard Agent Workflow
**Description:** Main execution workflow for coding agents
**Steps:**
1. Load agent from DB
2. Validate and get GitHub token
3. Create sandbox environment
4. Load conversation history
5. Create branch if first run
6. Run LLM with tools
7. Save response and track tokens
8. Commit changes if any
9. Cleanup sandbox
10. Mark complete

---

### F2.10 - Orchestrator Agent Workflow
**Description:** Multi-agent coordination workflow
**Differences from Standard:**
- Uses "high" reasoning effort
- Has `spawn_sub_agent` tool
- Extracts images from history
- Returns reasoning logs

---

### F2.11 - Workflow Step Functions
**Description:** Modular workflow steps
**Functions:**
- `loadAgent(agentId)`
- `validateAndGetToken(repoName)`
- `prepareSandbox(agentId, repo, token)`
- `loadPreviousMessages(agentId)`
- `createBranchIfNeeded(agent, sandbox)`
- `saveAgentResponse(agentId, text, usage)`
- `commitChangesIfNeeded(sandbox, agent, prompt)`
- `cleanupSandbox(sandbox)`

---

### F2.12 - Agent System Prompts
**Description:** System prompts for different agent types
**Prompts:**
- `AGENT_SYSTEM_PROMPT` - Standard coding agent
- `ORCHESTRATOR_AGENT_SYSTEM_PROMPT` - Multi-agent coordinator

---

## Category 3: Tool System

### F3.1 - Tool Entity
**Description:** Tool catalog data model
**Fields:**
- `id` - Primary key
- `displayName` - Human-readable name
- `slugName` - Unique identifier (e.g., "posthog/errors")
- `isModel` - Boolean (model config vs tool)
- `providerId` - Optional integration provider link

**Unique Constraint:** [displayName, provider, slugName]

---

### F3.2 - Dynamic Tool Building
**Description:** Runtime tool collection assembly
**Function:** `buildDynamicTools(agentId, toolSlugs, sandbox)`
**Logic:**
- Parse tool slugs
- Load matching tool builders
- Merge into single collection

---

### F3.3 - Sandbox Tool: listFiles
**Description:** List files in repository
**Input:** `{ relativePath: string }`
**Execution:** `ls -la {path}`

---

### F3.4 - Sandbox Tool: readFile
**Description:** Read file contents
**Input:** `{ relativeFilePath: string }`
**Execution:** `cat {path}`

---

### F3.5 - Sandbox Tool: updateFile
**Description:** Write/overwrite file
**Input:** `{ relativeFilePath: string, content: string }`
**Execution:** File write operation

---

### F3.6 - Sandbox Tool: grep
**Description:** Search in files
**Input:** `{ command: string }`
**Execution:** Bash grep command

---

### F3.7 - GitHub Tool: github_list_commits
**Description:** List recent commits
**Input:** `{ commitLimit: number (1-10, default 5) }`
**Execution:** `git log -n {limit} --pretty=format:'%H - %an, %ad : %s' --date=iso`

---

### F3.8 - GitHub Tool: github_view_commit
**Description:** Show commit diff
**Input:** `{ commitSha: string }`
**Execution:** `git show {sha}`

---

### F3.9 - PostHog Tools: Query Runner
**Slug:** `posthog/query_runner`
**Tools:**
- `query_run` - Execute HogQL query
- `query_generate_hogql_from_question` - Natural language to HogQL
- `docs_search` - Search documentation

---

### F3.10 - PostHog Tools: Insights
**Slug:** `posthog/insights`
**Tools:**
- `insight_create_from_query`
- `insight_delete`
- `insight_get`
- `insight_query`
- `insight_update`
- `insights_get_all`

---

### F3.11 - PostHog Tools: Errors
**Slug:** `posthog/errors`
**Tools:**
- `error_details` - Get error details
- `list_errors` - List recent errors

---

### F3.12 - PostHog Tools: Documentation
**Slug:** `posthog/documentation`
**Tools:**
- `docs_search` - Search PostHog docs

---

### F3.13 - Orchestrator Tool: spawn_sub_agent
**Description:** Create sub-agent for task delegation
**Input:** `{ prompt: string }`
**Execution:** Creates new agent via PrismProvider
**Returns:** New agent ID

---

### F3.14 - Tool Slug System
**Description:** Hierarchical tool identification
**Format:** `{provider}/{tool-name}`
**Examples:**
- `posthog/errors`
- `github/commits`
- `sandbox/files`

---

## Category 4: Sandbox Environment

### F4.1 - Vercel Sandbox Integration
**Description:** Isolated code execution environment
**Capabilities:**
- File operations (read, write, list)
- Command execution (bash)
- Git operations (clone, branch, commit)
- Ephemeral containers

---

### F4.2 - Sandbox Lifecycle Management
**Description:** Create, use, cleanup sandbox
**Functions:**
- `prepareSandbox(agentId, repo, token)` - Initialize
- `cleanupSandbox(sandbox)` - Destroy

---

### F4.3 - Sandbox Usage Tracking
**Description:** Track sandbox execution time
**Functions:**
- `canUseSandbox(orgId)` - Check availability
- `incrementSandboxTimeSeconds(orgId, seconds)` - Record usage

---

### F4.4 - Git Branch Creation
**Description:** Automatic feature branch creation
**Format:** `prism/{agent-id}-{sanitized-prompt}`
**Function:** `createBranchIfNeeded(agent, sandbox)`

---

### F4.5 - Auto-commit Changes
**Description:** Commit file changes after agent work
**Function:** `commitChangesIfNeeded(sandbox, agent, prompt)`
**Logic:**
- Detect file changes
- Generate commit message from prompt (truncated to 50 chars)
- Git commit

---

### F4.6 - Sandbox Command Execution
**Description:** Run bash commands in sandbox
**Method:** `sandbox.runCommand({ cmd, args })`
**Returns:** `{ stdout(), stderr() }`

---

## Category 5: Real-time Streaming

### F5.1 - Event Payload Types
**Description:** Event data structures
**Types:**
```typescript
type AgentEventPayload = {
  type: 'status' | 'error' | 'done' | string;
  phase?: string;
  step?: string;
  detail?: string;
  code?: string;
  reason?: string;
}
```

---

### F5.2 - Event Envelope
**Description:** Event wrapper with metadata
**Fields:**
- `id` - Auto-incrementing per agent
- `timestamp` - Event time
- `payload` - Event data

---

### F5.3 - In-Memory Event Store
**Description:** Per-agent event storage
**Structure:**
```typescript
Map<agentId, {
  events: AgentEventEnvelope[]  // Max 500
  listeners: Set<Function>
  lastActivity: Date
}>
```
**Cleanup:** 60-second interval, 5-minute inactivity threshold

---

### F5.4 - Event Publishing
**Description:** Emit events to listeners
**Functions:**
- `publishAgentEvent(agentId, payload)`
- `emitStatus(agentId, phase, step, detail?, metadata?)`
- `emitError(agentId, code, message)`
- `emitDone(agentId, reason)`

---

### F5.5 - SSE Endpoint
**Description:** Server-Sent Events HTTP endpoint
**Route:** `GET /sse/agent/:agentId`
**Features:**
- Event stream headers
- Last-event-id resumption
- Backlog replay
- Disconnect handling

---

### F5.6 - Reasoning Streamer
**Description:** Stream LLM reasoning to client
**Function:** `createReasoningStreamer(agentId)`
**Emits:** Status events with reasoning text

---

## Category 6: Workers System

### F6.1 - WorkerDefinition Entity
**Description:** Reusable agent template
**Fields:**
- `id` - Primary key
- `slug` - Unique per org
- `prompt` - Custom system prompt
- `organizationId` - Foreign key
- `key` - Webhook auth key
- `cloudProviders` - Array of provider configs
- `createdAt` - Timestamp

---

### F6.2 - WorkerDefinitionTool Entity
**Description:** Worker-to-tool association
**Fields:**
- `id` - Primary key
- `workerDefinitionId` - Foreign key
- `toolId` - Foreign key

---

### F6.3 - Workers tRPC Router
**Description:** Worker CRUD operations
**Procedures:**
- `workers.list()` - List org workers
- `workers.create({ slug, prompt, toolSlugs, cloudProviders, key })`
- `workers.update({ id, ... })`
- `workers.delete({ id })`

---

### F6.4 - Worker Slug Validation
**Description:** Unique slug per organization
**Validation:** Check slug doesn't exist for org

---

### F6.5 - Worker Tool Association
**Description:** Link tools to worker definitions
**Logic:**
- On create: Create WorkerDefinitionTool records
- On update: Delete old, create new

---

### F6.6 - Worker Provider Configuration
**Description:** Multi-provider support per worker
**Schema:**
```typescript
{
  provider: 'prism' | 'cursor' | 'jules',
  model?: string
}[]
```

---

## Category 7: Webhooks

### F7.1 - GitHub Webhook Handler
**Description:** Process GitHub events
**Route:** `POST /webhooks/github/events`
**Trigger:** Issue comments containing `--prism`
**Features:**
- HMAC-SHA256 signature verification
- Worker slug extraction
- Workspace creation

---

### F7.2 - PostHog Webhook Handler
**Description:** Process PostHog error events
**Route:** `POST /webhooks/posthog/issue`
**Auth:** Worker slug + key validation
**Action:** Create workspace with error context

---

### F7.3 - Cursor Completion Webhook
**Description:** Update agent status from Cursor
**Route:** `POST /webhooks/cursor/complete/:agentId`
**Input:** `{ status: 'FINISHED' | 'FAILED' }`

---

### F7.4 - Stripe Webhook Handler
**Description:** Process Stripe payment events
**Route:** `POST /webhooks/stripe`
**Events:**
- `invoice.paid` - Reset usage
- `customer.subscription.updated` - Tier change
- `customer.subscription.deleted` - Cancellation

---

### F7.5 - Webhook Signature Verification
**Description:** Validate webhook authenticity
**Implementation:**
```typescript
const expected = 'sha256=' + crypto
  .createHmac('sha256', secret)
  .update(payload)
  .digest('hex');
```

---

### F7.6 - Workspace Creation from Webhook
**Description:** Automated workspace + agent creation
**Function:** `createWorkspaceFromWebhook(params)`
**Steps:**
1. Find worker definition
2. Create workspace
3. Associate tools
4. Create agent

---

## Category 8: Integrations

### F8.1 - IntegrationProvider Entity
**Description:** Available integration types
**Fields:**
- `id` - Primary key
- `slug` - Unique identifier
- `displayName` - Human-readable name
- `logoUrl` - Icon URL

---

### F8.2 - IntegrationConnection Entity
**Description:** User's connected integrations
**Fields:**
- `id` - Primary key
- `organizationId` - Foreign key
- `providerId` - Foreign key
- `externalId` - External service ID
- `data` - Encrypted JSONB config
- `createdAt`, `updatedAt` - Timestamps

---

### F8.3 - Integrations tRPC Router
**Description:** Integration management API
**Procedures:**
- `integrations.list()` - List with connection status
- `integrations.connect({ providerSlug, data })`
- `integrations.disconnect({ id })`
- `integrations.repositories()` - GitHub repos
- `integrations.branches({ repositoryFullName })` - Repo branches

---

### F8.4 - GitHub App Integration
**Description:** GitHub repository access
**Features:**
- OAuth installation flow
- Installation token retrieval
- Repository listing
- Branch listing
- Webhook events

---

### F8.5 - GitHub Service
**Description:** GitHub token management
**Function:** `getGithubTokenForUser(organizationId)`
**Steps:**
1. Find integration connection
2. Get installation ID from encrypted config
3. Get installation token

---

### F8.6 - PostHog Integration
**Description:** Analytics and error tracking
**Storage:** API key in encrypted data
**Tools:** Query runner, insights, errors, documentation

---

### F8.7 - Cursor Integration
**Description:** External Cursor AI agent
**Auth:** API key (Basic auth)
**Features:** Agent creation, message retrieval

---

### F8.8 - Jules Integration
**Description:** External Jules AI agent
**Similar to Cursor integration**

---

### F8.9 - Slack Integration
**Description:** Team notifications
**Features:**
- User mapping (SlackUserMapping entity)
- Channel notifications
- Message threading

---

### F8.10 - Integration Data Encryption
**Description:** Secure credential storage
**Methods:**
- `setDataConfig(input)` - Encrypt and save
- `getDataConfig()` - Decrypt and return

---

## Category 9: LLM Integration

### F9.1 - OpenAI SDK Setup
**Description:** AI SDK configuration
**Model:** GPT-5-mini (or configurable)
**Features:**
- Reasoning mode support
- Tool calling
- Token tracking

---

### F9.2 - Token Usage Accumulator
**Description:** Track token consumption
**Structure:**
```typescript
{
  promptTokens: number
  completionTokens: number
  totalTokens: number
}
```

---

### F9.3 - Standard Agent LLM Runner
**Description:** Run LLM for standard agents
**Function:** `runAgentLLM(agentId, sandbox, toolSlugs, messages, usage)`
**Config:**
- Reasoning effort: "medium"
- Max steps: 32

---

### F9.4 - Orchestrator Agent LLM Runner
**Description:** Run LLM for orchestrator agents
**Function:** `runOrchestratorAgentLLM(...)`
**Config:**
- Reasoning effort: "high"
- Includes spawn_sub_agent tool
- Returns reasoning logs

---

### F9.5 - PostHog Tracing
**Description:** LLM call observability
**Integration:** PostHog SDK
**Tracks:** Calls, tokens, latency

---

### F9.6 - Token Cost Calculation
**Description:** Convert tokens to cost
**Formula:**
```typescript
const COST_PER_1K_TOKENS = 0.01;  // $0.01 per 1K
const costMicrodollars = Math.round(totalTokens * COST_PER_1K_TOKENS * 1000);
```

---

## Category 10: Billing & Usage

### F10.1 - Subscription Tiers
**Description:** Plan configuration
**Tiers:**
| Tier | Token Limit | Sandbox | Price |
|------|-------------|---------|-------|
| FREE | $5 | 2 hrs | $0/mo |
| PAID | $20 | 24 hrs | $20/mo |
| ENTERPRISE | Unlimited | Unlimited | Custom |

---

### F10.2 - Usage Tracking: Tokens
**Description:** Track token cost usage
**Functions:**
- `canSendMessage(orgId)` - Check limit
- `incrementTokenCostMicrodollars(orgId, amount)`

---

### F10.3 - Usage Tracking: Sandbox
**Description:** Track sandbox time usage
**Functions:**
- `canUseSandbox(orgId)` - Check limit
- `incrementSandboxTimeSeconds(orgId, seconds)`

---

### F10.4 - Billing Period Management
**Description:** Monthly billing cycle
**Function:** `resetBillingPeriodIfNeeded(org)`
**Logic:** Auto-reset for free tier, Stripe webhook for paid

---

### F10.5 - Usage Stats API
**Description:** Get usage statistics
**Function:** `getUsageStats(orgId)`
**Returns:**
```typescript
{
  tokenCostPercentage: number
  sandboxTimePercentage: number
  billingPeriodStart: Date
  billingPeriodEnd: Date
}
```

---

### F10.6 - Stripe Customer Management
**Description:** Customer creation and management
**Operations:**
- Create customer on signup
- Link to organization

---

### F10.7 - Stripe Subscription Management
**Description:** Handle subscriptions
**Operations:**
- Create subscription
- Update tier
- Cancel subscription

---

### F10.8 - Payment tRPC Router
**Description:** Billing API endpoints
**Procedures:**
- `payment.getPlans()` - List available plans
- `payment.createCheckout()` - Start subscription
- `payment.cancelSubscription()` - Cancel
- `payment.getUsage()` - Current usage

---

## Category 11: Frontend

### F11.1 - Home/Dashboard Page
**Path:** `/`
**Features:**
- Recent workspaces list
- Quick actions
- Usage summary

---

### F11.2 - Login Page
**Path:** `/login`
**Features:**
- WorkOS OAuth redirect
- Session handling

---

### F11.3 - Workspace Page
**Path:** `/workspace/:id`
**Features:**
- Agent chat interface
- Message display
- Tool call visualization
- SSE connection for real-time updates
- Send message input

---

### F11.4 - Workers Page
**Path:** `/workers`
**Features:**
- Worker list
- Create worker form
- Edit worker
- Delete worker

---

### F11.5 - Integrations Page
**Path:** `/integrations`
**Features:**
- Available integrations list
- Connected status
- Connect/disconnect buttons
- OAuth flows

---

### F11.6 - Usage Page
**Path:** `/usage`
**Features:**
- Token usage chart
- Sandbox time chart
- Billing period info
- Upgrade CTA

---

### F11.7 - Organization Settings Page
**Path:** `/organization`
**Features:**
- Org name editing
- Member management
- Subscription info

---

### F11.8 - Sidebar Component
**Description:** Navigation sidebar
**Features:**
- Page links
- Workspace list
- Current org display

---

### F11.9 - Message Display Component
**Description:** Chat message rendering
**Features:**
- User vs agent styling
- Markdown rendering
- Image display
- Timestamp

---

### F11.10 - Tool Call Visualization
**Description:** Show tool executions
**Features:**
- Tool name
- Arguments display
- Result/error display
- Collapsible details

---

### F11.11 - SSE Client Connection
**Description:** Real-time event subscription
**Features:**
- EventSource connection
- Reconnection logic
- Last-event-id resumption
- Event handling

---

### F11.12 - tRPC Client Setup
**Description:** Frontend API client
**Code:**
```typescript
export const trpc = createTRPCReact<AppRouter>();
```

---

## Implementation Priority Matrix

### P0 - Foundation (Must Have First)
| ID | Feature | Complexity |
|----|---------|------------|
| F1.1 | Database Setup | Medium |
| F1.2 | Organization Entity | Low |
| F1.3 | Workspace Entity | Low |
| F1.4 | Authentication System | High |
| F1.5 | tRPC API Setup | Medium |
| F2.1 | Agent Entity | Low |
| F2.2 | Agent Status Lifecycle | Low |
| F2.3 | Message Entity | Low |

### P1 - Core Features
| ID | Feature | Complexity |
|----|---------|------------|
| F2.5 | CloudProvider Interface | Medium |
| F2.6 | Prism Provider | High |
| F2.9 | Standard Agent Workflow | High |
| F4.1 | Vercel Sandbox Integration | High |
| F5.5 | SSE Endpoint | Medium |
| F9.1 | OpenAI SDK Setup | Medium |
| F11.3 | Workspace Page | High |

### P2 - Essential Tools
| ID | Feature | Complexity |
|----|---------|------------|
| F3.3-F3.6 | Sandbox Tools | Medium |
| F3.7-F3.8 | GitHub Tools | Low |
| F3.13 | spawn_sub_agent | Medium |
| F5.1-F5.4 | Event System | Medium |
| F8.4 | GitHub App Integration | High |

### P3 - Enhancements
| ID | Feature | Complexity |
|----|---------|------------|
| F6.1-F6.6 | Workers System | Medium |
| F7.1-F7.6 | Webhooks | Medium |
| F3.9-F3.12 | PostHog Tools | Medium |
| F2.7-F2.8 | Cursor/Jules Providers | Medium |

### P4 - Monetization
| ID | Feature | Complexity |
|----|---------|------------|
| F10.1-F10.8 | Billing System | High |
| F8.9 | Slack Integration | Medium |

---

## Quick Reference: Feature Dependencies

```
F1.1 (Database) ─┬─→ F1.2 (Organization)
                 ├─→ F1.3 (Workspace)
                 ├─→ F2.1 (Agent)
                 └─→ F2.3 (Message)

F1.4 (Auth) ────→ F1.5 (tRPC) ────→ All API Features

F4.1 (Sandbox) ──┬─→ F3.3-F3.6 (Sandbox Tools)
                 └─→ F2.9 (Agent Workflow)

F2.5 (Provider Interface) ─┬─→ F2.6 (Prism Provider)
                           ├─→ F2.7 (Cursor Provider)
                           └─→ F2.8 (Jules Provider)

F5.1-F5.4 (Events) ────→ F5.5 (SSE) ────→ F11.11 (SSE Client)

F6.1 (Workers) ────→ F7.1-F7.2 (Webhooks)
```

---

*Total: 76 features across 11 categories*
