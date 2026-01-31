import { useState, useCallback, useMemo, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import type {
  TaskFilterState,
  StatusFilter,
  DateRange,
  DateRangePreset,
  Task,
  ActiveFilter,
} from '../types/tasks';
import { DEFAULT_FILTER_STATE, STATUS_OPTIONS, DATE_RANGE_OPTIONS } from '../types/tasks';

interface UseTaskFiltersOptions {
  /** Debounce delay for search in milliseconds */
  searchDebounce?: number;
  /** Whether to sync filters with URL params */
  syncUrl?: boolean;
}

interface UseTaskFiltersResult {
  /** Current filter state */
  filters: TaskFilterState;
  /** Set a specific filter value */
  setFilter: <K extends keyof TaskFilterState>(key: K, value: TaskFilterState[K]) => void;
  /** Clear all filters */
  clearFilters: () => void;
  /** Remove a specific filter */
  removeFilter: (key: keyof TaskFilterState) => void;
  /** Whether any filters are active */
  hasActiveFilters: boolean;
  /** List of active filters for display */
  activeFilters: ActiveFilter[];
  /** Filter tasks based on current filters */
  filterTasks: (tasks: Task[]) => Task[];
  /** Loading state for when filters change */
  isFiltering: boolean;
  /** Debounced search value */
  debouncedSearch: string;
}

export function useTaskFilters(options: UseTaskFiltersOptions = {}): UseTaskFiltersResult {
  const { searchDebounce = 300, syncUrl = true } = options;
  const [searchParams, setSearchParams] = useSearchParams();
  const [filters, setFilters] = useState<TaskFilterState>(() => {
    if (!syncUrl) return DEFAULT_FILTER_STATE;
    return parseUrlParams(searchParams);
  });
  const [debouncedSearch, setDebouncedSearch] = useState(filters.search);
  const [isFiltering, setIsFiltering] = useState(false);

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(filters.search);
    }, searchDebounce);

    return () => clearTimeout(timer);
  }, [filters.search, searchDebounce]);

  // Sync filters to URL
  useEffect(() => {
    if (!syncUrl) return;

    const params = new URLSearchParams();

    if (filters.search) {
      params.set('q', filters.search);
    }
    if (filters.status !== 'all') {
      params.set('status', filters.status);
    }
    if (filters.dateRange.preset !== 'all') {
      params.set('date', filters.dateRange.preset);
      if (filters.dateRange.preset === 'custom') {
        if (filters.dateRange.startDate) {
          params.set('start', filters.dateRange.startDate.toISOString().split('T')[0]);
        }
        if (filters.dateRange.endDate) {
          params.set('end', filters.dateRange.endDate.toISOString().split('T')[0]);
        }
      }
    }
    if (filters.agentId) {
      params.set('agent', filters.agentId);
    }
    if (filters.conversationId) {
      params.set('conversation', filters.conversationId);
    }

    setSearchParams(params, { replace: true });
  }, [filters, syncUrl, setSearchParams]);

  // Set a specific filter
  const setFilter = useCallback(
    <K extends keyof TaskFilterState>(key: K, value: TaskFilterState[K]) => {
      setIsFiltering(true);
      setFilters((prev) => ({ ...prev, [key]: value }));
      // Reset filtering state after a short delay
      setTimeout(() => setIsFiltering(false), 100);
    },
    []
  );

  // Clear all filters
  const clearFilters = useCallback(() => {
    setIsFiltering(true);
    setFilters(DEFAULT_FILTER_STATE);
    setTimeout(() => setIsFiltering(false), 100);
  }, []);

  // Remove a specific filter
  const removeFilter = useCallback((key: keyof TaskFilterState) => {
    setIsFiltering(true);
    setFilters((prev) => ({
      ...prev,
      [key]: DEFAULT_FILTER_STATE[key],
    }));
    setTimeout(() => setIsFiltering(false), 100);
  }, []);

  // Check if any filters are active
  const hasActiveFilters = useMemo(() => {
    return (
      filters.search !== '' ||
      filters.status !== 'all' ||
      filters.dateRange.preset !== 'all' ||
      filters.agentId !== null ||
      filters.conversationId !== null
    );
  }, [filters]);

  // Get list of active filters for display
  const activeFilters = useMemo((): ActiveFilter[] => {
    const active: ActiveFilter[] = [];

    if (filters.search) {
      active.push({
        key: 'search',
        label: `Search: ${filters.search}`,
        value: filters.search,
      });
    }

    if (filters.status !== 'all') {
      const statusLabel = STATUS_OPTIONS.find((s) => s.value === filters.status)?.label || filters.status;
      active.push({
        key: 'status',
        label: statusLabel,
        value: filters.status,
      });
    }

    if (filters.dateRange.preset !== 'all') {
      let dateLabel: string;
      if (filters.dateRange.preset === 'custom' && filters.dateRange.startDate && filters.dateRange.endDate) {
        dateLabel = `${formatShortDate(filters.dateRange.startDate)} - ${formatShortDate(filters.dateRange.endDate)}`;
      } else {
        dateLabel = DATE_RANGE_OPTIONS.find((d) => d.value === filters.dateRange.preset)?.label || filters.dateRange.preset;
      }
      active.push({
        key: 'dateRange',
        label: dateLabel,
        value: filters.dateRange.preset,
      });
    }

    if (filters.agentId) {
      active.push({
        key: 'agentId',
        label: `Agent: ${filters.agentId}`,
        value: filters.agentId,
      });
    }

    if (filters.conversationId) {
      active.push({
        key: 'conversationId',
        label: `Conversation: ${filters.conversationId}`,
        value: filters.conversationId,
      });
    }

    return active;
  }, [filters]);

  // Filter tasks based on current filters
  const filterTasks = useCallback(
    (tasks: Task[]): Task[] => {
      return tasks.filter((task) => {
        // Search filter
        if (debouncedSearch) {
          const searchLower = debouncedSearch.toLowerCase();
          if (!task.prompt.toLowerCase().includes(searchLower)) {
            return false;
          }
        }

        // Status filter
        if (filters.status !== 'all' && task.status !== filters.status) {
          return false;
        }

        // Date range filter
        if (filters.dateRange.preset !== 'all') {
          const taskDate = new Date(task.createdAt);
          if (filters.dateRange.startDate && taskDate < filters.dateRange.startDate) {
            return false;
          }
          if (filters.dateRange.endDate && taskDate > filters.dateRange.endDate) {
            return false;
          }
        }

        // Agent filter
        if (filters.agentId && task.agentId !== filters.agentId) {
          return false;
        }

        // Conversation filter
        if (filters.conversationId && task.conversationId !== filters.conversationId) {
          return false;
        }

        return true;
      });
    },
    [debouncedSearch, filters]
  );

  return {
    filters,
    setFilter,
    clearFilters,
    removeFilter,
    hasActiveFilters,
    activeFilters,
    filterTasks,
    isFiltering,
    debouncedSearch,
  };
}

