import { memo } from 'react';
import { Search, Filter, X } from 'lucide-react';
import type { AgentFiltersState, AgentStatus } from './types';

interface AgentFiltersProps {
  filters: AgentFiltersState;
  onChange: (filters: AgentFiltersState) => void;
  availableModels?: string[];
}

const statusOptions: { value: AgentStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'All Statuses' },
  { value: 'running', label: 'Running' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'pending', label: 'Pending' },
  { value: 'cancelled', label: 'Cancelled' },
];

const dateRangeOptions: { value: AgentFiltersState['dateRange']; label: string }[] = [
  { value: 'all', label: 'All Time' },
  { value: 'today', label: 'Today' },
  { value: 'week', label: 'This Week' },
  { value: 'month', label: 'This Month' },
];

export const AgentFilters = memo(function AgentFilters({
  filters,
  onChange,
  availableModels = [],
}: AgentFiltersProps) {
  const handleStatusChange = (status: AgentStatus | 'all') => {
    onChange({ ...filters, status });
  };

  const handleModelChange = (model: string) => {
    onChange({ ...filters, model });
  };

  const handleDateRangeChange = (dateRange: AgentFiltersState['dateRange']) => {
    onChange({ ...filters, dateRange });
  };

  const handleSearchChange = (search: string) => {
    onChange({ ...filters, search });
  };

  const clearSearch = () => {
    onChange({ ...filters, search: '' });
  };

  const hasActiveFilters =
    filters.status !== 'all' ||
    filters.model !== 'all' ||
    filters.dateRange !== 'all' ||
    filters.search !== '';

  const clearAllFilters = () => {
    onChange({
      status: 'all',
      model: 'all',
      dateRange: 'all',
      search: '',
    });
  };

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-3">
        {/* Search */}
        <div className="relative flex-1 min-w-[200px]">
          <Search
            size={16}
            className="absolute left-3 top-1/2 -translate-y-1/2 text-editor-muted"
          />
          <input
            type="text"
            value={filters.search}
            onChange={(e) => handleSearchChange(e.target.value)}
            placeholder="Search agents..."
            className="w-full pl-9 pr-8 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text placeholder:text-editor-muted focus:outline-none focus:border-editor-accent/50 transition-colors"
          />
          {filters.search && (
            <button
              onClick={clearSearch}
              className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-editor-muted hover:text-editor-text transition-colors"
            >
              <X size={14} />
            </button>
          )}
        </div>

        {/* Status Filter */}
        <select
          value={filters.status}
          onChange={(e) => handleStatusChange(e.target.value as AgentStatus | 'all')}
          className="px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent/50 transition-colors cursor-pointer"
        >
          {statusOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>

        {/* Model Filter */}
        {availableModels.length > 0 && (
          <select
            value={filters.model}
            onChange={(e) => handleModelChange(e.target.value)}
            className="px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent/50 transition-colors cursor-pointer"
          >
            <option value="all">All Models</option>
            {availableModels.map((model) => (
              <option key={model} value={model}>
                {model}
              </option>
            ))}
          </select>
        )}

        {/* Date Range Filter */}
        <select
          value={filters.dateRange}
          onChange={(e) =>
            handleDateRangeChange(e.target.value as AgentFiltersState['dateRange'])
          }
          className="px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent/50 transition-colors cursor-pointer"
        >
          {dateRangeOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>

        {/* Clear All Button */}
        {hasActiveFilters && (
          <button
            onClick={clearAllFilters}
            className="flex items-center gap-1.5 px-3 py-2 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-bg border border-editor-border rounded-lg transition-colors"
          >
            <Filter size={14} />
            Clear Filters
          </button>
        )}
      </div>
    </div>
  );
});

export default AgentFilters;
