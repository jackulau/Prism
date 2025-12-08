import React, { useEffect, useState } from 'react';
import {
  Zap,
  Clock,
  Hash,
  TrendingUp,
  Activity,
  Cpu,
  Timer,
  BarChart3,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
import { useAppStore } from '../store';

interface MetricCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  unit?: string;
  trend?: 'up' | 'down' | 'neutral';
  highlight?: boolean;
  subValue?: string;
}

const MetricCard: React.FC<MetricCardProps> = ({
  icon,
  label,
  value,
  unit,
  trend,
  highlight,
  subValue,
}) => (
  <div
    className={`p-3 rounded-lg border transition-all ${
      highlight
        ? 'bg-editor-accent/10 border-editor-accent/30'
        : 'bg-editor-surface border-editor-border'
    }`}
  >
    <div className="flex items-center gap-2 mb-1">
      <span className={highlight ? 'text-editor-accent' : 'text-editor-muted'}>
        {icon}
      </span>
      <span className="text-xs text-editor-muted uppercase tracking-wide">{label}</span>
      {trend && trend !== 'neutral' && (
        <span className={trend === 'up' ? 'text-editor-success' : 'text-editor-error'}>
          {trend === 'up' ? (
            <TrendingUp className="w-3 h-3" />
          ) : (
            <TrendingUp className="w-3 h-3 rotate-180" />
          )}
        </span>
      )}
    </div>
    <div className="flex items-baseline gap-1">
      <span
        className={`text-xl font-semibold ${
          highlight ? 'gradient-text' : 'text-editor-text'
        }`}
      >
        {value}
      </span>
      {unit && <span className="text-sm text-editor-muted">{unit}</span>}
    </div>
    {subValue && (
      <div className="mt-1 text-xs text-editor-muted">{subValue}</div>
    )}
  </div>
);

interface ProgressBarProps {
  value: number;
  max: number;
  label: string;
  color?: string;
}

