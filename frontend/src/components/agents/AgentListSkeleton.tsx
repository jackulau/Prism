import { memo } from 'react';

interface AgentListSkeletonProps {
  count?: number;
}

function SkeletonCard() {
  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse">
      <div className="flex items-start gap-3">
        {/* Icon placeholder */}
        <div className="w-9 h-9 bg-editor-border rounded-lg flex-shrink-0" />

        <div className="flex-1 min-w-0">
          {/* Title */}
          <div className="h-4 bg-editor-border rounded w-3/4 mb-2" />

          {/* Status badges */}
          <div className="flex items-center gap-2">
            <div className="h-5 bg-editor-border rounded-full w-20" />
            <div className="h-5 bg-editor-border rounded w-16" />
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="mt-3 pt-3 border-t border-editor-border/50">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="h-3 bg-editor-border rounded w-12" />
            <div className="h-3 bg-editor-border rounded w-10" />
          </div>
          <div className="h-3 bg-editor-border rounded w-16" />
        </div>
      </div>
    </div>
  );
}

export const AgentListSkeleton = memo(function AgentListSkeleton({
  count = 6,
}: AgentListSkeletonProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      {Array.from({ length: count }).map((_, index) => (
        <SkeletonCard key={index} />
      ))}
    </div>
  );
});

export default AgentListSkeleton;