// Utility functions
function parseUrlParams(params: URLSearchParams): TaskFilterState {
  const search = params.get('q') || '';
  const status = (params.get('status') as StatusFilter) || 'all';
  const datePreset = (params.get('date') as DateRangePreset) || 'all';
  const startStr = params.get('start');
  const endStr = params.get('end');
  const agentId = params.get('agent') || null;
  const conversationId = params.get('conversation') || null;

  let dateRange: DateRange = { preset: 'all', startDate: null, endDate: null };

  if (datePreset === 'custom' && startStr && endStr) {
    dateRange = {
      preset: 'custom',
      startDate: new Date(startStr),
      endDate: new Date(endStr),
    };
  } else if (datePreset !== 'all') {
    dateRange = getDateRangeFromPreset(datePreset);
  }

  return {
    search,
    status,
    dateRange,
    agentId,
    conversationId,
  };
}

function getDateRangeFromPreset(preset: DateRangePreset): DateRange {
  const now = new Date();
  now.setHours(23, 59, 59, 999);

  switch (preset) {
    case 'today': {
      const start = new Date();
      start.setHours(0, 0, 0, 0);
      return { preset, startDate: start, endDate: now };
    }
    case '7d': {
      const start = new Date();
      start.setDate(start.getDate() - 7);
      start.setHours(0, 0, 0, 0);
      return { preset, startDate: start, endDate: now };
    }
    case '30d': {
      const start = new Date();
      start.setDate(start.getDate() - 30);
      start.setHours(0, 0, 0, 0);
      return { preset, startDate: start, endDate: now };
    }
    case 'all':
    default:
      return { preset: 'all', startDate: null, endDate: null };
  }
}

function formatShortDate(date: Date): string {
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  });
}
