package steps

import (
	"context"
	"fmt"

	"github.com/jacklau/prism/internal/workflow"
)

// LoadStep handles loading agent data from the database
type LoadStep struct {
	agentRepo  workflow.AgentRepository
	githubRepo workflow.GitHubRepository
}

// NewLoadStep creates a new load step handler
func NewLoadStep(agentRepo workflow.AgentRepository, githubRepo workflow.GitHubRepository) *LoadStep {
	return &LoadStep{
		agentRepo:  agentRepo,
		githubRepo: githubRepo,
	}
}

// LoadAgent loads an agent from the database (Step 1)
func (s *LoadStep) LoadAgent(ctx context.Context, agentID string) (*workflow.AgentData, error) {
	if agentID == "" {
		return nil, workflow.ErrAgentNotFound
	}

	agent, err := s.agentRepo.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent: %w", err)
	}

	if agent == nil {
		return nil, workflow.ErrAgentNotFound
	}

	// Validate required fields
	if agent.Provider == "" || agent.Model == "" {
		return nil, workflow.ErrInvalidAgentConfig
	}

	return agent, nil
}

// ValidateGitHubToken validates and retrieves the GitHub token for a user (Step 2)
func (s *LoadStep) ValidateGitHubToken(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", workflow.ErrGitHubTokenMissing
	}

	token, err := s.githubRepo.GetToken(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get github token: %w", err)
	}

	if token == "" {
		return "", workflow.ErrGitHubTokenMissing
	}

	// Validate the token is still valid
	valid, err := s.githubRepo.ValidateToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to validate github token: %w", err)
	}

	if !valid {
		return "", workflow.ErrGitHubTokenInvalid
	}

	return token, nil
}

// ValidateGitHubTokenOptional validates the GitHub token but returns empty if not configured
// Use this when GitHub integration is optional
func (s *LoadStep) ValidateGitHubTokenOptional(ctx context.Context, userID string) (string, error) {
	token, err := s.ValidateGitHubToken(ctx, userID)
	if err != nil {
		// If token is missing or invalid, return empty instead of error
		if err == workflow.ErrGitHubTokenMissing || err == workflow.ErrGitHubTokenInvalid {
			return "", nil
		}
		return "", err
	}
	return token, nil
}
