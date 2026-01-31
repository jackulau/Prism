import { useEffect, useRef, useCallback } from 'react';
import { Clock, Loader2, FolderOpen } from 'lucide-react';
import { BuildStatusBadge } from './BuildStatusBadge';
import type { Build } from '../../services/buildHistory';

interface BuildHistoryListProps {
  builds: Build[];
  selectedBuildId: string | null;
  isLoading: boolean;
  hasMore: boolean;
  onSelectBuild: (build: Build) => void;
  onLoadMore: () => void;
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}m ${remainingSeconds}s`;
}

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSeconds < 60) return 'just now';
  if (diffMinutes < 60) return `${diffMinutes}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return date.toLocaleDateString();
}

function truncateCommand(command: string, maxLength: number = 50): string {
  if (command.length <= maxLength) return command;
  return command.slice(0, maxLength - 3) + '...';
}

export function BuildHistoryList({
  builds,
  selectedBuildId,
  isLoading,
  hasMore,
  onSelectBuild,
  onLoadMore,
}: BuildHistoryListProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const loadMoreRef = useRef<HTMLDivElement>(null);

  // Intersection observer for infinite scroll
  const handleIntersection = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      const [entry] = entries;
      if (entry.isIntersecting && hasMore && !isLoading) {
        onLoadMore();
      }
    },
    [hasMore, isLoading, onLoadMore]
  );

  useEffect(() => {
    const observer = new IntersectionObserver(handleIntersection, {
      root: containerRef.current,
      threshold: 0.1,
    });

    if (loadMoreRef.current) {
      observer.observe(loadMoreRef.current);
    }

    return () => observer.disconnect();
  }, [handleIntersection]);

  if (builds.length === 0 && !isLoading) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center text-editor-muted p-8">
        <FolderOpen size={48} className="mb-4 opacity-50" />
        <p className="text-lg mb-1">No builds yet</p>
        <p className="text-sm opacity-75">Run a build to see it here</p>
      </div>
    );
  }

  return (
    <div ref={containerRef} className="flex-1 overflow-auto">
      <div className="divide-y divide-editor-border">
        {builds.map((build) => (
          <button
            key={build.id}
            onClick={() => onSelectBuild(build)}
            className={`w-full text-left px-4 py-3 hover:bg-editor-surface/50 transition-colors ${
              selectedBuildId === build.id ? 'bg-editor-surface' : ''
            }`}
          >
            <div className="flex items-start justify-between gap-3">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <BuildStatusBadge status={build.status} size="sm" />
                  {build.durationMs !== undefined && (
                    <span className="text-xs text-editor-muted flex items-center gap-1">
                      <Clock size={10} />
                      {formatDuration(build.durationMs)}
                    </span>
                  )}
                </div>
                <code className="text-sm font-mono text-editor-text block truncate">
                  {truncateCommand(build.command)}
                </code>
              </div>
              <span className="text-xs text-editor-muted whitespace-nowrap">
                {formatRelativeTime(build.startedAt)}
              </span>
            </div>
          </button>
        ))}
      </div>

      {/* Load more trigger / Loading indicator */}
      <div ref={loadMoreRef} className="py-4 flex items-center justify-center">
        {isLoading && (
          <div className="flex items-center gap-2 text-editor-muted text-sm">
            <Loader2 className="animate-spin" size={16} />
            Loading...
          </div>
        )}
      </div>
    </div>
  );
}
