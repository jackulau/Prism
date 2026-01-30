import { create } from 'zustand';
import type {
  AuditLogEntry,
  AuditLogFilter,
  ExportJob,
  ComplianceReport,
  RetentionPolicy,
} from '../types/audit';

interface AuditState {
  // Audit Logs
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

  // Actions - Audit Logs
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

const initialState = {
  // Audit Logs
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
};

export const useAuditStore = create<AuditState>((set) => ({
  ...initialState,

  // Audit Logs Actions
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
  reset: () => set(initialState),
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
