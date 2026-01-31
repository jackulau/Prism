import { Search, X, Calendar } from 'lucide-react';
import type { TaskFiltersState, TaskStatus } from '../../hooks/useTasks';

interface TaskFiltersProps {
  filters: TaskFiltersState;
  onChange: (filters: TaskFiltersState) => void;
  onClear: () => void;
  hasActiveFilters: boolean;
}

const statusOptions: Array<{ value: TaskStatus | ''; label: string }> = [
  { value: '', label: 'All Statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'running', label: 'Running' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'cancelled', label: 'Cancelled' },
];

export function TaskFilters({ filters, onChange, onClear, hasActiveFilters }: TaskFiltersProps) {
  const handleStatusChange = (status: TaskStatus | '') => {
    onChange({ ...filters, status });
  };

  const handleSearchChange = (search: string) => {
    onChange({ ...filters, search });
  };

  const handleDateFromChange = (dateFrom: string) => {
    onChange({ ...filters, dateFrom });
  };

  const handleDateToChange = (dateTo: string) => {
    onChange({ ...filters, dateTo });
  };

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
      <div className="flex flex-wrap items-center gap-4">
        {/* Search input */}
        <div className="relative flex-1 min-w-[200px]">
          <Search
            size={16}
            className="absolute left-3 top-1/2 -translate-y-1/2 text-editor-muted"
          />
          <input
            type="text"
            placeholder="Search tasks..."
            value={filters.search || ''}
            onChange={(e) => handleSearchChange(e.target.value)}
            className="w-full pl-9 pr-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text placeholder:text-editor-muted focus:outline-none focus:border-editor-accent"
          />
        </div>

        {/* Status filter */}
        <select
          value={filters.status || ''}
          onChange={(e) => handleStatusChange(e.target.value as TaskStatus | '')}
          className="px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent cursor-pointer"
        >
          {statusOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>

        {/* Date range filters */}
        <div className="flex items-center gap-2">
          <div className="relative">
            <Calendar
              size={14}
              className="absolute left-2.5 top-1/2 -translate-y-1/2 text-editor-muted pointer-events-none"
            />
            <input
              type="date"
              value={filters.dateFrom || ''}
              onChange={(e) => handleDateFromChange(e.target.value)}
              className="pl-8 pr-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent text-sm"
              placeholder="From"
            />
          </div>
          <span className="text-editor-muted">to</span>
          <div className="relative">
            <Calendar
              size={14}
              className="absolute left-2.5 top-1/2 -translate-y-1/2 text-editor-muted pointer-events-none"
            />
            <input
              type="date"
              value={filters.dateTo || ''}
              onChange={(e) => handleDateToChange(e.target.value)}
              className="pl-8 pr-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent text-sm"
              placeholder="To"
            />
          </div>
        </div>

        {/* Clear filters button */}
        {hasActiveFilters && (
          <button
            onClick={onClear}
            className="flex items-center gap-1.5 px-3 py-2 text-editor-muted hover:text-editor-text transition-colors"
          >
            <X size={14} />
            <span className="text-sm">Clear filters</span>
          </button>
        )}
      </div>
    </div>
  );
}
