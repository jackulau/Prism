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

  // MCP Server Management
  async getMCPServers() {
    return this.request<{
      servers: Array<{
        id: string;
        name: string;
        url: string;
        enabled: boolean;
        has_api_key?: boolean;
        manifest?: {
          name: string;
          version: string;
          description: string;
          tool_count: number;
        };
        created_at?: string;
        updated_at?: string;
        last_sync?: string;
        last_error?: string;
      }>;
    }>('/mcp/servers');
  }

  async addMCPServer(name: string, url: string, apiKey?: string) {
    return this.request<{
      id: string;
      name: string;
      url: string;
      enabled: boolean;
      manifest?: {
        name: string;
        version: string;
        description: string;
        tool_count: number;
      };
      created_at: string;
    }>('/mcp/servers', {
      method: 'POST',
      body: JSON.stringify({ name, url, api_key: apiKey }),
    });
  }

  async updateMCPServer(id: string, data: { name?: string; url?: string; apiKey?: string }) {
    return this.request<{ success: boolean }>(`/mcp/servers/${id}`, {
      method: 'PUT',
      body: JSON.stringify({
        name: data.name,
        url: data.url,
        api_key: data.apiKey,
      }),
    });
  }

  async removeMCPServer(id: string) {
    return this.request<{ success: boolean }>(`/mcp/servers/${id}`, {
      method: 'DELETE',
    });
  }

  async enableMCPServer(id: string) {
    return this.request<{ success: boolean; enabled: boolean }>(`/mcp/servers/${id}/enable`, {
      method: 'POST',
    });
  }

  async disableMCPServer(id: string) {
    return this.request<{ success: boolean; enabled: boolean }>(`/mcp/servers/${id}/disable`, {
      method: 'POST',
    });
  }

  async testMCPServer(id: string) {
    return this.request<{
      success: boolean;
      error?: string;
      manifest?: {
        name: string;
        version: string;
        description: string;
        tool_count: number;
      };
    }>(`/mcp/servers/${id}/test`, {
      method: 'POST',
    });
  }

  async refreshMCPServer(id: string) {
    return this.request<{
      success: boolean;
      last_sync?: string;
      manifest?: {
        name: string;
        version: string;
        description: string;
        tool_count: number;
      };
    }>(`/mcp/servers/${id}/refresh`, {
      method: 'POST',
    });
  }

  async reconnectMCPServer(id: string) {
    return this.request<{
      success: boolean;
      error?: string;
    }>(`/mcp/servers/${id}/reconnect`, {
      method: 'POST',
    });
  }

  async getMCPServerStatus(id: string) {
    return this.request<{
      connected: boolean;
      latency_ms?: number;
      error?: string;
      last_checked?: string;
    }>(`/mcp/servers/${id}/status`);
  }

  async getMCPServerTools(id: string) {
    return this.request<{
      tools: Array<{
        server_id: string;
        server_name: string;
        name: string;
        description: string;
        parameters?: Record<string, unknown>;
      }>;
    }>(`/mcp/servers/${id}/tools`);
  }

  async getMCPServerStats(id: string, timeRange: 'today' | 'week' | 'all' = 'all') {
    return this.request<{
      total_calls: number;
      successful_calls: number;
      failed_calls: number;
      average_response_ms: number;
    }>(`/mcp/servers/${id}/stats?range=${timeRange}`);
  }

  async getMCPTools() {
    return this.request<{
      tools: Array<{
        server_id: string;
        server_name: string;
        name: string;
        description: string;
        parameters?: Record<string, unknown>;
      }>;
    }>('/mcp/tools');
  }

  // API Key Management
  async listAPIKeys() {
    return this.request<{
      api_keys: Array<{
        id: string;
        name: string;
        prefix: string;
        scopes: string[];
        created_at: string;
        expires_at: string | null;
        last_used_at: string | null;
      }>;
    }>('/api-keys');
  }

  async createAPIKey(name: string, expiresInDays?: number, scopes?: string[]) {
    return this.request<{
      key: string;
      id: string;
      name: string;
      prefix: string;
      scopes: string[];
      created_at: string;
      expires_at: string | null;
      last_used_at: string | null;
    }>('/api-keys', {
      method: 'POST',
      body: JSON.stringify({
        name,
        expires_in_days: expiresInDays,
        scopes,
      }),
    });
  }

  async getAPIKey(id: string) {
    return this.request<{
      id: string;
      name: string;
      prefix: string;
      scopes: string[];
      created_at: string;
      expires_at: string | null;
      last_used_at: string | null;
    }>(`/api-keys/${id}`);
  }

  async deleteAPIKey(id: string) {
    return this.request<{ success: boolean; message: string }>(`/api-keys/${id}`, {
      method: 'DELETE',
    });
  }

  async updateAPIKeyName(id: string, name: string) {
    return this.request<{ success: boolean; message: string }>(`/api-keys/${id}/name`, {
      method: 'PUT',
      body: JSON.stringify({ name }),
    });
  }

  async rotateAPIKey(id: string) {
    return this.request<{
      key: string;
      id: string;
      name: string;
      prefix: string;
      scopes: string[];
      created_at: string;
      expires_at: string | null;
      last_used_at: string | null;
    }>(`/api-keys/${id}/rotate`, {
      method: 'POST',
    });
  }

  async listProviderKeyMetadata() {
    return this.request<{
      provider_keys: Array<{
        provider: string;
        is_active: boolean;
        created_at: string;
        last_used_at: string | null;
        use_count: number;
      }>;
    }>('/providers/keys/metadata');
  }

  // Audit Logs
  async getMyAuditLogs(params: { limit?: number; offset?: number }) {
    const queryParams = new URLSearchParams();
    if (params.limit) queryParams.set('limit', params.limit.toString());
    if (params.offset) queryParams.set('offset', params.offset.toString());

    return this.request<{
      logs: Array<{
        id: number;
        user_id: string | null;
        event_type: string;
        event_category: string;
        action: string;
        resource_type: string | null;
        resource_id: string | null;
        ip_address: string | null;
        user_agent: string | null;
        details: Record<string, unknown> | null;
        success: boolean;
        created_at: string;
      }>;
      limit: number;
      offset: number;
    }>(`/audit/logs/me?${queryParams.toString()}`);
  }

  async getAllAuditLogs(params: {
    category?: string;
    event_type?: string;
    start_date?: string;
    end_date?: string;
    success?: boolean;
    limit?: number;
    offset?: number;
  }) {
    const queryParams = new URLSearchParams();
    if (params.category) queryParams.set('category', params.category);
    if (params.event_type) queryParams.set('event_type', params.event_type);
    if (params.start_date) queryParams.set('start_date', params.start_date);
    if (params.end_date) queryParams.set('end_date', params.end_date);
    if (params.success !== undefined) queryParams.set('success', params.success.toString());
    if (params.limit) queryParams.set('limit', params.limit.toString());
    if (params.offset) queryParams.set('offset', params.offset.toString());

    return this.request<{
      logs: Array<{
        id: number;
        user_id: string | null;
        event_type: string;
        event_category: string;
        action: string;
        resource_type: string | null;
        resource_id: string | null;
        ip_address: string | null;
        user_agent: string | null;
        details: Record<string, unknown> | null;
        success: boolean;
        created_at: string;
      }>;
      total: number;
      limit: number;
      offset: number;
    }>(`/audit/logs?${queryParams.toString()}`);
  }

  async getAuditStats(period = '24h') {
    return this.request<{
      since: string;
      category_counts: Record<string, number>;
      auth_counts: Record<string, number>;
      provider_counts: Record<string, number>;
    }>(`/audit/stats?period=${period}`);
  }
}

export const apiService = new ApiService();
