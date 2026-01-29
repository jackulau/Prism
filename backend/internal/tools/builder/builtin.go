package builder

import (
	"github.com/jacklau/prism/internal/sandbox"
	"github.com/jacklau/prism/internal/tools"
	"github.com/jacklau/prism/internal/tools/builtin"
)

// RegisterBuiltinBuilders registers all built-in tool builders with the registry.
// This allows agents to dynamically select which tools they need.
func RegisterBuiltinBuilders(registry *BuilderRegistry) error {
	builders := []ToolBuilder{
		// File operations - sandbox provider
		NewSimpleBuilder("sandbox/file_read", func(s *sandbox.Service) tools.Tool {
			return builtin.NewFileReadTool(s)
		}),
		NewBuilderWithDeps("sandbox/file_write", nil, func(deps *BuilderDependencies) tools.Tool {
			return builtin.NewFileWriteTool(deps.Sandbox, deps.FileHistoryRepo)
		}),
		NewSimpleBuilder("sandbox/file_list", func(s *sandbox.Service) tools.Tool {
			return builtin.NewFileListTool(s)
		}),
		NewBuilderWithDeps("sandbox/file_delete", nil, func(deps *BuilderDependencies) tools.Tool {
			return builtin.NewFileDeleteTool(deps.Sandbox, deps.FileHistoryRepo)
		}),
		NewBuilderWithDeps("sandbox/file_rename", nil, func(deps *BuilderDependencies) tools.Tool {
			return builtin.NewFileRenameTool(deps.Sandbox, deps.FileHistoryRepo)
		}),
		NewSimpleBuilder("sandbox/file_mkdir", func(s *sandbox.Service) tools.Tool {
			return builtin.NewFileCreateDirectoryTool(s)
		}),

		// File history - requires FileHistoryRepo
		NewBuilderWithDeps("sandbox/file_history_list", nil, func(deps *BuilderDependencies) tools.Tool {
			if deps.FileHistoryRepo == nil {
				return nil
			}
			return builtin.NewFileHistoryListTool(deps.FileHistoryRepo)
		}),
		NewBuilderWithDeps("sandbox/file_history_get", nil, func(deps *BuilderDependencies) tools.Tool {
			if deps.FileHistoryRepo == nil {
				return nil
			}
			return builtin.NewFileHistoryGetTool(deps.FileHistoryRepo)
		}),
		NewBuilderWithDeps("sandbox/file_history_restore", nil, func(deps *BuilderDependencies) tools.Tool {
			if deps.FileHistoryRepo == nil {
				return nil
			}
			return builtin.NewFileHistoryRestoreTool(deps.Sandbox, deps.FileHistoryRepo)
		}),

		// Search tools - sandbox provider
		NewSimpleBuilder("sandbox/glob", func(s *sandbox.Service) tools.Tool {
			return builtin.NewGlobTool(s)
		}),
		NewSimpleBuilder("sandbox/grep", func(s *sandbox.Service) tools.Tool {
			return builtin.NewGrepTool(s)
		}),

		// Edit tools
		NewBuilderWithDeps("sandbox/edit", nil, func(deps *BuilderDependencies) tools.Tool {
			return builtin.NewEditTool(deps.Sandbox, deps.FileHistoryRepo)
		}),
		NewBuilderWithDeps("sandbox/multi_edit", nil, func(deps *BuilderDependencies) tools.Tool {
			return builtin.NewMultiEditTool(deps.Sandbox, deps.FileHistoryRepo)
		}),

		// Directory listing
		NewSimpleBuilder("sandbox/ls", func(s *sandbox.Service) tools.Tool {
			return builtin.NewLSTool(s)
		}),

		// Notebook tools
		NewSimpleBuilder("sandbox/notebook_read", func(s *sandbox.Service) tools.Tool {
			return builtin.NewNotebookReadTool(s)
		}),
		NewBuilderWithDeps("sandbox/notebook_edit", nil, func(deps *BuilderDependencies) tools.Tool {
			return builtin.NewNotebookEditTool(deps.Sandbox, deps.FileHistoryRepo)
		}),

		// Code execution - requires CodeRunner
		NewBuilderWithDeps("code/execute", nil, func(deps *BuilderDependencies) tools.Tool {
			if deps.CodeRunner == nil {
				return nil
			}
			return builtin.NewCodeExecutionTool(deps.CodeRunner)
		}),

		// Shell execution - requires sandbox
		NewBuilderWithDeps("shell/execute", nil, func(deps *BuilderDependencies) tools.Tool {
			config := builtin.DefaultShellExecConfig()
			return builtin.NewShellExecTool(deps.Sandbox, config)
		}),

		// Web tools
		NewBuilderWithDeps("web/fetch", nil, func(deps *BuilderDependencies) tools.Tool {
			config := builtin.WebFetchConfig{
				LLMManager: deps.LLMProvider,
			}
			return builtin.NewWebFetchTool(config)
		}),
		NewBuilderWithDeps("web/search", nil, func(deps *BuilderDependencies) tools.Tool {
			if deps.SerpAPIKey == "" && (deps.GoogleAPIKey == "" || deps.GoogleSearchCX == "") {
				return nil
			}
			config := builtin.WebSearchConfig{
				SerpAPIKey:     deps.SerpAPIKey,
				GoogleAPIKey:   deps.GoogleAPIKey,
				GoogleSearchCX: deps.GoogleSearchCX,
			}
			return builtin.NewWebSearchTool(config)
		}),

		// Image generation - requires OpenAI API key
		NewBuilderWithDeps("image/generate", nil, func(deps *BuilderDependencies) tools.Tool {
			if deps.OpenAIAPIKey == "" {
				return nil
			}
			return builtin.NewImageGenerationTool(deps.OpenAIAPIKey)
		}),

		// Database - requires DB
		NewBuilderWithDeps("database/query", nil, func(deps *BuilderDependencies) tools.Tool {
			if deps.DB == nil {
				return nil
			}
			return builtin.NewDatabaseQueryTool(deps.DB)
		}),

		// Todo tools - requires TodoRepo
		NewBuilderWithDeps("todo/read", nil, func(deps *BuilderDependencies) tools.Tool {
			if deps.TodoRepo == nil {
				return nil
			}
			return builtin.NewTodoReadTool(deps.Sandbox, deps.TodoRepo)
		}),
		NewBuilderWithDeps("todo/write", nil, func(deps *BuilderDependencies) tools.Tool {
			if deps.TodoRepo == nil {
				return nil
			}
			return builtin.NewTodoWriteTool(deps.Sandbox, deps.TodoRepo)
		}),
	}

	for _, builder := range builders {
		if err := registry.Register(builder); err != nil {
			return err
		}
	}

	return nil
}

