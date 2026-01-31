import { create } from 'zustand';
import type { FileHistoryEntry, FileHistoryStats, HistoryFilters } from '../types';

interface FileHistoryState {
  // Data state
  entries: FileHistoryEntry[];
  filteredEntries: FileHistoryEntry[];
  selectedEntry: FileHistoryEntry | null;
  historyContent: string | null;
  stats: FileHistoryStats | null;

  // Filter state
  filters: HistoryFilters;

  // UI state
  isLoading: boolean;
  isLoadingContent: boolean;
  isPanelOpen: boolean;
  error: string | null;
  currentFilePath: string | null;

  // Actions - Data
  setEntries: (entries: FileHistoryEntry[]) => void;
  setStats: (stats: FileHistoryStats | null) => void;
  setSelectedEntry: (entry: FileHistoryEntry | null) => void;
  setHistoryContent: (content: string | null) => void;
  clearHistory: () => void;

  // Actions - Filters
  setFilters: (filters: HistoryFilters) => void;
  clearFilters: () => void;
  applyFilters: () => void;

  // Actions - UI
  setIsLoading: (loading: boolean) => void;
  setIsLoadingContent: (loading: boolean) => void;
  setError: (error: string | null) => void;
  openPanel: (filePath?: string) => void;
  closePanel: () => void;
  togglePanel: () => void;

  // Actions - API integration
  loadHistory: (filePath?: string) => void;
  loadEntryContent: (entryId: string) => void;
}

const STORAGE_KEY = 'file-history-panel-open';

const getStoredPanelState = (): boolean => {
  if (typeof window === 'undefined') return false;
  const stored = localStorage.getItem(STORAGE_KEY);
  return stored === 'true';
};

const setStoredPanelState = (open: boolean) => {
  if (typeof window !== 'undefined') {
    localStorage.setItem(STORAGE_KEY, String(open));
  }
};

function applyFiltersToEntries(entries: FileHistoryEntry[], filters: HistoryFilters): FileHistoryEntry[] {
  return entries.filter((entry) => {
    // File path filter
    if (filters.filePath && entry.file_path !== filters.filePath) {
      return false;
    }

    // Operation type filter
    if (filters.operations?.length && !filters.operations.includes(entry.operation as 'create' | 'update' | 'delete')) {
      return false;
    }

    // Search query filter
    if (filters.searchQuery) {
      const query = filters.searchQuery.toLowerCase();
      const matchesPath = entry.file_path.toLowerCase().includes(query);
      const matchesDescription = entry.description?.toLowerCase().includes(query);
      const matchesAgent = entry.agent_name?.toLowerCase().includes(query);
      if (!matchesPath && !matchesDescription && !matchesAgent) {
        return false;
      }
    }

    // Date range filter
    if (filters.dateRange) {
      const entryDate = new Date(entry.created_at);
      if (entryDate < filters.dateRange.start || entryDate > filters.dateRange.end) {
        return false;
      }
    }

    return true;
  });
}

export const useFileHistoryStore = create<FileHistoryState>((set, get) => ({
  // Initial state
  entries: [],
  filteredEntries: [],
  selectedEntry: null,
  historyContent: null,
  stats: null,
  filters: {},
  isLoading: false,
  isLoadingContent: false,
  isPanelOpen: getStoredPanelState(),
  error: null,
  currentFilePath: null,

  // Data actions
  setEntries: (entries) => {
    const { filters } = get();
    set({
      entries,
      filteredEntries: applyFiltersToEntries(entries, filters),
      error: null,
    });
  },

  setStats: (stats) => set({ stats }),

  setSelectedEntry: (entry) => set({ selectedEntry: entry }),

  setHistoryContent: (content) => set({ historyContent: content }),

  clearHistory: () => set({
    entries: [],
    filteredEntries: [],
    selectedEntry: null,
    historyContent: null,
    stats: null,
    error: null,
  }),

  // Filter actions
  setFilters: (filters) => {
    const { entries } = get();
    set({
      filters,
      filteredEntries: applyFiltersToEntries(entries, filters),
    });
  },

  clearFilters: () => {
    const { entries } = get();
    set({
      filters: {},
      filteredEntries: entries,
    });
  },

  applyFilters: () => {
    const { entries, filters } = get();
    set({
      filteredEntries: applyFiltersToEntries(entries, filters),
    });
  },

  // UI actions
  setIsLoading: (loading) => set({ isLoading: loading }),

  setIsLoadingContent: (loading) => set({ isLoadingContent: loading }),

  setError: (error) => set({ error, isLoading: false, isLoadingContent: false }),

  openPanel: (filePath) => {
    setStoredPanelState(true);
    set({
      isPanelOpen: true,
      currentFilePath: filePath || null,
      error: null,
    });
    // Trigger loading when panel opens
    get().loadHistory(filePath);
  },

  closePanel: () => {
    setStoredPanelState(false);
    set({
      isPanelOpen: false,
      selectedEntry: null,
      historyContent: null,
    });
  },

  togglePanel: () => {
    const { isPanelOpen } = get();
    if (isPanelOpen) {
      get().closePanel();
    } else {
      get().openPanel();
    }
  },

  // API integration actions
  // These will be called and the actual WebSocket communication
  // will be handled by the websocket service
  loadHistory: (filePath) => {
    set({
      isLoading: true,
      error: null,
      currentFilePath: filePath || null,
    });
    // The actual WebSocket request will be made by the component
    // or via the websocket service. This just sets loading state.
    // Timeout fallback
    setTimeout(() => {
      if (get().isLoading) {
        set({ isLoading: false, error: 'Request timed out' });
      }
    }, 10000);
  },

  loadEntryContent: (_entryId) => {
    set({ isLoadingContent: true, error: null });
    // The actual WebSocket request will be made by the component
    // or via the websocket service. This just sets loading state.
    // Timeout fallback
    setTimeout(() => {
      if (get().isLoadingContent) {
        set({ isLoadingContent: false, error: 'Content request timed out' });
      }
    }, 10000);
  },
}));
