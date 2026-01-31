import React from 'react';
import {
  ArrowLeft,
  Copy,
  Check,
  Clock,
  Timer,
  Layers,
  GitBranch,
} from 'lucide-react';
import { useState } from 'react';

export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
export type ExecutionType = 'batch' | 'swarm';

export interface ExecutionDetailData {
  id: string;
  status: ExecutionStatus;
  type: ExecutionType;
  startedAt: Date | null;
  completedAt: Date | null;
  name?: string;
  description?: string;
  agentCount?: number;
  taskCount?: number;
}

interface ExecutionDetailProps {
  execution: ExecutionDetailData;
  onBack: () => void;
}

const statusStyles: Record<ExecutionStatus, { bg: string; text: string; label: string }> = {
  pending: { bg: 'bg-editor-muted/10', text: 'text-editor-muted', label: 'Pending' },
  running: { bg: 'bg-editor-warning/10', text: 'text-editor-warning', label: 'Running' },
  completed: { bg: 'bg-editor-success/10', text: 'text-editor-success', label: 'Completed' },
  failed: { bg: 'bg-editor-error/10', text: 'text-editor-error', label: 'Failed' },
  cancelled: { bg: 'bg-editor-muted/10', text: 'text-editor-muted', label: 'Cancelled' },
};

const typeIcons: Record<ExecutionType, React.ReactNode> = {
  batch: <Layers className="w-5 h-5" />,
  swarm: <GitBranch className="w-5 h-5" />,
};

function formatDuration(startedAt: Date | null, completedAt: Date | null): string {
  if (!startedAt) return '--';
  const end = completedAt || new Date();
  const durationMs = end.getTime() - startedAt.getTime();

  if (durationMs < 1000) return `${durationMs}ms`;
  if (durationMs < 60000) return `${(durationMs / 1000).toFixed(1)}s`;
  if (durationMs < 3600000) {
    const minutes = Math.floor(durationMs / 60000);
    const seconds = Math.floor((durationMs % 60000) / 1000);
    return `${minutes}m ${seconds}s`;
  }
  const hours = Math.floor(durationMs / 3600000);
  const minutes = Math.floor((durationMs % 3600000) / 60000);
  return `${hours}h ${minutes}m`;
}

function formatTimestamp(date: Date | null): string {
  if (!date) return '--';
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

export const ExecutionDetail: React.FC<ExecutionDetailProps> = ({ execution, onBack }) => {
  const [copied, setCopied] = useState(false);
  const statusStyle = statusStyles[execution.status];

  const copyId = async () => {
    await navigator.clipboard.writeText(execution.id);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="space-y-4">
      {/* Back button */}
      <button
        onClick={onBack}
        className="flex items-center gap-2 text-editor-muted hover:text-editor-text transition-colors group"
      >
        <ArrowLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
        <span className="text-sm">Back to list</span>
      </button>

      {/* Header */}
      <div className="bg-editor-surface border border-editor-border rounded-lg p-6">
        <div className="flex items-start justify-between gap-4">
          {/* Left side: Type icon and main info */}
          <div className="flex items-start gap-4">
            {/* Type icon */}
            <div className="p-3 bg-editor-accent/10 rounded-lg text-editor-accent">
              {typeIcons[execution.type]}
            </div>

            {/* Main info */}
            <div className="space-y-2">
              {/* Name or type label */}
              <h2 className="text-xl font-semibold text-editor-text">
                {execution.name || `${execution.type.charAt(0).toUpperCase() + execution.type.slice(1)} Execution`}
              </h2>

              {/* ID with copy */}
              <div className="flex items-center gap-2">
                <code className="text-sm text-editor-muted font-mono bg-editor-bg px-2 py-1 rounded">
                  {execution.id}
                </code>
                <button
                  onClick={copyId}
                  className="p-1 text-editor-muted hover:text-editor-accent transition-colors"
                  title="Copy ID"
                >
                  {copied ? (
                    <Check className="w-4 h-4 text-editor-success" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </button>
              </div>

              {/* Description if present */}
              {execution.description && (
                <p className="text-sm text-editor-muted max-w-xl">
                  {execution.description}
                </p>
              )}
            </div>
          </div>

          {/* Right side: Status badge */}
          <div
            className={`px-4 py-2 rounded-full text-sm font-medium ${statusStyle.bg} ${statusStyle.text}`}
          >
            {statusStyle.label}
          </div>
        </div>

        {/* Timestamps row */}
        <div className="mt-6 pt-4 border-t border-editor-border grid grid-cols-3 gap-4">
          {/* Start time */}
          <div className="flex items-center gap-3">
            <Clock className="w-4 h-4 text-editor-muted" />
            <div>
              <p className="text-xs text-editor-muted uppercase tracking-wide">Started</p>
              <p className="text-sm text-editor-text">{formatTimestamp(execution.startedAt)}</p>
            </div>
          </div>

          {/* End time */}
          <div className="flex items-center gap-3">
            <Clock className="w-4 h-4 text-editor-muted" />
            <div>
              <p className="text-xs text-editor-muted uppercase tracking-wide">Completed</p>
              <p className="text-sm text-editor-text">
                {execution.status === 'running' ? 'In progress...' : formatTimestamp(execution.completedAt)}
              </p>
            </div>
          </div>

          {/* Duration */}
          <div className="flex items-center gap-3">
            <Timer className="w-4 h-4 text-editor-muted" />
            <div>
              <p className="text-xs text-editor-muted uppercase tracking-wide">Duration</p>
              <p className="text-sm text-editor-text font-medium">
                {formatDuration(execution.startedAt, execution.completedAt)}
              </p>
            </div>
          </div>
        </div>

        {/* Counts row (if applicable) */}
        {(execution.agentCount !== undefined || execution.taskCount !== undefined) && (
          <div className="mt-4 flex items-center gap-6 text-sm">
            {execution.agentCount !== undefined && (
              <div className="flex items-center gap-2">
                <span className="text-editor-muted">Agents:</span>
                <span className="text-editor-text font-medium">{execution.agentCount}</span>
              </div>
            )}
            {execution.taskCount !== undefined && (
              <div className="flex items-center gap-2">
                <span className="text-editor-muted">Tasks:</span>
                <span className="text-editor-text font-medium">{execution.taskCount}</span>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