// NewDefaultBuilderRegistry creates a new registry with all built-in builders registered.
func NewDefaultBuilderRegistry() (*BuilderRegistry, error) {
	registry := NewBuilderRegistry()
	if err := RegisterBuiltinBuilders(registry); err != nil {
		return nil, err
	}
	return registry, nil
}

// BuiltinToolSlugs returns a list of all available built-in tool slugs.
func BuiltinToolSlugs() []string {
	return []string{
		// File operations
		"sandbox/file_read",
		"sandbox/file_write",
		"sandbox/file_list",
		"sandbox/file_delete",
		"sandbox/file_rename",
		"sandbox/file_mkdir",
		// File history
		"sandbox/file_history_list",
		"sandbox/file_history_get",
		"sandbox/file_history_restore",
		// Search
		"sandbox/glob",
		"sandbox/grep",
		// Edit
		"sandbox/edit",
		"sandbox/multi_edit",
		// Directory
		"sandbox/ls",
		// Notebook
		"sandbox/notebook_read",
		"sandbox/notebook_edit",
		// Code
		"code/execute",
		// Shell
		"shell/execute",
		// Web
		"web/fetch",
		"web/search",
		// Image
		"image/generate",
		// Database
		"database/query",
		// Todo
		"todo/read",
		"todo/write",
	}
}
