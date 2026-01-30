import { Clock, Hash, Folder, AlertTriangle, Info } from 'lucide-react';
import type { ActivityEvent } from '../../types/monitoring';

interface ActivityEventDetailsProps {
  event: ActivityEvent;
}

function formatDuration(startTime: number, endTime?: number): string {
  const end = endTime || Date.now();
  const durationMs = end - startTime;

  if (durationMs < 1000) {
    return `${durationMs}ms`;
  }
  if (durationMs < 60000) {
    return `${(durationMs / 1000).toFixed(1)}s`;
  }
  const minutes = Math.floor(durationMs / 60000);
  const seconds = Math.floor((durationMs % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

function formatFullTimestamp(timestamp: number): string {
  const date = new Date(timestamp);
  return date.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

export function ActivityEventDetails({ event }: ActivityEventDetailsProps) {
  const hasMetadata = event.metadata && Object.keys(event.metadata).length > 0;
  const isError = event.severity === 'error';
  const errorMessage = event.metadata?.error as string | undefined;
  const errorStack = event.metadata?.stack as string | undefined;
  const duration = event.metadata?.duration as number | undefined;
  const startTime = event.metadata?.startTime as number | undefined;
  const endTime = event.metadata?.endTime as number | undefined;

  return (
    <div className="px-3 pb-3 space-y-3 bg-editor-bg/50 border-t border-editor-border/50">
      {/* Basic info */}
      <div className="grid grid-cols-2 gap-2 text-xs pt-3">
        {/* Event ID */}
        <div className="flex items-center gap-2 text-editor-muted">
          <Hash size={12} />
          <span>ID:</span>
          <code className="text-editor-text font-mono bg-editor-surface px-1 rounded">
            {event.id.slice(0, 12)}...
          </code>
        </div>

        {/* Timestamp */}
        <div className="flex items-center gap-2 text-editor-muted">
          <Clock size={12} />
          <span>{formatFullTimestamp(event.timestamp)}</span>
        </div>

        {/* Agent ID if present */}
        {event.agentId && (
          <div className="flex items-center gap-2 text-editor-muted">
            <Info size={12} />
            <span>Agent:</span>
            <code className="text-editor-text font-mono bg-editor-surface px-1 rounded">
              {event.agentId.slice(0, 12)}...
            </code>
          </div>
        )}

        {/* Workspace ID if present */}
        {event.workspaceId && (
          <div className="flex items-center gap-2 text-editor-muted">
            <Folder size={12} />
            <span>Workspace:</span>
            <code className="text-editor-text font-mono bg-editor-surface px-1 rounded">
              {event.workspaceId.slice(0, 12)}...
            </code>
          </div>
        )}
      </div>

      {/* Duration for completed events */}
      {(duration !== undefined || (startTime && endTime)) && (
        <div className="flex items-center gap-2 text-xs text-editor-muted">
          <Clock size={12} />
          <span>Duration:</span>
          <span className="text-editor-text font-medium">
            {duration !== undefined ? formatDuration(0, duration) : formatDuration(startTime!, endTime)}
          </span>
        </div>
      )}

      {/* Error details */}
      {isError && errorMessage && (
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-xs text-editor-error">
            <AlertTriangle size={12} />
            <span className="font-medium">Error Details</span>
          </div>
          <div className="bg-editor-error/10 border border-editor-error/20 rounded-lg p-2">
            <p className="text-xs text-editor-error font-mono">{errorMessage}</p>
            {errorStack && (
              <pre className="text-xs text-editor-error/70 font-mono mt-2 whitespace-pre-wrap overflow-x-auto max-h-32 overflow-y-auto">
                {errorStack}
              </pre>
            )}
          </div>
        </div>
      )}

      {/* Full metadata (excluding error fields already shown) */}
      {hasMetadata && (
        <div className="space-y-2">
          <span className="text-xs text-editor-muted font-medium">Metadata</span>
          <div className="bg-editor-surface border border-editor-border rounded-lg p-2 overflow-hidden">
            <pre className="text-xs text-editor-text font-mono whitespace-pre-wrap overflow-x-auto max-h-48 overflow-y-auto">
              {JSON.stringify(
                Object.fromEntries(
                  Object.entries(event.metadata || {}).filter(
                    ([key]) => !['error', 'stack'].includes(key)
                  )
                ),
                null,
                2
              )}
            </pre>
          </div>
        </div>
      )}

      {/* Event type badge */}
      <div className="flex items-center gap-2">
        <span className="text-xs text-editor-muted">Type:</span>
        <span className="text-xs font-mono bg-editor-surface px-2 py-0.5 rounded text-editor-text">
          {event.type}
        </span>
      </div>
    </div>
  );
}
