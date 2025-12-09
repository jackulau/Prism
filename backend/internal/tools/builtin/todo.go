package builtin

import (
	"context"
	"fmt"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// TodoWriteTool manages task lists for the current workspace
type TodoWriteTool struct {
	sandbox  *sandbox.Service
	todoRepo *repository.TodoRepository
}

// NewTodoWriteTool creates a new todo write tool
func NewTodoWriteTool(sandbox *sandbox.Service, todoRepo *repository.TodoRepository) *TodoWriteTool {
	return &TodoWriteTool{sandbox: sandbox, todoRepo: todoRepo}
}

func (t *TodoWriteTool) Name() string {
	return "todo_write"
}

func (t *TodoWriteTool) Description() string {
	return `Create and manage a structured task list for the current workspace. Use this to track progress on complex multi-step tasks. Each todo has content (what to do), status (pending/in_progress/completed), and activeForm (present continuous description).`
}

func (t *TodoWriteTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"todos": {
				Type:        "array",
				Description: "Array of todo items. Each item should have 'content' (string), 'status' (pending/in_progress/completed), and 'activeForm' (string)",
			},
		},
		Required: []string{"todos"},
	}
}

// TodoItem represents a single todo item from the request
type TodoItem struct {
	Content    string `json:"content"`
	Status     string `json:"status"`
	ActiveForm string `json:"activeForm"`
}

func (t *TodoWriteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	if t.todoRepo == nil {
		return nil, fmt.Errorf("todo repository not available")
	}

	// Get workspace path
	workspacePath, err := t.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Parse todos
	todosRaw, ok := params["todos"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("todos parameter is required and must be an array")
	}

	todos := make([]repository.Todo, 0, len(todosRaw))
	for i, todoRaw := range todosRaw {
		todoMap, ok := todoRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("todo %d is not a valid object", i)
		}

		content, _ := todoMap["content"].(string)
		status, _ := todoMap["status"].(string)
		activeForm, _ := todoMap["activeForm"].(string)

		if content == "" {
			return nil, fmt.Errorf("todo %d: content is required", i)
		}

		// Validate status
		if status == "" {
			status = "pending"
		}
		if status != "pending" && status != "in_progress" && status != "completed" {
			return nil, fmt.Errorf("todo %d: status must be 'pending', 'in_progress', or 'completed'", i)
		}

		if activeForm == "" {
			activeForm = content
		}

		todos = append(todos, repository.Todo{
			UserID:        userID,
			WorkspacePath: workspacePath,
			Content:       content,
			Status:        status,
			ActiveForm:    activeForm,
		})
	}

	// Replace all todos for this workspace
	if err := t.todoRepo.ReplaceAll(userID, workspacePath, todos); err != nil {
		return nil, fmt.Errorf("failed to save todos: %w", err)
	}

	// Return the updated list
	return map[string]interface{}{
		"success": true,
		"count":   len(todos),
		"todos":   formatTodos(todos),
	}, nil
}

func formatTodos(todos []repository.Todo) []map[string]interface{} {
	result := make([]map[string]interface{}, len(todos))
	for i, todo := range todos {
		result[i] = map[string]interface{}{
			"content":     todo.Content,
			"status":      todo.Status,
			"active_form": todo.ActiveForm,
		}
	}
	return result
}

func (t *TodoWriteTool) RequiresConfirmation() bool {
	return false
}

// TodoReadTool reads the current task list
type TodoReadTool struct {
	sandbox  *sandbox.Service
	todoRepo *repository.TodoRepository
}

// NewTodoReadTool creates a new todo read tool
func NewTodoReadTool(sandbox *sandbox.Service, todoRepo *repository.TodoRepository) *TodoReadTool {
	return &TodoReadTool{sandbox: sandbox, todoRepo: todoRepo}
}

func (t *TodoReadTool) Name() string {
	return "todo_read"
}

func (t *TodoReadTool) Description() string {
	return "Read the current task list for the workspace. Returns all todos with their content, status, and active form."
}

func (t *TodoReadTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type:       "object",
		Properties: map[string]llm.JSONProperty{},
		Required:   []string{},
	}
}

func (t *TodoReadTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	if t.todoRepo == nil {
		return nil, fmt.Errorf("todo repository not available")
	}

	// Get workspace path
	workspacePath, err := t.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Get all todos for this workspace
	todos, err := t.todoRepo.GetAll(userID, workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}

	// Format for response
	result := make([]map[string]interface{}, len(todos))
	for i, todo := range todos {
		result[i] = map[string]interface{}{
			"content":     todo.Content,
			"status":      todo.Status,
			"active_form": todo.ActiveForm,
		}
	}

	return map[string]interface{}{
		"todos": result,
		"count": len(result),
	}, nil
}

func (t *TodoReadTool) RequiresConfirmation() bool {
	return false
}
