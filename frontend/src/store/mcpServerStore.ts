import { create } from 'zustand';
import { apiService } from '../services/api';

// Types for MCP Server data
export interface MCPServerManifest {
  name: string;
  version: string;
  description: string;
  tool_count: number;
}

export interface MCPServer {
  id: string;
  name: string;
  url: string;
  enabled: boolean;
  has_api_key?: boolean;
  manifest?: MCPServerManifest;
  created_at?: string;
  updated_at?: string;
  last_sync?: string;
  last_error?: string;
}

export interface MCPServerStatus {
  connected: boolean;
  latency_ms?: number;
  error?: string;
  last_checked?: string;
}

export interface MCPTool {
  server_id: string;
  server_name: string;
  name: string;
  description: string;
  parameters?: Record<string, unknown>;
}

export interface MCPServerStats {
  total_calls: number;
  successful_calls: number;
  failed_calls: number;
  average_response_ms: number;
}

interface MCPServerState {
  // Server list
  servers: MCPServer[];
  serversLoading: boolean;
  serversError: string | null;

  // Individual server status (keyed by server ID)
  serverStatuses: Record<string, MCPServerStatus>;

  // Tools (keyed by server ID)
  serverTools: Record<string, MCPTool[]>;
  toolsLoading: Record<string, boolean>;

  // Stats (keyed by server ID)
  serverStats: Record<string, MCPServerStats>;
  statsLoading: Record<string, boolean>;

  // Status polling
  pollingInterval: ReturnType<typeof setInterval> | null;

  // Actions
  setToken: (token: string | null) => void;
  fetchServers: () => Promise<void>;
  addServer: (name: string, url: string, apiKey?: string) => Promise<{ success: boolean; error?: string }>;
  updateServer: (id: string, data: { name?: string; url?: string; apiKey?: string }) => Promise<boolean>;
  removeServer: (id: string) => Promise<boolean>;
  enableServer: (id: string) => Promise<boolean>;
  disableServer: (id: string) => Promise<boolean>;
  testServer: (id: string) => Promise<{ success: boolean; error?: string }>;
  refreshServer: (id: string) => Promise<boolean>;
  reconnectServer: (id: string) => Promise<boolean>;
  fetchServerStatus: (id: string) => Promise<void>;
  fetchServerTools: (id: string) => Promise<void>;
  fetchServerStats: (id: string, timeRange?: 'today' | 'week' | 'all') => Promise<void>;
  startStatusPolling: (intervalMs?: number) => void;
  stopStatusPolling: () => void;
  clearError: () => void;
}

