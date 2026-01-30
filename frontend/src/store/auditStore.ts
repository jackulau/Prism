import { create } from 'zustand';
import { apiService } from '../services/api';

// Types matching backend API response
export interface AuditLog {
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
}

export interface AuditFilters {
  category?: string;
  event_type?: string;
  start_date?: string;
  end_date?: string;
  success?: boolean;
}

export interface AuditStats {
  since: string;
  category_counts: Record<string, number>;
  auth_counts: Record<string, number>;
  provider_counts: Record<string, number>;
}

interface AuditState {
  // User's own logs
  myLogs: AuditLog[];
  myLogsLoading: boolean;
  myLogsError: string | null;
  myLogsOffset: number;
  myLogsHasMore: boolean;

  // Admin all logs
  allLogs: AuditLog[];
  allLogsLoading: boolean;
  allLogsError: string | null;
  allLogsTotal: number;
  allLogsOffset: number;

  // Stats
  stats: AuditStats | null;
  statsLoading: boolean;
  statsError: string | null;

  // Filters
  filters: AuditFilters;

  // Actions
  fetchMyLogs: (reset?: boolean) => Promise<void>;
  fetchAllLogs: (reset?: boolean) => Promise<void>;
  fetchStats: (period?: string) => Promise<void>;
  setFilters: (filters: AuditFilters) => void;
  clearFilters: () => void;
  clearErrors: () => void;
}

const PAGE_SIZE = 20;

export const useAuditStore = create<AuditState>((set, get) => ({
  // User's logs initial state
  myLogs: [],
  myLogsLoading: false,
  myLogsError: null,
  myLogsOffset: 0,
  myLogsHasMore: true,

  // Admin logs initial state
  allLogs: [],
  allLogsLoading: false,
  allLogsError: null,
  allLogsTotal: 0,
  allLogsOffset: 0,

  // Stats initial state
  stats: null,
  statsLoading: false,
  statsError: null,

  // Filters
  filters: {},

  // Fetch current user's audit logs
  fetchMyLogs: async (reset = false) => {
    const { myLogsOffset, myLogs, myLogsLoading } = get();

    if (myLogsLoading) return;

    const offset = reset ? 0 : myLogsOffset;

    set({ myLogsLoading: true, myLogsError: null });

    try {
      const response = await apiService.getMyAuditLogs({
        limit: PAGE_SIZE,
        offset,
      });

      if (response.error) {
        set({ myLogsError: response.error, myLogsLoading: false });
        return;
      }

      const logs = response.data?.logs || [];

      set({
        myLogs: reset ? logs : [...myLogs, ...logs],
        myLogsOffset: offset + logs.length,
        myLogsHasMore: logs.length === PAGE_SIZE,
        myLogsLoading: false,
      });
    } catch (err) {
      set({
        myLogsError: err instanceof Error ? err.message : 'Failed to fetch logs',
        myLogsLoading: false,
      });
    }
  },

  // Fetch all audit logs (admin)
  fetchAllLogs: async (reset = false) => {
    const { allLogsOffset, allLogs, allLogsLoading, filters } = get();

    if (allLogsLoading) return;

    const offset = reset ? 0 : allLogsOffset;

    set({ allLogsLoading: true, allLogsError: null });

    try {
      const response = await apiService.getAllAuditLogs({
        ...filters,
        limit: PAGE_SIZE,
        offset,
      });

      if (response.error) {
        set({ allLogsError: response.error, allLogsLoading: false });
        return;
      }

      const data = response.data;
      const logs = data?.logs || [];

      set({
        allLogs: reset ? logs : [...allLogs, ...logs],
        allLogsTotal: data?.total || 0,
        allLogsOffset: offset + logs.length,
        allLogsLoading: false,
      });
    } catch (err) {
      set({
        allLogsError: err instanceof Error ? err.message : 'Failed to fetch logs',
        allLogsLoading: false,
      });
    }
  },

  // Fetch audit stats
  fetchStats: async (period = '24h') => {
    set({ statsLoading: true, statsError: null });

    try {
      const response = await apiService.getAuditStats(period);

      if (response.error) {
        set({ statsError: response.error, statsLoading: false });
        return;
      }

      set({
        stats: response.data || null,
        statsLoading: false,
      });
    } catch (err) {
      set({
        statsError: err instanceof Error ? err.message : 'Failed to fetch stats',
        statsLoading: false,
      });
    }
  },

  // Set filters
  setFilters: (filters) => {
    set({ filters });
  },

  // Clear filters
  clearFilters: () => {
    set({ filters: {} });
  },

  // Clear errors
  clearErrors: () => {
    set({
      myLogsError: null,
      allLogsError: null,
      statsError: null,
    });
  },
}));
