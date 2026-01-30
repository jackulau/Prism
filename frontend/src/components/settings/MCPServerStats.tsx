import { useState, useEffect } from 'react';
import { BarChart3, CheckCircle, XCircle, Clock, Activity } from 'lucide-react';
import { useMCPServerStore } from '../../store/mcpServerStore';

interface MCPServerStatsProps {
  serverId: string;
}

type TimeRange = 'today' | 'week' | 'all';

export function MCPServerStats({ serverId }: MCPServerStatsProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>('all');
  const { serverStats, statsLoading, fetchServerStats } = useMCPServerStore();
  const stats = serverStats[serverId];
  const loading = statsLoading[serverId];

  useEffect(() => {
    fetchServerStats(serverId, timeRange);
  }, [serverId, timeRange, fetchServerStats]);

  const timeRangeOptions: { value: TimeRange; label: string }[] = [
    { value: 'today', label: 'Today' },
    { value: 'week', label: 'This Week' },
    { value: 'all', label: 'All Time' },
  ];

  if (loading) {
    return (
      <div className="flex items-center gap-2 p-4 text-editor-muted">
        <div className="animate-spin rounded-full h-4 w-4 border-2 border-editor-muted border-t-transparent" />
        <span className="text-sm">Loading statistics...</span>
      </div>
    );
  }

  // Default stats if none available
  const displayStats = stats || {
    total_calls: 0,
    successful_calls: 0,
    failed_calls: 0,
    average_response_ms: 0,
  };

  const successRate =
    displayStats.total_calls > 0
      ? Math.round((displayStats.successful_calls / displayStats.total_calls) * 100)
      : 0;

  return (
    <div className="space-y-4">
      {/* Header with time range selector */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-editor-muted">
          <BarChart3 className="w-4 h-4" />
          <span>Usage Statistics</span>
        </div>
        <div className="flex items-center gap-1 bg-editor-bg rounded-lg p-0.5">
          {timeRangeOptions.map((option) => (
            <button
              key={option.value}
              onClick={() => setTimeRange(option.value)}
              className={`px-2.5 py-1 text-xs rounded-md transition-colors ${
                timeRange === option.value
                  ? 'bg-editor-surface text-editor-text'
                  : 'text-editor-muted hover:text-editor-text'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-2 gap-3">
        {/* Total calls */}
        <StatCard
          icon={<Activity className="w-4 h-4 text-editor-accent" />}
          label="Total Calls"
          value={displayStats.total_calls.toLocaleString()}
        />

        {/* Success rate */}
        <StatCard
          icon={<CheckCircle className="w-4 h-4 text-green-500" />}
          label="Success Rate"
          value={`${successRate}%`}
          subvalue={`${displayStats.successful_calls.toLocaleString()} successful`}
        />

        {/* Failed calls */}
        <StatCard
          icon={<XCircle className="w-4 h-4 text-red-400" />}
          label="Failed Calls"
          value={displayStats.failed_calls.toLocaleString()}
        />

        {/* Average response time */}
        <StatCard
          icon={<Clock className="w-4 h-4 text-yellow-500" />}
          label="Avg Response"
          value={
            displayStats.average_response_ms > 0
              ? `${displayStats.average_response_ms.toFixed(0)}ms`
              : 'N/A'
          }
        />
      </div>

      {/* Simple bar visualization */}
      {displayStats.total_calls > 0 && (
        <div className="space-y-1">
          <div className="flex items-center justify-between text-xs text-editor-muted">
            <span>Success vs Failure</span>
            <span>
              {displayStats.successful_calls} / {displayStats.failed_calls}
            </span>
          </div>
          <div className="h-2 bg-editor-bg rounded-full overflow-hidden flex">
            <div
              className="h-full bg-green-500 transition-all duration-300"
              style={{
                width: `${successRate}%`,
              }}
            />
            <div
              className="h-full bg-red-400 transition-all duration-300"
              style={{
                width: `${100 - successRate}%`,
              }}
            />
          </div>
        </div>
      )}

      {/* Empty state */}
      {displayStats.total_calls === 0 && (
        <div className="text-center py-4 text-editor-muted">
          <p className="text-sm">No usage data available for this period</p>
        </div>
      )}
    </div>
  );
}

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string;
  subvalue?: string;
}

function StatCard({ icon, label, value, subvalue }: StatCardProps) {
  return (
    <div className="p-3 bg-editor-bg rounded-lg">
      <div className="flex items-center gap-2 mb-1">
        {icon}
        <span className="text-xs text-editor-muted">{label}</span>
      </div>
      <p className="text-lg font-semibold text-editor-text">{value}</p>
      {subvalue && <p className="text-xs text-editor-muted">{subvalue}</p>}
    </div>
  );
}

export default MCPServerStats;
