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

      // Handle empty responses (204 No Content, etc.)
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
            // Response is not valid JSON
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

  async searchConversations(query: string, limit = 20) {
    return this.request<{
      conversations: Array<{
        id: string;
        title: string;
        provider: string;
        model: string;
        created_at: string;
        updated_at: string;
      }>;
    }>(`/conversations/search?q=${encodeURIComponent(query)}&limit=${limit}`);
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
        conversation_id?: string;
        parent_id?: string;
        role: string;
        content: string;
        thinking_content?: string;
        tool_calls?: unknown[];
        tool_call_id?: string;
        provider?: string;
        model?: string;
        status?: string;
        input_tokens?: number;
        output_tokens?: number;
        finish_reason?: string;
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

  async browseDirectories(path: string = '/') {
    return this.request<{
      current_path: string;
      parent_path: string;
      directories: Array<{ name: string; path: string }>;
    }>(`/workspace/browse?path=${encodeURIComponent(path)}`);
  }

  async openFolderPicker() {
    return this.request<{ success?: boolean; path?: string; cancelled?: boolean }>(
      '/workspace/pick-folder',
      { method: 'POST' }
    );
  }

  async listRecentWorkspaces() {
    return this.request<{
      workspaces: Array<{
        id: string;
        path: string;
        name: string;
        is_current: boolean;
        last_accessed_at?: string;
      }>;
    }>('/workspace/recent');
  }

  async removeWorkspace(id: string) {
    return this.request(`/workspace/${id}`, { method: 'DELETE' });
  }

  async setCurrentWorkspace(id: string) {
    return this.request<{ success: boolean; path: string }>(
      `/workspace/${id}/current`,
      { method: 'POST' }
    );
  }

  async renameSandboxFile(sourcePath: string, destPath: string) {
    return this.request<{ success: boolean }>('/sandbox/files/rename', {
      method: 'POST',
      body: JSON.stringify({ source_path: sourcePath, dest_path: destPath }),
    });
  }

  async createSandboxDirectory(path: string) {
    return this.request<{ success: boolean }>('/sandbox/files/mkdir', {
      method: 'POST',
      body: JSON.stringify({ path }),
    });
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

  // CloudProvider API methods
  async listCloudProviders() {
    return this.request<{
      providers: Array<{
        name: string;
        has_credentials: boolean;
      }>;
    }>('/cloud/providers');
  }

  async createCloudAgent(params: {
    provider: string;
    name?: string;
    system_prompt?: string;
    model?: string;
    tools?: string[];
    metadata?: Record<string, string>;
  }) {
    return this.request<{
      id: string;
      provider_id: string;
      provider_name: string;
      name: string;
      status: string;
      created_at: string;
      updated_at?: string;
      model?: string;
      system_prompt?: string;
    }>('/cloud/agents', {
      method: 'POST',
      body: JSON.stringify(params),
    });
  }

  async getCloudAgent(agentId: string) {
    return this.request<{
      id: string;
      provider_id: string;
      provider_name: string;
      name: string;
      status: string;
      created_at: string;
      updated_at?: string;
      model?: string;
      system_prompt?: string;
    }>(`/cloud/agents/${agentId}`);
  }

  async deleteCloudAgent(agentId: string) {
    return this.request(`/cloud/agents/${agentId}`, { method: 'DELETE' });
  }

  async getCloudAgentMessages(agentId: string) {
    return this.request<{
      messages: Array<{
        id: string;
        role: string;
        content: string;
        timestamp: string;
        tool_calls?: Array<{
          id: string;
          name: string;
          parameters: Record<string, unknown>;
          result?: unknown;
          status: string;
        }>;
        images?: Array<{
          url?: string;
          base64?: string;
          mime_type?: string;
        }>;
      }>;
    }>(`/cloud/agents/${agentId}/messages`);
  }

  async sendCloudAgentMessage(
    agentId: string,
    message: string,
    images?: Array<{ url?: string; base64?: string; mime_type?: string }>
  ) {
    return this.request<{ success: boolean }>(`/cloud/agents/${agentId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ message, images }),
    });
  }

  // SSO Authentication
  async ssoAuthorize(organization: string) {
    return this.request<{ authorization_url: string }>('/auth/sso/authorize', {
      method: 'POST',
      body: JSON.stringify({ organization }),
    });
  }

  async ssoCallback(code: string, state: string) {
    return this.request<{
      access_token: string;
      refresh_token: string;
      user: { id: string; email: string; created_at: string };
    }>('/auth/sso/callback', {
      method: 'POST',
      body: JSON.stringify({ code, state }),
    });
  }

  // MFA Authentication
  async mfaGetStatus() {
    return this.request<{ enabled: boolean }>('/auth/mfa/status');
  }

  async mfaStartSetup() {
    return this.request<{ secret: string; qr_url: string }>('/auth/mfa/setup', {
      method: 'POST',
    });
  }

  async mfaVerifySetup(code: string) {
    return this.request<{ backup_codes: string[] }>('/auth/mfa/verify-setup', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });
  }

  async mfaValidate(sessionToken: string, code: string) {
    return this.request<{
      access_token: string;
      refresh_token: string;
      user: { id: string; email: string; created_at: string };
    }>('/auth/mfa/validate', {
      method: 'POST',
      body: JSON.stringify({ session_token: sessionToken, code }),
    });
  }

  async mfaDisable(password: string, code: string) {
    return this.request<void>('/auth/mfa/disable', {
      method: 'POST',
      body: JSON.stringify({ password, code }),
    });
  }

  async mfaRegenerateBackupCodes(code: string) {
    return this.request<{ backup_codes: string[] }>('/auth/mfa/backup-codes', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });
  }

  async mfaVerifyBackupCode(sessionToken: string, code: string) {
    return this.request<{
      access_token: string;
      refresh_token: string;
      user: { id: string; email: string; created_at: string };
    }>('/auth/mfa/verify-backup', {
      method: 'POST',
      body: JSON.stringify({ session_token: sessionToken, code }),
    });
  }
}

export const apiService = new ApiService();
