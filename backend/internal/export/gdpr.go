package export

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// GDPRExportManifest contains metadata about a GDPR data export
type GDPRExportManifest struct {
	ExportID          string    `json:"export_id"`
	UserID            string    `json:"user_id"`
	UserEmail         string    `json:"user_email"`
	ExportedAt        time.Time `json:"exported_at"`
	RequestedAt       time.Time `json:"requested_at"`
	Format            string    `json:"format"`
	TotalRecords      int       `json:"total_records"`
	DataCategories    []string  `json:"data_categories"`
	IncludedFiles     []string  `json:"included_files"`
	ComplianceVersion string    `json:"compliance_version"`
}

// GDPRUserProfile contains basic user profile information
type GDPRUserProfile struct {
	ID                 string     `json:"id"`
	Email              string     `json:"email"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	GitHubUsername     string     `json:"github_username,omitempty"`
	GitHubConnectedAt  *time.Time `json:"github_connected_at,omitempty"`
	OrganizationID     string     `json:"organization_id,omitempty"`
	SSOProvider        string     `json:"sso_provider,omitempty"`
	SSOConnectionID    string     `json:"sso_connection_id,omitempty"`
}

// GDPRConversation represents a conversation export
type GDPRConversation struct {
	ID           string             `json:"id"`
	Title        string             `json:"title"`
	Provider     string             `json:"provider"`
	Model        string             `json:"model"`
	SystemPrompt string             `json:"system_prompt,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Messages     []GDPRMessage      `json:"messages"`
}

