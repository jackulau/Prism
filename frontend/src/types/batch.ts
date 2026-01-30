// Batch execution types for parallel agent processing

export type BatchTaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface BatchTask {
  id: string;
  prompt: string;
  systemPrompt?: string;
  status: BatchTaskStatus;
  progress?: number;
  result?: string;
  error?: string;
  startedAt?: Date;
  completedAt?: Date;
  duration?: number;
  tokensUsed?: number;
}

export interface BatchConfig {
  provider: string;
  model: string;
  maxConcurrent: number;
  timeout: number;
  temperature: number;
  maxTokens: number;
}

export type BatchExecutionStatus = 'idle' | 'running' | 'paused' | 'completed' | 'cancelled';

export interface BatchExecution {
  id: string;
  status: BatchExecutionStatus;
  config: BatchConfig;
  tasks: BatchTask[];
  startedAt?: Date;
  completedAt?: Date;
  totalTasks: number;
  completedTasks: number;
  failedTasks: number;
}

export interface BatchResult {
  taskId: string;
  prompt: string;
  status: BatchTaskStatus;
  result?: string;
  error?: string;
  duration?: number;
  tokensUsed?: number;
}
