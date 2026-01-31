/**
 * Task filtering types for the task queue UI.
 */

/** Task status values for filtering */
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

/** All possible status filter values including 'all' */
export type StatusFilter = TaskStatus | 'all';

/** Date range preset options */
export type DateRangePreset = 'today' | '7d' | '30d' | 'all' | 'custom';

/** Date range value for filtering */
export interface DateRange {
  preset: DateRangePreset;
  startDate: Date | null;
  endDate: Date | null;
}

/** Filter state for tasks */
export interface TaskFilterState {
  /** Search query for task prompts */
  search: string;
  /** Status filter (single selection or 'all') */
  status: StatusFilter;
  /** Date range for filtering */
  dateRange: DateRange;
  /** Agent ID filter (optional) */
  agentId: string | null;
  /** Conversation ID filter (optional) */
  conversationId: string | null;
}

/** Active filter for display as tag */
export interface ActiveFilter {
  key: keyof TaskFilterState;
  label: string;
  value: string;
}

/** Options for filter dropdowns */
export interface FilterOption<T = string> {
  value: T;
  label: string;
  count?: number;
}

/** Task data shape for filtering (matches backend API) */
export interface Task {
  id: string;
  prompt: string;
  status: TaskStatus;
  agentId?: string;
  agentName?: string;
  conversationId?: string;
  conversationName?: string;
  createdAt: Date | string;
  startedAt?: Date | string | null;
  completedAt?: Date | string | null;
  result?: unknown;
  error?: string;
}

/** Default filter state */
export const DEFAULT_FILTER_STATE: TaskFilterState = {
  search: '',
  status: 'all',
  dateRange: {
    preset: 'all',
    startDate: null,
    endDate: null,
  },
  agentId: null,
  conversationId: null,
};

/** Status filter options */
export const STATUS_OPTIONS: FilterOption<StatusFilter>[] = [
  { value: 'all', label: 'All Statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'running', label: 'Running' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'cancelled', label: 'Cancelled' },
];

/** Date range preset options */
export const DATE_RANGE_OPTIONS: FilterOption<DateRangePreset>[] = [
  { value: 'all', label: 'All Time' },
  { value: 'today', label: 'Today' },
  { value: '7d', label: 'Last 7 Days' },
  { value: '30d', label: 'Last 30 Days' },
  { value: 'custom', label: 'Custom Range' },
];
