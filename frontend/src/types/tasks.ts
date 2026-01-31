/**
 * Task queue related TypeScript types.
 * These types match the backend /api/v1/tasks/stats endpoint response.
 */

/** Possible task statuses */
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

/** Task statistics from the backend API */
export interface TaskStats {
  /** Total number of tasks */
  total: number;
  /** Number of pending tasks (queue depth) */
  pending: number;
  /** Number of currently running tasks */
  running: number;
  /** Number of completed tasks */
  completed: number;
  /** Number of failed tasks */
  failed: number;
  /** Number of cancelled tasks */
  cancelled: number;
  /** Timestamp of the stats snapshot */
  timestamp: string;
}

/** Response from the task stats endpoint */
export interface TaskStatsResponse {
  stats: TaskStats;
}
