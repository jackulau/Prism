---
id: cloudprovider-interface
name: CloudProvider Interface Definition
wave: 1
priority: 1
dependencies: []
estimated_hours: 2
tags:
- backend
- interface
- core
---

## Objective

Define the abstract `CloudProvider` interface that standardizes how the application interacts with cloud-based agent backends (Claude.ai, etc.).

## Context

The Prism codebase already has a well-established `llm.Provider` interface for LLM chat providers. The new `CloudProvider` interface serves a different purpose: it abstracts cloud-based agent services that can create, manage, and communicate with AI agents. This follows the existing pattern of interface-based provider abstraction.

**Key difference from llm.Provider:**
- `llm.Provider` handles direct LLM chat (local API calls)
- `CloudProvider` handles remote agent lifecycle management (cloud-hosted agents)

## Implementation

1. **Create new package**: `backend/internal/cloudprovider/`

2. **Create interface file**: `backend/internal/cloudprovider/provider.go`
   ```go
   package cloudprovider
   
   import (
       "context"
   )
   
   // CloudProvider defines the interface for cloud-based agent backends
   type CloudProvider interface {
       // Name returns the provider name (e.g., "claude-cloud", "openai-assistants")
       Name() string
       
       // CreateAgent creates a new agent with the given parameters
       CreateAgent(ctx context.Context, params CreateAgentParams) (*Agent, error)
       
       // GetAgent retrieves an agent by ID
       GetAgent(ctx context.Context, agentID string) (*Agent, error)
       
       // DeleteAgent removes an agent
       DeleteAgent(ctx context.Context, agentID string) error
       
       // GetMessages retrieves all messages for an agent's conversation
       GetMessages(ctx context.Context, agent *Agent) ([]ProviderMessage, error)
       
       // SendMessage sends a message to an agent and returns success status
       SendMessage(ctx context.Context, agent *Agent, message string, images []ImageData) (bool, error)
       
       // StreamMessages returns a channel for streaming agent responses
       StreamMessages(ctx context.Context, agent *Agent) (<-chan MessageChunk, error)
       
       // ValidateCredentials validates the provider credentials
       ValidateCredentials(ctx context.Context) error
       
       // HasCredentials returns whether credentials are configured
       HasCredentials() bool
   }
   ```

3. **Create types file**: `backend/internal/cloudprovider/types.go`
   - Define `Agent` struct (ID, ProviderID, Name, Status, CreatedAt, etc.)
   - Define `CreateAgentParams` (Name, SystemPrompt, Model, Tools, etc.)
   - Define `ProviderMessage` (ID, Role, Content, Timestamp, ToolCalls, etc.)
   - Define `ImageData` (URL, Base64, MimeType) - reuse from llm package if possible
   - Define `MessageChunk` for streaming responses
   - Define `AgentStatus` enum (active, idle, terminated)

4. **Create errors file**: `backend/internal/cloudprovider/errors.go`
   - Define standard errors: ErrAgentNotFound, ErrUnauthorized, ErrRateLimited, etc.

## Acceptance Criteria

- [ ] `CloudProvider` interface is defined with all required methods
- [ ] All supporting types are defined and documented
- [ ] Types follow existing Go conventions in the codebase
- [ ] No import cycles with existing packages
- [ ] Interface is generic enough to support multiple cloud providers

## Files to Create/Modify

- `backend/internal/cloudprovider/provider.go` - Interface definition
- `backend/internal/cloudprovider/types.go` - Supporting types
- `backend/internal/cloudprovider/errors.go` - Error definitions

## Integration Points

- **Provides**: CloudProvider interface for cloud agent implementations
- **Consumes**: None (core dependency)
- **Conflicts**: None - new package
