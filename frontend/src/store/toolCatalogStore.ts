import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type ViewMode = 'grid' | 'list';

interface ToolCatalogState {
  // Filter state
  searchQuery: string;
  selectedProvider: string | null;
  showModelsOnly: boolean;
  showBuiltinOnly: boolean;
  showCustomOnly: boolean;

  // Pagination
  page: number;
  pageSize: number;

  // Selection
  selectedToolId: string | null;

  // UI state
  isLoading: boolean;
  error: string | null;
  viewMode: ViewMode;

  // Actions
  setSearchQuery: (query: string) => void;
  setSelectedProvider: (provider: string | null) => void;
  setShowModelsOnly: (show: boolean) => void;
  setShowBuiltinOnly: (show: boolean) => void;
  setShowCustomOnly: (show: boolean) => void;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  setSelectedToolId: (id: string | null) => void;
  setViewMode: (mode: ViewMode) => void;
  setIsLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  resetFilters: () => void;
}

const DEFAULT_PAGE_SIZE = 20;

const initialFilterState = {
  searchQuery: '',
  selectedProvider: null,
  showModelsOnly: false,
  showBuiltinOnly: false,
  showCustomOnly: false,
  page: 1,
  selectedToolId: null,
  isLoading: false,
  error: null,
};

export const useToolCatalogStore = create<ToolCatalogState>()(
  persist(
    (set) => ({
      // Filter state
      ...initialFilterState,

      // User preferences (persisted)
      pageSize: DEFAULT_PAGE_SIZE,
      viewMode: 'grid' as ViewMode,

      // Actions
      setSearchQuery: (query) => set({ searchQuery: query, page: 1 }),
      setSelectedProvider: (provider) => set({ selectedProvider: provider, page: 1 }),
      setShowModelsOnly: (show) => set({ showModelsOnly: show, page: 1 }),
      setShowBuiltinOnly: (show) => set({ showBuiltinOnly: show, page: 1 }),
      setShowCustomOnly: (show) => set({ showCustomOnly: show, page: 1 }),
      setPage: (page) => set({ page }),
      setPageSize: (size) => set({ pageSize: size, page: 1 }),
      setSelectedToolId: (id) => set({ selectedToolId: id }),
      setViewMode: (mode) => set({ viewMode: mode }),
      setIsLoading: (loading) => set({ isLoading: loading }),
      setError: (error) => set({ error }),
      resetFilters: () => set(initialFilterState),
    }),
    {
      name: 'prism-tool-catalog',
      partialize: (state) => ({
        viewMode: state.viewMode,
        pageSize: state.pageSize,
      }),
    }
  )
);

// Selector functions for optimized re-renders
export const selectSearchQuery = (state: ToolCatalogState) => state.searchQuery;
export const selectSelectedProvider = (state: ToolCatalogState) => state.selectedProvider;
export const selectShowModelsOnly = (state: ToolCatalogState) => state.showModelsOnly;
export const selectShowBuiltinOnly = (state: ToolCatalogState) => state.showBuiltinOnly;
export const selectShowCustomOnly = (state: ToolCatalogState) => state.showCustomOnly;
export const selectPage = (state: ToolCatalogState) => state.page;
export const selectPageSize = (state: ToolCatalogState) => state.pageSize;
export const selectSelectedToolId = (state: ToolCatalogState) => state.selectedToolId;
export const selectViewMode = (state: ToolCatalogState) => state.viewMode;
export const selectIsLoading = (state: ToolCatalogState) => state.isLoading;
export const selectError = (state: ToolCatalogState) => state.error;

// Compound selectors
export const selectFilters = (state: ToolCatalogState) => ({
  searchQuery: state.searchQuery,
  selectedProvider: state.selectedProvider,
  showModelsOnly: state.showModelsOnly,
  showBuiltinOnly: state.showBuiltinOnly,
  showCustomOnly: state.showCustomOnly,
});

export const selectPagination = (state: ToolCatalogState) => ({
  page: state.page,
  pageSize: state.pageSize,
});

export const selectHasActiveFilters = (state: ToolCatalogState) =>
  state.searchQuery !== '' ||
  state.selectedProvider !== null ||
  state.showModelsOnly ||
  state.showBuiltinOnly ||
  state.showCustomOnly;