export const useMCPServerStore = create<MCPServerState>((set, get) => ({
  // Initial state
  servers: [],
  serversLoading: false,
  serversError: null,
  serverStatuses: {},
  serverTools: {},
  toolsLoading: {},
  serverStats: {},
  statsLoading: {},
  pollingInterval: null,

  setToken: (token) => {
    apiService.setToken(token);
  },

  fetchServers: async () => {
    set({ serversLoading: true, serversError: null });
    const response = await apiService.getMCPServers();
    if (response.data) {
      set({
        servers: response.data.servers,
        serversLoading: false,
      });
    } else {
      set({
        serversError: response.error || 'Failed to fetch MCP servers',
        serversLoading: false,
      });
    }
  },

  addServer: async (name, url, apiKey) => {
    const response = await apiService.addMCPServer(name, url, apiKey);
    if (response.data) {
      // Refresh the server list
      await get().fetchServers();
      return { success: true };
    }
    return { success: false, error: response.error };
  },

  updateServer: async (id, data) => {
    const response = await apiService.updateMCPServer(id, data);
    if (response.data) {
      await get().fetchServers();
      return true;
    }
    return false;
  },

  removeServer: async (id) => {
    const response = await apiService.removeMCPServer(id);
    if (response.data) {
      set((state) => ({
        servers: state.servers.filter((s) => s.id !== id),
        serverStatuses: Object.fromEntries(
          Object.entries(state.serverStatuses).filter(([key]) => key !== id)
        ),
        serverTools: Object.fromEntries(
          Object.entries(state.serverTools).filter(([key]) => key !== id)
        ),
        serverStats: Object.fromEntries(
          Object.entries(state.serverStats).filter(([key]) => key !== id)
        ),
      }));
      return true;
    }
    return false;
  },

  enableServer: async (id) => {
    const response = await apiService.enableMCPServer(id);
    if (response.data) {
      set((state) => ({
        servers: state.servers.map((s) =>
          s.id === id ? { ...s, enabled: true } : s
        ),
      }));
      return true;
    }
    return false;
  },

  disableServer: async (id) => {
    const response = await apiService.disableMCPServer(id);
    if (response.data) {
      set((state) => ({
        servers: state.servers.map((s) =>
          s.id === id ? { ...s, enabled: false } : s
        ),
      }));
      return true;
    }
    return false;
  },

  testServer: async (id) => {
    const response = await apiService.testMCPServer(id);
    if (response.data?.success) {
      set((state) => ({
        serverStatuses: {
          ...state.serverStatuses,
          [id]: {
            connected: true,
            last_checked: new Date().toISOString(),
          },
        },
      }));
      return { success: true };
    }
    set((state) => ({
      serverStatuses: {
        ...state.serverStatuses,
        [id]: {
          connected: false,
          error: response.error || response.data?.error,
          last_checked: new Date().toISOString(),
        },
      },
    }));
    return { success: false, error: response.error || response.data?.error };
  },

  refreshServer: async (id) => {
    const response = await apiService.refreshMCPServer(id);
    if (response.data?.success) {
      await get().fetchServers();
      return true;
    }
    return false;
  },

  reconnectServer: async (id) => {
    const response = await apiService.reconnectMCPServer(id);
    if (response.data?.success) {
      set((state) => ({
        serverStatuses: {
          ...state.serverStatuses,
          [id]: {
            connected: true,
            last_checked: new Date().toISOString(),
          },
        },
      }));
      await get().fetchServers();
      return true;
    }
    return false;
  },

  fetchServerStatus: async (id) => {
    const startTime = Date.now();
    const response = await apiService.getMCPServerStatus(id);
    const latency = Date.now() - startTime;

    if (response.data) {
      set((state) => ({
        serverStatuses: {
          ...state.serverStatuses,
          [id]: {
            connected: response.data!.connected,
            latency_ms: latency,
            error: response.data!.error,
            last_checked: new Date().toISOString(),
          },
        },
      }));
    } else {
      set((state) => ({
        serverStatuses: {
          ...state.serverStatuses,
          [id]: {
            connected: false,
            error: response.error,
            last_checked: new Date().toISOString(),
          },
        },
      }));
    }
  },

  fetchServerTools: async (id) => {
    set((state) => ({
      toolsLoading: { ...state.toolsLoading, [id]: true },
    }));

    const response = await apiService.getMCPServerTools(id);
    if (response.data) {
      set((state) => ({
        serverTools: { ...state.serverTools, [id]: response.data!.tools },
        toolsLoading: { ...state.toolsLoading, [id]: false },
      }));
    } else {
      set((state) => ({
        toolsLoading: { ...state.toolsLoading, [id]: false },
      }));
    }
  },

  fetchServerStats: async (id, timeRange = 'all') => {
    set((state) => ({
      statsLoading: { ...state.statsLoading, [id]: true },
    }));

    const response = await apiService.getMCPServerStats(id, timeRange);
    if (response.data) {
      set((state) => ({
        serverStats: { ...state.serverStats, [id]: response.data! },
        statsLoading: { ...state.statsLoading, [id]: false },
      }));
    } else {
      set((state) => ({
        statsLoading: { ...state.statsLoading, [id]: false },
      }));
    }
  },

  startStatusPolling: (intervalMs = 30000) => {
    const { pollingInterval, servers } = get();

    // Clear existing interval if any
    if (pollingInterval) {
      clearInterval(pollingInterval);
    }

    // Poll status for all servers
    const pollStatuses = async () => {
      const { servers } = get();
      await Promise.all(
        servers.filter((s) => s.enabled).map((s) => get().fetchServerStatus(s.id))
      );
    };

    // Initial poll
    pollStatuses();

    // Set up interval
    const interval = setInterval(pollStatuses, intervalMs);
    set({ pollingInterval: interval });
  },

  stopStatusPolling: () => {
    const { pollingInterval } = get();
    if (pollingInterval) {
      clearInterval(pollingInterval);
      set({ pollingInterval: null });
    }
  },

  clearError: () => set({ serversError: null }),
}));
