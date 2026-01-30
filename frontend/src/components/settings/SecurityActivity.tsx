import { useEffect, useState } from 'react';
import { useAuditStore, type AuditLog } from '../../store/auditStore';
import {
  getEventIcon,
  getEventLabel,
  getEventColor,
  getEventBgColor,
  formatRelativeTime,
  parseUserAgent,
  formatAuditDetails,
  getStatusIndicator,
} from '../../utils/auditHelpers';
import { ChevronDown, Loader2 } from 'lucide-react';

export function SecurityActivity() {
  const {
    myLogs,
    myLogsLoading,
    myLogsError,
    myLogsHasMore,
    fetchMyLogs,
  } = useAuditStore();

  const [expandedLog, setExpandedLog] = useState<number | null>(null);

  useEffect(() => {
    fetchMyLogs(true);
  }, [fetchMyLogs]);

  const handleLoadMore = () => {
    if (!myLogsLoading && myLogsHasMore) {
      fetchMyLogs(false);
    }
  };

  if (myLogsError) {
    return (
      <div className="text-red-400 text-sm p-4 bg-red-500/10 rounded-lg">
        Failed to load security activity: {myLogsError}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <p className="text-editor-muted text-sm">
        View your recent security events including logins, API key changes, and other account activity.
      </p>

      {myLogs.length === 0 && !myLogsLoading && (
        <div className="text-editor-muted text-sm text-center py-8">
          No security activity recorded yet.
        </div>
      )}

      <div className="space-y-3">
        {myLogs.map((log) => (
          <ActivityItem
            key={log.id}
            log={log}
            isExpanded={expandedLog === log.id}
            onToggle={() => setExpandedLog(expandedLog === log.id ? null : log.id)}
          />
        ))}
      </div>

      {myLogsLoading && (
        <div className="flex items-center justify-center py-4">
          <Loader2 className="w-5 h-5 animate-spin text-editor-muted" />
        </div>
      )}

      {myLogsHasMore && !myLogsLoading && myLogs.length > 0 && (
        <button
          onClick={handleLoadMore}
          className="w-full py-2 text-sm text-editor-accent hover:text-editor-accent/80 transition-colors"
        >
          Load more activity
        </button>
      )}
    </div>
  );
}

interface ActivityItemProps {
  log: AuditLog;
  isExpanded: boolean;
  onToggle: () => void;
}

function ActivityItem({ log, isExpanded, onToggle }: ActivityItemProps) {
  const Icon = getEventIcon(log.event_type);
  const colorClass = getEventColor(log.success, log.event_type);
  const bgClass = getEventBgColor(log.success);
  const { browser, os } = parseUserAgent(log.user_agent);
  const { icon: StatusIcon, className: statusClassName } = getStatusIndicator(log.success);
  const details = formatAuditDetails(log.details);

  return (
    <div
      className={`rounded-lg border border-editor-border overflow-hidden transition-colors ${
        !log.success ? 'border-red-500/30 bg-red-500/5' : 'bg-editor-bg'
      }`}
    >
      <button
        onClick={onToggle}
        className="w-full p-3 flex items-start gap-3 text-left hover:bg-editor-surface/50 transition-colors"
      >
        <div className={`p-2 rounded-lg ${bgClass}`}>
          <Icon className={`w-4 h-4 ${colorClass}`} />
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className={`font-medium ${colorClass}`}>
              {getEventLabel(log.event_type)}
            </span>
            <StatusIcon className={`w-3.5 h-3.5 ${statusClassName}`} />
          </div>
          <div className="text-xs text-editor-muted mt-0.5">
            {formatRelativeTime(log.created_at)}
            {log.ip_address && ` • IP: ${log.ip_address}`}
          </div>
          {details && (
            <div className="text-xs text-editor-muted mt-1 truncate">
              {details}
            </div>
          )}
        </div>

        <ChevronDown
          className={`w-4 h-4 text-editor-muted transition-transform flex-shrink-0 ${
            isExpanded ? 'rotate-180' : ''
          }`}
        />
      </button>

      {isExpanded && (
        <div className="px-3 pb-3 pt-0 border-t border-editor-border">
          <dl className="grid grid-cols-2 gap-2 text-xs mt-3">
            <div>
              <dt className="text-editor-muted">Time</dt>
              <dd className="text-editor-text">
                {new Date(log.created_at).toLocaleString()}
              </dd>
            </div>
            {log.ip_address && (
              <div>
                <dt className="text-editor-muted">IP Address</dt>
                <dd className="text-editor-text font-mono">{log.ip_address}</dd>
              </div>
            )}
            {log.user_agent && (
              <div>
                <dt className="text-editor-muted">Browser</dt>
                <dd className="text-editor-text">{browser} on {os}</dd>
              </div>
            )}
            <div>
              <dt className="text-editor-muted">Status</dt>
              <dd className={log.success ? 'text-green-400' : 'text-red-400'}>
                {log.success ? 'Success' : 'Failed'}
              </dd>
            </div>
            {log.action && (
              <div className="col-span-2">
                <dt className="text-editor-muted">Action</dt>
                <dd className="text-editor-text">{log.action}</dd>
              </div>
            )}
            {log.details && Object.keys(log.details).length > 0 && (
              <div className="col-span-2">
                <dt className="text-editor-muted">Details</dt>
                <dd className="text-editor-text font-mono text-[11px] bg-editor-surface p-2 rounded mt-1 overflow-x-auto">
                  {JSON.stringify(log.details, null, 2)}
                </dd>
              </div>
            )}
          </dl>
        </div>
      )}
    </div>
  );
}

export default SecurityActivity;
