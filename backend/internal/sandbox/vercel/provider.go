package vercel

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/sandbox"
)

// Provider implements sandbox.Provider for Vercel
type Provider struct {
	client     *APIClient
	sandboxes  map[string]*sandboxState
	mu         sync.RWMutex
}

// sandboxState tracks the state of a sandbox
type sandboxState struct {
	sandbox      *sandbox.Sandbox
	deploymentID string
	projectName  string
}

// NewProvider creates a new Vercel sandbox provider
func NewProvider(apiToken, teamID string) *Provider {
	return &Provider{
		client:    NewAPIClient(apiToken, teamID),
		sandboxes: make(map[string]*sandboxState),
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "vercel"
}

// CreateSandbox creates a new sandbox environment
func (p *Provider) CreateSandbox(ctx context.Context, opts *sandbox.CreateOptions) (*sandbox.Sandbox, error) {
	sandboxID := uuid.New().String()
	now := time.Now()

	// Generate a unique project name
	projectName := fmt.Sprintf("sandbox-%s", sandboxID[:8])

	sb := &sandbox.Sandbox{
		ID:        sandboxID,
		Provider:  "vercel",
		Status:    sandbox.SandboxStatusCreating,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata: map[string]interface{}{
			"framework":     string(opts.Framework),
			"node_version":  opts.NodeVersion,
			"build_command": opts.BuildCommand,
			"output_dir":    opts.OutputDir,
			"project_name":  projectName,
		},
	}

	// Store sandbox state
	p.mu.Lock()
	p.sandboxes[sandboxID] = &sandboxState{
		sandbox:     sb,
		projectName: projectName,
	}
	p.mu.Unlock()

	// Mark as ready (actual deployment happens in DeploySandbox)
	sb.Status = sandbox.SandboxStatusReady

	return sb, nil
}

// DeploySandbox deploys files to a sandbox
func (p *Provider) DeploySandbox(ctx context.Context, sandboxID string, files map[string][]byte) (*sandbox.DeployResult, error) {
	p.mu.Lock()
	state, ok := p.sandboxes[sandboxID]
	if !ok {
		p.mu.Unlock()
		return nil, fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	// Update status to deploying
	state.sandbox.Status = sandbox.SandboxStatusDeploying
	state.sandbox.UpdatedAt = time.Now()
	p.mu.Unlock()

	// Convert files to Vercel format
	deploymentFiles := make([]DeploymentFile, 0, len(files))
	for path, content := range files {
		// Determine if content should be base64 encoded
		// For binary files or large files, use base64
		encodedContent := base64.StdEncoding.EncodeToString(content)

		deploymentFiles = append(deploymentFiles, DeploymentFile{
			File: path,
			Data: encodedContent,
		})
	}

	// Build project settings from sandbox metadata
	var projectSettings *ProjectSettings
	if state.sandbox.Metadata != nil {
		projectSettings = &ProjectSettings{}
		if fw, ok := state.sandbox.Metadata["framework"].(string); ok {
			projectSettings.Framework = FrameworkToVercel(fw)
		}
		if bc, ok := state.sandbox.Metadata["build_command"].(string); ok {
			projectSettings.BuildCommand = bc
		}
		if od, ok := state.sandbox.Metadata["output_dir"].(string); ok {
			projectSettings.OutputDirectory = od
		}
		if nv, ok := state.sandbox.Metadata["node_version"].(string); ok {
			projectSettings.NodeVersion = nv
		}
	}

	// Create the deployment
	req := &CreateDeploymentRequest{
		Name:            state.projectName,
		Files:           deploymentFiles,
		ProjectSettings: projectSettings,
		Target:          "preview",
	}

	deployment, err := p.client.CreateDeployment(ctx, req)
	if err != nil {
		p.mu.Lock()
		state.sandbox.Status = sandbox.SandboxStatusFailed
		state.sandbox.UpdatedAt = time.Now()
		p.mu.Unlock()

		return &sandbox.DeployResult{
			Status: sandbox.DeploymentStatusError,
			Error:  err.Error(),
		}, err
	}

	// Store deployment ID
	p.mu.Lock()
	state.deploymentID = deployment.ID
	state.sandbox.PreviewURL = fmt.Sprintf("https://%s", deployment.URL)
	state.sandbox.UpdatedAt = time.Now()
	p.mu.Unlock()

	// Map Vercel state to our status
	status := sandbox.DeploymentStatus(StateToDeploymentStatus(deployment.State))
	now := time.Now()

	result := &sandbox.DeployResult{
		DeploymentID: deployment.ID,
		PreviewURL:   fmt.Sprintf("https://%s", deployment.URL),
		Status:       status,
		CreatedAt:    now,
	}

	// If deployment is already ready, update sandbox status
	if deployment.State == "READY" {
		p.mu.Lock()
		state.sandbox.Status = sandbox.SandboxStatusDeployed
		state.sandbox.UpdatedAt = time.Now()
		p.mu.Unlock()
		result.ReadyAt = &now
	}

	return result, nil
}

// GetSandbox retrieves the current state of a sandbox
func (p *Provider) GetSandbox(ctx context.Context, sandboxID string) (*sandbox.Sandbox, error) {
	p.mu.RLock()
	state, ok := p.sandboxes[sandboxID]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	// If we have a deployment, fetch its current status
	if state.deploymentID != "" {
		deployment, err := p.client.GetDeployment(ctx, state.deploymentID)
		if err == nil {
			p.mu.Lock()
			switch deployment.State {
			case "READY":
				state.sandbox.Status = sandbox.SandboxStatusDeployed
			case "ERROR":
				state.sandbox.Status = sandbox.SandboxStatusFailed
			case "BUILDING", "QUEUED":
				state.sandbox.Status = sandbox.SandboxStatusDeploying
			}
			state.sandbox.PreviewURL = fmt.Sprintf("https://%s", deployment.URL)
			state.sandbox.UpdatedAt = time.Now()
			p.mu.Unlock()
		}
	}

	return state.sandbox, nil
}

// GetLogs returns a channel of log entries for the sandbox
func (p *Provider) GetLogs(ctx context.Context, sandboxID string) (<-chan sandbox.LogEntry, error) {
	p.mu.RLock()
	state, ok := p.sandboxes[sandboxID]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	if state.deploymentID == "" {
		return nil, fmt.Errorf("no deployment exists for sandbox: %s", sandboxID)
	}

	logChan := make(chan sandbox.LogEntry, 100)

	go func() {
		defer close(logChan)

		// Poll for logs until context is done or deployment is complete
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		var lastLogTime int64

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logs, err := p.client.GetBuildLogs(ctx, state.deploymentID)
				if err != nil {
					logChan <- sandbox.LogEntry{
						Timestamp: time.Now(),
						Message:   fmt.Sprintf("Error fetching logs: %v", err),
						Level:     "error",
						Source:    "provider",
					}
					continue
				}

				for _, log := range logs {
					if log.Created > lastLogTime {
						lastLogTime = log.Created

						level := "info"
						if log.Type == "error" {
							level = "error"
						} else if log.Type == "warning" {
							level = "warn"
						}

						logChan <- sandbox.LogEntry{
							Timestamp: time.Unix(0, log.Created*int64(time.Millisecond)),
							Message:   log.Payload.Text,
							Level:     level,
							Source:    "build",
						}
					}
				}

				// Check if deployment is done
				deployment, err := p.client.GetDeployment(ctx, state.deploymentID)
				if err == nil {
					if deployment.State == "READY" || deployment.State == "ERROR" || deployment.State == "CANCELED" {
						return
					}
				}
			}
		}
	}()

	return logChan, nil
}

