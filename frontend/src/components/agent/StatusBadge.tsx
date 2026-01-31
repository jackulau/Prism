import React from 'react';
import {
  Clock,
  Loader2,
  Brain,
  CheckCircle,
  XCircle,
  Ban,
  type LucideIcon,
} from 'lucide-react';

export type AgentStatus =
  | 'pending'
  | 'running'
  | 'thinking'
  | 'completed'
  | 'failed'
  | 'cancelled';

export interface StatusBadgeProps {
  status: AgentStatus;
  size?: 'sm' | 'md';
  showIcon?: boolean;
  className?: string;
}

const statusConfig: Record<
  AgentStatus,
  {
    bg: string;
    text: string;
    border: string;
    label: string;
    Icon: LucideIcon;
    animate?: boolean;
  }
> = {
  pending: {
    bg: 'bg-gray-500/20',
    text: 'text-gray-400',
    border: 'border-gray-500/30',
    label: 'Pending',
    Icon: Clock,
  },
  running: {
    bg: 'bg-blue-500/20',
    text: 'text-blue-400',
    border: 'border-blue-500/30',
    label: 'Running',
    Icon: Loader2,
    animate: true,
  },
  thinking: {
    bg: 'bg-purple-500/20',
    text: 'text-purple-400',
    border: 'border-purple-500/30',
    label: 'Thinking',
    Icon: Brain,
    animate: true,
  },
  completed: {
    bg: 'bg-green-500/20',
    text: 'text-green-400',
    border: 'border-green-500/30',
    label: 'Complete',
    Icon: CheckCircle,
  },
  failed: {
    bg: 'bg-red-500/20',
    text: 'text-red-400',
    border: 'border-red-500/30',
    label: 'Failed',
    Icon: XCircle,
  },
  cancelled: {
    bg: 'bg-gray-500/20',
    text: 'text-gray-500',
    border: 'border-gray-500/30',
    label: 'Cancelled',
    Icon: Ban,
  },
};

const sizeConfig = {
  sm: {
    padding: 'px-1.5 py-0.5',
    text: 'text-xs',
    icon: 10,
    gap: 'gap-1',
  },
  md: {
    padding: 'px-2 py-1',
    text: 'text-xs',
    icon: 12,
    gap: 'gap-1.5',
  },
};

export const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  size = 'md',
  showIcon = true,
  className = '',
}) => {
  const config = statusConfig[status];
  const sizes = sizeConfig[size];
  const { Icon } = config;

  return (
    <span
      className={`
        inline-flex items-center ${sizes.gap} ${sizes.padding}
        rounded-full border font-medium
        ${config.bg} ${config.text} ${config.border}
        ${sizes.text} ${className}
      `}
    >
      {showIcon && (
        <Icon
          size={sizes.icon}
          className={config.animate ? 'animate-spin' : ''}
        />
      )}
      {config.label}
    </span>
  );
};

export default StatusBadge;
