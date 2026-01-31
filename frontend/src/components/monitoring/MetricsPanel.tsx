import { useEffect, useState, useRef } from 'react';
import { BarChart3, Users, Zap, Clock } from 'lucide-react';
import { useMonitoringMetrics } from '../../store/monitoringStore';

interface SparklineProps {
  data: number[];
  color: string;
  height?: number;
}

function Sparkline({ data, color, height = 24 }: SparklineProps) {
  if (data.length < 2) {
    return (
      <div
        className="flex items-center justify-center"
        style={{ height }}
        aria-label="Insufficient data for sparkline"
      >
        <div className="w-full h-0.5 bg-editor-border rounded" />
      </div>
    );
  }

  const max = Math.max(...data, 1);
  const min = Math.min(...data, 0);
  const range = max - min || 1;
  const width = 80;
  const stepX = width / (data.length - 1);

  const points = data
    .map((value, index) => {
      const x = index * stepX;
      const y = height - ((value - min) / range) * height;
      return `${x},${y}`;
    })
    .join(' ');

  return (
    <svg
      width={width}
      height={height}
      className="overflow-visible"
      role="img"
      aria-label="Metrics trend sparkline"
    >
      <polyline
        points={points}
        fill="none"
        stroke={color}
        strokeWidth={1.5}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

interface MetricCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  unit?: string;
  history: number[];
  color: string;
}

function MetricCard({ icon, label, value, unit, history, color }: MetricCardProps) {
  return (
    <div className="flex items-center justify-between p-3 bg-editor-bg rounded-lg">
      <div className="flex items-center gap-3">
        <div className="p-1.5 bg-editor-surface rounded-lg">{icon}</div>
        <div>
          <p className="text-xs text-editor-muted">{label}</p>
          <p className="text-lg font-semibold text-editor-text">
            {value}
            {unit && <span className="text-sm font-normal text-editor-muted ml-1">{unit}</span>}
          </p>
        </div>
      </div>
      <Sparkline data={history} color={color} />
    </div>
  );
}

const MAX_HISTORY_LENGTH = 12;

export function MetricsPanel() {
  const metrics = useMonitoringMetrics();
  const [connectionHistory, setConnectionHistory] = useState<number[]>([]);
  const [throughputHistory, setThroughputHistory] = useState<number[]>([]);
  const [latencyHistory, setLatencyHistory] = useState<number[]>([]);
  const prevMetricsRef = useRef(metrics);

  // Update history on metrics change
  useEffect(() => {
    const prev = prevMetricsRef.current;

    // Only update if metrics actually changed
    if (metrics.lastUpdated !== prev.lastUpdated) {
      setConnectionHistory((h) =>
        [...h, metrics.activeConnections].slice(-MAX_HISTORY_LENGTH)
      );
      setThroughputHistory((h) =>
        [...h, metrics.messagesThroughput].slice(-MAX_HISTORY_LENGTH)
      );
      setLatencyHistory((h) =>
        [...h, metrics.averageLatency].slice(-MAX_HISTORY_LENGTH)
      );
      prevMetricsRef.current = metrics;
    }
  }, [metrics]);

  // Initialize with current values
  useEffect(() => {
    setConnectionHistory([metrics.activeConnections]);
    setThroughputHistory([metrics.messagesThroughput]);
    setLatencyHistory([metrics.averageLatency]);
    // Only run once on mount
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const formatLatency = (ms: number): string => {
    if (ms < 1) return '<1';
    if (ms >= 1000) return `${(ms / 1000).toFixed(1)}s`;
    return Math.round(ms).toString();
  };

  const formatThroughput = (value: number): string => {
    if (value >= 1000) return `${(value / 1000).toFixed(1)}k`;
    return value.toFixed(1);
  };

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
      <div className="flex items-center gap-2 mb-4">
        <div className="p-1.5 bg-editor-warning/10 rounded-lg">
          <BarChart3 size={16} className="text-editor-warning" />
        </div>
        <h3 className="font-medium text-editor-text">Real-time Metrics</h3>
      </div>

      <div className="space-y-3">
        <MetricCard
          icon={<Users size={16} className="text-editor-accent" />}
          label="Active Connections"
          value={metrics.activeConnections}
          history={connectionHistory}
          color="var(--editor-accent, #60a5fa)"
        />

        <MetricCard
          icon={<Zap size={16} className="text-editor-success" />}
          label="Throughput"
          value={formatThroughput(metrics.messagesThroughput)}
          unit="msg/s"
          history={throughputHistory}
          color="var(--editor-success, #4ade80)"
        />

        <MetricCard
          icon={<Clock size={16} className="text-editor-warning" />}
          label="Avg Latency"
          value={formatLatency(metrics.averageLatency)}
          unit="ms"
          history={latencyHistory}
          color="var(--editor-warning, #facc15)"
        />
      </div>

      <div className="mt-4 pt-3 border-t border-editor-border">
        <div className="flex items-center justify-between text-xs text-editor-muted">
          <span>Messages: {metrics.messagesReceived} received / {metrics.messagesSent} sent</span>
          {metrics.errorsCount > 0 && (
            <span className="text-editor-error">{metrics.errorsCount} errors</span>
          )}
        </div>
      </div>
    </div>
  );
}
