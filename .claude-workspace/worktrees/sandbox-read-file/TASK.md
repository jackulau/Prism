---
id: sandbox-read-file
name: Sandbox Tool - readFile
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

Create the `readFile` sandbox tool that reads file contents from a repository using the `cat` command.

## Context

This tool allows agents to read the contents of files in a user's workspace. It provides the file content as a string, enabling agents to understand code, configuration, and documentation files.

## Implementation

1. **Tool Implementation** - Create `backend/internal/tools/builtin/read_file_sandbox.go`:
   ```go
   type ReadFileSandboxTool struct {
       sandbox *sandbox.Service
   }

   func NewReadFileSandboxTool(sandbox *sandbox.Service) *ReadFileSandboxTool

   func (t *ReadFileSandboxTool) Name() string { return "sandbox_read_file" }

   func (t *ReadFileSandboxTool) Description() string {
       return "Read the contents of a file in the repository"
   }

   func (t *ReadFileSandboxTool) Parameters() llm.JSONSchema {
       return llm.JSONSchema{
           Type: "object",
           Properties: map[string]llm.JSONProperty{
               "relativeFilePath": {
                   Type:        "string",
                   Description: "Relative path to the file within the repository",
               },
           },
           Required: []string{"relativeFilePath"},
       }
   }

   func (t *ReadFileSandboxTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
       // 1. Extract userID from context
       // 2. Get relativeFilePath from params
       // 3. Validate path (no directory traversal, no absolute paths)
       // 4. Get user's work directory
       // 5. Execute: cat {fullPath}
       // 6. Return content and metadata
   }

   func (t *ReadFileSandboxTool) RequiresConfirmation() bool { return false }
   ```

2. **Output Structure**:
   ```go
   type FileContent struct {
       Path     string `json:"path"`
       Content  string `json:"content"`
       Size     int64  `json:"size"`
       Encoding string `json:"encoding"` // "utf-8" or "binary"
   }
   ```

3. **Binary File Handling**:
   - Detect binary files (check for null bytes or use file extension)
   - For binary files, return base64-encoded content or error message
   - Set appropriate encoding field

4. **Size Limits**:
   - Implement maximum file size limit (e.g., 10MB)
   - Return error for files exceeding limit
   - Consider truncation option with offset parameter

5. **Path Validation**:
   - Reject paths containing `..`
   - Reject absolute paths
   - Ensure file exists before reading

6. **Registration** - Add to `backend/internal/tools/builtin/init.go`:
   ```go
   registry.Register(NewReadFileSandboxTool(sandbox))
   ```

## Acceptance Criteria

- [ ] Tool correctly reads file contents
- [ ] Returns structured response with path, content, size
- [ ] Path traversal attacks are prevented
- [ ] Binary files are handled appropriately
- [ ] Large files are handled with proper limits
- [ ] Non-existent files return clear error message
- [ ] Added to read-only tools list for auto-approval

## Files to Create/Modify

- `backend/internal/tools/builtin/read_file_sandbox.go` - Create (new file)
- `backend/internal/tools/builtin/init.go` - Register tool
- `backend/internal/tools/approval.go` - Add to ReadOnlyTools if appropriate

## Integration Points

- **Provides**: File reading capability for sandbox environments
- **Consumes**: Sandbox service for file operations
- **Conflicts**: Avoid modifying existing `file_read.go` - this is a new sandbox-specific variant
