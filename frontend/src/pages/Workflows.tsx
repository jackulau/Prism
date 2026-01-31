import { useState, useMemo } from 'react';
import { Plus, Layers } from 'lucide-react';
import { WorkflowFilters } from '../components/workflow/WorkflowFilters';
import { WorkflowCard } from '../components/workflow/WorkflowCard';
import { useWorkflows, type WorkflowStatus } from '../hooks/useWorkflows';

const ITEMS_PER_PAGE = 20;

export default function Workflows() {
  const [status, setStatus] = useState<WorkflowStatus | undefined>(undefined);
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState<'created' | 'updated' | 'name' | 'status'>('created');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [page, setPage] = useState(0);

  const { data: workflows, isLoading, error } = useWorkflows({
    status,
    name: searchQuery || undefined,
    limit: ITEMS_PER_PAGE,
    offset: page * ITEMS_PER_PAGE,
  });

  // Client-side sorting (backend should handle this ideally)
  const sortedWorkflows = useMemo(() => {
    if (!workflows) return [];

    const sorted = [...workflows].sort((a, b) => {
      let comparison = 0;

      switch (sortBy) {
        case 'name':
          comparison = a.name.localeCompare(b.name);
          break;
        case 'status':
          comparison = a.status.localeCompare(b.status);
          break;
        case 'updated':
          comparison = new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime();
          break;
        case 'created':
        default:
          comparison = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
          break;
      }

      return sortOrder === 'asc' ? comparison : -comparison;
    });

    return sorted;
  }, [workflows, sortBy, sortOrder]);

  const handleSortChange = (
    newSortBy: 'created' | 'updated' | 'name' | 'status',
    order: 'asc' | 'desc'
  ) => {
    setSortBy(newSortBy);
    setSortOrder(order);
  };

  const runningCount = workflows?.filter((w) => w.status === 'running').length ?? 0;

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-editor-text">Workflows</h1>
              {runningCount > 0 && (
                <span className="px-2 py-0.5 text-xs font-medium rounded-full bg-editor-warning/10 text-editor-warning">
                  {runningCount} running
                </span>
              )}
            </div>
            <p className="text-editor-muted">
              View and manage your workflow executions
            </p>
          </div>
          <button
            onClick={() => {
              // TODO: Navigate to workflow designer or show create modal
            }}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            <Plus size={18} />
            New Workflow
          </button>
        </div>

        {/* Filters */}
        <WorkflowFilters
          status={status}
          onStatusChange={(s) => {
            setStatus(s);
            setPage(0);
          }}
          searchQuery={searchQuery}
          onSearchChange={(q) => {
            setSearchQuery(q);
            setPage(0);
          }}
          sortBy={sortBy}
          sortOrder={sortOrder}
          onSortChange={handleSortChange}
        />

        {/* Content */}
        {isLoading ? (
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
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
        ) : error ? (
          <div className="bg-editor-error/10 border border-editor-error/20 rounded-lg p-6 text-center">
            <p className="text-editor-error">Failed to load workflows</p>
            <p className="text-sm text-editor-muted mt-1">
              {error instanceof Error ? error.message : 'Unknown error'}
            </p>
          </div>
        ) : sortedWorkflows.length === 0 ? (
          <EmptyState
            hasFilters={!!status || !!searchQuery}
            onClearFilters={() => {
              setStatus(undefined);
              setSearchQuery('');
            }}
          />
        ) : (
          <>
            {/* Workflow List */}
            <div className="space-y-4">
              {sortedWorkflows.map((workflow) => (
                <WorkflowCard key={workflow.id} workflow={workflow} />
              ))}
            </div>

            {/* Pagination */}
            {workflows && workflows.length >= ITEMS_PER_PAGE && (
              <div className="flex items-center justify-center gap-2 pt-4">
                <button
                  onClick={() => setPage((p) => Math.max(0, p - 1))}
                  disabled={page === 0}
                  className="px-4 py-2 text-sm bg-editor-surface border border-editor-border rounded-lg text-editor-muted hover:text-editor-text disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  Previous
                </button>
                <span className="px-4 py-2 text-sm text-editor-muted">
                  Page {page + 1}
                </span>
                <button
                  onClick={() => setPage((p) => p + 1)}
                  disabled={workflows.length < ITEMS_PER_PAGE}
                  className="px-4 py-2 text-sm bg-editor-surface border border-editor-border rounded-lg text-editor-muted hover:text-editor-text disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  Next
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

interface EmptyStateProps {
  hasFilters: boolean;
  onClearFilters: () => void;
}

function EmptyState({ hasFilters, onClearFilters }: EmptyStateProps) {
  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-12 text-center">
      <Layers className="w-16 h-16 text-editor-muted mx-auto mb-4" />
      {hasFilters ? (
        <>
          <h3 className="text-lg font-medium text-editor-text mb-2">
            No workflows found
          </h3>
          <p className="text-editor-muted mb-4">
            Try adjusting your filters to find what you're looking for.
          </p>
          <button
            onClick={onClearFilters}
            className="px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            Clear filters
          </button>
        </>
      ) : (
        <>
          <h3 className="text-lg font-medium text-editor-text mb-2">
            No workflows yet
          </h3>
          <p className="text-editor-muted mb-4">
            Create your first workflow to automate AI agent tasks.
          </p>
          <button
            onClick={() => {
              // TODO: Navigate to workflow designer
            }}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors mx-auto"
          >
            <Plus size={18} />
            Create Workflow
          </button>
        </>
      )}
    </div>
  );
}
