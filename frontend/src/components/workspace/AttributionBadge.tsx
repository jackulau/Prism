import { User, Bot, Workflow, Wrench } from 'lucide-react'
import type { AgentType } from '../../store/sandboxStore'

export interface AttributionBadgeProps {
  agentName?: string
  agentType?: AgentType
  toolName?: string
  onClick?: () => void
  size?: 'sm' | 'md'
}

const agentTypeConfig: Record<AgentType, { icon: typeof Bot; color: string; bgColor: string }> = {
  assistant: {
    icon: User,
    color: 'text-blue-400',
    bgColor: 'bg-blue-500/10 border-blue-500/30 hover:border-blue-500/50',
  },
  autonomous: {
    icon: Bot,
    color: 'text-purple-400',
    bgColor: 'bg-purple-500/10 border-purple-500/30 hover:border-purple-500/50',
  },
  workflow: {
    icon: Workflow,
    color: 'text-green-400',
    bgColor: 'bg-green-500/10 border-green-500/30 hover:border-green-500/50',
  },
}

const toolOnlyConfig = {
  icon: Wrench,
  color: 'text-editor-muted',
  bgColor: 'bg-editor-surface border-editor-border hover:border-editor-accent/50',
}

export function AttributionBadge({
  agentName,
  agentType,
  toolName,
  onClick,
  size = 'sm',
}: AttributionBadgeProps) {
  const hasAgent = agentName && agentType
  const config = hasAgent ? agentTypeConfig[agentType] : toolOnlyConfig
  const Icon = config.icon

  const sizeClasses = size === 'sm'
    ? 'px-1.5 py-0.5 text-xs gap-1'
    : 'px-2 py-1 text-sm gap-1.5'

  const iconSize = size === 'sm' ? 12 : 14

  // Build label text
  let label = ''
  if (agentName && toolName) {
    label = `${agentName} • ${toolName}`
  } else if (agentName) {
    label = agentName
  } else if (toolName) {
    label = toolName
  }

  if (!label) {
    return null
  }

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    onClick?.()
  }

  return (
    <button
      onClick={onClick ? handleClick : undefined}
      className={`
        inline-flex items-center rounded-md border transition-colors
        ${config.bgColor}
        ${sizeClasses}
        ${onClick ? 'cursor-pointer' : 'cursor-default'}
      `}
      title={`${hasAgent ? `${agentType} agent: ${agentName}` : ''}${toolName ? `${hasAgent ? ' using ' : 'Tool: '}${toolName}` : ''}`}
    >
      <Icon size={iconSize} className={config.color} />
      <span className={`${config.color} max-w-[150px] truncate`}>
        {label}
      </span>
    </button>
  )
}
