---
id: sandbox-list-files
name: Sandbox Tool - listFiles
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

Create the `listFiles` sandbox tool that lists files in a repository directory with detailed metadata.

## Context

This tool allows agents to explore the file structure of a user's workspace. It executes `ls -la` in the sandbox environment and returns structured file information. This is a foundational tool for file-based operations.

## Implementation

1. **Tool Implementation** - Create `backend/internal/tools/builtin/list_files_sandbox.go`:
   ```go
   type ListFilesSandboxTool struct {
       sandbox *sandbox.Service
   }

   func NewListFilesSandboxTool(sandbox *sandbox.Service) *ListFilesSandboxTool

   func (t *ListFilesSandboxTool) Name() string { return "sandbox_list_files" }

   func (t *ListFilesSandboxTool) Description() string {
       return "List files and directories in a repository path with detailed metadata"
   }

   func (t *ListFilesSandboxTool) Parameters() llm.JSONSchema {
       return llm.JSONSchema{
           Type: "object",
           Properties: map[string]llm.JSONProperty{
               "relativePath": {
                   Type:        "string",
                   Description: "Relative path within the repository to list (use '.' for root)",
               },
           },
           Required: []string{"relativePath"},
       }
   }

   func (t *ListFilesSandboxTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
       // 1. Extract userID from context
       // 2. Get relativePath from params
       // 3. Validate path (no directory traversal)
       // 4. Get user's work directory
       // 5. Execute: ls -la {fullPath}
       // 6. Parse output into structured format
       // 7. Return FileEntry slice
   }

   func (t *ListFilesSandboxTool) RequiresConfirmation() bool { return false }
   ```

2. **Output Structure**:
   ```go
   type FileEntry struct {
       Name        string `json:"name"`
       IsDirectory bool   `json:"isDirectory"`
       Size        int64  `json:"size"`
       Permissions string `json:"permissions"`
       ModTime     string `json:"modTime"`
   }
   ```

3. **Path Validation** - Ensure no directory traversal:
   - Reject paths containing `..`
   - Reject absolute paths
   - Normalize path using `filepath.Clean`

4. **Command Execution**:
   - Use sandbox's command execution capability
   - Parse `ls -la` output format
   - Handle errors gracefully (directory not found, permission denied)

5. **Registration** - Add to `backend/internal/tools/builtin/init.go`:
   ```go
   registry.Register(NewListFilesSandboxTool(sandbox))
   ```

## Acceptance Criteria

- [ ] Tool correctly lists files in specified directory
- [ ] Returns structured FileEntry objects with metadata
- [ ] Path traversal attacks are prevented
- [ ] Works with empty directories
- [ ] Handles non-existent paths gracefully with clear error
- [ ] Added to read-only tools list for auto-approval
- [ ] No breaking changes to existing file_list tool

## Files to Create/Modify

- `backend/internal/tools/builtin/list_files_sandbox.go` - Create (new file)
- `backend/internal/tools/builtin/init.go` - Register tool
- `backend/internal/tools/approval.go` - Add to ReadOnlyTools if appropriate

## Integration Points

- **Provides**: File listing capability for sandbox environments
- **Consumes**: Sandbox service for directory operations
- **Conflicts**: Avoid modifying existing `file_list.go` - this is a new sandbox-specific variant