// GDPRMessage represents a message in a conversation
type GDPRMessage struct {
	ID          string     `json:"id"`
	Role        string     `json:"role"`
	Content     string     `json:"content"`
	ToolCalls   string     `json:"tool_calls,omitempty"`
	ToolCallID  string     `json:"tool_call_id,omitempty"`
	TokensUsed  int        `json:"tokens_used,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// GDPRSettings contains user settings
type GDPRSettings struct {
	DefaultProvider string                 `json:"default_provider,omitempty"`
	DefaultModel    string                 `json:"default_model,omitempty"`
	Theme           string                 `json:"theme,omitempty"`
	CustomSettings  map[string]interface{} `json:"custom_settings,omitempty"`
}

// GDPRIntegration represents an integration configuration
type GDPRIntegration struct {
	Type      string    `json:"type"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GDPRWorkspace represents a workspace record
type GDPRWorkspace struct {
	ID             string     `json:"id"`
	Path           string     `json:"path"`
	Name           string     `json:"name,omitempty"`
	IsCurrent      bool       `json:"is_current"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// GDPRExportData contains all user data for GDPR export
type GDPRExportData struct {
	Manifest      GDPRExportManifest  `json:"manifest"`
	Profile       GDPRUserProfile     `json:"profile"`
	Settings      *GDPRSettings       `json:"settings,omitempty"`
	Conversations []GDPRConversation  `json:"conversations"`
	Integrations  []GDPRIntegration   `json:"integrations,omitempty"`
	Workspaces    []GDPRWorkspace     `json:"workspaces,omitempty"`
}

// GDPRDataProvider defines the interface for fetching user data for GDPR exports
type GDPRDataProvider interface {
	GetUserProfile(ctx context.Context, userID string) (*GDPRUserProfile, error)
	GetUserSettings(ctx context.Context, userID string) (*GDPRSettings, error)
	GetUserConversations(ctx context.Context, userID string) ([]GDPRConversation, error)
	GetUserIntegrations(ctx context.Context, userID string) ([]GDPRIntegration, error)
	GetUserWorkspaces(ctx context.Context, userID string) ([]GDPRWorkspace, error)
}

// GDPRExporter handles GDPR-compliant data exports
type GDPRExporter struct {
	provider  GDPRDataProvider
	exportDir string
}

// NewGDPRExporter creates a new GDPR exporter
func NewGDPRExporter(provider GDPRDataProvider, exportDir string) *GDPRExporter {
	return &GDPRExporter{
		provider:  provider,
		exportDir: exportDir,
	}
}

// Export creates a complete GDPR data export for a user
func (e *GDPRExporter) Export(ctx context.Context, job *ExportJob, progressCh chan<- int) error {
	userID := job.UserID

	// Step 1: Get user profile (10%)
	progressCh <- 5
	profile, err := e.provider.GetUserProfile(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}
	if profile == nil {
		return fmt.Errorf("user not found")
	}
	progressCh <- 10

	// Step 2: Get user settings (20%)
	settings, err := e.provider.GetUserSettings(ctx, userID)
	if err != nil {
		// Settings are optional, don't fail
		settings = nil
	}
	progressCh <- 20

	// Step 3: Get conversations (50%)
	conversations, err := e.provider.GetUserConversations(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get conversations: %w", err)
	}
	progressCh <- 50

	// Step 4: Get integrations (60%)
	integrations, err := e.provider.GetUserIntegrations(ctx, userID)
	if err != nil {
		integrations = nil
	}
	progressCh <- 60

	// Step 5: Get workspaces (70%)
	workspaces, err := e.provider.GetUserWorkspaces(ctx, userID)
	if err != nil {
		workspaces = nil
	}
	progressCh <- 70

	// Step 6: Create manifest
	totalRecords := 1 + len(conversations) // profile + conversations
	if settings != nil {
		totalRecords++
	}
	totalRecords += len(integrations)
	totalRecords += len(workspaces)

	categories := []string{"profile"}
	files := []string{"profile.json"}

	if settings != nil {
		categories = append(categories, "settings")
		files = append(files, "settings.json")
	}
	if len(conversations) > 0 {
		categories = append(categories, "conversations")
		files = append(files, "conversations.json")
	}
	if len(integrations) > 0 {
		categories = append(categories, "integrations")
		files = append(files, "integrations.json")
	}
	if len(workspaces) > 0 {
		categories = append(categories, "workspaces")
		files = append(files, "workspaces.json")
	}

	manifest := GDPRExportManifest{
		ExportID:          job.ID,
		UserID:            userID,
		UserEmail:         profile.Email,
		ExportedAt:        time.Now().UTC(),
		RequestedAt:       job.CreatedAt,
		Format:            "json",
		TotalRecords:      totalRecords,
		DataCategories:    categories,
		IncludedFiles:     append([]string{"manifest.json"}, files...),
		ComplianceVersion: "GDPR-2018",
	}

	// Step 7: Build export data
	exportData := GDPRExportData{
		Manifest:      manifest,
		Profile:       *profile,
		Settings:      settings,
		Conversations: conversations,
		Integrations:  integrations,
		Workspaces:    workspaces,
	}
	progressCh <- 80

	// Step 8: Create output file
	var filePath string
	switch job.Format {
	case FormatJSON:
		filePath = fmt.Sprintf("%s/%s.json", e.exportDir, job.ID)
		if err := WriteJSON(filePath, exportData); err != nil {
			return fmt.Errorf("failed to write JSON: %w", err)
		}
	case FormatZIP:
		filePath = fmt.Sprintf("%s/%s.zip", e.exportDir, job.ID)
		if err := e.createZipExport(filePath, exportData); err != nil {
			return fmt.Errorf("failed to create ZIP: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", job.Format)
	}

	job.FilePath = filePath
	progressCh <- 100

	return nil
}

// createZipExport creates a ZIP archive with separate files for each data category
func (e *GDPRExporter) createZipExport(zipPath string, data GDPRExportData) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Write manifest
	if err := writeJSONToZip(zipWriter, "manifest.json", data.Manifest); err != nil {
		return err
	}

	// Write profile
	if err := writeJSONToZip(zipWriter, "profile.json", data.Profile); err != nil {
		return err
	}

	// Write settings
	if data.Settings != nil {
		if err := writeJSONToZip(zipWriter, "settings.json", data.Settings); err != nil {
			return err
		}
	}

	// Write conversations
	if len(data.Conversations) > 0 {
		if err := writeJSONToZip(zipWriter, "conversations.json", data.Conversations); err != nil {
			return err
		}
	}

	// Write integrations
	if len(data.Integrations) > 0 {
		if err := writeJSONToZip(zipWriter, "integrations.json", data.Integrations); err != nil {
			return err
		}
	}

	// Write workspaces
	if len(data.Workspaces) > 0 {
		if err := writeJSONToZip(zipWriter, "workspaces.json", data.Workspaces); err != nil {
			return err
		}
	}

	// Write README
	readme := `GDPR Data Export
================

This archive contains your personal data as required under GDPR Article 15.

Contents:
- manifest.json: Export metadata and summary
- profile.json: Your user profile information
- settings.json: Your application settings (if any)
- conversations.json: Your chat conversations and messages
- integrations.json: Your configured integrations (if any)
- workspaces.json: Your workspace configurations (if any)

For questions about this data export, please contact our support team.

Export Date: ` + data.Manifest.ExportedAt.Format(time.RFC3339)

	writer, err := zipWriter.Create("README.txt")
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(readme))
	return err
}

// writeJSONToZip writes a JSON object to a file in a zip archive
func writeJSONToZip(zipWriter *zip.Writer, filename string, data interface{}) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create %s in zip: %w", filename, err)
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode %s: %w", filename, err)
	}

	return nil
}
