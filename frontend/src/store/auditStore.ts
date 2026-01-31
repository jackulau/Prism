import { create } from 'zustand';
import { apiService } from '../services/api';
import type {
  AuditLogEntry,
  AuditLogFilter,
  ExportJob,
  ComplianceReport,
  RetentionPolicy,
} from '../types/audit';

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
  // User's own logs (simple fetch)
  myLogs: AuditLog[];
  myLogsLoading: boolean;
  myLogsError: string | null;
  myLogsOffset: number;
  myLogsHasMore: boolean;

  // Admin all logs (simple fetch)
  allLogs: AuditLog[];
  allLogsLoading: boolean;
  allLogsError: string | null;
  allLogsTotal: number;
  allLogsOffset: number;

  // Stats
  stats: AuditStats | null;
  statsLoading: boolean;
  statsError: string | null;

  // Filters (simple)
  filters: AuditFilters;

  // Audit Logs (compliance/export view)
  auditLogs: AuditLogEntry[];
  auditLogsTotal: number;
  auditLogsPage: number;
  auditLogsPageSize: number;
  auditLogsHasMore: boolean;
  auditFilter: AuditLogFilter;
  auditLogsLoading: boolean;
  auditLogsError: string | null;

  // Export Jobs
  exportJobs: ExportJob[];
  exportJobsLoading: boolean;
  exportJobsError: string | null;

  // Compliance Reports
  reports: ComplianceReport[];
  reportsLoading: boolean;
  reportsError: string | null;

  // Retention Policies
  retentionPolicies: RetentionPolicy[];
  retentionLoading: boolean;
  retentionError: string | null;

  // Actions - Simple fetch
  fetchMyLogs: (reset?: boolean) => Promise<void>;
  fetchAllLogs: (reset?: boolean) => Promise<void>;
  fetchStats: (period?: string) => Promise<void>;
  setFilters: (filters: AuditFilters) => void;
  clearFilters: () => void;
  clearErrors: () => void;

  // Actions - Audit Logs (compliance view)
  setAuditLogs: (logs: AuditLogEntry[], total: number, hasMore: boolean) => void;
  appendAuditLogs: (logs: AuditLogEntry[], hasMore: boolean) => void;
  setAuditFilter: (filter: AuditLogFilter) => void;
  setAuditLogsPage: (page: number) => void;
  setAuditLogsLoading: (loading: boolean) => void;
  setAuditLogsError: (error: string | null) => void;
  clearAuditLogs: () => void;

  // Actions - Export Jobs
  setExportJobs: (jobs: ExportJob[]) => void;
  addExportJob: (job: ExportJob) => void;
  updateExportJob: (id: string, updates: Partial<ExportJob>) => void;
  removeExportJob: (id: string) => void;
  setExportJobsLoading: (loading: boolean) => void;
  setExportJobsError: (error: string | null) => void;

  // Actions - Reports
  setReports: (reports: ComplianceReport[]) => void;
  addReport: (report: ComplianceReport) => void;
  setReportsLoading: (loading: boolean) => void;
  setReportsError: (error: string | null) => void;

  // Actions - Retention Policies
  setRetentionPolicies: (policies: RetentionPolicy[]) => void;
  updateRetentionPolicy: (dataType: string, updates: Partial<RetentionPolicy>) => void;
  setRetentionLoading: (loading: boolean) => void;
  setRetentionError: (error: string | null) => void;

  // Reset
  reset: () => void;
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

  // Audit Logs (compliance view)
  auditLogs: [],
  auditLogsTotal: 0,
  auditLogsPage: 1,
  auditLogsPageSize: 50,
  auditLogsHasMore: false,
  auditFilter: {},
  auditLogsLoading: false,
  auditLogsError: null,

  // Export Jobs
  exportJobs: [],
  exportJobsLoading: false,
  exportJobsError: null,

  // Reports
  reports: [],
  reportsLoading: false,
  reportsError: null,

  // Retention Policies
  retentionPolicies: [],
  retentionLoading: false,
  retentionError: null,

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

  // Audit Logs Actions (compliance view)
  setAuditLogs: (logs, total, hasMore) =>
    set({
      auditLogs: logs,
      auditLogsTotal: total,
      auditLogsHasMore: hasMore,
      auditLogsError: null,
    }),

  appendAuditLogs: (logs, hasMore) =>
    set((state) => ({
      auditLogs: [...state.auditLogs, ...logs],
      auditLogsHasMore: hasMore,
      auditLogsPage: state.auditLogsPage + 1,
    })),

  setAuditFilter: (filter) =>
    set({
      auditFilter: filter,
      auditLogsPage: 1,
      auditLogs: [],
    }),

  setAuditLogsPage: (page) => set({ auditLogsPage: page }),

  setAuditLogsLoading: (loading) => set({ auditLogsLoading: loading }),

  setAuditLogsError: (error) => set({ auditLogsError: error }),

  clearAuditLogs: () =>
    set({
      auditLogs: [],
      auditLogsTotal: 0,
      auditLogsPage: 1,
      auditLogsHasMore: false,
      auditFilter: {},
    }),

  // Export Jobs Actions
  setExportJobs: (jobs) => set({ exportJobs: jobs, exportJobsError: null }),

  addExportJob: (job) =>
    set((state) => ({
      exportJobs: [job, ...state.exportJobs],
    })),

  updateExportJob: (id, updates) =>
    set((state) => ({
      exportJobs: state.exportJobs.map((job) =>
        job.id === id ? { ...job, ...updates } : job
      ),
    })),

  removeExportJob: (id) =>
    set((state) => ({
      exportJobs: state.exportJobs.filter((job) => job.id !== id),
    })),

  setExportJobsLoading: (loading) => set({ exportJobsLoading: loading }),

  setExportJobsError: (error) => set({ exportJobsError: error }),

  // Reports Actions
  setReports: (reports) => set({ reports, reportsError: null }),

  addReport: (report) =>
    set((state) => ({
      reports: [report, ...state.reports],
    })),

  setReportsLoading: (loading) => set({ reportsLoading: loading }),

  setReportsError: (error) => set({ reportsError: error }),

  // Retention Policies Actions
  setRetentionPolicies: (policies) =>
    set({ retentionPolicies: policies, retentionError: null }),

  updateRetentionPolicy: (dataType, updates) =>
    set((state) => ({
      retentionPolicies: state.retentionPolicies.map((policy) =>
        policy.dataType === dataType ? { ...policy, ...updates } : policy
      ),
    })),

  setRetentionLoading: (loading) => set({ retentionLoading: loading }),

  setRetentionError: (error) => set({ retentionError: error }),

  // Reset
  reset: () => set({
    myLogs: [],
    myLogsLoading: false,
    myLogsError: null,
    myLogsOffset: 0,
    myLogsHasMore: true,
    allLogs: [],
    allLogsLoading: false,
    allLogsError: null,
    allLogsTotal: 0,
    allLogsOffset: 0,
    stats: null,
    statsLoading: false,
    statsError: null,
    filters: {},
    auditLogs: [],
    auditLogsTotal: 0,
    auditLogsPage: 1,
    auditLogsPageSize: 50,
    auditLogsHasMore: false,
    auditFilter: {},
    auditLogsLoading: false,
    auditLogsError: null,
    exportJobs: [],
    exportJobsLoading: false,
    exportJobsError: null,
    reports: [],
    reportsLoading: false,
    reportsError: null,
    retentionPolicies: [],
    retentionLoading: false,
    retentionError: null,
  }),
}));

// Selectors
export const selectPendingExports = (state: AuditState) =>
  state.exportJobs.filter((job) => job.status === 'pending' || job.status === 'processing');

export const selectCompletedExports = (state: AuditState) =>
  state.exportJobs.filter((job) => job.status === 'complete');

export const selectActiveRetentionPolicies = (state: AuditState) =>
  state.retentionPolicies.filter((policy) => !policy.legalHold);

export const selectLegalHoldPolicies = (state: AuditState) =>
  state.retentionPolicies.filter((policy) => policy.legalHold);
