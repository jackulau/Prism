import { memo } from 'react';
import type { LucideIcon } from 'lucide-react';
import {
  Code,
  Eye,
  TestTube,
  ClipboardList,
  Sparkles,
  Search,
  FileText,
  BarChart2,
  Bug,
  Bot,
  Loader2,
  CheckCircle,
  XCircle,
  Clock,
  Pause,
  Square,
  MoreVertical,
} from 'lucide-react';
import type { SwarmAgent, SwarmAgentStatus } from '../../store/swarmStore';

// ============================================================================
// Types
// ============================================================================

interface SwarmAgentCardProps {
  agent: SwarmAgent;
  onViewDetails?: (agentId: string) => void;
  onStop?: (agentId: string) => void;
  compact?: boolean;
}

// ============================================================================
// Role Icon Mapping
// ============================================================================

const roleIcons: Record<string, LucideIcon> = {
  coder: Code,
  reviewer: Eye,
  tester: TestTube,
  planner: ClipboardList,
  synthesizer: Sparkles,
  researcher: Search,
  writer: FileText,
  analyst: BarChart2,
  debugger: Bug,
  general: Bot,
};

const getRoleIcon = (role: string): LucideIcon => {
  return roleIcons[role.toLowerCase()] || Bot;
};

// ============================================================================
// Status Configuration
// ============================================================================

interface StatusConfig {
  bgColor: string;
  textColor: string;
  borderColor: string;
  icon: LucideIcon;
  spin?: boolean;
  label: string;
  dotColor: string;
}

const statusConfig: Record<SwarmAgentStatus, StatusConfig> = {
  idle: {
    bgColor: 'bg-editor-muted/10',
    textColor: 'text-editor-muted',
    borderColor: 'border-editor-muted/30',
    icon: Clock,
    label: 'Idle',
    dotColor: 'bg-editor-muted',
  },
  running: {
    bgColor: 'bg-blue-500/10',
    textColor: 'text-blue-400',
    borderColor: 'border-blue-500/30',
    icon: Loader2,
    spin: true,
    label: 'Working',
    dotColor: 'bg-blue-400',
  },
  paused: {
    bgColor: 'bg-yellow-500/10',
    textColor: 'text-yellow-400',
    borderColor: 'border-yellow-500/30',
    icon: Pause,
    label: 'Paused',
    dotColor: 'bg-yellow-400',
  },
  completed: {
    bgColor: 'bg-editor-success/10',
    textColor: 'text-editor-success',
    borderColor: 'border-editor-success/30',
    icon: CheckCircle,
    label: 'Complete',
    dotColor: 'bg-editor-success',
  },
  failed: {
    bgColor: 'bg-editor-error/10',
    textColor: 'text-editor-error',
    borderColor: 'border-editor-error/30',
    icon: XCircle,
    label: 'Failed',
    dotColor: 'bg-editor-error',
  },
  waiting: {
    bgColor: 'bg-purple-500/10',
    textColor: 'text-purple-400',
    borderColor: 'border-purple-500/30',
    icon: Clock,
    label: 'Waiting',
    dotColor: 'bg-purple-400',
  },
};

// ============================================================================
// Role Colors
// ============================================================================

const roleColors: Record<string, string> = {
  coder: 'bg-blue-500/20 text-blue-400',
  reviewer: 'bg-purple-500/20 text-purple-400',
  tester: 'bg-green-500/20 text-green-400',
  planner: 'bg-orange-500/20 text-orange-400',
  synthesizer: 'bg-pink-500/20 text-pink-400',
  researcher: 'bg-cyan-500/20 text-cyan-400',
  writer: 'bg-amber-500/20 text-amber-400',
  analyst: 'bg-indigo-500/20 text-indigo-400',
  debugger: 'bg-red-500/20 text-red-400',
  general: 'bg-editor-muted/20 text-editor-muted',
};

const getRoleColor = (role: string) => {
  return roleColors[role.toLowerCase()] || roleColors.general;
};

