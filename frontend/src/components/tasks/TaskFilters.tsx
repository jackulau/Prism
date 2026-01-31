import { useCallback, useMemo } from 'react';
import { Search, X, Loader2, ListFilter } from 'lucide-react';
import { TaskFilterDropdown } from './TaskFilterDropdown';
import { DateRangePicker } from './DateRangePicker';
import type {
  TaskFilterState,
  StatusFilter,
  DateRange,
  ActiveFilter,
  Task,
  FilterOption,
} from '../../types/tasks';
import { STATUS_OPTIONS } from '../../types/tasks';

interface TaskFiltersProps {
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
  /** Whether filtering is in progress */
  isFiltering?: boolean;
  /** Tasks to extract agent options from */
  tasks?: Task[];
  /** Custom class name */
  className?: string;
}

export function TaskFilters({
  filters,
  setFilter,
  clearFilters,
  removeFilter,
  hasActiveFilters,
  activeFilters,
  isFiltering = false,
  tasks = [],
  className = '',
}: TaskFiltersProps) {
  // Extract unique agents from tasks for filter options
  const agentOptions = useMemo((): FilterOption<string>[] => {
    const agents = new Map<string, { name: string; count: number }>();
    tasks.forEach((task) => {
      if (task.agentId) {
        const existing = agents.get(task.agentId);
        if (existing) {
          existing.count++;
        } else {
          agents.set(task.agentId, {
            name: task.agentName || task.agentId,
            count: 1,
          });
        }
      }
    });

    const options: FilterOption<string>[] = [{ value: '', label: 'All Agents' }];
    agents.forEach((agent, id) => {
      options.push({
        value: id,
        label: agent.name,
        count: agent.count,
      });
    });

    return options;
  }, [tasks]);

  // Handle search input change
  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setFilter('search', e.target.value);
    },
    [setFilter]
  );

  // Handle status filter change
  const handleStatusChange = useCallback(
    (value: StatusFilter | StatusFilter[]) => {
      const status = Array.isArray(value) ? value[0] : value;
      setFilter('status', status);
    },
    [setFilter]
  );

  // Handle date range change
  const handleDateRangeChange = useCallback(
    (range: DateRange) => {
      setFilter('dateRange', range);
    },
    [setFilter]
  );

  // Handle agent filter change
  const handleAgentChange = useCallback(
    (value: string | string[]) => {
      const agentId = Array.isArray(value) ? value[0] : value;
      setFilter('agentId', agentId || null);
    },
    [setFilter]
  );

  // Get filter tag color based on type
  const getFilterTagColor = (key: keyof TaskFilterState) => {
    switch (key) {
      case 'status':
        return 'bg-editor-accent/10 text-editor-accent';
      case 'dateRange':
        return 'bg-editor-success/10 text-editor-success';
      case 'search':
        return 'bg-editor-warning/10 text-editor-warning';
      case 'agentId':
        return 'bg-editor-error/10 text-editor-error';
      default:
        return 'bg-editor-muted/10 text-editor-muted';
    }
  };

  return (
    <div className={`space-y-3 ${className}`}>
      {/* Filter bar */}
      <div className="flex flex-wrap gap-3 items-center bg-editor-surface border border-editor-border rounded-lg p-4">
        {/* Search input */}
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-editor-muted" />
          <input
            type="text"
            value={filters.search}
            onChange={handleSearchChange}
            placeholder="Search tasks..."
            className="w-full pl-10 pr-4 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
            aria-label="Search tasks"
          />
          {isFiltering && (
            <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-editor-accent animate-spin" />
          )}
        </div>

        {/* Status Filter */}
        <TaskFilterDropdown
          label="Status"
          options={STATUS_OPTIONS}
          value={filters.status}
          onChange={handleStatusChange}
        />

        {/* Date Range */}
        <DateRangePicker
          value={filters.dateRange}
          onChange={handleDateRangeChange}
        />

        {/* Agent Filter (only show if there are agents) */}
        {agentOptions.length > 1 && (
          <TaskFilterDropdown
            label="Agent"
            options={agentOptions}
            value={filters.agentId || ''}
            onChange={handleAgentChange}
            searchable={agentOptions.length > 5}
            searchPlaceholder="Search agents..."
          />
        )}

        {/* Clear Filters */}
        {hasActiveFilters && (
          <button
            onClick={clearFilters}
            className="flex items-center gap-1.5 px-3 py-2 text-sm text-editor-muted hover:text-editor-text transition-colors"
            aria-label="Clear all filters"
          >
            <X className="w-4 h-4" />
            <span className="hidden sm:inline">Clear all</span>
          </button>
        )}
      </div>

      {/* Active filter tags */}
      {hasActiveFilters && (
        <div className="flex flex-wrap items-center gap-2">
          <ListFilter className="w-4 h-4 text-editor-muted" />
          {activeFilters.map((filter) => (
            <span
              key={filter.key}
              className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-sm ${getFilterTagColor(
                filter.key
              )}`}
            >
              {filter.label}
              <button
                onClick={() => removeFilter(filter.key)}
                className="hover:opacity-70 transition-opacity"
                aria-label={`Remove ${filter.label} filter`}
              >
                <X className="w-3.5 h-3.5" />
              </button>
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

interface TaskFiltersEmptyStateProps {
  /** Whether filters are applied */
  hasFilters: boolean;
  /** Callback to clear filters */
  onClearFilters: () => void;
}

export function TaskFiltersEmptyState({
  hasFilters,
  onClearFilters,
}: TaskFiltersEmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <ListFilter className="w-12 h-12 text-editor-muted mb-4" />
      <h3 className="text-lg font-medium text-editor-text mb-2">
        {hasFilters ? 'No matching tasks' : 'No tasks yet'}
      </h3>
      <p className="text-editor-muted mb-4">
        {hasFilters
          ? 'Try adjusting your filters to find what you\'re looking for.'
          : 'Tasks will appear here when they are created.'}
      </p>
      {hasFilters && (
        <button
          onClick={onClearFilters}
          className="px-4 py-2 bg-editor-accent text-editor-bg rounded-lg hover:bg-editor-accent/90 transition-colors"
        >
          Clear all filters
        </button>
      )}
    </div>
  );
}

interface TaskFiltersLoadingProps {
  /** Number of skeleton items to show */
  count?: number;
}

export function TaskFiltersLoading({ count = 3 }: TaskFiltersLoadingProps) {
  return (
    <div className="space-y-4">
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse"
        >
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 bg-editor-border rounded-lg" />
            <div className="flex-1 space-y-2">
              <div className="h-4 bg-editor-border rounded w-1/4" />
              <div className="h-3 bg-editor-border rounded w-1/2" />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
