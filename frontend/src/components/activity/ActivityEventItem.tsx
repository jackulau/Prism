import { useState } from 'react';
import {
  Bot,
  Users,
  GitBranch,
  Wrench,
  Hammer,
  MessageSquare,
  AlertCircle,
  ChevronDown,
  ChevronRight,
  Wifi,
  WifiOff,
  Bell,
  Activity,
  User,
} from 'lucide-react';
import type { ActivityEvent, ActivityEventType, ActivityEventSeverity } from '../../types/monitoring';
import { ActivityEventDetails } from './ActivityEventDetails';

interface ActivityEventItemProps {
  event: ActivityEvent;
  isNew?: boolean;
}

type EventCategory = 'agent' | 'swarm' | 'workflow' | 'tool' | 'build' | 'chat' | 'connection' | 'notification' | 'system' | 'user' | 'error';

function getEventCategory(type: ActivityEventType): EventCategory {
  if (type.startsWith('agent.')) return 'agent';
  if (type.startsWith('swarm.')) return 'swarm';
  if (type.startsWith('build.')) return 'build';
  if (type.startsWith('connection.')) return 'connection';
  if (type.startsWith('notification.')) return 'notification';
  if (type.startsWith('system.')) return 'system';
  if (type.startsWith('user.')) return 'user';
  return 'system';
}

function getEventIcon(category: EventCategory) {
  switch (category) {
    case 'agent':
      return Bot;
    case 'swarm':
      return Users;
    case 'workflow':
      return GitBranch;
    case 'tool':
      return Wrench;
    case 'build':
      return Hammer;
    case 'chat':
      return MessageSquare;
    case 'connection':
      return Wifi;
    case 'notification':
      return Bell;
    case 'system':
      return Activity;
    case 'user':
      return User;
    case 'error':
      return AlertCircle;
    default:
      return Activity;
  }
}

function getCategoryColor(category: EventCategory): string {
  switch (category) {
    case 'agent':
      return 'text-blue-400 bg-blue-400/10';
    case 'swarm':
      return 'text-purple-400 bg-purple-400/10';
    case 'workflow':
      return 'text-green-400 bg-green-400/10';
    case 'tool':
      return 'text-orange-400 bg-orange-400/10';
    case 'build':
      return 'text-yellow-400 bg-yellow-400/10';
    case 'chat':
      return 'text-gray-400 bg-gray-400/10';
    case 'connection':
      return 'text-cyan-400 bg-cyan-400/10';
    case 'notification':
      return 'text-pink-400 bg-pink-400/10';
    case 'system':
      return 'text-editor-muted bg-editor-muted/10';
    case 'user':
      return 'text-indigo-400 bg-indigo-400/10';
    case 'error':
      return 'text-red-400 bg-red-400/10';
    default:
      return 'text-editor-muted bg-editor-muted/10';
  }
}

function getSeverityDotColor(severity: ActivityEventSeverity): string {
  switch (severity) {
    case 'success':
      return 'bg-editor-success';
    case 'warning':
      return 'bg-editor-warning';
    case 'error':
      return 'bg-editor-error';
    case 'info':
    default:
      return 'bg-editor-accent';
  }
}

function formatTimestamp(timestamp: number): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);

  if (diffSec < 60) {
    return 'just now';
  }
  if (diffMin < 60) {
    return `${diffMin}m ago`;
  }
  if (diffHour < 24) {
    return `${diffHour}h ago`;
  }

  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function formatPreciseTimestamp(timestamp: number): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

export function ActivityEventItem({ event, isNew = false }: ActivityEventItemProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const category = getEventCategory(event.type);
  const Icon = getEventIcon(category);
  const colorClass = getCategoryColor(category);
  const severityDotColor = getSeverityDotColor(event.severity);

  // Determine if event has error or needs special handling
  const isConnectionLost = event.type === 'connection.lost';
  const IconComponent = isConnectionLost ? WifiOff : Icon;

  return (
    <div
      className={`
        border-b border-editor-border last:border-b-0
        transition-all duration-300 ease-out
        ${isNew ? 'animate-slide-in-top bg-editor-accent/5' : ''}
      `}
    >
      <div
        className="flex items-start gap-3 p-3 hover:bg-editor-surface/50 cursor-pointer transition-colors"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        {/* Icon */}
        <div className={`p-1.5 rounded-lg shrink-0 ${colorClass}`}>
          <IconComponent size={14} />
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            {/* Severity indicator */}
            <div className={`w-1.5 h-1.5 rounded-full shrink-0 ${severityDotColor}`} />

            {/* Title */}
            <span className="font-medium text-sm text-editor-text truncate">
              {event.title}
            </span>
          </div>

          {/* Description */}
          {event.description && (
            <p className="text-xs text-editor-muted mt-0.5 line-clamp-1">
              {event.description}
            </p>
          )}
        </div>

        {/* Timestamp and expand */}
        <div className="flex items-center gap-2 shrink-0">
          <span
            className="text-xs text-editor-muted"
            title={formatPreciseTimestamp(event.timestamp)}
          >
            {formatTimestamp(event.timestamp)}
          </span>
          {isExpanded ? (
            <ChevronDown size={14} className="text-editor-muted" />
          ) : (
            <ChevronRight size={14} className="text-editor-muted" />
          )}
        </div>
      </div>

      {/* Expanded details */}
      {isExpanded && <ActivityEventDetails event={event} />}
    </div>
  );
}
