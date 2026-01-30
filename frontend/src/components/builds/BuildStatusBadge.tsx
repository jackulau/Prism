import { CheckCircle, XCircle, Clock, Loader2, Ban } from 'lucide-react';
import type { BuildStatus } from '../../services/buildHistory';

interface BuildStatusBadgeProps {
  status: BuildStatus;
  size?: 'sm' | 'md' | 'lg';
}

const statusConfig: Record<BuildStatus, {
  icon: typeof CheckCircle;
  label: string;
  className: string;
  animate?: boolean;
}> = {
  pending: {
    icon: Clock,
    label: 'Pending',
    className: 'bg-editor-muted/20 text-editor-muted',
    animate: true,
  },
  running: {
    icon: Loader2,
    label: 'Running',
    className: 'bg-editor-accent/20 text-editor-accent',
    animate: true,
  },
  success: {
    icon: CheckCircle,
    label: 'Success',
    className: 'bg-editor-success/20 text-editor-success',
  },
  failed: {
    icon: XCircle,
    label: 'Failed',
    className: 'bg-editor-error/20 text-editor-error',
  },
  cancelled: {
    icon: Ban,
    label: 'Cancelled',
    className: 'bg-editor-warning/20 text-editor-warning',
  },
};

const sizeConfig = {
  sm: { padding: 'px-1.5 py-0.5', text: 'text-xs', icon: 12, gap: 'gap-1' },
  md: { padding: 'px-2 py-1', text: 'text-sm', icon: 14, gap: 'gap-1.5' },
  lg: { padding: 'px-2.5 py-1.5', text: 'text-sm', icon: 16, gap: 'gap-2' },
};

export function BuildStatusBadge({ status, size = 'md' }: BuildStatusBadgeProps) {
  const config = statusConfig[status];
  const sizes = sizeConfig[size];
  const Icon = config.icon;

  return (
    <span
      className={`inline-flex items-center ${sizes.gap} ${sizes.padding} ${sizes.text} font-medium rounded-full ${config.className}`}
    >
      <Icon
        size={sizes.icon}
        className={config.animate ? (status === 'running' ? 'animate-spin' : 'animate-pulse') : ''}
      />
      <span>{config.label}</span>
    </span>
  );
}
