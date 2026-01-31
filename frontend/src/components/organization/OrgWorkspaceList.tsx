import { Briefcase, ChevronLeft, ChevronRight } from 'lucide-react';
import type { OrgWorkspace } from '../../types/organization';
import { OrgWorkspaceCard } from './OrgWorkspaceCard';

interface OrgWorkspaceListProps {
  workspaces: OrgWorkspace[];
  isLoading: boolean;
  total: number;
  hasMore: boolean;
  currentPage: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onEdit: (workspace: OrgWorkspace) => void;
  onDelete: (workspace: OrgWorkspace) => void;
  onOpen?: (workspace: OrgWorkspace) => void;
}

function LoadingSkeleton() {
  return (
    <div className="divide-y divide-editor-border">
      {[1, 2, 3].map((i) => (
        <div key={i} className="flex items-center gap-4 p-4 animate-pulse">
          <div className="w-10 h-10 bg-editor-surface rounded-lg" />
          <div className="flex-1 space-y-2">
            <div className="h-4 bg-editor-surface rounded w-1/3" />
            <div className="h-3 bg-editor-surface rounded w-1/2" />
          </div>
          <div className="flex gap-2">
            <div className="w-8 h-8 bg-editor-surface rounded" />
            <div className="w-8 h-8 bg-editor-surface rounded" />
          </div>
        </div>
      ))}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="p-8 text-center">
      <Briefcase className="w-12 h-12 text-editor-muted mx-auto mb-4" />
      <p className="text-editor-muted">No workspaces yet</p>
      <p className="text-sm text-editor-muted mt-1">
        Create a workspace to get started
      </p>
    </div>
  );
}

export function OrgWorkspaceList({
  workspaces,
  isLoading,
  total,
  hasMore,
  currentPage,
  pageSize,
  onPageChange,
  onEdit,
  onDelete,
  onOpen,
}: OrgWorkspaceListProps) {
  const totalPages = Math.ceil(total / pageSize);
  const canGoPrev = currentPage > 0;
  const canGoNext = hasMore || currentPage < totalPages - 1;

  if (isLoading && workspaces.length === 0) {
    return <LoadingSkeleton />;
  }

  if (workspaces.length === 0) {
    return <EmptyState />;
  }

  return (
    <div>
      <div className="divide-y divide-editor-border">
        {workspaces.map((workspace) => (
          <OrgWorkspaceCard
            key={workspace.id}
            workspace={workspace}
            onEdit={onEdit}
            onDelete={onDelete}
            onOpen={onOpen}
          />
        ))}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between px-4 py-3 border-t border-editor-border">
          <span className="text-sm text-editor-muted">
            Showing {currentPage * pageSize + 1}-
            {Math.min((currentPage + 1) * pageSize, total)} of {total}
          </span>

          <div className="flex items-center gap-2">
            <button
              onClick={() => onPageChange(currentPage - 1)}
              disabled={!canGoPrev}
              className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronLeft size={16} />
            </button>
            <span className="text-sm text-editor-text">
              Page {currentPage + 1} of {totalPages}
            </span>
            <button
              onClick={() => onPageChange(currentPage + 1)}
              disabled={!canGoNext}
              className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronRight size={16} />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
