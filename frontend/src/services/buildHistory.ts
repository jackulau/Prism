export type BuildStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled';

export interface Build {
  id: string;
  workspaceId?: string;
  orgWorkspaceId?: string;
  command: string;
  status: BuildStatus;
  exitCode?: number;
  startedAt: string;
  completedAt?: string;
  durationMs?: number;
}

export interface BuildLog {
  id: string;
  buildId: string;
  stream: 'stdout' | 'stderr';
  content: string;
  timestamp: string;
}

interface BuildListParams {
  limit?: number;
  offset?: number;
  status?: BuildStatus;
}

interface BuildListResponse {
  builds: Build[];
  total: number;
}

interface BuildLogsResponse {
  logs: BuildLog[];
}

const API_BASE_URL = '/api/v1';

class BuildHistoryService {
  private getAuthHeaders(): HeadersInit {
    const token = localStorage.getItem('token');
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  async list(params?: BuildListParams): Promise<{ data?: BuildListResponse; error?: string }> {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    if (params?.status) searchParams.set('status', params.status);

    const queryString = searchParams.toString();
    const url = `${API_BASE_URL}/builds${queryString ? `?${queryString}` : ''}`;

    try {
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...this.getAuthHeaders(),
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        return { error: errorData.error || 'Failed to fetch builds' };
      }

      const data = await response.json();
      return { data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }

  async get(id: string): Promise<{ data?: Build; error?: string }> {
    try {
      const response = await fetch(`${API_BASE_URL}/builds/${id}`, {
        headers: {
          'Content-Type': 'application/json',
          ...this.getAuthHeaders(),
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        return { error: errorData.error || 'Failed to fetch build' };
      }

      const data = await response.json();
      return { data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }

  async getLogs(id: string): Promise<{ data?: BuildLogsResponse; error?: string }> {
    try {
      const response = await fetch(`${API_BASE_URL}/builds/${id}/logs`, {
        headers: {
          'Content-Type': 'application/json',
          ...this.getAuthHeaders(),
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        return { error: errorData.error || 'Failed to fetch build logs' };
      }

      const data = await response.json();
      return { data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }

  async delete(id: string): Promise<{ success?: boolean; error?: string }> {
    try {
      const response = await fetch(`${API_BASE_URL}/builds/${id}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          ...this.getAuthHeaders(),
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        return { error: errorData.error || 'Failed to delete build' };
      }

      return { success: true };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }

  async cancel(id: string): Promise<{ data?: Build; error?: string }> {
    try {
      const response = await fetch(`${API_BASE_URL}/builds/${id}/cancel`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...this.getAuthHeaders(),
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        return { error: errorData.error || 'Failed to cancel build' };
      }

      const data = await response.json();
      return { data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }
}

export const buildHistoryService = new BuildHistoryService();
