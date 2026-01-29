---
id: sandbox-grep
name: Sandbox Tool - grep
wave: 1
priority: 1
dependencies: []
estimated_hours: 2
tags:
- backend
- tools
- sandbox
---

## Objective

Create the `grep` sandbox tool that searches for patterns in files within a repository.

## Context

This tool allows agents to search for text patterns across files in a user's workspace. It executes grep commands in the sandbox environment and returns structured search results. This is essential for code navigation and understanding codebases.

## Implementation

1. **Tool Implementation** - Create `backend/internal/tools/builtin/grep_sandbox.go`:
   ```go
   type GrepSandboxTool struct {
       sandbox *sandbox.Service
   }

   func NewGrepSandboxTool(sandbox *sandbox.Service) *GrepSandboxTool

   func (t *GrepSandboxTool) Name() string { return "sandbox_grep" }

   func (t *GrepSandboxTool) Description() string {
       return "Search for patterns in files within the repository using grep"
   }

   func (t *GrepSandboxTool) Parameters() llm.JSONSchema {
       return llm.JSONSchema{
           Type: "object",
           Properties: map[string]llm.JSONProperty{
               "command": {
                   Type:        "string",
                   Description: "The grep command to execute (e.g., 'grep -r \"pattern\" .')",
               },
           },
           Required: []string{"command"},
       }
   }

   func (t *GrepSandboxTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
       // 1. Extract userID from context
       // 2. Get command from params
       // 3. Validate command (must start with grep)
       // 4. Sanitize command (prevent command injection)
       // 5. Get user's work directory
       // 6. Execute grep command
       // 7. Parse output into structured format
       // 8. Return matches
   }

   func (t *GrepSandboxTool) RequiresConfirmation() bool { return false }
   ```

2. **Output Structure**:
   ```go
   type GrepMatch struct {
       File       string `json:"file"`
       LineNumber int    `json:"lineNumber"`
       Content    string `json:"content"`
   }

   type GrepResult struct {
       Matches    []GrepMatch `json:"matches"`
       TotalCount int         `json:"totalCount"`
       Truncated  bool        `json:"truncated"`
   }
   ```

3. **Command Validation**:
   - Ensure command starts with `grep`
   - Block dangerous flags or patterns
   - Prevent command injection (no `|`, `;`, `&&`, etc.)
   - Whitelist allowed grep flags: `-r`, `-n`, `-i`, `-l`, `-c`, `-w`, `-E`, `-P`

4. **Command Sanitization**:
   ```go
   func sanitizeGrepCommand(command string) (string, error) {
       // Parse command
       // Validate it's a grep command
       // Remove/block dangerous patterns
       // Return sanitized command
   }
   ```

5. **Output Limits**:
   - Limit number of matches returned (e.g., 1000)
   - Set truncated flag if results exceed limit
   - Include total count even if truncated

6. **Error Handling**:
   - Handle "no matches found" gracefully (exit code 1)
   - Handle invalid regex patterns
   - Handle permission denied errors

7. **Registration** - Add to `backend/internal/tools/builtin/init.go`:
   ```go
   registry.Register(NewGrepSandboxTool(sandbox))
   ```

## Acceptance Criteria

- [ ] Tool correctly executes grep commands
- [ ] Returns structured match results with file, line, content
- [ ] Command injection is prevented
- [ ] Dangerous flags are blocked
- [ ] Output is limited to prevent memory issues
- [ ] No matches returns empty result (not error)
- [ ] Added to read-only tools list for auto-approval
- [ ] Works with common grep flags (-r, -n, -i, etc.)

## Files to Create/Modify

- `backend/internal/tools/builtin/grep_sandbox.go` - Create (new file)
- `backend/internal/tools/builtin/init.go` - Register tool
- `backend/internal/tools/approval.go` - Add to ReadOnlyTools if appropriate

## Integration Points

- **Provides**: Pattern search capability for sandbox environments
- **Consumes**: Sandbox service for command execution
- **Conflicts**: Avoid modifying existing `grep.go` - this is a new sandbox-specific variant that takes raw command input
