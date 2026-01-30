import { create } from 'zustand';
import { buildHistoryService, Build, BuildLog, BuildStatus } from '../services/buildHistory';
import { toast } from './toastStore';

interface BuildHistoryFilter {
  status?: BuildStatus;
}

interface BuildHistoryState {
  builds: Build[];
  selectedBuild: Build | null;
  logs: BuildLog[];
  isLoading: boolean;
  isLoadingLogs: boolean;
  error: string | null;
  filter: BuildHistoryFilter;
  total: number;
  offset: number;
  limit: number;

  // Actions
  fetchBuilds: () => Promise<void>;
  fetchMoreBuilds: () => Promise<void>;
  fetchBuildLogs: (buildId: string) => Promise<void>;
  setSelectedBuild: (build: Build | null) => void;
  setFilter: (filter: BuildHistoryFilter) => void;
  deleteBuild: (id: string) => Promise<void>;
  cancelBuild: (id: string) => Promise<void>;
  refreshBuild: (id: string) => Promise<void>;
  clearError: () => void;
}

export const useBuildHistoryStore = create<BuildHistoryState>((set, get) => ({
  builds: [],
  selectedBuild: null,
  logs: [],
  isLoading: false,
  isLoadingLogs: false,
  error: null,
  filter: {},
  total: 0,
  offset: 0,
  limit: 20,

  fetchBuilds: async () => {
    const { filter, limit } = get();
    set({ isLoading: true, error: null, offset: 0 });

    const response = await buildHistoryService.list({
      limit,
      offset: 0,
      status: filter.status,
    });

    if (response.error) {
      set({ error: response.error, isLoading: false });
      toast.error(`Failed to load builds: ${response.error}`);
      return;
    }

    set({
      builds: response.data?.builds || [],
      total: response.data?.total || 0,
      isLoading: false,
    });
  },

  fetchMoreBuilds: async () => {
    const { filter, limit, offset, builds, total, isLoading } = get();

    if (isLoading || builds.length >= total) return;

    const newOffset = offset + limit;
    set({ isLoading: true, error: null });

    const response = await buildHistoryService.list({
      limit,
      offset: newOffset,
      status: filter.status,
    });

    if (response.error) {
      set({ error: response.error, isLoading: false });
      toast.error(`Failed to load more builds: ${response.error}`);
      return;
    }

    set({
      builds: [...builds, ...(response.data?.builds || [])],
      offset: newOffset,
      isLoading: false,
    });
  },

  fetchBuildLogs: async (buildId: string) => {
    set({ isLoadingLogs: true });

    const response = await buildHistoryService.getLogs(buildId);

    if (response.error) {
      set({ isLoadingLogs: false });
      toast.error(`Failed to load logs: ${response.error}`);
      return;
    }

    set({
      logs: response.data?.logs || [],
      isLoadingLogs: false,
    });
  },

  setSelectedBuild: (build) => {
    set({ selectedBuild: build, logs: [] });
    if (build) {
      get().fetchBuildLogs(build.id);
    }
  },

  setFilter: (filter) => {
    set({ filter });
    get().fetchBuilds();
  },

  deleteBuild: async (id: string) => {
    const response = await buildHistoryService.delete(id);

    if (response.error) {
      toast.error(`Failed to delete build: ${response.error}`);
      return;
    }

    const { builds, selectedBuild } = get();
    set({
      builds: builds.filter((b) => b.id !== id),
      selectedBuild: selectedBuild?.id === id ? null : selectedBuild,
      logs: selectedBuild?.id === id ? [] : get().logs,
    });
    toast.success('Build deleted');
  },

  cancelBuild: async (id: string) => {
    const response = await buildHistoryService.cancel(id);

    if (response.error) {
      toast.error(`Failed to cancel build: ${response.error}`);
      return;
    }

    if (response.data) {
      const { builds, selectedBuild } = get();
      set({
        builds: builds.map((b) => (b.id === id ? response.data! : b)),
        selectedBuild: selectedBuild?.id === id ? response.data : selectedBuild,
      });
      toast.success('Build cancelled');
    }
  },

  refreshBuild: async (id: string) => {
    const response = await buildHistoryService.get(id);

    if (response.error) {
      return;
    }

    if (response.data) {
      const { builds, selectedBuild } = get();
      set({
        builds: builds.map((b) => (b.id === id ? response.data! : b)),
        selectedBuild: selectedBuild?.id === id ? response.data : selectedBuild,
      });

      // Also refresh logs if this is the selected build
      if (selectedBuild?.id === id) {
        get().fetchBuildLogs(id);
      }
    }
  },

  clearError: () => set({ error: null }),
}));
