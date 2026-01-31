import { create } from 'zustand';
import type {
  OrgWorkspace,
  CreateOrgWorkspaceInput,
  UpdateOrgWorkspaceInput,
} from '../types/organization';
import { orgWorkspaceService } from '../services/orgWorkspace';

interface OrgWorkspaceState {
  workspaces: OrgWorkspace[];
  isLoading: boolean;
  error: string | null;
  selectedWorkspace: OrgWorkspace | null;
  total: number;
  hasMore: boolean;

  fetchWorkspaces: (orgId: string, limit?: number, offset?: number) => Promise<void>;
  createWorkspace: (orgId: string, data: CreateOrgWorkspaceInput) => Promise<OrgWorkspace>;
  updateWorkspace: (
    orgId: string,
    id: string,
    data: UpdateOrgWorkspaceInput
  ) => Promise<OrgWorkspace>;
  deleteWorkspace: (orgId: string, id: string) => Promise<void>;
  setSelectedWorkspace: (workspace: OrgWorkspace | null) => void;
  reset: () => void;
}

const initialState = {
  workspaces: [],
  isLoading: false,
  error: null,
  selectedWorkspace: null,
  total: 0,
  hasMore: false,
};

export const useOrgWorkspaceStore = create<OrgWorkspaceState>((set) => ({
  ...initialState,

  fetchWorkspaces: async (orgId: string, limit = 20, offset = 0) => {
    set({ isLoading: true, error: null });
    try {
      const response = await orgWorkspaceService.list(orgId, { limit, offset });
      set({
        workspaces: response.workspaces,
        total: response.total,
        hasMore: response.hasMore,
        isLoading: false,
      });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to fetch workspaces',
        isLoading: false,
      });
    }
  },

  createWorkspace: async (orgId: string, data: CreateOrgWorkspaceInput) => {
    set({ isLoading: true, error: null });
    try {
      const workspace = await orgWorkspaceService.create(orgId, data);
      set((state) => ({
        workspaces: [workspace, ...state.workspaces],
        total: state.total + 1,
        isLoading: false,
      }));
      return workspace;
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to create workspace',
        isLoading: false,
      });
      throw err;
    }
  },

  updateWorkspace: async (
    orgId: string,
    id: string,
    data: UpdateOrgWorkspaceInput
  ) => {
    set({ isLoading: true, error: null });
    try {
      const workspace = await orgWorkspaceService.update(orgId, id, data);
      set((state) => ({
        workspaces: state.workspaces.map((w) => (w.id === id ? workspace : w)),
        selectedWorkspace:
          state.selectedWorkspace?.id === id ? workspace : state.selectedWorkspace,
        isLoading: false,
      }));
      return workspace;
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update workspace',
        isLoading: false,
      });
      throw err;
    }
  },

  deleteWorkspace: async (orgId: string, id: string) => {
    set({ isLoading: true, error: null });
    try {
      await orgWorkspaceService.delete(orgId, id);
      set((state) => ({
        workspaces: state.workspaces.filter((w) => w.id !== id),
        total: state.total - 1,
        selectedWorkspace:
          state.selectedWorkspace?.id === id ? null : state.selectedWorkspace,
        isLoading: false,
      }));
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to delete workspace',
        isLoading: false,
      });
      throw err;
    }
  },

  setSelectedWorkspace: (workspace: OrgWorkspace | null) => {
    set({ selectedWorkspace: workspace });
  },

  reset: () => set(initialState),
}));
