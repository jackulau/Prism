const API_BASE_URL = '/api/v1';

export interface RemoteConfig {
  port?: number;
  password?: string;
  allowedIPs?: string[];
}

export interface RemoteStatus {
  enabled: boolean;
  port: number;
  password: string | null;
  connectionUrl: string | null;
  tlsEnabled: boolean;
  allowedIPs: string[];
}

export interface ConnectionInfo {
  publicIP: string | null;
  localIPs: string[];
  port: number;
  connectionUrl: string;
  tlsEnabled: boolean;
}

export interface RemoteSession {
  id: string;
  clientIP: string;
  connectedAt: string;
  lastActivity: string;
  bytesIn: number;
  bytesOut: number;
}

interface ApiResponse<T> {
  data?: T;
  error?: string;
}

class RemoteAccessService {
  private token: string | null = null;

  setToken(token: string | null) {
    this.token = token;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers,
      });

      const data = await response.json();

      if (!response.ok) {
        return { error: data.error || 'An error occurred' };
      }

      return { data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }

  async getStatus(): Promise<ApiResponse<RemoteStatus>> {
    return this.request<RemoteStatus>('/remote/status');
  }

  async enable(config: RemoteConfig): Promise<ApiResponse<RemoteStatus>> {
    return this.request<RemoteStatus>('/remote/enable', {
      method: 'POST',
      body: JSON.stringify(config),
    });
  }

  async disable(): Promise<ApiResponse<{ success: boolean }>> {
    return this.request<{ success: boolean }>('/remote/disable', {
      method: 'POST',
    });
  }

  async regeneratePassword(): Promise<ApiResponse<{ password: string }>> {
    return this.request<{ password: string }>('/remote/password/regenerate', {
      method: 'POST',
    });
  }

  async getSessions(): Promise<ApiResponse<{ sessions: RemoteSession[] }>> {
    return this.request<{ sessions: RemoteSession[] }>('/remote/sessions');
  }

  async kickSession(id: string): Promise<ApiResponse<{ success: boolean }>> {
    return this.request<{ success: boolean }>(`/remote/sessions/${id}/kick`, {
      method: 'POST',
    });
  }

  async getConnectionInfo(): Promise<ApiResponse<ConnectionInfo>> {
    return this.request<ConnectionInfo>('/remote/connection-info');
  }

  async updateConfig(config: RemoteConfig): Promise<ApiResponse<RemoteStatus>> {
    return this.request<RemoteStatus>('/remote/config', {
      method: 'PATCH',
      body: JSON.stringify(config),
    });
  }
}

export const remoteAccessService = new RemoteAccessService();
