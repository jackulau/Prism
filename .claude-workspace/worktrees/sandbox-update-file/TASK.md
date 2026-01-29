---
id: sandbox-update-file
name: Sandbox Tool - updateFile
wave: 1
priority: 1
dependencies: []
estimated_hours: 3
tags:
- backend
- tools
- sandbox
---

## Objective

Create the `updateFile` sandbox tool that writes or overwrites file contents in a repository.

## Context

This tool allows agents to modify files in a user's workspace. It's a write operation that requires user confirmation before execution. The tool should support creating new files and overwriting existing ones, with proper backup/history tracking.

## Implementation

1. **Tool Implementation** - Create `backend/internal/tools/builtin/update_file_sandbox.go`:
   ```go
   type UpdateFileSandboxTool struct {
       sandbox     *sandbox.Service
       historyRepo *repository.FileHistoryRepository
   }

   func NewUpdateFileSandboxTool(
       sandbox *sandbox.Service,
       historyRepo *repository.FileHistoryRepository,
   ) *UpdateFileSandboxTool

   func (t *UpdateFileSandboxTool) Name() string { return "sandbox_update_file" }

   func (t *UpdateFileSandboxTool) Description() string {
       return "Write or overwrite a file in the repository with new content"
   }

   func (t *UpdateFileSandboxTool) Parameters() llm.JSONSchema {
       return llm.JSONSchema{
           Type: "object",
           Properties: map[string]llm.JSONProperty{
               "relativeFilePath": {
                   Type:        "string",
                   Description: "Relative path to the file within the repository",
               },
               "content": {
                   Type:        "string",
                   Description: "The new content to write to the file",
               },
           },
           Required: []string{"relativeFilePath", "content"},
       }
   }

   func (t *UpdateFileSandboxTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
       // 1. Extract userID from context
       // 2. Get relativeFilePath and content from params
       // 3. Validate path (no directory traversal, no absolute paths)
       // 4. Get user's work directory
       // 5. Check if file exists (for history tracking)
       // 6. If exists, save current content to file history
       // 7. Write new content to file
       // 8. Return success response with metadata
   }

   func (t *UpdateFileSandboxTool) RequiresConfirmation() bool { return true }
   ```

2. **Output Structure**:
   ```go
   type UpdateResult struct {
       Path      string `json:"path"`
       Created   bool   `json:"created"`   // true if new file
       Modified  bool   `json:"modified"`  // true if existing file
       Size      int64  `json:"size"`
       Timestamp string `json:"timestamp"`
   }
   ```

3. **Directory Creation**:
   - Automatically create parent directories if they don't exist
   - Use `os.MkdirAll` with appropriate permissions

4. **File History Integration**:
   - Before overwriting, save current content to FileHistoryRepository
   - Track operation type: "create" or "update"
   - Include userID for audit trail

5. **Path Validation**:
   - Reject paths containing `..`
   - Reject absolute paths
   - Validate filename (no special characters that could cause issues)

6. **Permission Handling**:
   - Set appropriate file permissions (0644 for files)
   - Set appropriate directory permissions (0755 for directories)

7. **Registration** - Add to `backend/internal/tools/builtin/init.go`:
   ```go
   registry.Register(NewUpdateFileSandboxTool(sandbox, config.FileHistoryRepo))
   ```

## Acceptance Criteria

- [ ] Tool correctly creates new files
- [ ] Tool correctly overwrites existing files
- [ ] Parent directories are created automatically
- [ ] File history is recorded for overwrites
- [ ] Path traversal attacks are prevented
- [ ] RequiresConfirmation returns true (write operation)
- [ ] Returns clear success/error responses
- [ ] Proper file permissions are set

## Files to Create/Modify

- `backend/internal/tools/builtin/update_file_sandbox.go` - Create (new file)
- `backend/internal/tools/builtin/init.go` - Register tool

## Integration Points

- **Provides**: File writing capability for sandbox environments
- **Consumes**: Sandbox service, FileHistoryRepository
- **Conflicts**: Avoid modifying existing `file_write.go` - this is a new sandbox-specific variant
