import { useState, useMemo, useCallback } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { ResultsFilter } from './ResultsFilter';
import { ResultsListItem } from './ResultsListItem';
import { ResultsEmpty } from './ResultsEmpty';
import {
  ExecutionResult,
  ResultsFilters,
  ResultsPagination,
  DEFAULT_FILTERS,
  DateRangePreset,
} from './types';

interface ResultsListProps {
  results: ExecutionResult[];
  isLoading?: boolean;
  selectedId?: string;
  onSelect?: (result: ExecutionResult) => void;
  onRunAgents?: () => void;
  pageSize?: number;
}

function ResultsListSkeleton() {
  return (
    <div className="space-y-4">
      {[1, 2, 3].map((i) => (
        <div
          key={i}
          className="bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse"
        >
          <div className="flex items-start gap-4">
            <div className="w-6 h-6 bg-editor-border rounded" />
            <div className="w-10 h-10 bg-editor-border rounded-lg" />
            <div className="flex-1 space-y-2">
              <div className="flex items-center gap-3">
                <div className="h-5 bg-editor-border rounded w-48" />
                <div className="h-5 bg-editor-border rounded w-16" />
              </div>
              <div className="h-4 bg-editor-border rounded w-32" />
            </div>
            <div className="flex items-center gap-6">
              <div className="h-4 bg-editor-border rounded w-20" />
              <div className="h-4 bg-editor-border rounded w-16" />
              <div className="h-4 bg-editor-border rounded w-16" />
              <div className="h-4 bg-editor-border rounded w-16" />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

function getDateRangeFilter(preset: DateRangePreset, customStart?: Date, customEnd?: Date) {
  const now = new Date();
  switch (preset) {
    case 'today':
      const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
      return { start: todayStart, end: now };
    case 'week':
      const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
      return { start: weekAgo, end: now };
    case 'month':
      const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
      return { start: monthAgo, end: now };
    case 'custom':
      return { start: customStart || now, end: customEnd || now };
    default:
      return { start: new Date(0), end: now };
  }
}

export function ResultsList({
  results,
  isLoading,
  selectedId,
  onSelect,
  onRunAgents,
  pageSize = 10,
}: ResultsListProps) {
  const [filters, setFilters] = useState<ResultsFilters>(DEFAULT_FILTERS);
  const [pagination, setPagination] = useState<ResultsPagination>({
    page: 1,
    pageSize,
    total: 0,
  });

  const filteredResults = useMemo(() => {
    let filtered = [...results];

    // Apply date range filter
    const dateRange = getDateRangeFilter(
      filters.dateRange,
      filters.customDateStart,
      filters.customDateEnd
    );
    filtered = filtered.filter(
      (r) => r.startedAt >= dateRange.start && r.startedAt <= dateRange.end
    );

    // Apply status filter
    if (filters.statuses.length > 0) {
      filtered = filtered.filter((r) => filters.statuses.includes(r.status));
    }

    // Apply type filter
    if (filters.types.length > 0) {
      filtered = filtered.filter((r) => filters.types.includes(r.type));
    }

    // Apply search filter
    if (filters.searchQuery) {
      const query = filters.searchQuery.toLowerCase();
      filtered = filtered.filter(
        (r) =>
          r.id.toLowerCase().includes(query) ||
          (r.name && r.name.toLowerCase().includes(query))
      );
    }

    // Apply sorting
    filtered.sort((a, b) => {
      let comparison = 0;
      switch (filters.sortField) {
        case 'date':
          comparison = a.startedAt.getTime() - b.startedAt.getTime();
          break;
        case 'duration':
          comparison = (a.durationMs || 0) - (b.durationMs || 0);
          break;
        case 'tokens':
          comparison = a.totalTokens - b.totalTokens;
          break;
        case 'status':
          comparison = a.status.localeCompare(b.status);
          break;
      }
      return filters.sortDirection === 'asc' ? comparison : -comparison;
    });

    return filtered;
  }, [results, filters]);

  const paginatedResults = useMemo(() => {
    const start = (pagination.page - 1) * pagination.pageSize;
    const end = start + pagination.pageSize;
    return filteredResults.slice(start, end);
  }, [filteredResults, pagination.page, pagination.pageSize]);

  const totalPages = Math.ceil(filteredResults.length / pagination.pageSize);

  const handleFiltersChange = useCallback((newFilters: ResultsFilters) => {
    setFilters(newFilters);
    setPagination((prev) => ({ ...prev, page: 1 }));
  }, []);

  const handleClearFilters = useCallback(() => {
    setFilters(DEFAULT_FILTERS);
    setPagination((prev) => ({ ...prev, page: 1 }));
  }, []);

  const goToPage = (page: number) => {
    if (page >= 1 && page <= totalPages) {
      setPagination((prev) => ({ ...prev, page }));
    }
  };

  const hasActiveFilters =
    filters.statuses.length > 0 ||
    filters.types.length > 0 ||
    filters.dateRange !== DEFAULT_FILTERS.dateRange ||
    filters.searchQuery !== '';

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-12 bg-editor-surface border border-editor-border rounded-lg animate-pulse" />
        <ResultsListSkeleton />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Filter Bar */}
      <ResultsFilter
        filters={filters}
        onFiltersChange={handleFiltersChange}
        resultCount={filteredResults.length}
      />

      {/* Results List or Empty State */}
      {filteredResults.length === 0 ? (
        <ResultsEmpty
          hasFilters={hasActiveFilters}
          onClearFilters={hasActiveFilters ? handleClearFilters : undefined}
          onRunAgents={onRunAgents}
        />
      ) : (
        <>
          {/* Results */}
          <div className="space-y-3">
            {paginatedResults.map((result) => (
              <ResultsListItem
                key={result.id}
                result={result}
                isSelected={result.id === selectedId}
                onClick={onSelect}
              />
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between pt-4 border-t border-editor-border">
              <div className="text-sm text-editor-muted">
                Showing {(pagination.page - 1) * pagination.pageSize + 1}-
                {Math.min(pagination.page * pagination.pageSize, filteredResults.length)} of{' '}
                {filteredResults.length}
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={() => goToPage(pagination.page - 1)}
                  disabled={pagination.page === 1}
                  className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface/80 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <ChevronLeft size={18} />
                </button>

                <div className="flex items-center gap-1">
                  {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                    let pageNum;
                    if (totalPages <= 5) {
                      pageNum = i + 1;
                    } else if (pagination.page <= 3) {
                      pageNum = i + 1;
                    } else if (pagination.page >= totalPages - 2) {
                      pageNum = totalPages - 4 + i;
                    } else {
                      pageNum = pagination.page - 2 + i;
                    }

                    return (
                      <button
                        key={pageNum}
                        onClick={() => goToPage(pageNum)}
                        className={`w-8 h-8 text-sm rounded-lg transition-colors ${
                          pagination.page === pageNum
                            ? 'bg-editor-accent text-white'
                            : 'text-editor-muted hover:text-editor-text hover:bg-editor-surface/80'
                        }`}
                      >
                        {pageNum}
                      </button>
                    );
                  })}
                </div>

                <button
                  onClick={() => goToPage(pagination.page + 1)}
                  disabled={pagination.page === totalPages}
                  className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface/80 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <ChevronRight size={18} />
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
