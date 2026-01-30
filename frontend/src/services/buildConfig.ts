const API_BASE_URL = '/api/v1';

// Types
export interface BuildConfig {
  id: string;
  workspaceId?: string;
  orgWorkspaceId?: string;
  name: string;
  description?: string;
  isDefault: boolean;
  commands: BuildCommand[];
  envVars: BuildEnvVar[];
  createdAt: string;
  updatedAt: string;
}

export interface BuildCommand {
  id: string;
  configId: string;
  name: string;
  command: string;
  workingDirectory?: string;
  runOrder: number;
  isEnabled: boolean;
}

export interface BuildEnvVar {
  id: string;
  configId: string;
  key: string;
  value: string; // Masked for secrets in responses
  isSecret: boolean;
}

export interface CreateBuildConfigInput {
  workspaceId?: string;
  orgWorkspaceId?: string;
  name: string;
  description?: string;
}

export interface UpdateBuildConfigInput {
  name?: string;
  description?: string;
}

export interface CreateCommandInput {
  name: string;
  command: string;
  workingDirectory?: string;
  runOrder?: number;
  isEnabled?: boolean;
}

export interface UpdateCommandInput {
  name?: string;
  command?: string;
  workingDirectory?: string;
  isEnabled?: boolean;
}

export interface SetEnvVarInput {
  key: string;
  value: string;
  isSecret?: boolean;
}

interface ApiResponse<T> {
  data?: T;
  error?: string;
}

class BuildConfigService {
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

      const contentLength = response.headers.get('Content-Length');
      const contentType = response.headers.get('Content-Type');
      const hasJsonContent = contentType?.includes('application/json');
      const hasContent = contentLength !== '0' && contentLength !== null;

      let data: T | undefined;
      if (hasContent || hasJsonContent) {
        const text = await response.text();
        if (text) {
          try {
            data = JSON.parse(text);
          } catch {
            if (!response.ok) {
              return { error: text || 'An error occurred' };
            }
          }
        }
      }

      if (!response.ok) {
        return { error: (data as { error?: string })?.error || 'An error occurred' };
      }

      return { data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }

  // Config operations
  async list(workspaceId?: string): Promise<ApiResponse<{ configs: BuildConfig[] }>> {
    const params = workspaceId ? `?workspaceId=${encodeURIComponent(workspaceId)}` : '';
    return this.request(`/build-configs${params}`);
  }

  async create(data: CreateBuildConfigInput): Promise<ApiResponse<BuildConfig>> {
    return this.request('/build-configs', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async get(id: string): Promise<ApiResponse<BuildConfig>> {
    return this.request(`/build-configs/${id}`);
  }

  async update(id: string, data: UpdateBuildConfigInput): Promise<ApiResponse<BuildConfig>> {
    return this.request(`/build-configs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async delete(id: string): Promise<ApiResponse<{ success: boolean }>> {
    return this.request(`/build-configs/${id}`, {
      method: 'DELETE',
    });
  }

  async setDefault(id: string): Promise<ApiResponse<{ success: boolean }>> {
    return this.request(`/build-configs/${id}/default`, {
      method: 'POST',
    });
  }

  // Command operations
  async addCommand(configId: string, data: CreateCommandInput): Promise<ApiResponse<BuildCommand>> {
    return this.request(`/build-configs/${configId}/commands`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateCommand(
    configId: string,
    cmdId: string,
    data: UpdateCommandInput
  ): Promise<ApiResponse<BuildCommand>> {
    return this.request(`/build-configs/${configId}/commands/${cmdId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteCommand(configId: string, cmdId: string): Promise<ApiResponse<{ success: boolean }>> {
    return this.request(`/build-configs/${configId}/commands/${cmdId}`, {
      method: 'DELETE',
    });
  }

  async reorderCommands(configId: string, order: string[]): Promise<ApiResponse<{ success: boolean }>> {
    return this.request(`/build-configs/${configId}/commands/order`, {
      method: 'PUT',
      body: JSON.stringify({ order }),
    });
  }

  // Env var operations
  async setEnvVar(configId: string, data: SetEnvVarInput): Promise<ApiResponse<BuildEnvVar>> {
    return this.request(`/build-configs/${configId}/env`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async deleteEnvVar(configId: string, key: string): Promise<ApiResponse<{ success: boolean }>> {
    return this.request(`/build-configs/${configId}/env/${encodeURIComponent(key)}`, {
      method: 'DELETE',
    });
  }

  async getEnvVars(configId: string): Promise<ApiResponse<{ envVars: BuildEnvVar[] }>> {
    return this.request(`/build-configs/${configId}/env`);
  }
}

export const buildConfigService = new BuildConfigService();
