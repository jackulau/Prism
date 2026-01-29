---
id: dynamic-tool-builder
name: Dynamic Tool Building Function
wave: 2
priority: 2
dependencies:
- tool-entity
estimated_hours: 4
tags:
- backend
- tools
- runtime
---

## Objective

Implement the `buildDynamicTools` function for runtime tool collection assembly based on agent configuration and tool slugs.

## Context

The dynamic tool builder enables agents to have customizable toolsets. Given an agent ID and list of tool slugs, it assembles the appropriate tools at runtime. This is essential for flexible agent configurations where different agents may need different tool combinations.

## Implementation

1. **Tool Builder Interface** - Create `backend/internal/tools/builder/builder.go`:
   ```go
   type ToolBuilder interface {
       // Slug returns the unique identifier for this tool builder
       Slug() string
       // Build creates a Tool instance with the given sandbox
       Build(sandbox *sandbox.Service) tools.Tool
       // DependsOn returns slugs of tools this builder depends on
       DependsOn() []string
   }
   ```

2. **Tool Builder Registry** - Create `backend/internal/tools/builder/registry.go`:
   ```go
   type BuilderRegistry struct {
       builders map[string]ToolBuilder
       mu       sync.RWMutex
   }

   func (r *BuilderRegistry) Register(builder ToolBuilder) error
   func (r *BuilderRegistry) Get(slug string) (ToolBuilder, bool)
   func (r *BuilderRegistry) List() []ToolBuilder
   ```

3. **Dynamic Builder Function** - Create `backend/internal/tools/builder/dynamic.go`:
   ```go
   // buildDynamicTools assembles a tool collection for an agent
   func BuildDynamicTools(
       agentID string,
       toolSlugs []string,
       sandbox *sandbox.Service,
       registry *BuilderRegistry,
   ) (*tools.Registry, error) {
       // 1. Parse and validate tool slugs
       // 2. Resolve dependencies (topological sort)
       // 3. Load matching tool builders
       // 4. Build tools and add to new registry
       // 5. Return assembled tool collection
   }
   ```

4. **Slug Parser** - Handle slugs like "posthog/errors", "sandbox/listFiles":
   ```go
   type ParsedSlug struct {
       Provider string // e.g., "posthog", "sandbox"
       Tool     string // e.g., "errors", "listFiles"
   }

   func ParseSlug(slug string) (*ParsedSlug, error)
   ```

5. **Register Built-in Builders** - Update `backend/internal/tools/builtin/init.go`:
   - Create builders for existing tools
   - Register them with the BuilderRegistry

## Acceptance Criteria

- [ ] ToolBuilder interface is defined and documented
- [ ] BuilderRegistry supports thread-safe registration and lookup
- [ ] BuildDynamicTools correctly parses slugs and assembles tools
- [ ] Dependency resolution works correctly (topological sort)
- [ ] Error handling for missing or invalid slugs
- [ ] Existing tool registration remains unchanged (backwards compatible)

## Files to Create/Modify

- `backend/internal/tools/builder/builder.go` - Create (new file)
- `backend/internal/tools/builder/registry.go` - Create (new file)
- `backend/internal/tools/builder/dynamic.go` - Create (new file)
- `backend/internal/tools/builder/slug.go` - Create (new file)
- `backend/internal/tools/builtin/init.go` - Add builder registration

## Integration Points

- **Provides**: Dynamic tool assembly for agents
- **Consumes**: Tool entity (from tool-entity task), sandbox service, tool registry
- **Conflicts**: Avoid modifying core tools/registry.go structure
