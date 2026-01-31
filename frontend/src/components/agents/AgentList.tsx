import { useState, useMemo, useCallback } from 'react';
import { Bot, RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react';
import { AgentCard } from './AgentCard';
import { AgentFilters } from './AgentFilters';
import { AgentListSkeleton } from './AgentListSkeleton';
import type { AgentExecution, AgentFiltersState } from './types';

interface AgentListProps {
  agents: AgentExecution[];
  isLoading?: boolean;
  error?: Error | null;
  onRefresh?: () => void;
  onAgentClick?: (agent: AgentExecution) => void;
  showFilters?: boolean;
  pageSize?: number;
  emptyMessage?: string;
  emptyDescription?: string;
  title?: string;
  className?: string;
}

const defaultFilters: AgentFiltersState = {
  status: 'all',
  model: 'all',
  dateRange: 'all',
  search: '',
};

function filterAgents(
  agents: AgentExecution[],
  filters: AgentFiltersState
): AgentExecution[] {
  return agents.filter((agent) => {
    // Status filter
    if (filters.status !== 'all' && agent.status !== filters.status) {
      return false;
    }

    // Model filter
    if (filters.model !== 'all' && agent.model !== filters.model) {
      return false;
    }

    // Date range filter
    if (filters.dateRange !== 'all') {
      const createdAt = new Date(agent.created_at);
      const now = new Date();
      const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());

      switch (filters.dateRange) {
        case 'today':
          if (createdAt < startOfToday) return false;
          break;
        case 'week': {
          const startOfWeek = new Date(startOfToday);
          startOfWeek.setDate(startOfWeek.getDate() - startOfWeek.getDay());
          if (createdAt < startOfWeek) return false;
          break;
        }
        case 'month': {
          const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
          if (createdAt < startOfMonth) return false;
          break;
        }
      }
    }

    // Search filter
    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      const nameMatch = (agent.name || '').toLowerCase().includes(searchLower);
      const taskMatch = (agent.task || '').toLowerCase().includes(searchLower);
      const modelMatch = (agent.model || '').toLowerCase().includes(searchLower);
      if (!nameMatch && !taskMatch && !modelMatch) {
        return false;
      }
    }

    return true;
  });
}

function getAvailableModels(agents: AgentExecution[]): string[] {
  const models = new Set<string>();
  agents.forEach((agent) => {
    if (agent.model) {
      models.add(agent.model);
    }
  });
  return Array.from(models).sort();
}

