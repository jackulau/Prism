package builtin

import (
	"database/sql"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/sandbox"
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
}

// RegisterAll registers all built-in tools with the registry
func RegisterAll(registry *tools.Registry, sandbox *sandbox.Service, runner *coderunner.Runner, db *sql.DB, config Config) error {
	// File operation tools
	if err := registry.Register(NewFileReadTool(sandbox)); err != nil {
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

	return nil
}
