/**
 * Types for batch task execution
 */

/** Priority level for batch tasks */
export type BatchTaskPriority = 'low' | 'normal' | 'high';

/** Status of a batch task during execution */
export type BatchTaskStatus = 'pending' | 'running' | 'completed' | 'failed';

/** Single batch task item */
export interface BatchTask {
  id: string;
  prompt: string;
  context?: string;
  priority: BatchTaskPriority;
  status: BatchTaskStatus;
  createdAt: Date;
  result?: string;
  error?: string;
}

/** Form data for creating/editing a batch task */
export interface BatchTaskFormData {
  prompt: string;
  context?: string;
  priority: BatchTaskPriority;
}
