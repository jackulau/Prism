import { useEffect, useState } from 'react';
import { useAuditStore } from '../../store/auditStore';
import {
  LogIn,
  AlertTriangle,
  Key,
  Shield,
  Loader2,
} from 'lucide-react';

interface AuditStatsCardProps {
  className?: string;
}

export function AuditStatsCard({ className = '' }: AuditStatsCardProps) {
  const { stats, statsLoading, statsError, fetchStats } = useAuditStore();
  const [period, setPeriod] = useState('24h');

  useEffect(() => {
    fetchStats(period);
  }, [fetchStats, period]);

  const periods = [
    { value: '1h', label: '1 hour' },
    { value: '24h', label: '24 hours' },
    { value: '7d', label: '7 days' },
    { value: '30d', label: '30 days' },
  ];

  if (statsError) {
    return (
      <div className={`bg-editor-surface border border-editor-border rounded-lg p-4 ${className}`}>
        <div className="text-red-400 text-sm">
          Failed to load stats: {statsError}
        </div>
      </div>
    );
  }

  // Calculate stats
  const loginSuccess = stats?.auth_counts?.login || 0;
  const loginFailed = stats?.auth_counts?.login_failed || 0;
  const totalLogins = loginSuccess + loginFailed;
  const failureRate = totalLogins > 0 ? Math.round((loginFailed / totalLogins) * 100) : 0;

  const apiKeyCreated = stats?.provider_counts?.provider_key_set || 0;
  const apiKeyDeleted = stats?.provider_counts?.provider_key_deleted || 0;

  const mfaEnabled = stats?.auth_counts?.mfa_enabled || 0;
  const mfaDisabled = stats?.auth_counts?.mfa_disabled || 0;

  const totalEvents = Object.values(stats?.category_counts || {}).reduce((a, b) => a + b, 0);

  return (
    <div className={`bg-editor-surface border border-editor-border rounded-lg ${className}`}>
      <div className="p-4 border-b border-editor-border flex items-center justify-between">
        <h3 className="font-medium">Security Overview</h3>
        <select
          value={period}
          onChange={(e) => setPeriod(e.target.value)}
          className="text-xs bg-editor-bg border border-editor-border rounded px-2 py-1"
        >
          {periods.map((p) => (
            <option key={p.value} value={p.value}>
              Last {p.label}
            </option>
          ))}
        </select>
      </div>

      {statsLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-6 h-6 animate-spin text-editor-muted" />
        </div>
      ) : (
        <div className="p-4 grid grid-cols-2 lg:grid-cols-4 gap-4">
          {/* Login Attempts */}
          <StatItem
            icon={LogIn}
            label="Login Attempts"
            value={totalLogins}
            subValue={`${loginSuccess} success, ${loginFailed} failed`}
            trend={failureRate > 10 ? 'warning' : 'neutral'}
          />

          {/* Failure Rate */}
          <StatItem
            icon={AlertTriangle}
            label="Failure Rate"
            value={`${failureRate}%`}
            subValue="authentication failures"
            trend={failureRate > 20 ? 'bad' : failureRate > 10 ? 'warning' : 'good'}
          />

          {/* API Key Activity */}
          <StatItem
            icon={Key}
            label="API Key Changes"
            value={apiKeyCreated + apiKeyDeleted}
            subValue={`${apiKeyCreated} added, ${apiKeyDeleted} removed`}
            trend="neutral"
          />

          {/* MFA Activity */}
          <StatItem
            icon={Shield}
            label="MFA Activity"
            value={mfaEnabled + mfaDisabled}
            subValue={`${mfaEnabled} enabled, ${mfaDisabled} disabled`}
            trend={mfaEnabled > mfaDisabled ? 'good' : mfaEnabled < mfaDisabled ? 'warning' : 'neutral'}
          />
        </div>
      )}

      {/* Category breakdown */}
      {!statsLoading && stats && (
        <div className="px-4 pb-4 pt-0">
          <div className="border-t border-editor-border pt-4">
            <div className="flex items-center justify-between text-xs text-editor-muted mb-2">
              <span>Events by Category</span>
              <span>{totalEvents} total</span>
            </div>
            <div className="flex gap-1 h-2 rounded overflow-hidden bg-editor-bg">
              {Object.entries(stats.category_counts || {}).map(([category, count]) => {
                const percentage = totalEvents > 0 ? (count / totalEvents) * 100 : 0;
                return (
                  <div
                    key={category}
                    className={getCategoryColor(category)}
                    style={{ width: `${percentage}%` }}
                    title={`${category}: ${count} events (${Math.round(percentage)}%)`}
                  />
                );
              })}
            </div>
            <div className="flex flex-wrap gap-3 mt-2">
              {Object.entries(stats.category_counts || {}).map(([category, count]) => (
                <div key={category} className="flex items-center gap-1.5 text-xs">
                  <div className={`w-2 h-2 rounded-full ${getCategoryColor(category)}`} />
                  <span className="text-editor-muted capitalize">
                    {category.replace(/_/g, ' ')}: {count}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

interface StatItemProps {
  icon: typeof LogIn;
  label: string;
  value: number | string;
  subValue: string;
  trend: 'good' | 'warning' | 'bad' | 'neutral';
}

function StatItem({ icon: Icon, label, value, subValue, trend }: StatItemProps) {
  const trendColors = {
    good: 'text-green-400',
    warning: 'text-yellow-400',
    bad: 'text-red-400',
    neutral: 'text-editor-muted',
  };

  const trendBgColors = {
    good: 'bg-green-500/10',
    warning: 'bg-yellow-500/10',
    bad: 'bg-red-500/10',
    neutral: 'bg-editor-bg',
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <div className={`p-1.5 rounded ${trendBgColors[trend]}`}>
          <Icon className={`w-4 h-4 ${trendColors[trend]}`} />
        </div>
        <span className="text-xs text-editor-muted">{label}</span>
      </div>
      <div className={`text-2xl font-bold ${trendColors[trend]}`}>{value}</div>
      <div className="text-xs text-editor-muted">{subValue}</div>
    </div>
  );
}

function getCategoryColor(category: string): string {
  const colors: Record<string, string> = {
    authentication: 'bg-blue-500',
    api_key: 'bg-purple-500',
    session: 'bg-green-500',
    settings: 'bg-yellow-500',
    mfa: 'bg-cyan-500',
    provider: 'bg-orange-500',
    github: 'bg-gray-500',
  };
  return colors[category] || 'bg-editor-muted';
}

export default AuditStatsCard;
