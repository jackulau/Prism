package builtin

import (
	"database/sql"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
	"github.com/jacklau/prism/internal/security"
	"github.com/jacklau/prism/internal/services/coderunner"
	"github.com/jacklau/prism/internal/tools"
)

// Config holds configuration for built-in tools
type Config struct {
	// Search API configuration
	SerpAPIKey     string
	GoogleAPIKey   string
	GoogleSearchCX string

	// OpenAI API key for image generation
	OpenAIAPIKey string

	// File history repository for tracking file changes
	FileHistoryRepo *repository.FileHistoryRepository

	// Shell execution configuration
	ShellExecConfig *ShellExecConfig

	// Todo repository for task tracking
	TodoRepo *repository.TodoRepository

	// LLM provider for WebFetch AI analysis (optional)
	LLMProvider llm.Provider

	// PostHog tools configuration (for query runner)
	PostHogConfig *PostHogConfig
}

// RegisterAll registers all built-in tools with the registry
func RegisterAll(registry *tools.Registry, sandbox *sandbox.Service, runner *coderunner.Runner, db *sql.DB, config Config) error {
	// File operation tools
	if err := registry.Register(NewFileReadTool(sandbox)); err != nil {
		return err
	}
	if err := registry.Register(NewReadFileSandboxTool(sandbox)); err != nil {
		return err
	}
	if err := registry.Register(NewFileWriteTool(sandbox, config.FileHistoryRepo)); err != nil {
		return err
	}
	if err := registry.Register(NewFileListTool(sandbox)); err != nil {
		return err
	}
	if err := registry.Register(NewFileDeleteTool(sandbox, config.FileHistoryRepo)); err != nil {
		return err
	}
	if err := registry.Register(NewFileRenameTool(sandbox, config.FileHistoryRepo)); err != nil {
		return err
	}
	if err := registry.Register(NewFileCreateDirectoryTool(sandbox)); err != nil {
		return err
	}

	// Sandbox-specific file update tool
	if err := registry.Register(NewUpdateFileSandboxTool(sandbox, config.FileHistoryRepo)); err != nil {
		return err
	}

	// File history tools (for viewing and restoring previous versions)
	if config.FileHistoryRepo != nil {
		if err := registry.Register(NewFileHistoryListTool(config.FileHistoryRepo)); err != nil {
			return err
		}
		if err := registry.Register(NewFileHistoryGetTool(config.FileHistoryRepo)); err != nil {
			return err
		}
		if err := registry.Register(NewFileHistoryRestoreTool(sandbox, config.FileHistoryRepo)); err != nil {
			return err
		}
	}

	// Code execution tool
	if err := registry.Register(NewCodeExecutionTool(runner)); err != nil {
		return err
	}

	// Shell execution tool for running commands like npm, git, pip, etc.
	shellConfig := DefaultShellExecConfig()
	if config.ShellExecConfig != nil {
		shellConfig = *config.ShellExecConfig
	}
	shellExecTool := NewShellExecTool(sandbox, shellConfig)

	// Background shell manager for background execution
	backgroundMgr := NewBackgroundShellManager(sandbox, shellConfig)
	shellExecTool.SetBackgroundManager(backgroundMgr)

	if err := registry.Register(shellExecTool); err != nil {
		return err
	}

	// Background shell tools (bash_output, kill_shell)
	if err := registry.Register(NewBashOutputTool(backgroundMgr)); err != nil {
		return err
	}
	if err := registry.Register(NewKillShellTool(backgroundMgr)); err != nil {
		return err
	}

	// Glob tool for file pattern matching
	if err := registry.Register(NewGlobTool(sandbox)); err != nil {
		return err
	}

	// Grep tool for content search
	if err := registry.Register(NewGrepTool(sandbox)); err != nil {
		return err
	}

	// Sandbox grep tool for raw grep command execution
	if err := registry.Register(NewGrepSandboxTool(sandbox)); err != nil {
		return err
	}

	// Edit tool for precise string replacement
	if err := registry.Register(NewEditTool(sandbox, config.FileHistoryRepo)); err != nil {
		return err
	}

	// MultiEdit tool for batch edits
	if err := registry.Register(NewMultiEditTool(sandbox, config.FileHistoryRepo)); err != nil {
		return err
	}

	// LS tool for enhanced directory listing
	if err := registry.Register(NewLSTool(sandbox)); err != nil {
		return err
	}

	// Sandbox list files tool for ls -la based file listing
	if err := registry.Register(NewListFilesSandboxTool(sandbox)); err != nil {
		return err
	}

	// WebFetch tool for fetching web content
	webFetchConfig := WebFetchConfig{
		LLMManager: config.LLMProvider,
	}
	if err := registry.Register(NewWebFetchTool(webFetchConfig)); err != nil {
		return err
	}

	// Notebook tools for Jupyter notebook support
	if err := registry.Register(NewNotebookReadTool(sandbox)); err != nil {
		return err
	}
	if err := registry.Register(NewNotebookEditTool(sandbox, config.FileHistoryRepo)); err != nil {
		return err
	}

	// Todo tools for task tracking
	if config.TodoRepo != nil {
		if err := registry.Register(NewTodoReadTool(sandbox, config.TodoRepo)); err != nil {
			return err
		}
		if err := registry.Register(NewTodoWriteTool(sandbox, config.TodoRepo)); err != nil {
			return err
		}
	}

	// Web search tool (only if configured)
	if config.SerpAPIKey != "" || (config.GoogleAPIKey != "" && config.GoogleSearchCX != "") {
		searchConfig := WebSearchConfig{
			SerpAPIKey:     config.SerpAPIKey,
			GoogleAPIKey:   config.GoogleAPIKey,
			GoogleSearchCX: config.GoogleSearchCX,
		}
		if err := registry.Register(NewWebSearchTool(searchConfig)); err != nil {
			return err
		}
	}

	// Image generation tool (only if configured)
	if config.OpenAIAPIKey != "" {
		if err := registry.Register(NewImageGenerationTool(config.OpenAIAPIKey)); err != nil {
			return err
		}
	}

	// Database query tool
	if db != nil {
		if err := registry.Register(NewDatabaseQueryTool(db)); err != nil {
			return err
		}
	}

	// PostHog tools (query runner, query generator, docs search)
	if config.PostHogConfig != nil && config.PostHogConfig.APIKey != "" {
		if err := registry.Register(NewPostHogQueryRunTool(*config.PostHogConfig)); err != nil {
			return err
		}
		if err := registry.Register(NewPostHogGenerateQueryTool(*config.PostHogConfig)); err != nil {
			return err
		}
	}
	// Docs search doesn't require API key - always register it
	if err := registry.Register(NewPostHogDocsSearchTool()); err != nil {
		return err
	}

	return nil
}
