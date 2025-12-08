const API_BASE_URL = '/api/v1';

interface ApiResponse<T> {
  data?: T;
  error?: string;
}

class ApiService {
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

  // Auth
  async register(email: string, password: string) {
    return this.request<{
      access_token: string;
      refresh_token: string;
      expires_at: string;
      user: { id: string; email: string; created_at: string };
    }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  }

  async login(email: string, password: string) {
    return this.request<{
      access_token: string;
      refresh_token: string;
      expires_at: string;
      user: { id: string; email: string; created_at: string };
    }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  }

  async logout() {
    return this.request('/auth/logout', { method: 'POST' });
  }

  async refreshToken(refreshToken: string) {
    return this.request<{
      access_token: string;
      refresh_token: string;
      expires_at: string;
      user: { id: string; email: string; created_at: string };
    }>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
  }

  async getMe() {
    return this.request<{ id: string; email: string; created_at: string }>('/auth/me');
  }

  // Conversations
  async listConversations(limit = 50, offset = 0) {
    return this.request<{
      conversations: Array<{
        id: string;
        title: string;
        provider: string;
        model: string;
        created_at: string;
        updated_at: string;
      }>;
    }>(`/conversations?limit=${limit}&offset=${offset}`);
  }

  async createConversation(provider: string, model: string, systemPrompt?: string) {
    return this.request<{
      id: string;
      title: string;
      provider: string;
      model: string;
      created_at: string;
      updated_at: string;
    }>('/conversations', {
      method: 'POST',
      body: JSON.stringify({ provider, model, system_prompt: systemPrompt }),
    });
  }

  async getConversation(id: string) {
    return this.request<{
      id: string;
      title: string;
      provider: string;
      model: string;
      system_prompt: string;
      created_at: string;
      updated_at: string;
    }>(`/conversations/${id}`);
  }

  async updateConversation(id: string, title: string) {
    return this.request(`/conversations/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ title }),
    });
  }

  async deleteConversation(id: string) {
    return this.request(`/conversations/${id}`, { method: 'DELETE' });
  }

  async getMessages(conversationId: string) {
    return this.request<{
      messages: Array<{
        id: string;
        role: string;
        content: string;
        tool_calls?: unknown[];
        tool_call_id?: string;
        created_at: string;
      }>;
    }>(`/conversations/${conversationId}/messages`);
  }

  // Providers
  async listProviders() {
    return this.request<{
      providers: Array<{
        name: string;
        models: Array<{
          id: string;
          name: string;
          context_window: number;
          supports_tools: boolean;
          supports_vision: boolean;
        }>;
        supports_tools: boolean;
        supports_vision: boolean;
      }>;
    }>('/providers');
  }

  async setProviderKey(provider: string, apiKey: string) {
    return this.request(`/providers/${provider}/key`, {
      method: 'POST',
      body: JSON.stringify({ api_key: apiKey }),
    });
  }

  async deleteProviderKey(provider: string) {
    return this.request(`/providers/${provider}/key`, { method: 'DELETE' });
  }

  async validateProviderKey(provider: string, apiKey: string) {
    return this.request<{ valid: boolean }>(`/providers/${provider}/validate`, {
      method: 'POST',
      body: JSON.stringify({ api_key: apiKey }),
    });
  }

  // Sandbox
  async getSandboxFiles() {
    return this.request<{
      files: Array<{
        name: string;
        path: string;
        is_directory: boolean;
        children?: Array<{
          name: string;
          path: string;
          is_directory: boolean;
        }>;
        size?: number;
        modified?: number;
      }>;
    }>('/sandbox/files');
  }

  async getSandboxFile(path: string) {
    return this.request<{
      path: string;
      content: string;
    }>(`/sandbox/files/${encodeURIComponent(path)}`);
  }

  async writeSandboxFile(path: string, content: string) {
    return this.request('/sandbox/files', {
      method: 'POST',
      body: JSON.stringify({ path, content }),
    });
  }

  async deleteSandboxFile(path: string) {
    return this.request(`/sandbox/files/${encodeURIComponent(path)}`, {
      method: 'DELETE',
    });
  }

  // Workspace/Project Management
  async setWorkspaceDirectory(directory: string) {
    return this.request<{ path: string; success: boolean }>('/workspace/directory', {
      method: 'POST',
      body: JSON.stringify({ directory }),
    });
  }

  async getWorkspaceDirectory() {
    return this.request<{ path: string }>('/workspace/directory');
  }

  // GitHub Integration
  async getGitHubStatus() {
    return this.request<{
      connected: boolean;
      username: string;
      connected_at: string;
    }>('/github/status');
  }

  async getGitHubAuthUrl() {
    return this.request<{ url: string }>('/oauth/github/authorize');
  }

  async getGitHubRepos() {
    return this.request<{
      repos: Array<{
        id: number;
        name: string;
        full_name: string;
        description: string;
        private: boolean;
        html_url: string;
        clone_url: string;
        default_branch: string;
        updated_at: string;
      }>;
    }>('/github/repos');
  }

  async cloneGitHubRepo(repoUrl: string, branch?: string) {
    return this.request<{
      success: boolean;
      path: string;
      message: string;
    }>('/github/clone', {
      method: 'POST',
      body: JSON.stringify({ repo_url: repoUrl, branch }),
    });
  }

  async disconnectGitHub() {
    return this.request('/github/disconnect', { method: 'DELETE' });
  }
}

export const apiService = new ApiService();
