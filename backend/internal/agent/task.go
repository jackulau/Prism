package agent

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// TaskPriority represents the priority level of a task
type TaskPriority int

const (
	TaskPriorityLow    TaskPriority = 0
	TaskPriorityNormal TaskPriority = 1
	TaskPriorityHigh   TaskPriority = 2
	TaskPriorityUrgent TaskPriority = 3
)

// Task represents a unit of work to be executed by an agent
type Task struct {
	ID          string                 `json:"id"`
	Prompt      string                 `json:"prompt"`
	Context     string                 `json:"context,omitempty"`
	Priority    TaskPriority           `json:"priority"`
	AgentConfig *AgentConfig           `json:"agent_config,omitempty"` // Optional: override default agent config
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	Timeout     time.Duration          `json:"timeout,omitempty"` // Optional timeout for task execution

	// Callback configuration
	CallbackURL  string            `json:"callback_url,omitempty"`
	CallbackData map[string]string `json:"callback_data,omitempty"`
}

// TaskOption is a functional option for configuring a task
type TaskOption func(*Task)

// NewTask creates a new task with the given prompt and options
func NewTask(prompt string, opts ...TaskOption) *Task {
	task := &Task{
		ID:        uuid.New().String(),
		Prompt:    prompt,
		Priority:  TaskPriorityNormal,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(task)
	}

	return task
}

// WithTaskID sets a custom task ID
func WithTaskID(id string) TaskOption {
	return func(t *Task) {
		t.ID = id
	}
}

// WithContext sets the task context
func WithContext(context string) TaskOption {
	return func(t *Task) {
		t.Context = context
	}
}

// WithPriority sets the task priority
func WithPriority(priority TaskPriority) TaskOption {
	return func(t *Task) {
		t.Priority = priority
	}
}

// WithAgentConfig sets a custom agent configuration for this task
func WithAgentConfig(config *AgentConfig) TaskOption {
	return func(t *Task) {
		t.AgentConfig = config
	}
}

// WithMetadata sets the task metadata
func WithMetadata(metadata map[string]interface{}) TaskOption {
	return func(t *Task) {
		t.Metadata = metadata
	}
}

// WithTimeout sets the task timeout
func WithTimeout(timeout time.Duration) TaskOption {
	return func(t *Task) {
		t.Timeout = timeout
	}
}

// WithCallback sets the callback URL and data
func WithCallback(url string, data map[string]string) TaskOption {
	return func(t *Task) {
		t.CallbackURL = url
		t.CallbackData = data
	}
}

// BatchTask represents a collection of tasks to be executed together
type BatchTask struct {
	ID          string                 `json:"id"`
	Tasks       []*Task                `json:"tasks"`
	Parallel    bool                   `json:"parallel"` // If true, execute tasks in parallel
	MaxParallel int                    `json:"max_parallel,omitempty"` // Max concurrent tasks (0 = unlimited)
	Status      TaskStatus             `json:"status"`
	Results     []*AgentResult         `json:"results,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// NewBatchTask creates a new batch task
func NewBatchTask(tasks []*Task, parallel bool, maxParallel int) *BatchTask {
	return &BatchTask{
		ID:          uuid.New().String(),
		Tasks:       tasks,
		Parallel:    parallel,
		MaxParallel: maxParallel,
		Status:      TaskStatusPending,
		Results:     make([]*AgentResult, 0, len(tasks)),
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
	}
}

// TaskQueue represents a priority queue for tasks
type TaskQueue struct {
	tasks    []*Task
	capacity int
}

// NewTaskQueue creates a new task queue with the given capacity
func NewTaskQueue(capacity int) *TaskQueue {
	return &TaskQueue{
		tasks:    make([]*Task, 0, capacity),
		capacity: capacity,
	}
}

// Push adds a task to the queue based on priority
func (q *TaskQueue) Push(task *Task) bool {
	if len(q.tasks) >= q.capacity {
		return false
	}

	// Find insertion point based on priority (higher priority first)
	insertIdx := len(q.tasks)
	for i, t := range q.tasks {
		if task.Priority > t.Priority {
			insertIdx = i
			break
		}
	}

	// Insert at the correct position
	q.tasks = append(q.tasks, nil)
	copy(q.tasks[insertIdx+1:], q.tasks[insertIdx:])
	q.tasks[insertIdx] = task

	return true
}

// Pop removes and returns the highest priority task
func (q *TaskQueue) Pop() *Task {
	if len(q.tasks) == 0 {
		return nil
	}

	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task
}

// Peek returns the highest priority task without removing it
func (q *TaskQueue) Peek() *Task {
	if len(q.tasks) == 0 {
		return nil
	}
	return q.tasks[0]
}

// Len returns the number of tasks in the queue
func (q *TaskQueue) Len() int {
	return len(q.tasks)
}

// IsEmpty returns true if the queue is empty
func (q *TaskQueue) IsEmpty() bool {
	return len(q.tasks) == 0
}

// IsFull returns true if the queue is at capacity
func (q *TaskQueue) IsFull() bool {
	return len(q.tasks) >= q.capacity
}

// Clear removes all tasks from the queue
func (q *TaskQueue) Clear() {
	q.tasks = q.tasks[:0]
}
