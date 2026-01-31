import { useState, useRef, useEffect } from 'react';
import {
  Calendar,
  ChevronDown,
  Filter,
  Search,
  SortAsc,
  SortDesc,
  X,
} from 'lucide-react';
import {
  ResultsFilters,
  DEFAULT_FILTERS,
  STATUS_OPTIONS,
  TYPE_OPTIONS,
  DATE_RANGE_OPTIONS,
  SORT_OPTIONS,
  ExecutionStatus,
  ExecutionType,
  DateRangePreset,
  SortField,
} from './types';

interface ResultsFilterProps {
  filters: ResultsFilters;
  onFiltersChange: (filters: ResultsFilters) => void;
  resultCount?: number;
}

export function ResultsFilter({ filters, onFiltersChange, resultCount }: ResultsFilterProps) {
  const [showStatusDropdown, setShowStatusDropdown] = useState(false);
  const [showTypeDropdown, setShowTypeDropdown] = useState(false);
  const [showDateDropdown, setShowDateDropdown] = useState(false);
  const [showSortDropdown, setShowSortDropdown] = useState(false);

  const statusRef = useRef<HTMLDivElement>(null);
  const typeRef = useRef<HTMLDivElement>(null);
  const dateRef = useRef<HTMLDivElement>(null);
  const sortRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (statusRef.current && !statusRef.current.contains(event.target as Node)) {
        setShowStatusDropdown(false);
      }
      if (typeRef.current && !typeRef.current.contains(event.target as Node)) {
        setShowTypeDropdown(false);
      }
      if (dateRef.current && !dateRef.current.contains(event.target as Node)) {
        setShowDateDropdown(false);
      }
      if (sortRef.current && !sortRef.current.contains(event.target as Node)) {
        setShowSortDropdown(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const toggleStatus = (status: ExecutionStatus) => {
    const newStatuses = filters.statuses.includes(status)
      ? filters.statuses.filter((s) => s !== status)
      : [...filters.statuses, status];
    onFiltersChange({ ...filters, statuses: newStatuses });
  };

  const toggleType = (type: ExecutionType) => {
    const newTypes = filters.types.includes(type)
      ? filters.types.filter((t) => t !== type)
      : [...filters.types, type];
    onFiltersChange({ ...filters, types: newTypes });
  };

  const setDateRange = (dateRange: DateRangePreset) => {
    onFiltersChange({ ...filters, dateRange });
    setShowDateDropdown(false);
  };

  const setSort = (sortField: SortField) => {
    if (filters.sortField === sortField) {
      onFiltersChange({
        ...filters,
        sortDirection: filters.sortDirection === 'asc' ? 'desc' : 'asc',
      });
    } else {
      onFiltersChange({ ...filters, sortField, sortDirection: 'desc' });
    }
    setShowSortDropdown(false);
  };

  const hasActiveFilters =
    filters.statuses.length > 0 ||
    filters.types.length > 0 ||
    filters.dateRange !== DEFAULT_FILTERS.dateRange ||
    filters.searchQuery !== '';

  const clearFilters = () => {
    onFiltersChange(DEFAULT_FILTERS);
  };

  const getActiveFiltersCount = () => {
    let count = 0;
    if (filters.statuses.length > 0) count++;
    if (filters.types.length > 0) count++;
    if (filters.dateRange !== DEFAULT_FILTERS.dateRange) count++;
    if (filters.searchQuery !== '') count++;
    return count;
  };

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-3">
        {/* Search Input */}
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-editor-muted" />
          <input
            type="text"
            placeholder="Search by ID or name..."
            value={filters.searchQuery}
            onChange={(e) => onFiltersChange({ ...filters, searchQuery: e.target.value })}
            className="w-full pl-10 pr-4 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder:text-editor-muted focus:outline-none focus:border-editor-accent/50 transition-colors"
          />
          {filters.searchQuery && (
            <button
              onClick={() => onFiltersChange({ ...filters, searchQuery: '' })}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-editor-muted hover:text-editor-text transition-colors"
            >
              <X size={14} />
            </button>
          )}
        </div>

        {/* Date Range Dropdown */}
        <div ref={dateRef} className="relative">
          <button
            onClick={() => setShowDateDropdown(!showDateDropdown)}
            className={`flex items-center gap-2 px-3 py-2 border rounded-lg transition-colors ${
              filters.dateRange !== DEFAULT_FILTERS.dateRange
                ? 'bg-editor-accent/10 border-editor-accent/30 text-editor-accent'
                : 'bg-editor-surface border-editor-border text-editor-text hover:border-editor-accent/30'
            }`}
          >
            <Calendar size={16} />
            <span className="text-sm">
              {DATE_RANGE_OPTIONS.find((d) => d.value === filters.dateRange)?.label}
            </span>
            <ChevronDown size={14} />
          </button>

          {showDateDropdown && (
            <div className="absolute top-full left-0 mt-1 w-40 bg-editor-surface border border-editor-border rounded-lg shadow-lg z-10 py-1">
              {DATE_RANGE_OPTIONS.map((option) => (
                <button
                  key={option.value}
                  onClick={() => setDateRange(option.value)}
                  className={`w-full px-3 py-2 text-left text-sm transition-colors ${
                    filters.dateRange === option.value
                      ? 'bg-editor-accent/10 text-editor-accent'
                      : 'text-editor-text hover:bg-editor-surface/80'
                  }`}
                >
                  {option.label}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Status Filter Dropdown */}
        <div ref={statusRef} className="relative">
          <button
            onClick={() => setShowStatusDropdown(!showStatusDropdown)}
            className={`flex items-center gap-2 px-3 py-2 border rounded-lg transition-colors ${
              filters.statuses.length > 0
                ? 'bg-editor-accent/10 border-editor-accent/30 text-editor-accent'
                : 'bg-editor-surface border-editor-border text-editor-text hover:border-editor-accent/30'
            }`}
          >
            <Filter size={16} />
            <span className="text-sm">
              Status
              {filters.statuses.length > 0 && (
                <span className="ml-1 px-1.5 py-0.5 bg-editor-accent/20 rounded text-xs">
                  {filters.statuses.length}
                </span>
              )}
            </span>
            <ChevronDown size={14} />
          </button>

          {showStatusDropdown && (
            <div className="absolute top-full left-0 mt-1 w-48 bg-editor-surface border border-editor-border rounded-lg shadow-lg z-10 py-1">
              {STATUS_OPTIONS.map((option) => (
                <button
                  key={option.value}
                  onClick={() => toggleStatus(option.value)}
                  className="w-full px-3 py-2 text-left text-sm text-editor-text hover:bg-editor-surface/80 flex items-center gap-2 transition-colors"
                >
                  <div
                    className={`w-4 h-4 border rounded flex items-center justify-center ${
                      filters.statuses.includes(option.value)
                        ? 'bg-editor-accent border-editor-accent'
                        : 'border-editor-border'
                    }`}
                  >
                    {filters.statuses.includes(option.value) && (
                      <svg
                        className="w-3 h-3 text-white"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    )}
                  </div>
                  {option.label}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Type Filter Dropdown */}
        <div ref={typeRef} className="relative">
          <button
            onClick={() => setShowTypeDropdown(!showTypeDropdown)}
            className={`flex items-center gap-2 px-3 py-2 border rounded-lg transition-colors ${
              filters.types.length > 0
                ? 'bg-editor-accent/10 border-editor-accent/30 text-editor-accent'
                : 'bg-editor-surface border-editor-border text-editor-text hover:border-editor-accent/30'
            }`}
          >
            <span className="text-sm">
              Type
              {filters.types.length > 0 && (
                <span className="ml-1 px-1.5 py-0.5 bg-editor-accent/20 rounded text-xs">
                  {filters.types.length}
                </span>
              )}
            </span>
            <ChevronDown size={14} />
          </button>

          {showTypeDropdown && (
            <div className="absolute top-full left-0 mt-1 w-36 bg-editor-surface border border-editor-border rounded-lg shadow-lg z-10 py-1">
              {TYPE_OPTIONS.map((option) => (
                <button
                  key={option.value}
                  onClick={() => toggleType(option.value)}
                  className="w-full px-3 py-2 text-left text-sm text-editor-text hover:bg-editor-surface/80 flex items-center gap-2 transition-colors"
                >
                  <div
                    className={`w-4 h-4 border rounded flex items-center justify-center ${
                      filters.types.includes(option.value)
                        ? 'bg-editor-accent border-editor-accent'
                        : 'border-editor-border'
                    }`}
                  >
                    {filters.types.includes(option.value) && (
                      <svg
                        className="w-3 h-3 text-white"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    )}
                  </div>
                  {option.label}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Sort Dropdown */}
        <div ref={sortRef} className="relative">
          <button
            onClick={() => setShowSortDropdown(!showSortDropdown)}
            className="flex items-center gap-2 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text hover:border-editor-accent/30 transition-colors"
          >
            {filters.sortDirection === 'asc' ? <SortAsc size={16} /> : <SortDesc size={16} />}
            <span className="text-sm">
              {SORT_OPTIONS.find((s) => s.value === filters.sortField)?.label}
            </span>
            <ChevronDown size={14} />
          </button>

          {showSortDropdown && (
            <div className="absolute top-full right-0 mt-1 w-36 bg-editor-surface border border-editor-border rounded-lg shadow-lg z-10 py-1">
              {SORT_OPTIONS.map((option) => (
                <button
                  key={option.value}
                  onClick={() => setSort(option.value)}
                  className={`w-full px-3 py-2 text-left text-sm flex items-center justify-between transition-colors ${
                    filters.sortField === option.value
                      ? 'bg-editor-accent/10 text-editor-accent'
                      : 'text-editor-text hover:bg-editor-surface/80'
                  }`}
                >
                  {option.label}
                  {filters.sortField === option.value && (
                    <span className="text-xs text-editor-muted">
                      {filters.sortDirection === 'asc' ? 'Asc' : 'Desc'}
                    </span>
                  )}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Clear Filters */}
        {hasActiveFilters && (
          <button
            onClick={clearFilters}
            className="flex items-center gap-1 px-3 py-2 text-sm text-editor-muted hover:text-editor-error transition-colors"
          >
            <X size={14} />
            Clear ({getActiveFiltersCount()})
          </button>
        )}
      </div>

      {/* Results Count */}
      {resultCount !== undefined && (
        <div className="text-sm text-editor-muted">
          {resultCount} {resultCount === 1 ? 'result' : 'results'} found
        </div>
      )}
    </div>
  );
}
