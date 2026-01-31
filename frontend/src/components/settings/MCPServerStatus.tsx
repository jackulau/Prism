import { useEffect } from 'react';
import { CheckCircle, XCircle, AlertCircle, RefreshCw, Clock } from 'lucide-react';
import { useMCPServerStore } from '../../store/mcpServerStore';

interface MCPServerStatusProps {
  serverId: string;
  serverName: string;
  enabled: boolean;
  lastError?: string;
  compact?: boolean;
}

export function MCPServerStatus({
  serverId,
  serverName: _serverName,  // Available for debugging/logging
  enabled,
  lastError,
  compact = false,
}: MCPServerStatusProps) {
  const { serverStatuses, fetchServerStatus, reconnectServer, testServer } = useMCPServerStore();
  const status = serverStatuses[serverId];

  useEffect(() => {
    if (enabled) {
      fetchServerStatus(serverId);
    }
  }, [serverId, enabled, fetchServerStatus]);

  const handleReconnect = async () => {
    await reconnectServer(serverId);
  };

  const handleTest = async () => {
    await testServer(serverId);
  };

  // Determine the display status
  const getStatusDisplay = (): {
    icon: React.ReactNode;
    text: string;
    color: string;
    bgColor: string;
  } => {
    if (!enabled) {
      return {
        icon: <AlertCircle className="w-4 h-4" />,
        text: 'Disabled',
        color: 'text-editor-muted',
        bgColor: 'bg-editor-muted/10',
      };
    }

    if (!status) {
      return {
        icon: <Clock className="w-4 h-4 animate-pulse" />,
        text: 'Checking...',
        color: 'text-yellow-500',
        bgColor: 'bg-yellow-500/10',
      };
    }

    if (status.connected) {
      return {
        icon: <CheckCircle className="w-4 h-4" />,
        text: 'Connected',
        color: 'text-green-500',
        bgColor: 'bg-green-500/10',
      };
    }

    return {
      icon: <XCircle className="w-4 h-4" />,
      text: 'Disconnected',
      color: 'text-red-400',
      bgColor: 'bg-red-400/10',
    };
  };

  const { icon, text, color, bgColor } = getStatusDisplay();

  if (compact) {
    return (
      <div className={`flex items-center gap-1.5 px-2 py-1 rounded-full ${bgColor} ${color}`}>
        {icon}
        <span className="text-xs font-medium">{text}</span>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {/* Status indicator */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full ${bgColor} ${color}`}>
            {icon}
            <span className="text-sm font-medium">{text}</span>
          </div>
          {status?.latency_ms !== undefined && status.connected && (
            <span className="text-xs text-editor-muted">
              {status.latency_ms}ms
            </span>
          )}
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-2">
          <button
            onClick={handleTest}
            className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded transition-colors"
            title="Test connection"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
          {!status?.connected && enabled && (
            <button
              onClick={handleReconnect}
              className="px-3 py-1 text-xs bg-primary text-white rounded hover:bg-primary/90 transition-colors"
            >
              Reconnect
            </button>
          )}
        </div>
      </div>

      {/* Error message */}
      {(status?.error || lastError) && !status?.connected && (
        <div className="p-2 bg-red-400/10 border border-red-400/20 rounded-lg">
          <p className="text-xs text-red-400">
            {status?.error || lastError}
          </p>
        </div>
      )}

      {/* Last checked timestamp */}
      {status?.last_checked && (
        <p className="text-xs text-editor-muted">
          Last checked: {new Date(status.last_checked).toLocaleTimeString()}
        </p>
      )}
    </div>
  );
}

export default MCPServerStatus;
