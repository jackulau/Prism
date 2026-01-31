import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

// API base URL
function getApiUrl(): string {
  if (import.meta.env.VITE_API_URL) {
    return import.meta.env.VITE_API_URL;
  }
  if (import.meta.env.DEV) {
    return '';
  }
  return 'http://localhost:3001';
}

// Helper to get auth headers
function getAuthHeaders(): HeadersInit {
  const token = localStorage.getItem('prism_token');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

// Types matching backend workflow package
export type WorkflowStatus = 'pending' | 'running' | 'paused' | 'completed' | 'failed' | 'cancelled';
export type StepStatus = 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
export type StepType = 'agent' | 'tool' | 'condition' | 'parallel' | 'wait' | 'transform';

export interface StepResult {
  step_id: string;
  step_name: string;
  status: StepStatus;
  output?: unknown;
  error?: string;
  duration: number;
  retry_count?: number;
  started_at: string;
  completed_at: string;
  metadata?: Record<string, unknown>;
}

export interface Step {
  id: string;
  name: string;
  description?: string;
  type: StepType;
  config: Record<string, unknown>;
  condition?: {
    type: string;
    expression?: string;
    state_key?: string;
    operator?: string;
    value?: string;
  };
  on_success?: string;
  on_failure?: string;
  timeout?: number;
  retry_policy?: {
    max_retries: number;
    delay: number;
    backoff_type?: string;
    max_delay?: number;
  };
}

export interface Workflow {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  steps: Step[];
  status: WorkflowStatus;
  current_step: number;
  state?: Record<string, unknown>;
  error?: string;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
  step_results?: StepResult[];
}

export interface WorkflowFilter {
  status?: WorkflowStatus;
  name?: string;
  limit?: number;
  offset?: number;
}

// Query keys
export const workflowKeys = {
  all: ['workflows'] as const,
  lists: () => [...workflowKeys.all, 'list'] as const,
  list: (filters: WorkflowFilter) => [...workflowKeys.lists(), filters] as const,
  details: () => [...workflowKeys.all, 'detail'] as const,
  detail: (id: string) => [...workflowKeys.details(), id] as const,
  state: (id: string) => [...workflowKeys.all, 'state', id] as const,
};

// Fetch workflows list
async function fetchWorkflows(filters: WorkflowFilter): Promise<{ workflows: Workflow[] }> {
  const params = new URLSearchParams();
  if (filters.status) params.set('status', filters.status);
  if (filters.name) params.set('name', filters.name);
  if (filters.limit) params.set('limit', String(filters.limit));
  if (filters.offset) params.set('offset', String(filters.offset));

  const response = await fetch(`${getApiUrl()}/api/v1/workflows?${params}`, {
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
  });

  if (!response.ok) {
    throw new Error('Failed to fetch workflows');
  }

  return response.json();
}

// Fetch single workflow
async function fetchWorkflow(id: string): Promise<{ workflow: Workflow }> {
  const response = await fetch(`${getApiUrl()}/api/v1/workflows/${id}`, {
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
  });

  if (!response.ok) {
    throw new Error('Failed to fetch workflow');
  }

  return response.json();
}

// Fetch workflow state
async function fetchWorkflowState(id: string): Promise<{
  state: Record<string, unknown>;
  status: WorkflowStatus;
  current_step: number;
}> {
  const response = await fetch(`${getApiUrl()}/api/v1/workflows/${id}/state`, {
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
  });

  if (!response.ok) {
    throw new Error('Failed to fetch workflow state');
  }

  return response.json();
}

// Start workflow
async function startWorkflow(id: string, state?: Record<string, unknown>): Promise<void> {
  const response = await fetch(`${getApiUrl()}/api/v1/workflows/${id}/start`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    body: state ? JSON.stringify({ state }) : undefined,
  });

  if (!response.ok) {
    const data = await response.json();
    throw new Error(data.error || 'Failed to start workflow');
  }
}

// Pause workflow
async function pauseWorkflow(id: string): Promise<void> {
  const response = await fetch(`${getApiUrl()}/api/v1/workflows/${id}/pause`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
  });

  if (!response.ok) {
    const data = await response.json();
    throw new Error(data.error || 'Failed to pause workflow');
  }
}

// Resume workflow
async function resumeWorkflow(id: string): Promise<void> {
  const response = await fetch(`${getApiUrl()}/api/v1/workflows/${id}/resume`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
  });

  if (!response.ok) {
    const data = await response.json();
    throw new Error(data.error || 'Failed to resume workflow');
  }
}

// Cancel workflow
async function cancelWorkflow(id: string): Promise<void> {
  const response = await fetch(`${getApiUrl()}/api/v1/workflows/${id}/cancel`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
  });

  if (!response.ok) {
    const data = await response.json();
    throw new Error(data.error || 'Failed to cancel workflow');
  }
}

// Hooks

export function useWorkflows(filters: WorkflowFilter = {}) {
  return useQuery({
    queryKey: workflowKeys.list(filters),
    queryFn: () => fetchWorkflows(filters),
    select: (data) => data.workflows,
    refetchInterval: 5000, // Poll every 5 seconds for running workflows
  });
}

export function useWorkflow(id: string | undefined) {
  return useQuery({
    queryKey: workflowKeys.detail(id!),
    queryFn: () => fetchWorkflow(id!),
    select: (data) => data.workflow,
    enabled: !!id,
    refetchInterval: (query) => {
      // Poll more frequently for running workflows
      // Note: query.state.data has raw queryFn result type, not selected type
      const raw = query.state.data as { workflow: Workflow } | undefined;
      const workflow = raw?.workflow;
      if (workflow && (workflow.status === 'running' || workflow.status === 'paused')) {
        return 2000;
      }
      return false;
    },
  });
}

export function useWorkflowState(id: string | undefined) {
  return useQuery({
    queryKey: workflowKeys.state(id!),
    queryFn: () => fetchWorkflowState(id!),
    enabled: !!id,
    refetchInterval: 2000, // Poll every 2 seconds
  });
}

export function useStartWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, state }: { id: string; state?: Record<string, unknown> }) =>
      startWorkflow(id, state),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() });
    },
  });
}

export function usePauseWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => pauseWorkflow(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() });
    },
  });
}

export function useResumeWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => resumeWorkflow(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() });
    },
  });
}

export function useCancelWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => cancelWorkflow(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() });
    },
  });
}

// Helper to calculate workflow duration
export function getWorkflowDuration(workflow: Workflow): number | null {
  if (!workflow.started_at) return null;

  const start = new Date(workflow.started_at).getTime();
  const end = workflow.completed_at
    ? new Date(workflow.completed_at).getTime()
    : Date.now();

  return end - start;
}

// Format duration for display
export function formatDuration(ms: number | null): string {
  if (ms === null) return '-';

  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  }
  return `${seconds}s`;
}

// Format relative time
export function formatRelativeTime(date: string): string {
  const now = Date.now();
  const then = new Date(date).getTime();
  const diff = now - then;

  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days} day${days > 1 ? 's' : ''} ago`;
  if (hours > 0) return `${hours} hour${hours > 1 ? 's' : ''} ago`;
  if (minutes > 0) return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
  return 'just now';
}
