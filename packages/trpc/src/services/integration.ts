import type {
  IntegrationStatus,
  DiscordSettings,
  SlackSettings,
  PostHogSettings,
  MCPServer,
  StdioMCPServer,
  MCPTool,
} from '../routers/integrations/schemas.js';

// Base URL for the Go backend API
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080/api/v1';

interface FetchOptions {
  token: string;
  method?: string;
  body?: unknown;
}

async function fetchAPI<T>(
  path: string,
  options: FetchOptions
): Promise<T> {
  const { token, method = 'GET', body } = options;

  const response = await fetch(`${API_BASE_URL}${path}`, {
    method,
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Request failed' })) as { error?: string };
    throw new Error(errorData.error || `Request failed with status ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export const integrationService = {
  // Get all integration statuses
  async getStatus(token: string): Promise<IntegrationStatus> {
    const data = await fetchAPI<{
      discord: { enabled: boolean; connected: boolean };
      slack: { enabled: boolean; connected: boolean; channel_id?: string };
      posthog: { enabled: boolean; connected: boolean };
    }>('/integrations/status', { token });

    return {
      discord: {
        type: 'discord',
        enabled: data.discord.enabled,
        connected: data.discord.connected,
      },
      slack: {
        type: 'slack',
        enabled: data.slack.enabled,
        connected: data.slack.connected,
        channelId: data.slack.channel_id,
      },
      posthog: {
        type: 'posthog',
        enabled: data.posthog.enabled,
        connected: data.posthog.connected,
      },
    };
  },

  // Discord
  async setDiscord(
    token: string,
    webhookUrl: string,
    enabled: boolean
  ): Promise<DiscordSettings> {
    const data = await fetchAPI<{ enabled: boolean; connected: boolean }>(
      '/integrations/discord',
      {
        token,
        method: 'POST',
        body: { webhook_url: webhookUrl, enabled },
      }
    );

    return {
      type: 'discord',
      enabled: data.enabled,
      connected: data.connected,
    };
  },

  async deleteDiscord(token: string): Promise<void> {
    await fetchAPI('/integrations/discord', { token, method: 'DELETE' });
  },

  // Slack
  async setSlack(
    token: string,
    webhookUrl: string,
    channelId: string | undefined,
    enabled: boolean
  ): Promise<SlackSettings> {
    const data = await fetchAPI<{
      enabled: boolean;
      connected: boolean;
      channel_id?: string;
    }>('/integrations/slack', {
      token,
      method: 'POST',
      body: { webhook_url: webhookUrl, channel_id: channelId, enabled },
    });

    return {
      type: 'slack',
      enabled: data.enabled,
      connected: data.connected,
      channelId: data.channel_id,
    };
  },

  async deleteSlack(token: string): Promise<void> {
    await fetchAPI('/integrations/slack', { token, method: 'DELETE' });
  },

  // PostHog
  async setPostHog(token: string, enabled: boolean): Promise<PostHogSettings> {
    const data = await fetchAPI<{ enabled: boolean; connected: boolean }>(
      '/integrations/posthog',
      {
        token,
        method: 'POST',
        body: { enabled },
      }
    );

    return {
      type: 'posthog',
      enabled: data.enabled,
      connected: data.connected,
    };
  },

  async deletePostHog(token: string): Promise<void> {
    await fetchAPI('/integrations/posthog', { token, method: 'DELETE' });
  },

  // MCP HTTP Servers
  async listMcpServers(token: string): Promise<MCPServer[]> {
    const data = await fetchAPI<{
      servers: Array<{
        id: string;
        name: string;
        url: string;
        enabled: boolean;
        has_api_key: boolean;
        created_at: string;
        updated_at: string;
        last_sync: string | null;
        last_error: string | null;
        manifest?: {
          name: string;
          version: string;
          description?: string;
          tool_count: number;
        };
      }>;
    }>('/mcp/servers', { token });

    return data.servers.map((s) => ({
      id: s.id,
      name: s.name,
      url: s.url,
      enabled: s.enabled,
      hasApiKey: s.has_api_key,
      createdAt: s.created_at,
      updatedAt: s.updated_at,
      lastSync: s.last_sync,
      lastError: s.last_error,
      manifest: s.manifest
        ? {
            name: s.manifest.name,
            version: s.manifest.version,
            description: s.manifest.description,
            toolCount: s.manifest.tool_count,
          }
        : null,
    }));
  },

  async getMcpServer(token: string, id: string): Promise<MCPServer> {
    const data = await fetchAPI<{
      id: string;
      name: string;
      url: string;
      enabled: boolean;
      has_api_key: boolean;
      created_at: string;
      updated_at: string;
      last_sync: string | null;
      last_error: string | null;
      manifest?: {
        name: string;
        version: string;
        description?: string;
        tool_count: number;
      };
    }>(`/mcp/servers/${id}`, { token });

    return {
      id: data.id,
      name: data.name,
      url: data.url,
      enabled: data.enabled,
      hasApiKey: data.has_api_key,
      createdAt: data.created_at,
      updatedAt: data.updated_at,
      lastSync: data.last_sync,
      lastError: data.last_error,
      manifest: data.manifest
        ? {
            name: data.manifest.name,
            version: data.manifest.version,
            description: data.manifest.description,
            toolCount: data.manifest.tool_count,
          }
        : null,
    };
  },

  async addMcpServer(
    token: string,
    name: string,
    url: string,
    apiKey?: string
  ): Promise<MCPServer> {
    const data = await fetchAPI<{
      id: string;
      name: string;
      url: string;
      enabled: boolean;
      created_at: string;
      manifest?: {
        name: string;
        version: string;
        description?: string;
        tool_count: number;
      };
    }>('/mcp/servers', {
      token,
      method: 'POST',
      body: { name, url, api_key: apiKey },
    });

    return {
      id: data.id,
      name: data.name,
      url: data.url,
      enabled: data.enabled,
      hasApiKey: !!apiKey,
      createdAt: data.created_at,
      updatedAt: data.created_at,
      lastSync: data.created_at,
      lastError: null,
      manifest: data.manifest
        ? {
            name: data.manifest.name,
            version: data.manifest.version,
            description: data.manifest.description,
            toolCount: data.manifest.tool_count,
          }
        : null,
    };
  },

  async updateMcpServer(
    token: string,
    id: string,
    updates: { name?: string; url?: string; apiKey?: string }
  ): Promise<void> {
    await fetchAPI(`/mcp/servers/${id}`, {
      token,
      method: 'PUT',
      body: {
        name: updates.name,
        url: updates.url,
        api_key: updates.apiKey,
      },
    });
  },

  async removeMcpServer(token: string, id: string): Promise<void> {
    await fetchAPI(`/mcp/servers/${id}`, { token, method: 'DELETE' });
  },

  async testMcpServer(
    token: string,
    id: string
  ): Promise<{ success: boolean; error?: string }> {
    return fetchAPI(`/mcp/servers/${id}/test`, { token, method: 'POST' });
  },

  async refreshMcpServer(token: string, id: string): Promise<void> {
    await fetchAPI(`/mcp/servers/${id}/refresh`, { token, method: 'POST' });
  },

  async enableMcpServer(token: string, id: string): Promise<void> {
    await fetchAPI(`/mcp/servers/${id}/enable`, { token, method: 'POST' });
  },

  async disableMcpServer(token: string, id: string): Promise<void> {
    await fetchAPI(`/mcp/servers/${id}/disable`, { token, method: 'POST' });
  },

  async listMcpTools(token: string): Promise<MCPTool[]> {
    const data = await fetchAPI<{
      tools: Array<{
        server_id: string;
        server_name: string;
        name: string;
        description?: string;
        parameters?: Record<string, unknown>;
      }>;
    }>('/mcp/tools', { token });

    return data.tools.map((t) => ({
      serverId: t.server_id,
      serverName: t.server_name,
      name: t.name,
      description: t.description,
      parameters: t.parameters,
    }));
  },

  // MCP Stdio Servers
  async listStdioServers(token: string): Promise<StdioMCPServer[]> {
    const data = await fetchAPI<{
      servers: Array<{
        id: string;
        name: string;
        command: string;
        args: string[];
        env?: string[];
        enabled: boolean;
        running: boolean;
        tool_count: number;
        created_at: string;
        updated_at: string;
        last_error: string | null;
      }>;
    }>('/mcp/stdio/servers', { token });

    return data.servers.map((s) => ({
      id: s.id,
      name: s.name,
      command: s.command,
      args: s.args,
      env: s.env,
      enabled: s.enabled,
      running: s.running,
      toolCount: s.tool_count,
      createdAt: s.created_at,
      updatedAt: s.updated_at,
      lastError: s.last_error,
    }));
  },

  async getStdioServer(token: string, id: string): Promise<StdioMCPServer> {
    const data = await fetchAPI<{
      server: {
        id: string;
        name: string;
        command: string;
        args: string[];
        env?: string[];
        enabled: boolean;
        running: boolean;
        tools: Array<{
          name: string;
          description?: string;
          parameters?: Record<string, unknown>;
        }>;
        created_at: string;
        updated_at: string;
        last_error: string | null;
      };
    }>(`/mcp/stdio/servers/${id}`, { token });

    const s = data.server;
    return {
      id: s.id,
      name: s.name,
      command: s.command,
      args: s.args,
      env: s.env,
      enabled: s.enabled,
      running: s.running,
      toolCount: s.tools?.length || 0,
      createdAt: s.created_at,
      updatedAt: s.updated_at,
      lastError: s.last_error,
    };
  },

  async addStdioServer(
    token: string,
    name: string,
    command: string,
    args: string[],
    env?: string[]
  ): Promise<StdioMCPServer> {
    const data = await fetchAPI<{
      server: {
        id: string;
        name: string;
        command: string;
        args: string[];
        env?: string[];
        enabled: boolean;
        running: boolean;
        tool_count: number;
        last_error?: string;
      };
    }>('/mcp/stdio/servers', {
      token,
      method: 'POST',
      body: { name, command, args, env },
    });

    const s = data.server;
    return {
      id: s.id,
      name: s.name,
      command: s.command,
      args: s.args,
      env: s.env,
      enabled: s.enabled,
      running: s.running,
      toolCount: s.tool_count || 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      lastError: s.last_error || null,
    };
  },

  async updateStdioServer(
    token: string,
    id: string,
    updates: {
      name?: string;
      command?: string;
      args?: string[];
      env?: string[];
      enabled?: boolean;
    }
  ): Promise<void> {
    await fetchAPI(`/mcp/stdio/servers/${id}`, {
      token,
      method: 'PUT',
      body: updates,
    });
  },

  async removeStdioServer(token: string, id: string): Promise<void> {
    await fetchAPI(`/mcp/stdio/servers/${id}`, { token, method: 'DELETE' });
  },

  async startStdioServer(token: string, id: string): Promise<void> {
    await fetchAPI(`/mcp/stdio/servers/${id}/start`, { token, method: 'POST' });
  },

  async stopStdioServer(token: string, id: string): Promise<void> {
    await fetchAPI(`/mcp/stdio/servers/${id}/stop`, { token, method: 'POST' });
  },

  async restartStdioServer(token: string, id: string): Promise<void> {
    await fetchAPI(`/mcp/stdio/servers/${id}/restart`, { token, method: 'POST' });
  },
};
