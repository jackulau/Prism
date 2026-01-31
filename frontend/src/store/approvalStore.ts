import { create } from 'zustand';
import { approvalService } from '../services/approval';
import type {
  ApprovalRequest,
  ApprovalWorkflow,
  ApprovalStats,
  ApprovalEvent,
} from '../types/approval';

interface ApprovalState {
  // Pending approvals
  pendingApprovals: ApprovalRequest[];
  isPendingLoading: boolean;
  pendingError: string | null;

  // Approval history
  approvalHistory: ApprovalRequest[];
  isHistoryLoading: boolean;
  historyError: string | null;
  historyTotal: number;
  historyPage: number;

  // Workflows
  workflows: ApprovalWorkflow[];
  isWorkflowsLoading: boolean;
  workflowsError: string | null;

  // Stats
  stats: ApprovalStats | null;
  isStatsLoading: boolean;

  // Selected approval for detail view
  selectedApproval: ApprovalRequest | null;

  // Actions
  loadPendingApprovals: () => Promise<void>;
  loadApprovalHistory: (page?: number) => Promise<void>;
  loadWorkflows: () => Promise<void>;
  loadStats: () => Promise<void>;
  submitDecision: (requestId: string, decision: 'approved' | 'rejected', comment?: string) => Promise<boolean>;
  selectApproval: (approval: ApprovalRequest | null) => void;
  createWorkflow: (workflow: Omit<ApprovalWorkflow, 'id' | 'createdAt' | 'updatedAt' | 'createdBy'>) => Promise<ApprovalWorkflow | null>;
  updateWorkflow: (id: string, updates: Partial<ApprovalWorkflow>) => Promise<boolean>;
  deleteWorkflow: (id: string) => Promise<boolean>;
  toggleWorkflowActive: (id: string, isActive: boolean) => Promise<boolean>;
  handleApprovalEvent: (event: ApprovalEvent) => void;
  clearErrors: () => void;
}

