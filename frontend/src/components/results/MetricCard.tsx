import { TrendingUp, TrendingDown } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

export type MetricCardVariant = 'default' | 'success' | 'warning' | 'error';
export type TrendDirection = 'up' | 'down' | 'neutral';

export interface MetricCardProps {
  /** Lucide icon component to display */
  icon: LucideIcon;
  /** Card title/label */
  title: string;
  /** Primary value to display prominently */
  value: string | number;
  /** Unit suffix for the value (e.g., "tokens", "ms") */
  unit?: string;
  /** Secondary/comparison value (optional) */
  secondaryValue?: string;
  /** Trend direction for indicator */
  trend?: TrendDirection;
  /** Trend label (e.g., "+12% from last week") */
  trendLabel?: string;
  /** Color variant */
  variant?: MetricCardVariant;
  /** Loading state */
  loading?: boolean;
  /** Additional CSS classes */
  className?: string;
}

const variantStyles: Record<MetricCardVariant, { bg: string; text: string; iconBg: string }> = {
  default: {
    bg: 'bg-editor-surface border-editor-border',
    text: 'text-editor-text',
    iconBg: 'bg-editor-accent/10 text-editor-accent',
  },
  success: {
    bg: 'bg-editor-surface border-editor-success/30',
    text: 'text-editor-success',
    iconBg: 'bg-editor-success/10 text-editor-success',
  },
  warning: {
    bg: 'bg-editor-surface border-editor-warning/30',
    text: 'text-editor-warning',
    iconBg: 'bg-editor-warning/10 text-editor-warning',
  },
  error: {
    bg: 'bg-editor-surface border-editor-error/30',
    text: 'text-editor-error',
    iconBg: 'bg-editor-error/10 text-editor-error',
  },
};

export function MetricCardSkeleton({ className = '' }: { className?: string }) {
  return (
    <div
      className={`bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse ${className}`}
    >
      <div className="flex items-center gap-3 mb-3">
        <div className="w-10 h-10 bg-editor-border rounded-lg" />
        <div className="h-4 bg-editor-border rounded w-20" />
      </div>
      <div className="h-8 bg-editor-border rounded w-24 mb-2" />
      <div className="h-3 bg-editor-border rounded w-16" />
    </div>
  );
}

export function MetricCard({
  icon: Icon,
  title,
  value,
  unit,
  secondaryValue,
  trend,
  trendLabel,
  variant = 'default',
  loading = false,
  className = '',
}: MetricCardProps) {
  if (loading) {
    return <MetricCardSkeleton className={className} />;
  }

  const styles = variantStyles[variant];
  const TrendIcon = trend === 'up' ? TrendingUp : trend === 'down' ? TrendingDown : null;
  const trendColor =
    trend === 'up'
      ? 'text-editor-success'
      : trend === 'down'
        ? 'text-editor-error'
        : 'text-editor-muted';

  return (
    <div className={`${styles.bg} border rounded-lg p-4 ${className}`}>
      <div className="flex items-center gap-3 mb-3">
        <div className={`p-2.5 rounded-lg ${styles.iconBg}`}>
          <Icon size={18} />
        </div>
        <span className="text-sm font-medium text-editor-muted uppercase tracking-wide">
          {title}
        </span>
      </div>

      <div className="flex items-baseline gap-1.5 mb-1">
        <span className={`text-2xl font-semibold ${styles.text}`}>{value}</span>
        {unit && <span className="text-sm text-editor-muted">{unit}</span>}
      </div>

      {(secondaryValue || trend) && (
        <div className="flex items-center gap-2 mt-2">
          {secondaryValue && (
            <span className="text-xs text-editor-muted">{secondaryValue}</span>
          )}
          {TrendIcon && (
            <div className={`flex items-center gap-1 ${trendColor}`}>
              <TrendIcon size={14} />
              {trendLabel && <span className="text-xs">{trendLabel}</span>}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default MetricCard;
