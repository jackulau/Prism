import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { ToolType } from '../types/tools';

interface ToolCatalogState {
  // Filter state
  searchQuery: string;
  selectedProvider: string | null;
  selectedType: ToolType;

  // Pagination (for future use)
  page: number;
  pageSize: number;

  // Selection
  selectedToolId: string | null;

  // UI state
  viewMode: 'grid' | 'list';

  // Actions
  setSearchQuery: (query: string) => void;
  setSelectedProvider: (provider: string | null) => void;
  setSelectedType: (type: ToolType) => void;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  setSelectedToolId: (id: string | null) => void;
  setViewMode: (mode: 'grid' | 'list') => void;
  resetFilters: () => void;
}

const initialFilters = {
  searchQuery: '',
  selectedProvider: null,
  selectedType: 'all' as ToolType,
  page: 1,
  pageSize: 24,
  selectedToolId: null,
};

export const useToolCatalogStore = create<ToolCatalogState>()(
  persist(
    (set) => ({
      // Initial state
      ...initialFilters,
      viewMode: 'grid',

      // Actions
      setSearchQuery: (query) => set({ searchQuery: query, page: 1 }),
      setSelectedProvider: (provider) => set({ selectedProvider: provider, page: 1 }),
      setSelectedType: (type) => set({ selectedType: type, page: 1 }),
      setPage: (page) => set({ page }),
      setPageSize: (size) => set({ pageSize: size, page: 1 }),
      setSelectedToolId: (id) => set({ selectedToolId: id }),
      setViewMode: (mode) => set({ viewMode: mode }),
      resetFilters: () => set(initialFilters),
    }),
    {
      name: 'tool-catalog-preferences',
      partialize: (state) => ({
        viewMode: state.viewMode,
        pageSize: state.pageSize,
      }),
    }
  )
);

// Selector hooks for optimized re-renders
export const useToolFilters = () =>
  useToolCatalogStore((state) => ({
    searchQuery: state.searchQuery,
    selectedProvider: state.selectedProvider,
    selectedType: state.selectedType,
  }));

export const useToolSelection = () =>
  useToolCatalogStore((state) => ({
    selectedToolId: state.selectedToolId,
    setSelectedToolId: state.setSelectedToolId,
  }));

export const useToolViewMode = () =>
  useToolCatalogStore((state) => ({
    viewMode: state.viewMode,
    setViewMode: state.setViewMode,
  }));
