import { memo } from 'react'
import {
  Clock,
  Loader2,
  CheckCircle,
  XCircle,
  StopCircle,
} from 'lucide-react'

export type AgentStatus = 'idle' | 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'

interface AgentStatusBadgeProps {
  status: AgentStatus
  size?: 'sm' | 'md' | 'lg'
  showLabel?: boolean
}

const statusConfig: Record<AgentStatus, {
  color: string
  icon: typeof Clock
  spin: boolean
  label: string
}> = {
  idle: {
    color: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
    icon: Clock,
    spin: false,
    label: 'Idle',
  },
  pending: {
    color: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    icon: Clock,
    spin: false,
    label: 'Pending',
  },
  running: {
    color: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    icon: Loader2,
    spin: true,
    label: 'Running',
  },
  completed: {
    color: 'bg-green-500/20 text-green-400 border-green-500/30',
    icon: CheckCircle,
    spin: false,
    label: 'Completed',
  },
  failed: {
    color: 'bg-red-500/20 text-red-400 border-red-500/30',
    icon: XCircle,
    spin: false,
    label: 'Failed',
  },
  cancelled: {
    color: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
    icon: StopCircle,
    spin: false,
    label: 'Cancelled',
  },
}

const sizeConfig = {
  sm: {
    badge: 'px-1.5 py-0.5 text-xs',
    icon: 10,
    gap: 'gap-1',
  },
  md: {
    badge: 'px-2 py-1 text-xs',
    icon: 12,
    gap: 'gap-1.5',
  },
  lg: {
    badge: 'px-2.5 py-1.5 text-sm',
    icon: 14,
    gap: 'gap-2',
  },
}

export const AgentStatusBadge = memo(function AgentStatusBadge({
  status,
  size = 'md',
  showLabel = true,
}: AgentStatusBadgeProps) {
  const config = statusConfig[status]
  const sizes = sizeConfig[size]
  const StatusIcon = config.icon

  return (
    <span
      className={`inline-flex items-center ${sizes.gap} ${sizes.badge} rounded-full font-medium border ${config.color} transition-colors`}
      role="status"
      aria-label={config.label}
    >
      <StatusIcon
        size={sizes.icon}
        className={config.spin ? 'animate-spin' : ''}
        aria-hidden="true"
      />
      {showLabel && <span>{config.label}</span>}
    </span>
  )
})

export default AgentStatusBadge
