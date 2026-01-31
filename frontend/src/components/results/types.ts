export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'partially_completed' | 'failed' | 'cancelled';

export type ExecutionType = 'batch' | 'swarm';

export type DateRangePreset = 'today' | 'week' | 'month' | 'custom';

export type SortField = 'date' | 'duration' | 'tokens' | 'status';
export type SortDirection = 'asc' | 'desc';

export interface ResultsFilters {
  dateRange: DateRangePreset;
  customDateStart?: Date;
  customDateEnd?: Date;
  statuses: ExecutionStatus[];
  types: ExecutionType[];
  sortField: SortField;
  sortDirection: SortDirection;
  searchQuery: string;
}

export interface ExecutionResult {
  id: string;
  name?: string;
  type: ExecutionType;
  status: ExecutionStatus;
  startedAt: Date;
  completedAt?: Date;
  durationMs?: number;
  taskCount: number;
  agentCount: number;
  totalTokens: number;
  cost?: number;
  error?: string;
  results?: AgentResult[];
}

export interface AgentResult {
  agentId: string;
  taskId: string;
  success: boolean;
  output?: string;
  error?: string;
  durationMs: number;
  usage?: {
    promptTokens: number;
    completionTokens: number;
    totalTokens: number;
  };
}

export interface ResultsPagination {
  page: number;
  pageSize: number;
  total: number;
}

export const DEFAULT_FILTERS: ResultsFilters = {
  dateRange: 'week',
  statuses: [],
  types: [],
  sortField: 'date',
  sortDirection: 'desc',
  searchQuery: '',
};

export const STATUS_OPTIONS: { value: ExecutionStatus; label: string }[] = [
  { value: 'pending', label: 'Pending' },
  { value: 'running', label: 'Running' },
  { value: 'completed', label: 'Completed' },
  { value: 'partially_completed', label: 'Partial' },
  { value: 'failed', label: 'Failed' },
  { value: 'cancelled', label: 'Cancelled' },
];

export const TYPE_OPTIONS: { value: ExecutionType; label: string }[] = [
  { value: 'batch', label: 'Batch' },
  { value: 'swarm', label: 'Swarm' },
];

export const DATE_RANGE_OPTIONS: { value: DateRangePreset; label: string }[] = [
  { value: 'today', label: 'Today' },
  { value: 'week', label: 'This Week' },
  { value: 'month', label: 'This Month' },
  { value: 'custom', label: 'Custom' },
];

export const SORT_OPTIONS: { value: SortField; label: string }[] = [
  { value: 'date', label: 'Date' },
  { value: 'duration', label: 'Duration' },
  { value: 'tokens', label: 'Tokens' },
  { value: 'status', label: 'Status' },
];
