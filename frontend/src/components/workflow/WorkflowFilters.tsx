import { useState } from 'react';
import { Search, X, ArrowUpDown } from 'lucide-react';
import type { WorkflowStatus } from '../../hooks/useWorkflows';

interface WorkflowFiltersProps {
  status: WorkflowStatus | undefined;
  onStatusChange: (status: WorkflowStatus | undefined) => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  sortBy: 'created' | 'updated' | 'name' | 'status';
  sortOrder: 'asc' | 'desc';
  onSortChange: (sortBy: 'created' | 'updated' | 'name' | 'status', order: 'asc' | 'desc') => void;
}

const STATUS_TABS: { value: WorkflowStatus | undefined; label: string }[] = [
  { value: undefined, label: 'All' },
  { value: 'running', label: 'Running' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'paused', label: 'Paused' },
  { value: 'pending', label: 'Pending' },
];

const SORT_OPTIONS: { value: 'created' | 'updated' | 'name' | 'status'; label: string }[] = [
  { value: 'created', label: 'Created' },
  { value: 'updated', label: 'Updated' },
  { value: 'name', label: 'Name' },
  { value: 'status', label: 'Status' },
];

export function WorkflowFilters({
  status,
  onStatusChange,
  searchQuery,
  onSearchChange,
  sortBy,
  sortOrder,
  onSortChange,
}: WorkflowFiltersProps) {
  const [showSortMenu, setShowSortMenu] = useState(false);

  const handleSortSelect = (newSortBy: 'created' | 'updated' | 'name' | 'status') => {
    if (newSortBy === sortBy) {
      // Toggle order if same field
      onSortChange(sortBy, sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      // Default to descending for new field
      onSortChange(newSortBy, 'desc');
    }
    setShowSortMenu(false);
  };

  return (
    <div className="space-y-4">
      {/* Status Tabs */}
      <div className="flex items-center gap-1 overflow-x-auto">
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value ?? 'all'}
            onClick={() => onStatusChange(tab.value)}
            className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors whitespace-nowrap ${
              status === tab.value
                ? 'bg-editor-accent text-white'
                : 'text-editor-muted hover:text-editor-text hover:bg-editor-surface'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Search and Sort */}
      <div className="flex items-center gap-4">
        {/* Search Input */}
        <div className="flex-1 relative">
          <Search
            size={16}
            className="absolute left-3 top-1/2 -translate-y-1/2 text-editor-muted"
          />
          <input
            type="text"
            placeholder="Search workflows..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            className="w-full pl-9 pr-9 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
          />
          {searchQuery && (
            <button
              onClick={() => onSearchChange('')}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-editor-muted hover:text-editor-text"
            >
              <X size={16} />
            </button>
          )}
        </div>

        {/* Sort Dropdown */}
        <div className="relative">
          <button
            onClick={() => setShowSortMenu(!showSortMenu)}
            className="flex items-center gap-2 px-4 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-muted hover:text-editor-text transition-colors"
          >
            <ArrowUpDown size={16} />
            <span className="text-sm">
              {SORT_OPTIONS.find((opt) => opt.value === sortBy)?.label}
            </span>
            <span className="text-xs">{sortOrder === 'asc' ? '\u2191' : '\u2193'}</span>
          </button>

          {showSortMenu && (
            <>
              <div
                className="fixed inset-0 z-10"
                onClick={() => setShowSortMenu(false)}
              />
              <div className="absolute right-0 mt-1 w-40 bg-editor-surface border border-editor-border rounded-lg shadow-lg z-20 overflow-hidden">
                {SORT_OPTIONS.map((option) => (
                  <button
                    key={option.value}
                    onClick={() => handleSortSelect(option.value)}
                    className={`w-full flex items-center justify-between px-4 py-2 text-sm transition-colors ${
                      sortBy === option.value
                        ? 'bg-editor-accent/10 text-editor-accent'
                        : 'text-editor-muted hover:text-editor-text hover:bg-editor-hover'
                    }`}
                  >
                    <span>{option.label}</span>
                    {sortBy === option.value && (
                      <span className="text-xs">{sortOrder === 'asc' ? '\u2191' : '\u2193'}</span>
                    )}
                  </button>
                ))}
              </div>
            </>
          )}
        </div>

        {/* Clear Filters */}
        {(status || searchQuery) && (
          <button
            onClick={() => {
              onStatusChange(undefined);
              onSearchChange('');
            }}
            className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text transition-colors"
          >
            Clear filters
          </button>
        )}
      </div>
    </div>
  );
}
