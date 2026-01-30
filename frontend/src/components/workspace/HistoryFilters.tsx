import { useState, useCallback } from 'react';
import { Search, Filter, X, Calendar } from 'lucide-react';
import type { HistoryFilters, FileHistoryOperation } from '../../types';

interface HistoryFiltersProps {
  onFilterChange: (filters: HistoryFilters) => void;
  availableFiles: string[];
  currentFilters: HistoryFilters;
}

const OPERATION_OPTIONS: { value: FileHistoryOperation; label: string; color: string }[] = [
  { value: 'create', label: 'Created', color: 'bg-green-500/20 text-green-400 border-green-500/30' },
  { value: 'update', label: 'Updated', color: 'bg-blue-500/20 text-blue-400 border-blue-500/30' },
  { value: 'delete', label: 'Deleted', color: 'bg-red-500/20 text-red-400 border-red-500/30' },
];

export function HistoryFilters({ onFilterChange, availableFiles, currentFilters }: HistoryFiltersProps) {
  const [showFilters, setShowFilters] = useState(false);
  const [searchQuery, setSearchQuery] = useState(currentFilters.searchQuery || '');

  const handleSearchChange = useCallback((value: string) => {
    setSearchQuery(value);
    onFilterChange({ ...currentFilters, searchQuery: value || undefined });
  }, [currentFilters, onFilterChange]);

  const handleOperationToggle = useCallback((operation: FileHistoryOperation) => {
    const current = currentFilters.operations || [];
    const newOperations = current.includes(operation)
      ? current.filter(op => op !== operation)
      : [...current, operation];

    onFilterChange({
      ...currentFilters,
      operations: newOperations.length > 0 ? newOperations : undefined,
    });
  }, [currentFilters, onFilterChange]);

  const handleFilePathChange = useCallback((filePath: string) => {
    onFilterChange({
      ...currentFilters,
      filePath: filePath || undefined,
    });
  }, [currentFilters, onFilterChange]);

  const handleClearFilters = useCallback(() => {
    setSearchQuery('');
    onFilterChange({});
  }, [onFilterChange]);

  const hasActiveFilters = !!(
    currentFilters.filePath ||
    currentFilters.operations?.length ||
    currentFilters.searchQuery ||
    currentFilters.dateRange
  );

  return (
    <div className="border-b border-editor-border bg-editor-surface/30">
      {/* Search bar */}
      <div className="flex items-center gap-2 p-2">
        <div className="flex-1 relative">
          <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-editor-muted" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => handleSearchChange(e.target.value)}
            placeholder="Search files..."
            className="w-full pl-7 pr-3 py-1.5 text-xs bg-editor-bg border border-editor-border rounded focus:outline-none focus:border-editor-accent text-editor-text placeholder:text-editor-muted"
          />
          {searchQuery && (
            <button
              onClick={() => handleSearchChange('')}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-editor-muted hover:text-editor-text"
            >
              <X size={12} />
            </button>
          )}
        </div>
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`p-1.5 rounded transition-colors ${
            showFilters || hasActiveFilters
              ? 'bg-editor-accent/20 text-editor-accent'
              : 'text-editor-muted hover:text-editor-text hover:bg-editor-border/50'
          }`}
          title="Toggle filters"
        >
          <Filter size={14} />
        </button>
        {hasActiveFilters && (
          <button
            onClick={handleClearFilters}
            className="p-1.5 rounded text-editor-muted hover:text-editor-error hover:bg-editor-error/10 transition-colors"
            title="Clear all filters"
          >
            <X size={14} />
          </button>
        )}
      </div>

      {/* Expanded filters */}
      {showFilters && (
        <div className="px-2 pb-2 space-y-2">
          {/* Operation type filter */}
          <div>
            <label className="text-xs text-editor-muted mb-1 block">Operation Type</label>
            <div className="flex gap-1 flex-wrap">
              {OPERATION_OPTIONS.map((op) => {
                const isSelected = currentFilters.operations?.includes(op.value);
                return (
                  <button
                    key={op.value}
                    onClick={() => handleOperationToggle(op.value)}
                    className={`px-2 py-1 text-xs rounded border transition-colors ${
                      isSelected
                        ? op.color
                        : 'bg-editor-bg border-editor-border text-editor-muted hover:text-editor-text hover:border-editor-text/30'
                    }`}
                  >
                    {op.label}
                  </button>
                );
              })}
            </div>
          </div>

          {/* File path filter */}
          {availableFiles.length > 0 && (
            <div>
              <label className="text-xs text-editor-muted mb-1 block">File</label>
              <select
                value={currentFilters.filePath || ''}
                onChange={(e) => handleFilePathChange(e.target.value)}
                className="w-full px-2 py-1.5 text-xs bg-editor-bg border border-editor-border rounded focus:outline-none focus:border-editor-accent text-editor-text"
              >
                <option value="">All files</option>
                {availableFiles.map((file) => (
                  <option key={file} value={file}>
                    {file.split('/').pop() || file}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Date range filter placeholder */}
          <div>
            <label className="text-xs text-editor-muted mb-1 block">Date Range</label>
            <div className="flex items-center gap-2">
              <button
                className="flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-xs bg-editor-bg border border-editor-border rounded text-editor-muted hover:text-editor-text hover:border-editor-text/30 transition-colors"
                disabled
                title="Date range filter coming soon"
              >
                <Calendar size={12} />
                <span>Select dates</span>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Active filter chips */}
      {hasActiveFilters && !showFilters && (
        <div className="px-2 pb-2 flex gap-1 flex-wrap">
          {currentFilters.operations?.map((op) => {
            const opConfig = OPERATION_OPTIONS.find(o => o.value === op);
            return (
              <span
                key={op}
                className={`inline-flex items-center gap-1 px-1.5 py-0.5 text-xs rounded ${opConfig?.color}`}
              >
                {opConfig?.label}
                <button
                  onClick={() => handleOperationToggle(op)}
                  className="hover:opacity-80"
                >
                  <X size={10} />
                </button>
              </span>
            );
          })}
          {currentFilters.filePath && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-xs rounded bg-editor-accent/20 text-editor-accent">
              {currentFilters.filePath.split('/').pop()}
              <button
                onClick={() => handleFilePathChange('')}
                className="hover:opacity-80"
              >
                <X size={10} />
              </button>
            </span>
          )}
        </div>
      )}
    </div>
  );
}
