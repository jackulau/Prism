package agent

import (
	"encoding/json"

	"github.com/jacklau/prism/internal/database/repository"
)

// RepositoryPersister adapts the AgentTaskRepository to the TaskPersister interface
type RepositoryPersister struct {
	repo *repository.AgentTaskRepository
}

// NewRepositoryPersister creates a new repository-based task persister
func NewRepositoryPersister(repo *repository.AgentTaskRepository) *RepositoryPersister {
	return &RepositoryPersister{repo: repo}
}

// CreateTask persists a task to the database
func (p *RepositoryPersister) CreateTask(task *Task) error {
	// Convert agent config to map
	var agentConfig map[string]interface{}
	if task.AgentConfig != nil {
		agentConfig = map[string]interface{}{
			"provider":     task.AgentConfig.Provider,
			"model":        task.AgentConfig.Model,
			"temperature":  task.AgentConfig.Temperature,
			"max_tokens":   task.AgentConfig.MaxTokens,
			"system_prompt": task.AgentConfig.SystemPrompt,
		}
		if task.AgentConfig.Name != "" {
			agentConfig["name"] = task.AgentConfig.Name
		}
		if task.AgentConfig.ID != "" {
			agentConfig["id"] = task.AgentConfig.ID
		}
	}

	// Convert CallbackData if it exists
	var callbackData map[string]string
	if task.CallbackData != nil {
		callbackData = task.CallbackData
	}

	dbTask := &repository.AgentTask{
		ID:           task.ID,
		UserID:       task.UserID,
		Prompt:       task.Prompt,
		Context:      task.Context,
		Priority:     int(task.Priority),
		Status:       string(task.Status),
		AgentConfig:  agentConfig,
		Metadata:     task.Metadata,
		CallbackURL:  task.CallbackURL,
		CallbackData: callbackData,
		CreatedAt:    task.CreatedAt,
	}

	return p.repo.Create(dbTask)
}

// UpdateTaskStatus updates the status of a task in the database
func (p *RepositoryPersister) UpdateTaskStatus(taskID string, status TaskStatus) error {
	return p.repo.UpdateStatus(taskID, string(status))
}

// SetTaskResult sets the result of a completed task
func (p *RepositoryPersister) SetTaskResult(taskID string, result map[string]interface{}) error {
	return p.repo.SetResult(taskID, result)
}

// SetTaskError sets the error message for a failed task
func (p *RepositoryPersister) SetTaskError(taskID string, errorMsg string) error {
	return p.repo.SetError(taskID, errorMsg)
}

// GetPendingTasks retrieves pending tasks from the database for recovery
func (p *RepositoryPersister) GetPendingTasks(limit int) ([]*Task, error) {
	dbTasks, err := p.repo.ListPending(limit)
	if err != nil {
		return nil, err
	}

	tasks := make([]*Task, len(dbTasks))
	for i, dbTask := range dbTasks {
		tasks[i] = p.convertToAgentTask(dbTask)
	}

	return tasks, nil
}

// convertToAgentTask converts a repository task to an agent task
func (p *RepositoryPersister) convertToAgentTask(dbTask *repository.AgentTask) *Task {
	task := &Task{
		ID:          dbTask.ID,
		UserID:      dbTask.UserID,
		Prompt:      dbTask.Prompt,
		Context:     dbTask.Context,
		Priority:    TaskPriority(dbTask.Priority),
		Status:      TaskStatus(dbTask.Status),
		Metadata:    dbTask.Metadata,
		Result:      dbTask.Result,
		Error:       dbTask.Error,
		CallbackURL: dbTask.CallbackURL,
		CreatedAt:   dbTask.CreatedAt,
		StartedAt:   dbTask.StartedAt,
		CompletedAt: dbTask.CompletedAt,
	}

	// Convert callback data
	if dbTask.CallbackData != nil {
		task.CallbackData = dbTask.CallbackData
	}

	// Convert agent config from map
	if dbTask.AgentConfig != nil {
		config := &AgentConfig{}
		if provider, ok := dbTask.AgentConfig["provider"].(string); ok {
			config.Provider = provider
		}
		if model, ok := dbTask.AgentConfig["model"].(string); ok {
			config.Model = model
		}
		if temp, ok := dbTask.AgentConfig["temperature"].(float64); ok {
			config.Temperature = temp
		}
		if maxTokens, ok := dbTask.AgentConfig["max_tokens"].(float64); ok {
			config.MaxTokens = int(maxTokens)
		}
		if systemPrompt, ok := dbTask.AgentConfig["system_prompt"].(string); ok {
			config.SystemPrompt = systemPrompt
		}
		if name, ok := dbTask.AgentConfig["name"].(string); ok {
			config.Name = name
		}
		if id, ok := dbTask.AgentConfig["id"].(string); ok {
			config.ID = id
		}

		// Handle case where agent_config might be stored as JSON string
		if configStr, ok := dbTask.AgentConfig["_raw"].(string); ok {
			json.Unmarshal([]byte(configStr), config)
		}

		task.AgentConfig = config
	}

	return task
}