export const useApprovalStore = create<ApprovalState>((set, get) => ({
  // Initial state
  pendingApprovals: [],
  isPendingLoading: false,
  pendingError: null,

  approvalHistory: [],
  isHistoryLoading: false,
  historyError: null,
  historyTotal: 0,
  historyPage: 1,

  workflows: [],
  isWorkflowsLoading: false,
  workflowsError: null,

  stats: null,
  isStatsLoading: false,

  selectedApproval: null,

  // Actions
  loadPendingApprovals: async () => {
    set({ isPendingLoading: true, pendingError: null });
    try {
      const response = await approvalService.getPendingApprovals();
      if (response.error) {
        set({ pendingError: response.error, isPendingLoading: false });
        return;
      }
      set({
        pendingApprovals: response.data?.approvals || [],
        isPendingLoading: false,
      });
    } catch (err) {
      set({
        pendingError: err instanceof Error ? err.message : 'Failed to load approvals',
        isPendingLoading: false,
      });
    }
  },

  loadApprovalHistory: async (page = 1) => {
    set({ isHistoryLoading: true, historyError: null });
    try {
      const response = await approvalService.getApprovalHistory({ page, pageSize: 20 });
      if (response.error) {
        set({ historyError: response.error, isHistoryLoading: false });
        return;
      }
      set({
        approvalHistory: response.data?.approvals || [],
        historyTotal: response.data?.total || 0,
        historyPage: page,
        isHistoryLoading: false,
      });
    } catch (err) {
      set({
        historyError: err instanceof Error ? err.message : 'Failed to load history',
        isHistoryLoading: false,
      });
    }
  },

  loadWorkflows: async () => {
    set({ isWorkflowsLoading: true, workflowsError: null });
    try {
      const response = await approvalService.listWorkflows();
      if (response.error) {
        set({ workflowsError: response.error, isWorkflowsLoading: false });
        return;
      }
      set({
        workflows: response.data?.workflows || [],
        isWorkflowsLoading: false,
      });
    } catch (err) {
      set({
        workflowsError: err instanceof Error ? err.message : 'Failed to load workflows',
        isWorkflowsLoading: false,
      });
    }
  },

  loadStats: async () => {
    set({ isStatsLoading: true });
    try {
      const response = await approvalService.getApprovalStats();
      if (response.data) {
        set({ stats: response.data, isStatsLoading: false });
      } else {
        set({ isStatsLoading: false });
      }
    } catch {
      set({ isStatsLoading: false });
    }
  },

  submitDecision: async (requestId, decision, comment) => {
    try {
      const response = await approvalService.submitDecision({
        requestId,
        decision,
        comment,
      });
      if (response.error) {
        return false;
      }

      // Update local state
      const state = get();
      const updatedApproval = response.data;
      if (updatedApproval) {
        // Remove from pending
        set({
          pendingApprovals: state.pendingApprovals.filter(a => a.id !== requestId),
          // Add to history
          approvalHistory: [updatedApproval, ...state.approvalHistory],
          selectedApproval: null,
        });

        // Reload stats
        get().loadStats();
      }

      return true;
    } catch {
      return false;
    }
  },

  selectApproval: (approval) => {
    set({ selectedApproval: approval });
  },

  createWorkflow: async (workflow) => {
    try {
      const response = await approvalService.createWorkflow({
        name: workflow.name,
        description: workflow.description,
        triggerType: workflow.triggerType,
        triggerConditions: workflow.triggerConditions,
        steps: workflow.steps.map(({ id: _, ...step }) => step),
      });
      if (response.error || !response.data) {
        return null;
      }

      set(state => ({
        workflows: [...state.workflows, response.data!],
      }));

      return response.data;
    } catch {
      return null;
    }
  },

  updateWorkflow: async (id, updates) => {
    try {
      const response = await approvalService.updateWorkflow(id, {
        name: updates.name,
        description: updates.description,
        triggerConditions: updates.triggerConditions,
        steps: updates.steps?.map(({ id: _, ...step }) => step),
        isActive: updates.isActive,
      });
      if (response.error || !response.data) {
        return false;
      }

      set(state => ({
        workflows: state.workflows.map(w =>
          w.id === id ? response.data! : w
        ),
      }));

      return true;
    } catch {
      return false;
    }
  },

  deleteWorkflow: async (id) => {
    try {
      const response = await approvalService.deleteWorkflow(id);
      if (response.error) {
        return false;
      }

      set(state => ({
        workflows: state.workflows.filter(w => w.id !== id),
      }));

      return true;
    } catch {
      return false;
    }
  },

  toggleWorkflowActive: async (id, isActive) => {
    try {
      const response = await approvalService.toggleWorkflow(id, isActive);
      if (response.error || !response.data) {
        return false;
      }

      set(state => ({
        workflows: state.workflows.map(w =>
          w.id === id ? { ...w, isActive } : w
        ),
      }));

      return true;
    } catch {
      return false;
    }
  },

  handleApprovalEvent: (event) => {
    const state = get();

    switch (event.type) {
      case 'approval.created':
        // Add new approval to pending list if it's pending
        if (event.approval.status === 'pending') {
          set({
            pendingApprovals: [event.approval, ...state.pendingApprovals],
          });
        }
        break;

      case 'approval.updated':
      case 'approval.decided':
        if (event.approval.status === 'pending') {
          // Update in pending list
          set({
            pendingApprovals: state.pendingApprovals.map(a =>
              a.id === event.approval.id ? event.approval : a
            ),
          });
        } else {
          // Remove from pending, add to history
          set({
            pendingApprovals: state.pendingApprovals.filter(a => a.id !== event.approval.id),
            approvalHistory: [event.approval, ...state.approvalHistory.filter(a => a.id !== event.approval.id)],
          });
        }

        // Update selected approval if it's the same one
        if (state.selectedApproval?.id === event.approval.id) {
          set({ selectedApproval: event.approval });
        }
        break;

      case 'approval.escalated':
        // Update the approval with escalation info
        set({
          pendingApprovals: state.pendingApprovals.map(a =>
            a.id === event.approval.id ? event.approval : a
          ),
        });

        if (state.selectedApproval?.id === event.approval.id) {
          set({ selectedApproval: event.approval });
        }
        break;

      case 'approval.expired':
        // Move from pending to history
        set({
          pendingApprovals: state.pendingApprovals.filter(a => a.id !== event.approval.id),
          approvalHistory: [event.approval, ...state.approvalHistory],
        });
        break;
    }

    // Reload stats after any event
    get().loadStats();
  },

  clearErrors: () => {
    set({
      pendingError: null,
      historyError: null,
      workflowsError: null,
    });
  },
}));

// Selector hooks for convenience
export const usePendingApprovals = () => useApprovalStore(state => ({
  approvals: state.pendingApprovals,
  isLoading: state.isPendingLoading,
  error: state.pendingError,
  refresh: state.loadPendingApprovals,
}));

export const useApprovalHistory = () => useApprovalStore(state => ({
  approvals: state.approvalHistory,
  isLoading: state.isHistoryLoading,
  error: state.historyError,
  total: state.historyTotal,
  page: state.historyPage,
  loadPage: state.loadApprovalHistory,
}));

export const useApprovalWorkflows = () => useApprovalStore(state => ({
  workflows: state.workflows,
  isLoading: state.isWorkflowsLoading,
  error: state.workflowsError,
  refresh: state.loadWorkflows,
  create: state.createWorkflow,
  update: state.updateWorkflow,
  delete: state.deleteWorkflow,
  toggleActive: state.toggleWorkflowActive,
}));

export const useApprovalStats = () => useApprovalStore(state => ({
  stats: state.stats,
  isLoading: state.isStatsLoading,
  refresh: state.loadStats,
}));

export const usePendingCount = () => useApprovalStore(state => state.pendingApprovals.length);