// DeleteSandbox deletes a sandbox and its associated resources
func (p *Provider) DeleteSandbox(ctx context.Context, sandboxID string) error {
	p.mu.Lock()
	state, ok := p.sandboxes[sandboxID]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	deploymentID := state.deploymentID
	delete(p.sandboxes, sandboxID)
	p.mu.Unlock()

	// Delete the deployment if it exists
	if deploymentID != "" {
		if err := p.client.DeleteDeployment(ctx, deploymentID); err != nil {
			// Log but don't fail if we can't delete the deployment
			// It may have already been deleted or expired
			return nil
		}
	}

	return nil
}

// GetPreviewURL returns the preview URL for a sandbox
func (p *Provider) GetPreviewURL(sandboxID string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	state, ok := p.sandboxes[sandboxID]
	if !ok {
		return ""
	}

	return state.sandbox.PreviewURL
}

// WaitForDeployment waits for a deployment to complete and returns its status
func (p *Provider) WaitForDeployment(ctx context.Context, sandboxID string, timeout time.Duration) (*sandbox.DeployResult, error) {
	p.mu.RLock()
	state, ok := p.sandboxes[sandboxID]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	if state.deploymentID == "" {
		return nil, fmt.Errorf("no deployment exists for sandbox: %s", sandboxID)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for deployment")
		case <-ticker.C:
			deployment, err := p.client.GetDeployment(ctx, state.deploymentID)
			if err != nil {
				continue
			}

			status := sandbox.DeploymentStatus(StateToDeploymentStatus(deployment.State))

			if deployment.State == "READY" || deployment.State == "ERROR" || deployment.State == "CANCELED" {
				now := time.Now()
				result := &sandbox.DeployResult{
					DeploymentID: deployment.ID,
					PreviewURL:   fmt.Sprintf("https://%s", deployment.URL),
					Status:       status,
					CreatedAt:    time.Unix(0, deployment.CreatedAt*int64(time.Millisecond)),
				}

				if deployment.State == "READY" {
					result.ReadyAt = &now
				}
				if deployment.State == "ERROR" {
					result.Error = "Deployment failed"
					if deployment.AliasError != nil {
						result.Error = deployment.AliasError.Message
					}
				}

				// Update sandbox state
				p.mu.Lock()
				if deployment.State == "READY" {
					state.sandbox.Status = sandbox.SandboxStatusDeployed
				} else if deployment.State == "ERROR" {
					state.sandbox.Status = sandbox.SandboxStatusFailed
				}
				state.sandbox.UpdatedAt = time.Now()
				p.mu.Unlock()

				return result, nil
			}
		}
	}
}
