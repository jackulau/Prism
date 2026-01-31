import React from 'react';
import type { WorkflowStepStatus } from '../../types';

export interface WorkflowNodeData {
  id: string;
  name: string;
  type: string;
  description?: string;
  status?: WorkflowStepStatus;
  duration?: number;
  error?: string;
  retryCount?: number;
}

interface WorkflowNodeProps {
  data: WorkflowNodeData;
  isSelected?: boolean;
  onClick?: () => void;
  className?: string;
}

const NODE_TYPE_ICONS: Record<string, React.ReactNode> = {
  llm: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
    </svg>
  ),
  tool: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
  ),
  condition: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  loop: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
    </svg>
  ),
  transform: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
    </svg>
  ),
  input: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
    </svg>
  ),
  output: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
    </svg>
  ),
  default: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
    </svg>
  ),
};

const STATUS_RING_COLORS: Record<WorkflowStepStatus, string> = {
  pending: 'ring-gray-600',
  running: 'ring-blue-500',
  completed: 'ring-green-500',
  failed: 'ring-red-500',
  skipped: 'ring-gray-500',
  retrying: 'ring-yellow-500',
};

const STATUS_BG_COLORS: Record<WorkflowStepStatus, string> = {
  pending: 'bg-gray-800',
  running: 'bg-gray-800',
  completed: 'bg-gray-800',
  failed: 'bg-red-900/20',
  skipped: 'bg-gray-800/50',
  retrying: 'bg-yellow-900/20',
};

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

function StatusOverlay({ status, retryCount }: { status: WorkflowStepStatus; retryCount?: number }) {
  switch (status) {
    case 'running':
      return (
        <div className="absolute -top-1 -right-1">
          <div className="relative w-5 h-5">
            <div className="absolute inset-0 rounded-full border-2 border-blue-400 border-t-transparent animate-spin" />
            <div className="absolute inset-0 rounded-full bg-blue-400/20 animate-pulse" />
          </div>
        </div>
      );

    case 'completed':
      return (
        <div className="absolute -top-1 -right-1 w-5 h-5 bg-green-500 rounded-full flex items-center justify-center shadow-lg shadow-green-500/30">
          <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
          </svg>
        </div>
      );

    case 'failed':
      return (
        <div className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 rounded-full flex items-center justify-center shadow-lg shadow-red-500/30">
          <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
      );

    case 'skipped':
      return (
        <div className="absolute -top-1 -right-1 w-5 h-5 bg-gray-600 rounded-full flex items-center justify-center">
          <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M5 5l7 7-7 7" />
          </svg>
        </div>
      );

    case 'retrying':
      return (
        <div className="absolute -top-1 -right-1">
          <div className="w-5 h-5 bg-yellow-500 rounded-full flex items-center justify-center animate-spin shadow-lg shadow-yellow-500/30">
            <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </div>
          {retryCount !== undefined && retryCount > 0 && (
            <div className="absolute -bottom-1 -right-1 w-3.5 h-3.5 bg-yellow-600 rounded-full flex items-center justify-center text-[8px] text-white font-bold">
              {retryCount}
            </div>
          )}
        </div>
      );

    default:
      return null;
  }
}

export function WorkflowNode({
  data,
  isSelected = false,
  onClick,
  className = '',
}: WorkflowNodeProps) {
  const icon = NODE_TYPE_ICONS[data.type] || NODE_TYPE_ICONS.default;
  const status = data.status || 'pending';
  const ringColor = STATUS_RING_COLORS[status];
  const bgColor = STATUS_BG_COLORS[status];

  const isRunning = status === 'running';
  const hasError = status === 'failed' && data.error;

  return (
    <div
      className={`
        relative min-w-40 p-3 rounded-lg border transition-all cursor-pointer
        ${bgColor}
        ${isSelected ? 'border-blue-500 ring-2 ring-blue-500/30' : 'border-gray-700 hover:border-gray-600'}
        ${isRunning ? `ring-2 ${ringColor} ring-opacity-50` : ''}
        ${className}
      `}
      onClick={onClick}
    >
      {/* Running pulse effect */}
      {isRunning && (
        <div className="absolute inset-0 rounded-lg bg-blue-400/5 animate-pulse pointer-events-none" />
      )}

      {/* Status overlay */}
      <StatusOverlay status={status} retryCount={data.retryCount} />

      {/* Node content */}
      <div className="flex items-start gap-2">
        {/* Type icon */}
        <div className={`
          flex-shrink-0 p-1.5 rounded
          ${isRunning ? 'bg-blue-500/20 text-blue-400' : 'bg-gray-700 text-gray-400'}
        `}>
          {icon}
        </div>

        <div className="flex-1 min-w-0">
          {/* Name */}
          <div className="text-sm font-medium text-gray-200 truncate">
            {data.name}
          </div>

          {/* Type badge */}
          <div className="text-xs text-gray-500 mt-0.5">
            {data.type}
          </div>

          {/* Description */}
          {data.description && (
            <div className="text-xs text-gray-500 mt-1 line-clamp-2">
              {data.description}
            </div>
          )}
        </div>
      </div>

      {/* Duration badge */}
      {data.duration !== undefined && status !== 'running' && status !== 'pending' && (
        <div className="mt-2 text-xs text-gray-500 flex items-center gap-1">
          <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {formatDuration(data.duration)}
        </div>
      )}

      {/* Error message (truncated) */}
      {hasError && (
        <div className="mt-2 text-xs text-red-400 truncate" title={data.error}>
          {data.error}
        </div>
      )}
    </div>
  );
}

export default WorkflowNode;
