import { create } from 'zustand';
import { apiService } from '../services/api';

export interface Session {
  id: string;
  user_id: string;
  ip_address: string;
  user_agent: string;
  device_name: string;
  created_at: string;
  last_activity_at: string;
  is_current: boolean;
}

interface SessionState {
  sessions: Session[];
  isLoading: boolean;
  error: string | null;
  idleWarningVisible: boolean;
  idleCountdown: number | null;

  // Actions
  fetchSessions: () => Promise<void>;
  terminateSession: (id: string) => Promise<void>;
  terminateOthers: () => Promise<void>;
  terminateAll: () => Promise<void>;
  showIdleWarning: (countdown: number) => void;
  hideIdleWarning: () => void;
  updateIdleCountdown: (countdown: number | null) => void;
  extendSession: () => Promise<void>;
  clearError: () => void;
}

export const useSessionStore = create<SessionState>((set, get) => ({
  sessions: [],
  isLoading: false,
  error: null,
  idleWarningVisible: false,
  idleCountdown: null,

  fetchSessions: async () => {
    set({ isLoading: true, error: null });
    const response = await apiService.listSessions();
    if (response.data) {
      set({ sessions: response.data.sessions, isLoading: false });
    } else {
      set({ error: response.error || 'Failed to fetch sessions', isLoading: false });
    }
  },

  terminateSession: async (id: string) => {
    const response = await apiService.terminateSession(id);
    if (response.error) {
      set({ error: response.error });
      return;
    }
    // Remove terminated session from list
    set((state) => ({
      sessions: state.sessions.filter((s) => s.id !== id),
    }));
  },

  terminateOthers: async () => {
    set({ isLoading: true, error: null });
    const response = await apiService.terminateOtherSessions();
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    // Keep only the current session
    set((state) => ({
      sessions: state.sessions.filter((s) => s.is_current),
      isLoading: false,
    }));
  },

  terminateAll: async () => {
    set({ isLoading: true, error: null });
    const response = await apiService.terminateAllSessions();
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({ sessions: [], isLoading: false });
  },

  showIdleWarning: (countdown: number) => {
    set({ idleWarningVisible: true, idleCountdown: countdown });
  },

  hideIdleWarning: () => {
    set({ idleWarningVisible: false, idleCountdown: null });
  },

  updateIdleCountdown: (countdown: number | null) => {
    set({ idleCountdown: countdown });
  },

  extendSession: async () => {
    // Calling any authenticated endpoint extends the session
    // We use getMe as a lightweight endpoint for this purpose
    await apiService.getMe();
    get().hideIdleWarning();
  },

  clearError: () => set({ error: null }),
}));
