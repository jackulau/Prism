/**
 * Batch Execution Types
 *
 * Type definitions for managing parallel batch execution of agent tasks.
 */

/**
 * Status of a batch execution
 */
export type BatchStatus =
  | 'pending'           // Batch created but not started
  | 'running'           // Batch is actively executing
  | 'completed'         // All tasks completed successfully
  | 'partially_completed' // Some tasks succeeded, some failed
  | 'failed'            // All tasks failed or batch-level error
  | 'cancelled';        // Batch was cancelled by user

/**
 * Status of an individual task within a batch
 */
export type BatchTaskStatus =
  | 'pending'           // Task waiting to be executed
  | 'queued'            // Task in execution queue
  | 'running'           // Task currently executing
  | 'completed'         // Task completed successfully
  | 'failed'            // Task failed
  | 'cancelled';        // Task was cancelled

/**
 * Configuration for a batch task
 */
export interface BatchTaskConfig {
  /** Unique identifier for the task */
  id: string;
  /** Display name for the task */
  name: string;
  /** The prompt/message to send to the agent */
  prompt: string;
  /** Optional conversation ID to use (creates new if not provided) */
  conversationId?: string;
  /** Provider to use for this task */
  provider?: string;
  /** Model to use for this task */
  model?: string;
  /** Optional metadata for the task */
  metadata?: Record<string, unknown>;
  /** Priority for execution (higher = earlier) */
  priority?: number;
  /** Maximum retries on failure */
  maxRetries?: number;
}

/**
 * A task within a batch execution
 */
export interface BatchTask extends BatchTaskConfig {
  /** Current status of the task */
  status: BatchTaskStatus;
  /** Number of retry attempts made */
  retryCount: number;
  /** Error message if task failed */
  error?: string;
  /** Task output/result if completed */
  output?: string;
  /** When the task started executing */
  startedAt?: Date;
  /** When the task finished */
  completedAt?: Date;
  /** Token usage for this task */
  tokenUsage?: {
    input: number;
    output: number;
    total: number;
  };
}

/**
 * Progress information for a batch
 */
export interface BatchProgressInfo {
  /** Total number of tasks in the batch */
  total: number;
  /** Number of tasks completed (success or failure) */
  completed: number;
  /** Number of tasks that succeeded */
  succeeded: number;
  /** Number of tasks that failed */
  failed: number;
  /** Number of tasks currently running */
  running: number;
  /** Number of tasks pending */
  pending: number;
  /** Progress percentage (0-100) */
  percentage: number;
  /** Estimated time remaining in milliseconds */
  estimatedTimeRemaining?: number;
  /** Average time per task in milliseconds */
  averageTaskTime?: number;
}

/**
 * Result of a completed batch execution
 */
export interface BatchResult {
  /** Unique identifier for this batch execution */
  batchId: string;
  /** Final status of the batch */
  status: BatchStatus;
  /** All tasks and their final states */
  tasks: BatchTask[];
  /** Progress info at completion */
  progress: BatchProgressInfo;
  /** When the batch started */
  startedAt: Date;
  /** When the batch completed */
  completedAt: Date;
  /** Total execution time in milliseconds */
  totalDuration: number;
  /** Combined token usage across all tasks */
  totalTokenUsage: {
    input: number;
    output: number;
    total: number;
  };
  /** Batch-level error if any */
  error?: string;
}

/**
 * Configuration for starting a batch execution
 */
export interface BatchExecutionConfig {
  /** Maximum concurrent tasks to run */
  maxConcurrency?: number;
  /** Whether to stop all tasks on first failure */
  stopOnFirstFailure?: boolean;
  /** Global timeout for the entire batch in milliseconds */
  timeoutMs?: number;
  /** Per-task timeout in milliseconds */
  taskTimeoutMs?: number;
  /** Default provider for all tasks */
  defaultProvider?: string;
  /** Default model for all tasks */
  defaultModel?: string;
}

/**
 * State of a batch execution
 */
export interface BatchExecutionState {
  /** Unique identifier for this batch */
  batchId: string | null;
  /** Current status of the batch */
  status: BatchStatus;
  /** All tasks in the batch */
  tasks: Map<string, BatchTask>;
  /** Ordered list of task IDs for display */
  taskOrder: string[];
  /** Progress information */
  progress: BatchProgressInfo;
  /** Batch configuration */
  config: BatchExecutionConfig;
  /** When the batch was created */
  createdAt: Date | null;
  /** When the batch started executing */
  startedAt: Date | null;
  /** When the batch completed */
  completedAt: Date | null;
  /** Batch-level error message */
  error: string | null;
}

/**
 * WebSocket message types for batch execution
 */
export type BatchMessageType =
  | 'agent.run_parallel'       // Start batch execution
  | 'agent.batch_progress'     // Batch progress update
  | 'agent.batch_completed'    // Batch finished
  | 'agent.task_started'       // Individual task started
  | 'agent.task_progress'      // Individual task progress
  | 'agent.task_completed'     // Individual task finished
  | 'agent.batch_stop';        // Stop batch execution

/**
 * WebSocket message for batch operations (outgoing)
 */
export interface BatchWSMessageOutgoing {
  type: 'agent.run_parallel' | 'agent.batch_stop';
  batch_id?: string;
  tasks?: BatchTaskConfig[];
  config?: BatchExecutionConfig;
}

/**
 * WebSocket message for batch updates (incoming)
 */
export interface BatchWSMessageIncoming {
  type: BatchMessageType;
  batch_id: string;
  task_id?: string;
  status?: BatchStatus | BatchTaskStatus;
  progress?: Partial<BatchProgressInfo>;
  error?: string;
  output?: string;
  token_usage?: {
    input: number;
    output: number;
    total: number;
  };
  timestamp?: string;
}

/**
 * Batch execution history entry
 */
export interface BatchHistoryEntry {
  /** Batch ID */
  id: string;
  /** Batch status */
  status: BatchStatus;
  /** Number of tasks */
  taskCount: number;
  /** Number of successful tasks */
  successCount: number;
  /** Number of failed tasks */
  failCount: number;
  /** When the batch was created */
  createdAt: Date;
  /** When the batch completed */
  completedAt?: Date;
  /** Total duration in ms */
  duration?: number;
}