// ============================================================================
// Component
// ============================================================================

export const SwarmAgentCard = memo(function SwarmAgentCard({
  agent,
  onViewDetails,
  onStop,
  compact = false,
}: SwarmAgentCardProps) {
  const config = statusConfig[agent.status];
  const StatusIcon = config.icon;
  const RoleIcon = getRoleIcon(agent.role);

  const formatRoleName = (role: string) => {
    return role.charAt(0).toUpperCase() + role.slice(1).toLowerCase();
  };

  if (compact) {
    return (
      <div
        className={`
          flex items-center gap-2 p-2 rounded-lg border
          bg-editor-surface/50 border-editor-border
          hover:border-editor-border/80 transition-all
        `}
      >
        {/* Status dot */}
        <span className={`w-2 h-2 rounded-full ${config.dotColor} ${config.spin ? 'animate-pulse' : ''}`} />

        {/* Role icon */}
        <div className={`p-1 rounded ${getRoleColor(agent.role)}`}>
          <RoleIcon size={12} />
        </div>

        {/* Agent name */}
        <span className="text-xs text-editor-text truncate flex-1">{agent.name}</span>

        {/* Status */}
        <span className={`text-xs ${config.textColor}`}>{config.label}</span>
      </div>
    );
  }

  return (
    <div
      className={`
        rounded-lg border bg-editor-surface border-editor-border
        hover:border-editor-accent/30 transition-all
        overflow-hidden
      `}
    >
      {/* Header */}
      <div className="flex items-start justify-between p-4 pb-3">
        <div className="flex items-start gap-3">
          {/* Role icon */}
          <div className={`p-2 rounded-lg ${getRoleColor(agent.role)}`}>
            <RoleIcon size={20} />
          </div>

          {/* Agent info */}
          <div className="min-w-0">
            <h3 className="font-medium text-editor-text truncate">{agent.name}</h3>
            <p className="text-xs text-editor-muted mt-0.5">{formatRoleName(agent.role)}</p>
          </div>
        </div>

        {/* Status indicator */}
        <div
          className={`
            flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium
            ${config.bgColor} ${config.textColor} border ${config.borderColor}
          `}
        >
          <StatusIcon size={12} className={config.spin ? 'animate-spin' : ''} />
          {config.label}
        </div>
      </div>

      {/* Description/Activity */}
      {agent.description && (
        <div className="px-4 pb-3">
          <p className="text-sm text-editor-muted line-clamp-2">{agent.description}</p>
        </div>
      )}

      {/* Model info */}
      <div className="px-4 pb-3">
        <div className="flex items-center gap-2 text-xs text-editor-muted">
          <span className="px-1.5 py-0.5 bg-editor-bg rounded text-editor-muted/80">
            {agent.provider}
          </span>
          <span className="truncate">{agent.model}</span>
        </div>
      </div>

      {/* Actions */}
      {(onViewDetails || onStop) && (
        <div className="flex items-center gap-2 px-4 py-3 border-t border-editor-border bg-editor-bg/30">
          {onViewDetails && (
            <button
              onClick={() => onViewDetails(agent.id)}
              className="flex-1 px-3 py-1.5 text-xs text-editor-text bg-editor-surface border border-editor-border rounded-lg hover:border-editor-accent/50 hover:bg-editor-surface/80 transition-colors"
            >
              View Details
            </button>
          )}
          {onStop && agent.status === 'running' && (
            <button
              onClick={() => onStop(agent.id)}
              className="flex items-center justify-center gap-1 px-3 py-1.5 text-xs text-editor-error bg-editor-error/10 border border-editor-error/30 rounded-lg hover:bg-editor-error/20 transition-colors"
            >
              <Square size={10} />
              Stop
            </button>
          )}
          {!onStop && !onViewDetails && (
            <button className="p-1.5 text-editor-muted hover:text-editor-text transition-colors">
              <MoreVertical size={14} />
            </button>
          )}
        </div>
      )}
    </div>
  );
});

export default SwarmAgentCard;
