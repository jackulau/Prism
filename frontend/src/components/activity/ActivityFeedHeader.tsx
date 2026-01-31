import { useState, useEffect, useCallback } from 'react';
import { Search, Pause, Play, Trash2, Filter, X } from 'lucide-react';
import type { EventTypeFilter, EventStatusFilter } from './ActivityFeed';

interface ActivityFeedHeaderProps {
  typeFilter: EventTypeFilter;
  statusFilter: EventStatusFilter;
  searchQuery: string;
  isPaused: boolean;
  isLive: boolean;
  onTypeFilterChange: (filter: EventTypeFilter) => void;
  onStatusFilterChange: (filter: EventStatusFilter) => void;
  onSearchChange: (query: string) => void;
  onTogglePause: () => void;
  onClear: () => void;
}

const TYPE_FILTER_OPTIONS: { value: EventTypeFilter; label: string }[] = [
  { value: 'all', label: 'All Types' },
  { value: 'agent', label: 'Agent' },
  { value: 'swarm', label: 'Swarm' },
  { value: 'build', label: 'Build' },
  { value: 'connection', label: 'Connection' },
  { value: 'notification', label: 'Notification' },
  { value: 'system', label: 'System' },
];

const STATUS_FILTER_OPTIONS: { value: EventStatusFilter; label: string }[] = [
  { value: 'all', label: 'All Status' },
  { value: 'success', label: 'Success' },
  { value: 'info', label: 'Info' },
  { value: 'warning', label: 'Warning' },
  { value: 'error', label: 'Error' },
];

const SEARCH_DEBOUNCE_MS = 300;

export function ActivityFeedHeader({
  typeFilter,
  statusFilter,
  searchQuery,
  isPaused,
  isLive,
  onTypeFilterChange,
  onStatusFilterChange,
  onSearchChange,
  onTogglePause,
  onClear,
}: ActivityFeedHeaderProps) {
  const [localSearch, setLocalSearch] = useState(searchQuery);
  const [showFilters, setShowFilters] = useState(false);

  // Debounce search
  useEffect(() => {
    const timeout = setTimeout(() => {
      onSearchChange(localSearch);
    }, SEARCH_DEBOUNCE_MS);

    return () => clearTimeout(timeout);
  }, [localSearch, onSearchChange]);

  const clearSearch = useCallback(() => {
    setLocalSearch('');
    onSearchChange('');
  }, [onSearchChange]);

  const hasActiveFilters = typeFilter !== 'all' || statusFilter !== 'all';

  return (
    <div className="border-b border-editor-border bg-editor-surface/50">
      {/* Title row */}
      <div className="flex items-center justify-between px-3 py-2">
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-semibold text-editor-text">Activity Feed</h3>
          {/* Live indicator */}
          <div className="flex items-center gap-1.5">
            <div
              className={`w-2 h-2 rounded-full ${
                isLive ? 'bg-editor-success animate-pulse' : 'bg-editor-muted'
              }`}
            />
            <span className="text-xs text-editor-muted">
              {isLive ? 'Live' : isPaused ? 'Paused' : 'Offline'}
            </span>
          </div>
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-1">
          {/* Filter toggle */}
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`p-1.5 rounded hover:bg-editor-surface transition-colors ${
              hasActiveFilters || showFilters ? 'text-editor-accent' : 'text-editor-muted'
            }`}
            title="Toggle filters"
          >
            <Filter size={14} />
          </button>

          {/* Pause/Resume */}
          <button
            onClick={onTogglePause}
            className={`p-1.5 rounded hover:bg-editor-surface transition-colors ${
              isPaused ? 'text-editor-warning' : 'text-editor-muted'
            }`}
            title={isPaused ? 'Resume' : 'Pause'}
          >
            {isPaused ? <Play size={14} /> : <Pause size={14} />}
          </button>

          {/* Clear */}
          <button
            onClick={onClear}
            className="p-1.5 rounded text-editor-muted hover:text-editor-error hover:bg-editor-surface transition-colors"
            title="Clear all events"
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>

      {/* Filter/Search row (collapsible) */}
      {showFilters && (
        <div className="px-3 py-2 space-y-2 border-t border-editor-border/50">
          {/* Search */}
          <div className="relative">
            <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-editor-muted" />
            <input
              type="text"
              value={localSearch}
              onChange={(e) => setLocalSearch(e.target.value)}
              placeholder="Search events..."
              className="w-full pl-7 pr-7 py-1.5 text-xs bg-editor-bg border border-editor-border rounded focus:outline-none focus:border-editor-accent text-editor-text placeholder:text-editor-muted"
            />
            {localSearch && (
              <button
                onClick={clearSearch}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-editor-muted hover:text-editor-text"
              >
                <X size={12} />
              </button>
            )}
          </div>

          {/* Filter dropdowns */}
          <div className="flex items-center gap-2">
            {/* Type filter */}
            <select
              value={typeFilter}
              onChange={(e) => onTypeFilterChange(e.target.value as EventTypeFilter)}
              className="flex-1 px-2 py-1.5 text-xs bg-editor-bg border border-editor-border rounded focus:outline-none focus:border-editor-accent text-editor-text cursor-pointer"
            >
              {TYPE_FILTER_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>

            {/* Status filter */}
            <select
              value={statusFilter}
              onChange={(e) => onStatusFilterChange(e.target.value as EventStatusFilter)}
              className="flex-1 px-2 py-1.5 text-xs bg-editor-bg border border-editor-border rounded focus:outline-none focus:border-editor-accent text-editor-text cursor-pointer"
            >
              {STATUS_FILTER_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
        </div>
      )}
    </div>
  );
}
