import { useState } from 'react';
import { ChevronDown, ChevronUp, RefreshCw, MonitorDot } from 'lucide-react';
import { SystemStatusCard } from './SystemStatusCard';
import { ActiveAgentsPanel } from './ActiveAgentsPanel';
import { MetricsPanel } from './MetricsPanel';
import { useMonitoringStore } from '../../store/monitoringStore';

interface SystemStatusDashboardProps {
  defaultCollapsed?: boolean;
}

export function SystemStatusDashboard({ defaultCollapsed = false }: SystemStatusDashboardProps) {
  const [isCollapsed, setIsCollapsed] = useState(defaultCollapsed);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const recordHeartbeat = useMonitoringStore((state) => state.recordHeartbeat);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    // Trigger a heartbeat to simulate refresh
    recordHeartbeat();
    // Brief delay for visual feedback
    await new Promise((resolve) => setTimeout(resolve, 500));
    setIsRefreshing(false);
  };

  return (
    <section className="space-y-4" aria-label="System Status Dashboard">
      <div className="flex items-center justify-between">
        <button
          onClick={() => setIsCollapsed(!isCollapsed)}
          className="flex items-center gap-2 text-lg font-semibold text-editor-text hover:text-editor-accent transition-colors"
          aria-expanded={!isCollapsed}
          aria-controls="system-status-content"
        >
          <MonitorDot size={20} className="text-editor-accent" />
          <span>System Status</span>
          {isCollapsed ? <ChevronDown size={18} /> : <ChevronUp size={18} />}
        </button>
        {!isCollapsed && (
          <button
            onClick={handleRefresh}
            disabled={isRefreshing}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text bg-editor-surface border border-editor-border rounded-lg transition-colors disabled:opacity-50"
            aria-label="Refresh status"
          >
            <RefreshCw
              size={14}
              className={isRefreshing ? 'animate-spin' : ''}
            />
            <span>Refresh</span>
          </button>
        )}
      </div>

      {!isCollapsed && (
        <div
          id="system-status-content"
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
        >
          <SystemStatusCard />
          <ActiveAgentsPanel />
          <MetricsPanel />
        </div>
      )}
    </section>
  );
}
