import { useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  User,
  Clock,
  Globe,
  Monitor,
  ExternalLink,
} from 'lucide-react';
import type { AuditLogEntry as AuditLogEntryType, AuditActionType } from '../../types/audit';

interface AuditLogEntryProps {
  entry: AuditLogEntryType;
  onResourceClick?: (resourceType: string, resourceId: string) => void;
}

const actionLabels: Record<AuditActionType, string> = {
  'user.login': 'User Login',
  'user.logout': 'User Logout',
  'user.create': 'User Created',
  'user.update': 'User Updated',
  'user.delete': 'User Deleted',
  'workspace.create': 'Workspace Created',
  'workspace.update': 'Workspace Updated',
  'workspace.delete': 'Workspace Deleted',
  'conversation.create': 'Conversation Created',
  'conversation.delete': 'Conversation Deleted',
  'api_key.create': 'API Key Created',
  'api_key.revoke': 'API Key Revoked',
  'settings.update': 'Settings Updated',
  'export.request': 'Export Requested',
  'export.download': 'Export Downloaded',
  'member.invite': 'Member Invited',
  'member.remove': 'Member Removed',
  'role.change': 'Role Changed',
};

const actionColors: Record<string, string> = {
  login: 'text-editor-success bg-editor-success/10',
  logout: 'text-editor-muted bg-editor-surface',
  create: 'text-editor-accent bg-editor-accent/10',
  update: 'text-editor-warning bg-editor-warning/10',
  delete: 'text-editor-error bg-editor-error/10',
  revoke: 'text-editor-error bg-editor-error/10',
  invite: 'text-editor-accent bg-editor-accent/10',
  remove: 'text-editor-error bg-editor-error/10',
  request: 'text-editor-accent bg-editor-accent/10',
  download: 'text-editor-success bg-editor-success/10',
  change: 'text-editor-warning bg-editor-warning/10',
};

function getActionColor(action: string): string {
  const actionPart = action.split('.')[1] || '';
  return actionColors[actionPart] || 'text-editor-text bg-editor-surface';
}

function formatTimestamp(date: Date): string {
  const d = new Date(date);
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

function formatRelativeTime(date: Date): string {
  const now = new Date();
  const d = new Date(date);
  const diffMs = now.getTime() - d.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return formatTimestamp(date);
}

export function AuditLogEntry({ entry, onResourceClick }: AuditLogEntryProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const hasDetails = entry.details && (entry.details.before || entry.details.after || entry.details.metadata);

  return (
    <div className="border border-editor-border rounded-lg bg-editor-surface overflow-hidden">
      {/* Main Row */}
      <div
        className={`flex items-center gap-4 p-4 ${hasDetails ? 'cursor-pointer hover:bg-editor-bg/50' : ''}`}
        onClick={() => hasDetails && setIsExpanded(!isExpanded)}
      >
        {/* Expand Icon */}
        <div className="w-5 flex-shrink-0">
          {hasDetails ? (
            isExpanded ? (
              <ChevronDown size={16} className="text-editor-muted" />
            ) : (
              <ChevronRight size={16} className="text-editor-muted" />
            )
          ) : null}
        </div>

        {/* Timestamp */}
        <div className="w-36 flex-shrink-0">
          <div className="flex items-center gap-1.5 text-sm text-editor-text">
            <Clock size={14} className="text-editor-muted" />
            <span title={formatTimestamp(entry.timestamp)}>
              {formatRelativeTime(entry.timestamp)}
            </span>
          </div>
        </div>

        {/* Action Badge */}
        <div className="w-40 flex-shrink-0">
          <span className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${getActionColor(entry.action)}`}>
            {actionLabels[entry.action] || entry.action}
          </span>
        </div>

        {/* Actor */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <User size={14} className="text-editor-muted flex-shrink-0" />
            <span className="text-sm text-editor-text truncate">
              {entry.actor.name || entry.actor.email}
            </span>
          </div>
        </div>

        {/* Resource */}
        <div className="flex-1 min-w-0">
          {entry.resource.id && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                onResourceClick?.(entry.resource.type, entry.resource.id);
              }}
              className="flex items-center gap-1.5 text-sm text-editor-accent hover:underline"
            >
              <span className="truncate">
                {entry.resource.name || `${entry.resource.type}:${entry.resource.id.slice(0, 8)}`}
              </span>
              <ExternalLink size={12} />
            </button>
          )}
        </div>

        {/* IP Address */}
        <div className="w-32 flex-shrink-0 text-right">
          {entry.ipAddress && (
            <div className="flex items-center gap-1.5 text-xs text-editor-muted justify-end">
              <Globe size={12} />
              <span>{entry.ipAddress}</span>
            </div>
          )}
        </div>
      </div>

      {/* Expanded Details */}
      {isExpanded && hasDetails && (
        <div className="border-t border-editor-border bg-editor-bg p-4 space-y-4">
          {/* Session Info */}
          {(entry.sessionId || entry.userAgent) && (
            <div className="space-y-2">
              <h4 className="text-xs font-medium text-editor-muted uppercase tracking-wide">
                Session Info
              </h4>
              <div className="grid grid-cols-2 gap-4 text-sm">
                {entry.sessionId && (
                  <div>
                    <span className="text-editor-muted">Session ID:</span>{' '}
                    <code className="text-editor-accent">{entry.sessionId.slice(0, 16)}...</code>
                  </div>
                )}
                {entry.userAgent && (
                  <div className="flex items-center gap-1.5">
                    <Monitor size={14} className="text-editor-muted" />
                    <span className="truncate text-editor-muted" title={entry.userAgent}>
                      {entry.userAgent.length > 50
                        ? `${entry.userAgent.slice(0, 50)}...`
                        : entry.userAgent}
                    </span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Before/After State */}
          {(entry.details?.before || entry.details?.after) && (
            <div className="grid grid-cols-2 gap-4">
              {entry.details.before && (
                <div className="space-y-2">
                  <h4 className="text-xs font-medium text-editor-muted uppercase tracking-wide">
                    Before
                  </h4>
                  <pre className="text-xs bg-editor-surface p-3 rounded-lg overflow-x-auto border border-editor-border">
                    {JSON.stringify(entry.details.before, null, 2)}
                  </pre>
                </div>
              )}
              {entry.details.after && (
                <div className="space-y-2">
                  <h4 className="text-xs font-medium text-editor-muted uppercase tracking-wide">
                    After
                  </h4>
                  <pre className="text-xs bg-editor-surface p-3 rounded-lg overflow-x-auto border border-editor-border">
                    {JSON.stringify(entry.details.after, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          )}

          {/* Additional Metadata */}
          {entry.details?.metadata && Object.keys(entry.details.metadata).length > 0 && (
            <div className="space-y-2">
              <h4 className="text-xs font-medium text-editor-muted uppercase tracking-wide">
                Additional Details
              </h4>
              <pre className="text-xs bg-editor-surface p-3 rounded-lg overflow-x-auto border border-editor-border">
                {JSON.stringify(entry.details.metadata, null, 2)}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