const ProgressBar: React.FC<ProgressBarProps> = ({ value, max, label, color = 'bg-editor-accent' }) => {
  const percentage = max > 0 ? Math.min((value / max) * 100, 100) : 0;

  return (
    <div className="space-y-1">
      <div className="flex justify-between text-xs">
        <span className="text-editor-muted">{label}</span>
        <span className="text-editor-text">{value.toLocaleString()} / {max.toLocaleString()}</span>
      </div>
      <div className="h-2 bg-editor-surface rounded-full overflow-hidden">
        <div
          className={`h-full ${color} transition-all duration-300 rounded-full`}
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
};

const LiveIndicator: React.FC<{ active: boolean }> = ({ active }) => (
  <div className="flex items-center gap-2">
    <div
      className={`w-2 h-2 rounded-full ${
        active ? 'bg-editor-success animate-pulse' : 'bg-editor-muted'
      }`}
    />
    <span className={`text-xs ${active ? 'text-editor-success' : 'text-editor-muted'}`}>
      {active ? 'Generating' : 'Idle'}
    </span>
  </div>
);

export const MetricsPanel: React.FC = () => {
  const { metrics, isMetricsPanelOpen } = useAppStore();
  const [elapsedDisplay, setElapsedDisplay] = useState('0.0');
  const [isExpanded, setIsExpanded] = useState(true);

  // Update elapsed time display in real-time during generation
  useEffect(() => {
    if (!metrics.isGenerating || !metrics.startTime) {
      if (metrics.elapsedTime) {
        setElapsedDisplay((metrics.elapsedTime / 1000).toFixed(1));
      }
      return;
    }

    const interval = setInterval(() => {
      const elapsed = (performance.now() - metrics.startTime!) / 1000;
      setElapsedDisplay(elapsed.toFixed(1));
    }, 100);

    return () => clearInterval(interval);
  }, [metrics.isGenerating, metrics.startTime, metrics.elapsedTime]);

  if (!isMetricsPanelOpen) return null;

  const formatTime = (ms: number | null): string => {
    if (ms === null) return '--';
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const estimatedMaxTokens = 4096;

  return (
    <div className="bg-editor-bg border-l border-editor-border h-full flex flex-col">
      {/* Header */}
      <div
        className="flex items-center justify-between px-4 py-3 border-b border-editor-border cursor-pointer hover:bg-editor-surface/50"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center gap-2">
          <BarChart3 className="w-4 h-4 text-editor-accent" />
          <span className="text-sm font-medium text-editor-text">Generation Metrics</span>
        </div>
        <div className="flex items-center gap-3">
          <LiveIndicator active={metrics.isGenerating} />
          {isExpanded ? (
            <ChevronUp className="w-4 h-4 text-editor-muted" />
          ) : (
            <ChevronDown className="w-4 h-4 text-editor-muted" />
          )}
        </div>
      </div>

      {isExpanded && (
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {/* Primary Metrics */}
          <div className="grid grid-cols-2 gap-3">
            <MetricCard
              icon={<Zap className="w-4 h-4" />}
              label="Tokens/sec"
              value={metrics.tokensPerSecond.toFixed(1)}
              unit="t/s"
              highlight={metrics.isGenerating}
              trend={metrics.tokensPerSecond > 30 ? 'up' : 'neutral'}
            />
            <MetricCard
              icon={<Hash className="w-4 h-4" />}
              label="Total Tokens"
              value={metrics.tokenCount.toLocaleString()}
              subValue={`~${metrics.charCount.toLocaleString()} chars`}
            />
          </div>

          {/* Timing Metrics */}
          <div className="grid grid-cols-2 gap-3">
            <MetricCard
              icon={<Timer className="w-4 h-4" />}
              label="Time to First"
              value={formatTime(metrics.timeToFirstToken)}
              highlight={metrics.timeToFirstToken !== null && metrics.timeToFirstToken < 500}
            />
            <MetricCard
              icon={<Clock className="w-4 h-4" />}
              label="Elapsed"
              value={elapsedDisplay}
              unit="s"
            />
          </div>

          {/* Token Progress */}
          <div className="space-y-3">
            <h4 className="text-xs text-editor-muted uppercase tracking-wide flex items-center gap-2">
              <Activity className="w-3 h-3" />
              Generation Progress
            </h4>
            <ProgressBar
              value={metrics.tokenCount}
              max={estimatedMaxTokens}
              label="Token Usage"
              color="bg-editor-accent"
            />
          </div>

          {/* Detailed Stats */}
          <div className="space-y-3">
            <h4 className="text-xs text-editor-muted uppercase tracking-wide flex items-center gap-2">
              <Cpu className="w-3 h-3" />
              Detailed Statistics
            </h4>
            <div className="bg-editor-surface rounded-lg p-3 space-y-2 font-mono text-xs">
              <div className="flex justify-between">
                <span className="text-editor-muted">Start Time</span>
                <span className="text-editor-text">
                  {metrics.startTime
                    ? new Date(metrics.startTime).toLocaleTimeString()
                    : '--'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-editor-muted">First Token At</span>
                <span className="text-editor-text">
                  {metrics.firstTokenTime
                    ? new Date(metrics.firstTokenTime).toLocaleTimeString()
                    : '--'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-editor-muted">End Time</span>
                <span className="text-editor-text">
                  {metrics.endTime
                    ? new Date(metrics.endTime).toLocaleTimeString()
                    : metrics.isGenerating
                    ? 'In progress...'
                    : '--'}
                </span>
              </div>
              <div className="border-t border-editor-border my-2" />
              <div className="flex justify-between">
                <span className="text-editor-muted">Avg Token Size</span>
                <span className="text-editor-text">
                  {metrics.tokenCount > 0
                    ? `${(metrics.charCount / metrics.tokenCount).toFixed(2)} chars`
                    : '--'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-editor-muted">Est. Cost</span>
                <span className="text-editor-text">
                  {metrics.tokenCount > 0
                    ? `$${((metrics.tokenCount / 1000) * 0.002).toFixed(4)}`
                    : '--'}
                </span>
              </div>
            </div>
          </div>

          {/* Real-time Speed Graph Placeholder */}
          {metrics.isGenerating && (
            <div className="space-y-2">
              <h4 className="text-xs text-editor-muted uppercase tracking-wide">
                Live Speed
              </h4>
              <div className="h-16 bg-editor-surface rounded-lg flex items-end justify-center gap-1 p-2 overflow-hidden">
                {Array.from({ length: 20 }).map((_, i) => {
                  const height = Math.random() * 80 + 20;
                  return (
                    <div
                      key={i}
                      className="w-2 bg-editor-accent/60 rounded-t transition-all duration-150"
                      style={{
                        height: `${height}%`,
                        animationDelay: `${i * 50}ms`,
                      }}
                    />
                  );
                })}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