export function AgentList({
  agents,
  isLoading = false,
  error = null,
  onRefresh,
  onAgentClick,
  showFilters = true,
  pageSize = 12,
  emptyMessage = 'No agent executions',
  emptyDescription = 'Run an agent to see executions here',
  title,
  className = '',
}: AgentListProps) {
  const [filters, setFilters] = useState<AgentFiltersState>(defaultFilters);
  const [currentPage, setCurrentPage] = useState(1);

  const availableModels = useMemo(() => getAvailableModels(agents), [agents]);

  const filteredAgents = useMemo(
    () => filterAgents(agents, filters),
    [agents, filters]
  );

  const totalPages = Math.ceil(filteredAgents.length / pageSize);
  const startIndex = (currentPage - 1) * pageSize;
  const paginatedAgents = filteredAgents.slice(startIndex, startIndex + pageSize);

  const handleFiltersChange = useCallback((newFilters: AgentFiltersState) => {
    setFilters(newFilters);
    setCurrentPage(1);
  }, []);

  const goToPage = useCallback((page: number) => {
    setCurrentPage(Math.max(1, Math.min(page, totalPages)));
  }, [totalPages]);

  // Loading state
  if (isLoading) {
    return (
      <div className={`space-y-4 ${className}`}>
        {title && (
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-editor-text">{title}</h2>
          </div>
        )}
        {showFilters && (
          <AgentFilters
            filters={filters}
            onChange={handleFiltersChange}
            availableModels={availableModels}
          />
        )}
        <AgentListSkeleton count={pageSize > 6 ? 6 : pageSize} />
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className={`space-y-4 ${className}`}>
        {title && (
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-editor-text">{title}</h2>
          </div>
        )}
        <div className="bg-editor-surface border border-editor-error/20 rounded-lg p-6 text-center">
          <p className="text-editor-error">Failed to load agents</p>
          <p className="text-sm text-editor-muted mt-1">{error.message}</p>
          {onRefresh && (
            <button
              onClick={onRefresh}
              className="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-editor-accent/10 text-editor-accent rounded-lg hover:bg-editor-accent/20 transition-colors"
            >
              <RefreshCw size={16} />
              Try Again
            </button>
          )}
        </div>
      </div>
    );
  }

  // Empty state (no agents at all)
  if (agents.length === 0) {
    return (
      <div className={`space-y-4 ${className}`}>
        {title && (
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-editor-text">{title}</h2>
          </div>
        )}
        <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
          <Bot className="w-12 h-12 text-editor-muted mx-auto mb-4" />
          <h3 className="text-lg font-medium text-editor-text mb-2">{emptyMessage}</h3>
          <p className="text-editor-muted">{emptyDescription}</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Header */}
      {(title || onRefresh) && (
        <div className="flex items-center justify-between">
          {title && <h2 className="text-lg font-semibold text-editor-text">{title}</h2>}
          {onRefresh && (
            <button
              onClick={onRefresh}
              className="flex items-center gap-2 text-sm text-editor-muted hover:text-editor-text transition-colors"
            >
              <RefreshCw size={16} />
              Refresh
            </button>
          )}
        </div>
      )}

      {/* Filters */}
      {showFilters && (
        <AgentFilters
          filters={filters}
          onChange={handleFiltersChange}
          availableModels={availableModels}
        />
      )}

      {/* Empty filtered state */}
      {filteredAgents.length === 0 && (
        <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
          <Bot className="w-12 h-12 text-editor-muted mx-auto mb-4" />
          <h3 className="text-lg font-medium text-editor-text mb-2">No matching agents</h3>
          <p className="text-editor-muted">
            Try adjusting your filters to see more results
          </p>
        </div>
      )}

      {/* Agent Grid */}
      {filteredAgents.length > 0 && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {paginatedAgents.map((agent) => (
              <AgentCard key={agent.id} agent={agent} onClick={onAgentClick} />
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between pt-4 border-t border-editor-border">
              <p className="text-sm text-editor-muted">
                Showing {startIndex + 1}-{Math.min(startIndex + pageSize, filteredAgents.length)}{' '}
                of {filteredAgents.length} agents
              </p>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => goToPage(currentPage - 1)}
                  disabled={currentPage === 1}
                  className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <ChevronLeft size={18} />
                </button>
                <div className="flex items-center gap-1">
                  {Array.from({ length: Math.min(5, totalPages) }).map((_, i) => {
                    let pageNum: number;
                    if (totalPages <= 5) {
                      pageNum = i + 1;
                    } else if (currentPage <= 3) {
                      pageNum = i + 1;
                    } else if (currentPage >= totalPages - 2) {
                      pageNum = totalPages - 4 + i;
                    } else {
                      pageNum = currentPage - 2 + i;
                    }

                    return (
                      <button
                        key={pageNum}
                        onClick={() => goToPage(pageNum)}
                        className={`w-8 h-8 text-sm rounded-lg transition-colors ${
                          currentPage === pageNum
                            ? 'bg-editor-accent text-white'
                            : 'text-editor-muted hover:text-editor-text hover:bg-editor-bg'
                        }`}
                      >
                        {pageNum}
                      </button>
                    );
                  })}
                </div>
                <button
                  onClick={() => goToPage(currentPage + 1)}
                  disabled={currentPage === totalPages}
                  className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
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

export default AgentList;
