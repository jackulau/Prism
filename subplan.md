# additional features

## Complete Feature Analysis & Implementation Guide

> **Primary Language**: TypeScript (86%), CSS (13.5%)
> **Architecture**: Full-stack monorepo (Frontend + Backend + Orchestrator)

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Core Concepts](#core-concepts)
4. [Workers System](#workers-system)
5. [Agent System](#agent-system)
6. [Provider System](#provider-system)
7. [Tool System](#tool-system)
8. [Sandbox Environment](#sandbox-environment)
9. [Workflow Orchestration](#workflow-orchestration)
10. [LLM Integration](#llm-integration)
11. [Real-time Streaming (SSE)](#real-time-streaming-sse)
12. [Database Schema](#database-schema)
13. [Authentication](#authentication)
14. [Integrations](#integrations)
15. [Payment & Billing](#payment--billing)
16. [Webhooks](#webhooks)
17. [API Routes (tRPC)](#api-routes-trpc)
18. [Frontend Architecture](#frontend-architecture)
19. [Implementation Checklist](#implementation-checklist)

---

## Overview

Prism is an **asynchronous coding agent platform** that allows developers to run long-running AI-powered tasks on their codebases in the cloud. The key differentiator is **true autonomy** — agents have access to external integrations and analytics data (not just the codebase), enabling richer context for problem-solving without constant human micromanagement.

### Key Value Propositions

1. **Asynchronous Execution** - Tasks run in the cloud without blocking
2. **External Integrations** - Access to analytics, error tracking, and external services
3. **Worker Customization** - Define specialized agents with custom prompts and tools
4. **Multi-Provider Support** - Native Prism, Cursor, and Jules integration
5. **Webhook-Driven Automation** - Trigger agents from GitHub issues, PostHog errors, etc.

---

## Architecture

### Tech Stack

| Component | Technology |
|-----------|------------|
| **Frontend** | React + TypeScript + Vite + Bun |
| **Backend** | TypeScript + Nitro + Bun |
| **Database** | PostgreSQL + TypeORM |
| **API** | tRPC (type-safe RPC) |
| **Auth** | WorkOS (SSO/OAuth) |
| **Payments** | Stripe |
| **Sandbox** | Vercel Sandbox |
| **LLM** | OpenAI GPT-5-mini with reasoning |
| **Analytics** | PostHog |
| **Real-time** | Server-Sent Events (SSE) |

### Directory Structure

```
prism/
├── frontend/
│   ├── src/
│   │   ├── assets/svgs/
│   │   ├── components/
│   │   │   ├── CreateBranch/
│   │   │   └── Sidebar/
│   │   ├── features/repositories/
│   │   ├── lib/
│   │   │   ├── trpc.ts          # tRPC client
│   │   │   ├── types.ts         # Shared types
│   │   │   └── useAuth.tsx      # Auth hook
│   │   ├── pages/
│   │   │   ├── Home/
│   │   │   ├── Integrations/
│   │   │   ├── Login/
│   │   │   ├── Organization/
│   │   │   ├── Usage/
│   │   │   ├── Workers/
│   │   │   └── Workspace/
│   │   ├── App.tsx
│   │   └── main.tsx
│   └── package.json
│
├── backend/
│   ├── src/
│   │   ├── auth/
│   │   │   ├── auth.ts          # WorkOS setup
│   │   │   └── routes.ts        # Auth routes
│   │   ├── db/
│   │   │   ├── entities/        # TypeORM entities
│   │   │   ├── data-source.ts   # DB connection
│   │   │   └── seed.ts          # Default data
│   │   ├── express/
│   │   │   └── webhooks.ts      # Webhook handlers
│   │   ├── payment/
│   │   │   ├── model-pricing.ts
│   │   │   ├── plans.ts
│   │   │   ├── stripe.ts
│   │   │   ├── usage.ts
│   │   │   └── webhook.ts
│   │   ├── providers/
│   │   │   ├── base.ts          # Provider interface
│   │   │   ├── prism.ts         # Native provider
│   │   │   ├── cursor.ts        # Cursor AI
│   │   │   └── jules.ts         # Jules AI
│   │   ├── services/
│   │   │   ├── githubService.ts
│   │   │   └── organizationService.ts
│   │   ├── stream/
│   │   │   ├── events.ts        # Event system
│   │   │   └── sse.ts           # SSE router
│   │   ├── tools/
│   │   │   ├── github/
│   │   │   │   └── commits.ts
│   │   │   ├── posthog/
│   │   │   │   └── index.ts
│   │   │   ├── dynamic.ts       # Dynamic tool builder
│   │   │   ├── primaryAgent.ts  # Orchestrator tools
│   │   │   └── sandboxTools.ts  # File/code tools
│   │   ├── trpc/
│   │   │   ├── routers/
│   │   │   │   ├── integrations.ts
│   │   │   │   ├── organization.ts
│   │   │   │   ├── payment.ts
│   │   │   │   ├── workers.ts
│   │   │   │   └── workspace.ts
│   │   │   ├── context.ts
│   │   │   ├── router.ts
│   │   │   └── trpc.ts
│   │   ├── utils/
│   │   └── workflows/
│   │       ├── helpers/
│   │       ├── agent.ts         # Main workflows
│   │       ├── llm.ts           # LLM integration
│   │       ├── prompts.ts       # System prompts
│   │       └── steps.ts         # Workflow steps
│   └── package.json
```

---

## Core Concepts

### 1. Organizations
Top-level tenant containing users, workspaces, and billing.

### 2. Workspaces
Container for agent sessions, linked to a GitHub repository.

### 3. Workers
Customizable agent templates with specific prompts and tools.

### 4. Agents
Individual AI execution instances (can be standard or orchestrator).

### 5. Messages
Conversation history with token tracking and tool calls.

---

## Workers System

Workers are **webhook-triggered workspace environments** that enable specialized, autonomous agents.

### Worker Definition Entity

```typescript
interface WorkerDefinition {
  id: number;
  slug: string;                              // Unique per org
  prompt: string;                            // Custom system prompt
  organizationId: number;
  key: string | null;                        // API key for webhooks
  cloudProviders: Array<{                    // Provider configs
    provider: 'prism' | 'cursor' | 'jules';
    model?: string;
  }>;
  tools: WorkerDefinitionTool[];             // Associated tools
  createdAt: Date;
}
```

### Worker Features

1. **Custom Prompts** - Define specialized instructions
2. **Selective Tools** - Only include relevant tools (avoid context bloat)
3. **Webhook Triggers** - Automated via GitHub issues, PostHog errors
4. **Multi-Provider** - Can use Prism, Cursor, or Jules
5. **API Key Auth** - Secure webhook authentication

### Worker tRPC Operations

```typescript
// List all workers
workers.list()

// Create worker
workers.create({
  slug: "error-fixer",
  prompt: "Fix errors reported by PostHog...",
  toolSlugs: ["posthog/errors", "github/commits"],
  cloudProviders: [{ provider: "prism", model: "gpt-5-mini" }],
  key: "optional-webhook-key"
})

// Update worker
workers.update({ id, slug, prompt, toolSlugs, cloudProviders, key })

// Delete worker
workers.delete({ id })
```

---

## Agent System

Agents are AI execution instances that perform tasks autonomously.

### Agent Entity

```typescript
enum AgentStatus {
  PENDING = 'PENDING',
  RUNNING = 'RUNNING',
  COMPLETED = 'COMPLETED',
  FAILED = 'FAILED'
}

enum ProviderType {
  PRISM = 'PRISM',
  CURSOR = 'CURSOR',
  JULES = 'JULES'
}

interface Agent {
  id: number;
  workspace: Workspace;
  status: AgentStatus;
  providerType: ProviderType;
  conversationId: string;
  url: string;
  githubBranchName: string | null;
  name: string;                    // Default: "Untitled"
  model: string | null;
  sandboxId: string | null;
  isOrchestratorAgent: boolean;    // Can spawn sub-agents
  createdAt: Date;
  updatedAt: Date;
  messages: Message[];
}
```

### Agent Types

#### 1. Standard Agent
- Performs coding tasks directly
- Has sandbox access for file operations
- Creates branches and commits

#### 2. Orchestrator Agent
- Coordinates multiple sub-agents
- Uses "high" reasoning effort
- Can spawn sub-agents via `spawn_sub_agent` tool
- Strategic task decomposition

### Agent Lifecycle

```
PENDING → RUNNING → COMPLETED
              ↓
           FAILED
```

---

## Provider System

Providers abstract different AI agent backends.

### CloudProvider Interface

```typescript
interface CloudProvider {
  // Create new agent
  createAgent(params: {
    organizationId: number;
    workspace: Workspace;
    repositoryFullName: string;
    message: string;
    toolSlugs: string[];
    baseBranch: string;
    model?: string;
    isOrchestratorAgent: boolean;
    images: MessageImage[];
  }): Promise<Agent>;

  // Get conversation history
  getMessages(agent: Agent): Promise<ProviderMessage[]>;

  // Send follow-up message
  sendMessage(
    agent: Agent,
    message: string,
    images: MessageImage[]
  ): Promise<boolean>;
}
```

### Supported Providers

#### 1. Prism Provider (Native)
- Full sandbox control
- Token tracking and cost calculation
- Internal workflow execution
- SSE streaming support

#### 2. Cursor Provider
- External Cursor AI API integration
- Basic auth with API key
- Branch/URL tracking from response
- Endpoints:
  - `POST https://api.cursor.com/v0/agents` - Create
  - `GET /messages` - History
  - `POST /followup` - Continue

#### 3. Jules Provider
- Similar to Cursor integration
- External Jules AI API

---

## Tool System

Tools provide agents with capabilities beyond just code generation.

### Tool Categories

#### 1. Sandbox Tools (File Operations)

```typescript
function sandboxTools(agentId: number, sandbox: Sandbox) {
  return {
    listFiles: tool({
      description: "List files in the repository",
      inputSchema: z.object({
        relativePath: z.string()
      }),
      execute: async ({ relativePath }) => {
        const result = await sandbox.runCommand({
          cmd: 'bash',
          args: ['-c', `ls -la ${relativePath}`]
        });
        return result.stdout();
      }
    }),

    readFile: tool({
      description: "Read a file from the repository",
      inputSchema: z.object({
        relativeFilePath: z.string()
      }),
      execute: async ({ relativeFilePath }) => {
        const result = await sandbox.runCommand({
          cmd: 'cat',
          args: [relativeFilePath]
        });
        return result.stdout();
      }
    }),

    updateFile: tool({
      description: "Overwrite a file with new content",
      inputSchema: z.object({
        relativeFilePath: z.string(),
        content: z.string()
      }),
      execute: async ({ relativeFilePath, content }) => {
        // Write file content
      }
    }),

    grep: tool({
      description: "Run grep in the repository",
      inputSchema: z.object({
        command: z.string()
      }),
      execute: async ({ command }) => {
        const result = await sandbox.runCommand({
          cmd: 'bash',
          args: ['-c', command]
        });
        return result.stdout();
      }
    })
  };
}
```

#### 2. GitHub Tools

```typescript
function buildCommitTools(sandbox: Sandbox) {
  return {
    github_list_commits: tool({
      description: "List recent git commits (max 10)",
      inputSchema: z.object({
        commitLimit: z.number().min(1).max(10).default(5)
      }),
      execute: async ({ commitLimit }) => {
        const result = await sandbox.runCommand({
          cmd: 'bash',
          args: ['-c', `git log -n ${commitLimit} --pretty=format:'%H - %an, %ad : %s' --date=iso`]
        });
        return result.stdout();
      }
    }),

    github_view_commit: tool({
      description: "Show detailed diff for a specific commit",
      inputSchema: z.object({
        commitSha: z.string()
      }),
      execute: async ({ commitSha }) => {
        const result = await sandbox.runCommand({
          cmd: 'bash',
          args: ['-c', `git show ${commitSha}`]
        });
        return result.stdout();
      }
    })
  };
}
```

#### 3. PostHog Tools

| Tool Slug | Tools Included |
|-----------|----------------|
| `posthog/query_runner` | `query_run`, `query_generate_hogql_from_question`, `docs_search` |
| `posthog/insights` | `insight_create_from_query`, `insight_delete`, `insight_get`, `insight_query`, `insight_update`, `insights_get_all` |
| `posthog/errors` | `error_details`, `list_errors` |
| `posthog/documentation` | `docs_search` |

#### 4. Orchestrator Tools

```typescript
function buildOrchestratorAgentTools(params: {
  agentId: number;
  organizationId: number;
  workspace: Workspace;
  repositoryFullName: string;
  baseBranch: string;
  toolSlugs: string[];
  images: MessageImage[];
}) {
  return {
    spawn_sub_agent: tool({
      description: "Spawn a new agent with a given prompt",
      inputSchema: z.object({
        prompt: z.string().describe("The prompt to spawn the agent with")
      }),
      execute: async ({ prompt }) => {
        const agent = await new PrismProvider().createAgent({
          organizationId,
          workspace,
          repositoryFullName,
          message: prompt,
          toolSlugs,
          baseBranch,
          isOrchestratorAgent: false,
          images
        });
        return agent.id;
      }
    })
  };
}
```

### Dynamic Tool Building

```typescript
async function buildDynamicTools(
  agentId: number,
  toolSlugs: string[],
  sandbox: Sandbox
): Promise<ToolCollection> {
  const tools: ToolCollection = {};

  // PostHog tools
  if (toolSlugs.includes('posthog/query_runner')) {
    Object.assign(tools, await buildQueryRunnerTools(agentId));
  }
  if (toolSlugs.includes('posthog/insights')) {
    Object.assign(tools, await buildInsightTools(agentId));
  }
  if (toolSlugs.includes('posthog/errors')) {
    Object.assign(tools, await buildErrorTools(agentId));
  }
  if (toolSlugs.includes('posthog/documentation')) {
    Object.assign(tools, await buildDocumentationTools(agentId));
  }

  // GitHub tools
  if (toolSlugs.includes('github/commits')) {
    Object.assign(tools, buildCommitTools(sandbox));
  }

  return tools;
}
```

### Tool Entity

```typescript
interface Tool {
  id: number;
  displayName: string;
  slugName: string;                   // e.g., "posthog/errors"
  isModel: boolean;                   // Is this a model config?
  provider: IntegrationProvider | null;
  workerDefinitionTools: WorkerDefinitionTool[];
}
```

### ToolCall Entity (Execution Tracking)

```typescript
interface ToolCall {
  id: number;
  agent: Agent;
  message: Message;
  createdAt: Date;
  toolName: string;                   // Indexed for queries
  arguments: Record<string, unknown>;
  result: string;
  status: string;                     // Default: 'success'
}
```

---

## Sandbox Environment

Prism uses **Vercel Sandbox** for isolated code execution.

### Sandbox Capabilities

1. **File Operations** - Read, write, list files
2. **Command Execution** - Run bash commands
3. **Git Operations** - Clone, branch, commit
4. **Isolated Environment** - Secure, ephemeral containers

### Sandbox Lifecycle

```
prepareSandbox() → [Agent Work] → cleanupSandbox()
```

### Usage Tracking

```typescript
// Check if sandbox time available
canUseSandbox(organizationId): Promise<{ allowed: boolean; message?: string }>

// Increment usage
incrementSandboxTimeSeconds(organizationId, seconds): Promise<void>
```

---

## Workflow Orchestration

### Main Workflows

#### 1. Standard Agent Workflow

```typescript
async function runAgentWorkflow(payload: AgentJobPayload) {
  const { agentId, prompt, repositoryFullName, toolSlugs, baseBranch } = payload;

  try {
    // 1. Load agent from DB
    const agent = await loadAgent(agentId);

    // 2. Validate and get GitHub token
    const token = await validateAndGetToken(repositoryFullName);

    // 3. Create sandbox environment
    const sandbox = await prepareSandbox(agentId, repositoryFullName, token);

    // 4. Load conversation history
    const messages = await loadPreviousMessages(agentId);

    // 5. Create branch if first run
    await createBranchIfNeeded(agent, sandbox);

    // 6. Run LLM with tools
    const { text, usage } = await runAgentLLM(
      agentId,
      sandbox,
      toolSlugs,
      messages,
      usageAccumulator
    );

    // 7. Save response and track tokens
    await saveAgentResponse(agentId, text, usage);

    // 8. Commit changes if any
    await commitChangesIfNeeded(sandbox, agent, prompt);

    // 9. Cleanup
    await cleanupSandbox(sandbox);

    // 10. Mark complete
    await markAgentComplete(agentId);

  } catch (error) {
    await markAgentFailed(agentId, error);
  }
}
```

#### 2. Orchestrator Agent Workflow

```typescript
async function runOrchestratorAgentWorkflow(payload: AgentJobPayload) {
  // Similar to standard workflow but:
  // - Uses "high" reasoning effort
  // - Has spawn_sub_agent tool
  // - Extracts images from history
  // - Returns reasoning logs
}
```

### Workflow Steps

```typescript
// Step functions
loadAgent(agentId): Promise<Agent>
validateAndGetToken(repoName): Promise<string>
prepareSandbox(agentId, repo, token): Promise<Sandbox>
loadPreviousMessages(agentId): Promise<Message[]>
createBranchIfNeeded(agent, sandbox): Promise<void>
saveAgentResponse(agentId, text, usage): Promise<void>
commitChangesIfNeeded(sandbox, agent, prompt): Promise<void>
cleanupSandbox(sandbox): Promise<void>
markAgentComplete(agentId): Promise<void>
markAgentFailed(agentId, error): Promise<void>
```

---

## LLM Integration

### Configuration

```typescript
// Model: OpenAI GPT-5-mini with reasoning
import { openai } from '@ai-sdk/openai';

const model = openai('gpt-5-mini', {
  reasoning: {
    effort: 'medium' | 'high',  // Standard vs Orchestrator
    summary: 'concise'
  }
});
```

### Token Tracking

```typescript
interface TokenUsageAccumulator {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
}
```

### Standard Agent LLM

```typescript
async function runAgentLLM(
  agentId: number,
  sandbox: Sandbox,
  toolSlugs: string[],
  messages: Message[],
  usageAccumulator: TokenUsageAccumulator
) {
  // Configure model with medium reasoning
  // Combine sandbox tools + dynamic tools
  // Track tokens across all steps
  // Cap at 32 steps via stepCountIs(32)
  // Return final text and usage
}
```

### Orchestrator Agent LLM

```typescript
async function runOrchestratorAgentLLM(
  agentId: number,
  sandbox: Sandbox,
  toolSlugs: string[],
  messages: Message[],
  workspace: Workspace,
  repositoryFullName: string,
  baseBranch: string,
  images: MessageImage[],
  usageAccumulator: TokenUsageAccumulator
) {
  // Configure model with high reasoning
  // Include orchestrator tools (spawn_sub_agent)
  // Extract user images from history
  // Return text + reasoning logs
}
```

### System Prompts

```typescript
// Standard Agent
const AGENT_SYSTEM_PROMPT = `
You are Prism, an asynchronous coding agent. You work on GitHub repositories,
read code, make changes, and explain your steps succinctly.
If the user requests git operations, prefer using tools (update_file, list_files,
read_file, grep).
Avoid destructive operations. Return concise reasoning and resulting changes.
`;

// Orchestrator Agent
const ORCHESTRATOR_AGENT_SYSTEM_PROMPT = `
You are the primary agent of a coding agent, Prism. Prism is an Asynchronous
coding agent platform...
[Instructions for multi-agent orchestration]
Under no circumstances should you finish without creating sub agents,
unless there is truly no further work to be done.
`;
```

---

## Real-time Streaming (SSE)

### Event System Architecture

```typescript
// Event payload types
type AgentEventPayload = {
  type: 'status' | 'error' | 'done' | string;
  phase?: string;
  step?: string;
  detail?: string;
  code?: string;
  reason?: string;
};

// Event envelope (with ID for resumption)
interface AgentEventEnvelope {
  id: number;           // Auto-incrementing
  timestamp: Date;
  payload: AgentEventPayload;
}

// In-memory event store
const eventStates = new Map<number, {
  events: AgentEventEnvelope[];  // Max 500
  listeners: Set<Function>;
  lastActivity: Date;
}>();
```

### Event Publishing

```typescript
// Publish event to all listeners
function publishAgentEvent(agentId: number, payload: AgentEventPayload): void

// Convenience emitters
function emitStatus(agentId, phase, step, detail?, metadata?): Promise<void>
function emitError(agentId, code, message): Promise<void>
function emitDone(agentId, reason): Promise<void>
```

### SSE Endpoint

```typescript
// GET /sse/agent/:agentId
router.get('/agent/:agentId', async (req, res) => {
  // 1. Set SSE headers
  res.setHeader('Content-Type', 'text/event-stream');
  res.setHeader('Cache-Control', 'no-cache');
  res.flushHeaders();

  // 2. Handle last-event-id for resumption
  const lastEventId = req.headers['last-event-id'] || req.query['last-event-id'];

  // 3. Send backlog if resuming
  if (lastEventId !== '$') {
    const history = readHistorySince(agentId, lastEventId);
    for (const event of history) {
      sendEvent(res, event);
    }
  }

  // 4. Subscribe to new events
  const unsubscribe = subscribeToAgentEvents(agentId, (event) => {
    sendEvent(res, event);
    if (event.payload.type === 'done') {
      cleanup();
    }
  });

  // 5. Handle disconnect
  req.on('close', cleanup);
});
```

### Reasoning Streamer

```typescript
// Stream intermediate reasoning from LLM
function createReasoningStreamer(agentId: number) {
  return (step: ReasoningStep) => {
    const text = extractReasoningText(step);
    if (text) {
      emitStatus(agentId, 'reasoning', 'thinking', text);
    }
  };
}
```

---

## Database Schema

### Entity Relationship Diagram

```
Organization (1) ──── (*) OrganizationMember
     │
     ├──── (*) Workspace ──── (*) Agent ──── (*) Message ──── (*) ToolCall
     │           │
     │           └──── (*) WorkspaceTool ──── (1) Tool
     │
     ├──── (*) WorkerDefinition ──── (*) WorkerDefinitionTool ──── (1) Tool
     │
     ├──── (*) IntegrationConnection ──── (1) IntegrationProvider
     │
     └──── (*) SlackUserMapping
```

### All Entities

#### Organization

```typescript
interface Organization {
  id: number;
  workosOrganizationId: string;          // Unique, indexed
  name: string;                          // Default: "Personal Workspace"

  // Stripe
  stripeCustomerId?: string;
  stripeSubscriptionId?: string;
  subscriptionTier: 'FREE' | 'PAID' | 'ENTERPRISE';
  subscriptionStatus: 'ACTIVE' | 'CANCELED' | 'PAST_DUE' | 'INCOMPLETE';
  cancelAtPeriodEnd: boolean;

  // Usage tracking
  tokenCostUsedMicrodollars: number;     // Default: 0
  tokenCostLimitMicrodollars: number;    // Default: 5,000,000 ($5)
  sandboxTimeUsedSeconds: number;        // Default: 0
  sandboxTimeLimitSeconds: number;       // Default: 900 (15 min)

  // Billing period
  billingPeriodStart: Date;
  billingPeriodEnd: Date;

  createdAt: Date;
  updatedAt: Date;
  members: OrganizationMember[];
}
```

#### Workspace

```typescript
interface Workspace {
  id: number;
  name: string;                          // Default: "Untitled"
  organizationId: number;
  organization: Organization;
  githubRepositoryName: string;
  workerId?: number;                     // Link to worker template
  currentBranch: string;                 // Default: "main", indexed
  slackChannelId?: string;
  slackMessageTs?: string;
  providerAgents: Agent[];
  createdAt: Date;
}
```

#### Agent

```typescript
interface Agent {
  id: number;
  workspace: Workspace;
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED';
  providerType: 'PRISM' | 'CURSOR' | 'JULES';
  conversationId: string;
  url: string;
  githubBranchName?: string;
  name: string;                          // Default: "Untitled"
  model?: string;
  sandboxId?: string;
  isOrchestratorAgent: boolean;          // Default: false
  createdAt: Date;
  updatedAt: Date;
  messages: Message[];
}
```

#### Message

```typescript
interface MessageImage {
  data: string;      // Base64
  mimeType: string;  // e.g., "image/png"
}

interface Message {
  id: number;
  createdAt: Date;
  agent: Agent;
  content: string;
  sender: 'USER' | 'AGENT';
  images: MessageImage[];                // Default: []

  // Token tracking
  promptTokens: number;                  // Default: 0
  completionTokens: number;              // Default: 0
  totalTokens: number;                   // Default: 0
  costMicrodollars: number;              // Default: 0

  // Metadata
  error?: string;
  model: string;                         // Default: "gpt-5-mini"
  sandboxDurationMs: number;             // Default: 0
  toolCalls: ToolCall[];
}
```

#### WorkerDefinition

```typescript
interface WorkerDefinition {
  id: number;
  slug: string;                          // Unique per org
  prompt: string;
  organizationId: number;
  organization: Organization;
  key?: string;                          // Webhook auth key
  cloudProviders: Array<{
    provider: string;
    model?: string;
  }>;
  tools: WorkerDefinitionTool[];
  createdAt: Date;
}
```

#### Tool

```typescript
interface Tool {
  id: number;
  displayName: string;
  slugName: string;                      // e.g., "posthog/errors"
  isModel: boolean;
  provider?: IntegrationProvider;
  workerDefinitionTools: WorkerDefinitionTool[];
}
// Unique constraint: [displayName, provider, slugName]
```

#### IntegrationConnection

```typescript
interface IntegrationConnection {
  id: number;
  organizationId: number;
  organization: Organization;
  provider: IntegrationProvider;
  externalId: string;                    // Default: ""
  data: Record<string, string>;          // JSONB, encrypted
  createdAt: Date;
  updatedAt: Date;

  // Methods
  setDataConfig(input: Record<string, string>): void;  // Encrypts
  getDataConfig(): Record<string, string>;             // Decrypts
}
// Unique constraint: [organizationId, provider, externalId]
```

#### ToolCall

```typescript
interface ToolCall {
  id: number;
  agent: Agent;
  message: Message;
  createdAt: Date;
  toolName: string;                      // Indexed
  arguments: Record<string, unknown>;    // JSON
  result: string;                        // Default: ""
  status: string;                        // Default: "success"
}
```

### Database Configuration

```typescript
// PostgreSQL connection
const AppDataSource = new DataSource({
  type: 'postgresql',
  url: `postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/${PGDATABASE}`,
  ssl: PGSSL === 'true' ? { rejectUnauthorized: false } : false,
  synchronize: NODE_ENV !== 'production',
  entities: [
    Agent, IntegrationConnection, IntegrationProvider, Message,
    Organization, OrganizationMember, SlackUserMapping, Tool,
    ToolCall, WorkerDefinition, WorkerDefinitionTool, Workspace,
    WorkspaceTool
  ]
});
```

---

## Authentication

### WorkOS Integration

```typescript
import { WorkOS } from '@workos-inc/node';

// Environment variables
const WORKOS_API_KEY = process.env.WORKOS_API_KEY;
const WORKOS_CLIENT_ID = process.env.WORKOS_CLIENT_ID;
const WORKOS_COOKIE_PASSWORD = process.env.WORKOS_COOKIE_PASSWORD;
const WORKOS_REDIRECT_URI = process.env.WORKOS_REDIRECT_URI || 'http://localhost:5001/api/auth/callback';

// Initialize
export const workos = new WorkOS(WORKOS_API_KEY, {
  clientId: WORKOS_CLIENT_ID
});

export const COOKIE_NAME = 'wos-session';
```

### Auth Flow

1. User clicks login
2. Redirect to WorkOS OAuth
3. Callback with auth code
4. Exchange for session token
5. Set secure cookie
6. Load organization context

---

## Integrations

### Supported Integrations

| Integration | Slug | Features |
|------------|------|----------|
| **GitHub App** | `github_app` | Repos, branches, commits, webhooks |
| **PostHog** | `posthog` | Errors, insights, analytics, docs |
| **Cursor** | `cursor` | External agent API |
| **Jules** | `jules` | External agent API |
| **Slack** | `slack` | Notifications, user mapping |

### Integration tRPC Router

```typescript
// List all integrations with connection status
integrations.list()

// Connect integration
integrations.connect({
  providerSlug: 'github_app',
  data: { installation_id: '12345' }
})

// Disconnect
integrations.disconnect({ id })

// Get GitHub repos
integrations.repositories()

// Get repo branches
integrations.branches({ repositoryFullName: 'owner/repo' })
```

### GitHub Service

```typescript
async function getGithubTokenForUser(organizationId: number): Promise<string> {
  // 1. Find integration connection
  const connection = await connectionRepository.findOne({
    where: { organizationId, provider: { slug: 'github_app' } }
  });

  // 2. Get installation ID from encrypted config
  const installationId = connection.getDataConfig().installation_id;

  // 3. Get installation token
  return await getInstallationToken(installationId);
}
```

---

## Payment & Billing

### Subscription Tiers

| Tier | Token Limit | Sandbox Time | Price |
|------|-------------|--------------|-------|
| **Lite (Free)** | $5.00 | 2 hours | $0/mo |
| **Pro** | $20.00 | 24 hours | $20/mo |
| **Enterprise** | Unlimited | Unlimited | Custom |

### Plan Configuration

```typescript
interface PlanConfig {
  tier: SubscriptionTier;
  name: string;
  description: string;
  tokenCostLimitMicrodollars: number;  // -1 for unlimited
  sandboxTimeLimitSeconds: number;      // -1 for unlimited
  priceMonthly: number;
  stripePriceId?: string;
  perks: PlanPerk[];
}

const PLANS: Record<SubscriptionTier, PlanConfig> = {
  FREE: {
    tier: 'FREE',
    name: 'Lite',
    tokenCostLimitMicrodollars: 5_000_000,  // $5
    sandboxTimeLimitSeconds: 7200,           // 2 hours
    priceMonthly: 0,
    perks: [
      { name: 'Slack Bot', included: true },
      { name: 'Priority Support', included: false },
      { name: 'Premium Models', included: false }
    ]
  },
  PAID: {
    tier: 'PAID',
    name: 'Pro',
    tokenCostLimitMicrodollars: 20_000_000,  // $20
    sandboxTimeLimitSeconds: 86400,           // 24 hours
    priceMonthly: 20,
    stripePriceId: 'price_xxx',
    perks: [
      { name: 'Slack Bot', included: true },
      { name: 'Priority Support', included: true },
      { name: 'Premium Models', included: true }
    ]
  },
  ENTERPRISE: {
    tier: 'ENTERPRISE',
    name: 'Enterprise',
    tokenCostLimitMicrodollars: -1,
    sandboxTimeLimitSeconds: -1,
    priceMonthly: -1,  // Custom
    perks: [
      { name: 'Dedicated Slack Channel', included: true },
      { name: 'Custom Features', included: true }
    ]
  }
};
```

### Usage Tracking

```typescript
// Check if can send message
async function canSendMessage(orgId: number): Promise<{
  allowed: boolean;
  message?: string;
}> {
  const org = await getOrganization(orgId);
  await resetBillingPeriodIfNeeded(org);

  if (org.tokenCostUsedMicrodollars >= org.tokenCostLimitMicrodollars) {
    return {
      allowed: false,
      message: `Cost limit reached. Used $${used}/$${limit}`
    };
  }
  return { allowed: true };
}

// Increment token cost
async function incrementTokenCostMicrodollars(
  orgId: number,
  microdollars: number
): Promise<void>

// Check sandbox availability
async function canUseSandbox(orgId: number): Promise<{
  allowed: boolean;
  message?: string;
}>

// Increment sandbox time
async function incrementSandboxTimeSeconds(
  orgId: number,
  seconds: number
): Promise<void>

// Auto-reset billing period (free tier)
async function resetBillingPeriodIfNeeded(org: Organization): Promise<void>

// Get usage stats
async function getUsageStats(orgId: number): Promise<{
  tokenCostPercentage: number;
  sandboxTimePercentage: number;
  billingPeriodStart: Date;
  billingPeriodEnd: Date;
}>
```

### Stripe Integration

- Customer creation
- Subscription management
- Webhook handling for:
  - `invoice.paid` - Reset usage
  - `customer.subscription.updated` - Tier changes
  - `customer.subscription.deleted` - Cancellation

---

## Webhooks

### Webhook Endpoints

#### 1. GitHub Events (`/webhooks/github/events`)

Triggers on issue comments containing `--prism`:

```typescript
// Signature verification
const signature = req.headers['x-hub-signature-256'];
const expected = 'sha256=' + crypto
  .createHmac('sha256', GITHUB_WEBHOOK_SECRET)
  .update(payload)
  .digest('hex');

// Payload processing
if (event === 'issue_comment' && body.comment.body.includes('--prism')) {
  // Extract worker slug from: --prism worker-slug
  const workerSlug = extractWorkerSlug(body.comment.body);

  // Create workspace from webhook
  await createWorkspaceFromWebhook({
    workerSlug,
    repositoryFullName: body.repository.full_name,
    prompt: body.comment.body,
    organizationId
  });
}
```

#### 2. PostHog Issue (`/webhooks/posthog/issue`)

Triggers on new errors:

```typescript
// Auth via worker key
const { slug, key } = req.query;
const worker = await findWorker(slug);
if (worker.key !== key) {
  return res.status(403).send('Invalid key');
}

// Create workspace with error context
await createWorkspaceFromWebhook({
  workerSlug: slug,
  prompt: `New PostHog error: ${body.error}`,
  ...
});
```

#### 3. Cursor Completion (`/webhooks/cursor/complete/:agentId`)

```typescript
const { status } = req.body;  // 'FINISHED' | 'FAILED'

await updateAgentStatus(agentId, status === 'FINISHED' ? 'COMPLETED' : 'FAILED');
```

#### 4. Stripe (`/webhooks/stripe`)

Handled by dedicated `handleStripeWebhook` function.

### Workspace Creation from Webhook

```typescript
async function createWorkspaceFromWebhook(params: {
  workerSlug: string;
  repositoryFullName: string;
  prompt: string;
  organizationId: number;
}) {
  // 1. Find worker definition
  const worker = await findWorker(params.workerSlug);

  // 2. Create workspace
  const workspace = await createWorkspace({
    name: generateTitle(params.prompt),
    githubRepositoryName: params.repositoryFullName,
    organizationId: params.organizationId,
    workerId: worker.id
  });

  // 3. Associate tools
  for (const toolSlug of worker.toolSlugs) {
    await createWorkspaceTool(workspace.id, toolSlug);
  }

  // 4. Create agent from provider config
  const provider = worker.cloudProviders[0];
  await createAgent({
    workspace,
    provider,
    prompt: params.prompt,
    toolSlugs: worker.toolSlugs,
    isOrchestratorAgent: false
  });
}
```

---

## API Routes (tRPC)

### Router Structure

```typescript
// Main router
const appRouter = router({
  workspace: workspaceRouter,
  workers: workersRouter,
  integrations: integrationsRouter,
  organization: organizationRouter,
  payment: paymentRouter
});
```

### Workspace Router

```typescript
workspaceRouter = router({
  // List workspaces
  list: protectedProcedure.query(async ({ ctx }) => {
    return getWorkspaces(ctx.organizationId);
  }),

  // Get single workspace
  get: protectedProcedure
    .input(z.object({ id: z.number() }))
    .query(async ({ input }) => {
      return getWorkspace(input.id);
    }),

  // Create workspace with agent
  create: protectedProcedure
    .input(z.object({
      repositoryFullName: z.string(),
      toolSlugs: z.array(z.string()),
      cloudProviders: z.array(providerConfig),
      baseBranch: z.string(),
      message: z.string(),
      sub_agents: z.boolean().optional(),
      images: z.array(imageSchema).optional()
    }))
    .mutation(async ({ input, ctx }) => {
      // Check usage limits
      const canSend = await canSendMessage(ctx.organizationId);
      if (!canSend.allowed) throw new TRPCError({ code: 'FORBIDDEN' });

      // Generate title via LLM
      const name = await generateWorkspaceTitle(input.message);

      // Create workspace and agent
      // ...
    }),

  // Get agent messages
  messages: protectedProcedure
    .input(z.object({ agentId: z.number() }))
    .query(async ({ input }) => {
      const agent = await getAgent(input.agentId);
      const provider = getProvider(agent.providerType);
      return provider.getMessages(agent);
    }),

  // Send follow-up message
  sendMessage: protectedProcedure
    .input(z.object({
      agentId: z.number(),
      message: z.string(),
      images: z.array(imageSchema).optional()
    }))
    .mutation(async ({ input, ctx }) => {
      // Check limits and send
    }),

  // Get agent status
  agentStatus: protectedProcedure
    .input(z.object({ agentId: z.number() }))
    .query(async ({ input }) => {
      return getAgentStatus(input.agentId);
    }),

  // Create/get branch name
  createBranch: protectedProcedure
    .input(z.object({ agentId: z.number() }))
    .mutation(async ({ input }) => {
      return createOrGetBranch(input.agentId);
    })
});
```

### Workers Router

```typescript
workersRouter = router({
  list: protectedProcedure.query(...),
  create: protectedProcedure.input(workerInput).mutation(...),
  update: protectedProcedure.input(workerUpdateInput).mutation(...),
  delete: protectedProcedure.input(z.object({ id: z.number() })).mutation(...)
});
```

### Integrations Router

```typescript
integrationsRouter = router({
  list: protectedProcedure.query(...),
  connect: protectedProcedure.input(connectInput).mutation(...),
  disconnect: protectedProcedure.input(z.object({ id: z.number() })).mutation(...),
  repositories: protectedProcedure.query(...),
  branches: protectedProcedure.input(z.object({ repositoryFullName: z.string() })).query(...)
});
```

---

## Frontend Architecture

### Tech Stack

- **React** with TypeScript
- **Vite** for bundling
- **Bun** as package manager
- **tRPC** for API calls

### Page Structure

| Page | Purpose |
|------|---------|
| `/` (Home) | Dashboard |
| `/login` | Authentication |
| `/organization` | Org settings |
| `/workspace/:id` | Agent chat interface |
| `/workers` | Worker management |
| `/integrations` | Integration setup |
| `/usage` | Billing & usage stats |

### Key Components

```
components/
├── CreateBranch/    # Branch creation UI
└── Sidebar/         # Navigation sidebar
```

### Libraries

```typescript
// lib/trpc.ts - tRPC client setup
export const trpc = createTRPCReact<AppRouter>();

// lib/types.ts - Shared type definitions

// lib/useAuth.tsx - Authentication hook
export function useAuth() {
  // Session management
  // Organization context
  // Login/logout functions
}
```

---

## Implementation Checklist

### Phase 1: Core Infrastructure

- [ ] **Database Setup**
  - [ ] PostgreSQL with TypeORM
  - [ ] All 13 entities
  - [ ] Data source configuration
  - [ ] Seed data

- [ ] **Authentication**
  - [ ] WorkOS integration
  - [ ] Session management
  - [ ] Organization context

- [ ] **tRPC Setup**
  - [ ] Router configuration
  - [ ] Context with auth
  - [ ] Error handling

### Phase 2: Agent System

- [ ] **Agent Management**
  - [ ] Agent entity and status tracking
  - [ ] Provider interface
  - [ ] Prism provider implementation

- [ ] **Sandbox Environment**
  - [ ] Vercel Sandbox integration
  - [ ] Sandbox tools (file ops)
  - [ ] Usage tracking

- [ ] **Workflow Orchestration**
  - [ ] Standard agent workflow
  - [ ] Orchestrator workflow
  - [ ] Step functions

- [ ] **LLM Integration**
  - [ ] OpenAI SDK setup
  - [ ] Token tracking
  - [ ] Reasoning configuration

### Phase 3: Real-time Features

- [ ] **Event System**
  - [ ] In-memory event store
  - [ ] Event publishing
  - [ ] Cleanup logic

- [ ] **SSE Streaming**
  - [ ] SSE endpoint
  - [ ] Event resumption
  - [ ] Client integration

### Phase 4: Tools

- [ ] **Sandbox Tools**
  - [ ] listFiles
  - [ ] readFile
  - [ ] updateFile
  - [ ] grep

- [ ] **GitHub Tools**
  - [ ] github_list_commits
  - [ ] github_view_commit

- [ ] **PostHog Tools**
  - [ ] Query runner
  - [ ] Insights
  - [ ] Errors
  - [ ] Documentation

- [ ] **Orchestrator Tools**
  - [ ] spawn_sub_agent

- [ ] **Dynamic Tool Building**
  - [ ] Slug-based loading
  - [ ] Tool collection merging

### Phase 5: Workers

- [ ] **Worker Definitions**
  - [ ] CRUD operations
  - [ ] Tool associations
  - [ ] Provider configs

- [ ] **Webhook Triggers**
  - [ ] GitHub issue comments
  - [ ] PostHog errors
  - [ ] Signature verification

### Phase 6: Integrations

- [ ] **GitHub App**
  - [ ] OAuth flow
  - [ ] Installation tokens
  - [ ] Repository/branch listing

- [ ] **PostHog**
  - [ ] API key storage
  - [ ] Toolkit integration

- [ ] **Cursor/Jules**
  - [ ] API integration
  - [ ] Provider implementation

- [ ] **Slack** (optional)
  - [ ] User mapping
  - [ ] Notifications

### Phase 7: Billing

- [ ] **Usage Tracking**
  - [ ] Token costs
  - [ ] Sandbox time
  - [ ] Billing periods

- [ ] **Subscription Plans**
  - [ ] Plan configuration
  - [ ] Limit enforcement

- [ ] **Stripe Integration**
  - [ ] Customer management
  - [ ] Subscriptions
  - [ ] Webhooks

### Phase 8: Frontend

- [ ] **Pages**
  - [ ] Home/Dashboard
  - [ ] Workspace (chat)
  - [ ] Workers
  - [ ] Integrations
  - [ ] Usage

- [ ] **Components**
  - [ ] Sidebar
  - [ ] Message display
  - [ ] Tool call visualization
  - [ ] SSE connection

---

## Key Implementation Notes

### 1. Token Cost Calculation
Use microdollars (1/1,000,000 of a dollar) for precision:
```typescript
const COST_PER_1K_TOKENS = 0.01;  // $0.01 per 1K tokens
const costMicrodollars = Math.round(totalTokens * COST_PER_1K_TOKENS * 1000);
```

### 2. Branch Naming
Format: `prism/{agent-id}-{sanitized-prompt}`

### 3. Event ID Format
Auto-incrementing integers per agent, stored in envelope.

### 4. Tool Slugs
Format: `{provider}/{tool-name}` (e.g., `posthog/errors`, `github/commits`)

### 5. Encryption
Use `setDataConfig`/`getDataConfig` for sensitive data in IntegrationConnection.

---

## Environment Variables

```bash
# Database
PGHOST=localhost
PGPORT=5432
PGUSER=postgres
PGPASSWORD=secret
PGDATABASE=prism
PGSSL=false

# Auth
WORKOS_API_KEY=sk_xxx
WORKOS_CLIENT_ID=client_xxx
WORKOS_COOKIE_PASSWORD=32-char-secret
WORKOS_REDIRECT_URI=http://localhost:5001/api/auth/callback

# LLM
OPENAI_API_KEY=sk-xxx

# Analytics
POSTHOG_API_KEY=phc_xxx

# Payments
STRIPE_SECRET_KEY=sk_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx

# GitHub
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----...
GITHUB_WEBHOOK_SECRET=xxx
```

---

*This document provides a comprehensive analysis of the Prism platform for implementation reference.*
