import { useEffect, useState } from 'react';
import { Activity, Wifi, WifiOff, Radio, Clock } from 'lucide-react';
import { useMonitoringStore, useSystemHealthStatus, useConnectionStatus } from '../../store/monitoringStore';

function formatTimeAgo(timestamp: number | null): string {
  if (!timestamp) return 'Never';
  const seconds = Math.floor((Date.now() - timestamp) / 1000);
  if (seconds < 5) return 'Just now';
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  return `${hours}h ago`;
}

interface HealthIndicatorProps {
  status: 'healthy' | 'degraded' | 'unhealthy';
}

function HealthIndicator({ status }: HealthIndicatorProps) {
  const colors = {
    healthy: 'bg-editor-success',
    degraded: 'bg-editor-warning',
    unhealthy: 'bg-editor-error',
  };

  const labels = {
    healthy: 'Healthy',
    degraded: 'Degraded',
    unhealthy: 'Unhealthy',
  };

  return (
    <div className="flex items-center gap-2">
      <span className={`relative flex h-3 w-3`}>
        <span
          className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${colors[status]}`}
        />
        <span
          className={`relative inline-flex rounded-full h-3 w-3 ${colors[status]}`}
        />
      </span>
      <span className="text-sm font-medium text-editor-text">{labels[status]}</span>
    </div>
  );
}

interface ConnectionBadgeProps {
  type: 'ws' | 'sse';
  status: 'connecting' | 'connected' | 'disconnected' | 'error';
}

function ConnectionBadge({ type, status }: ConnectionBadgeProps) {
  const label = type === 'ws' ? 'WebSocket' : 'SSE';

  const statusStyles = {
    connecting: 'bg-editor-warning/20 text-editor-warning border-editor-warning/30',
    connected: 'bg-editor-success/20 text-editor-success border-editor-success/30',
    disconnected: 'bg-editor-muted/20 text-editor-muted border-editor-border',
    error: 'bg-editor-error/20 text-editor-error border-editor-error/30',
  };

  const Icon = status === 'connected' ? Wifi : status === 'error' ? WifiOff : Radio;

  return (
    <div
      className={`flex items-center gap-1.5 px-2 py-1 rounded-full border text-xs font-medium ${statusStyles[status]}`}
    >
      <Icon size={12} />
      <span>{label}</span>
    </div>
  );
}

export function SystemStatusCard() {
  const systemHealth = useSystemHealthStatus();
  const connectionStatus = useConnectionStatus();
  const lastHeartbeat = useMonitoringStore((state) => state.lastHeartbeat);
  const [timeAgo, setTimeAgo] = useState(() => formatTimeAgo(lastHeartbeat));
  const [isPulsing, setIsPulsing] = useState(false);

  // Update time ago every second
  useEffect(() => {
    const interval = setInterval(() => {
      setTimeAgo(formatTimeAgo(lastHeartbeat));
    }, 1000);
    return () => clearInterval(interval);
  }, [lastHeartbeat]);

  // Pulse effect when heartbeat updates
  useEffect(() => {
    if (lastHeartbeat) {
      setIsPulsing(true);
      const timeout = setTimeout(() => setIsPulsing(false), 500);
      return () => clearTimeout(timeout);
    }
  }, [lastHeartbeat]);

  return (
    <div
      className={`bg-editor-surface border border-editor-border rounded-lg p-4 transition-all ${
        isPulsing ? 'ring-2 ring-editor-accent/50' : ''
      }`}
    >
      <div className="flex items-center gap-2 mb-4">
        <div className="p-1.5 bg-editor-accent/10 rounded-lg">
          <Activity size={16} className="text-editor-accent" />
        </div>
        <h3 className="font-medium text-editor-text">System Status</h3>
      </div>

      <div className="space-y-4">
        {/* Health Indicator */}
        <div className="flex items-center justify-between">
          <span className="text-sm text-editor-muted">Overall Health</span>
          <HealthIndicator status={systemHealth} />
        </div>

        {/* Connection Status */}
        <div className="space-y-2">
          <span className="text-sm text-editor-muted">Connections</span>
          <div className="flex flex-wrap gap-2">
            <ConnectionBadge type="ws" status={connectionStatus.ws} />
            <ConnectionBadge type="sse" status={connectionStatus.sse} />
          </div>
        </div>

        {/* Last Heartbeat */}
        <div className="flex items-center justify-between pt-2 border-t border-editor-border">
          <div className="flex items-center gap-1.5 text-sm text-editor-muted">
            <Clock size={14} />
            <span>Last heartbeat</span>
          </div>
          <span className="text-sm text-editor-text">{timeAgo}</span>
        </div>
      </div>
    </div>
  );
}
