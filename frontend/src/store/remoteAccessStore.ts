import { create } from 'zustand';
import {
  remoteAccessService,
  type RemoteSession,
  type RemoteConfig,
} from '../services/remoteAccess';

interface RemoteAccessState {
  // Status
  enabled: boolean;
  port: number;
  password: string | null;
  connectionUrl: string | null;
  tlsEnabled: boolean;
  allowedIPs: string[];

  // Connection info
  publicIP: string | null;
  localIPs: string[];

  // Sessions
  sessions: RemoteSession[];

  // Loading states
  loading: boolean;
  sessionsLoading: boolean;
  error: string | null;

  // Actions
  setToken: (token: string | null) => void;
  fetchStatus: () => Promise<void>;
  fetchConnectionInfo: () => Promise<void>;
  fetchSessions: () => Promise<void>;
  enable: (config?: RemoteConfig) => Promise<boolean>;
  disable: () => Promise<boolean>;
  regeneratePassword: () => Promise<string | null>;
  kickSession: (id: string) => Promise<boolean>;
  updateConfig: (config: RemoteConfig) => Promise<boolean>;
  clearError: () => void;
}

export const useRemoteAccessStore = create<RemoteAccessState>((set, get) => ({
  // Initial state
  enabled: false,
  port: 8443,
  password: null,
  connectionUrl: null,
  tlsEnabled: false,
  allowedIPs: [],
  publicIP: null,
  localIPs: [],
  sessions: [],
  loading: false,
  sessionsLoading: false,
  error: null,

  setToken: (token) => {
    remoteAccessService.setToken(token);
  },

  fetchStatus: async () => {
    set({ loading: true, error: null });
    const response = await remoteAccessService.getStatus();
    if (response.data) {
      set({
        enabled: response.data.enabled,
        port: response.data.port,
        password: response.data.password,
        connectionUrl: response.data.connectionUrl,
        tlsEnabled: response.data.tlsEnabled,
        allowedIPs: response.data.allowedIPs,
        loading: false,
      });
    } else {
      set({ error: response.error || 'Failed to fetch status', loading: false });
    }
  },

  fetchConnectionInfo: async () => {
    const response = await remoteAccessService.getConnectionInfo();
    if (response.data) {
      set({
        publicIP: response.data.publicIP,
        localIPs: response.data.localIPs,
        port: response.data.port,
        connectionUrl: response.data.connectionUrl,
        tlsEnabled: response.data.tlsEnabled,
      });
    }
  },

  fetchSessions: async () => {
    set({ sessionsLoading: true });
    const response = await remoteAccessService.getSessions();
    if (response.data) {
      set({ sessions: response.data.sessions, sessionsLoading: false });
    } else {
      set({ sessionsLoading: false });
    }
  },

  enable: async (config) => {
    set({ loading: true, error: null });
    const response = await remoteAccessService.enable(config || {});
    if (response.data) {
      set({
        enabled: response.data.enabled,
        port: response.data.port,
        password: response.data.password,
        connectionUrl: response.data.connectionUrl,
        tlsEnabled: response.data.tlsEnabled,
        allowedIPs: response.data.allowedIPs,
        loading: false,
      });
      // Fetch connection info after enabling
      get().fetchConnectionInfo();
      return true;
    } else {
      set({ error: response.error || 'Failed to enable remote access', loading: false });
      return false;
    }
  },

  disable: async () => {
    set({ loading: true, error: null });
    const response = await remoteAccessService.disable();
    if (response.data?.success) {
      set({
        enabled: false,
        password: null,
        connectionUrl: null,
        sessions: [],
        loading: false,
      });
      return true;
    } else {
      set({ error: response.error || 'Failed to disable remote access', loading: false });
      return false;
    }
  },

  regeneratePassword: async () => {
    set({ loading: true, error: null });
    const response = await remoteAccessService.regeneratePassword();
    if (response.data) {
      set({ password: response.data.password, loading: false });
      return response.data.password;
    } else {
      set({ error: response.error || 'Failed to regenerate password', loading: false });
      return null;
    }
  },

  kickSession: async (id) => {
    const response = await remoteAccessService.kickSession(id);
    if (response.data?.success) {
      set((state) => ({
        sessions: state.sessions.filter((s) => s.id !== id),
      }));
      return true;
    }
    return false;
  },

  updateConfig: async (config) => {
    set({ loading: true, error: null });
    const response = await remoteAccessService.updateConfig(config);
    if (response.data) {
      set({
        port: response.data.port,
        allowedIPs: response.data.allowedIPs,
        loading: false,
      });
      return true;
    } else {
      set({ error: response.error || 'Failed to update config', loading: false });
      return false;
    }
  },

  clearError: () => set({ error: null }),
}));
