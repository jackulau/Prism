import { useState } from 'react';
import { Copy, Check, Trash2, XCircle, Clock, Calendar } from 'lucide-react';
import { BuildStatusBadge } from './BuildStatusBadge';
import { ConfirmDialog } from '../ConfirmDialog';
import type { Build } from '../../services/buildHistory';

interface BuildDetailHeaderProps {
  build: Build;
  onCancel: () => void;
  onDelete: () => void;
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}m ${remainingSeconds}s`;
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleString();
}

export function BuildDetailHeader({ build, onCancel, onDelete }: BuildDetailHeaderProps) {
  const [copied, setCopied] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const handleCopyCommand = async () => {
    await navigator.clipboard.writeText(build.command);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const isRunning = build.status === 'running' || build.status === 'pending';

  return (
    <>
      <div className="border-b border-editor-border bg-editor-bg">
        {/* Command section */}
        <div className="px-4 py-3 border-b border-editor-border">
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-2">
                <span className="text-xs text-editor-muted font-medium uppercase tracking-wide">
                  Command
                </span>
                <button
                  onClick={handleCopyCommand}
                  className="p-1 rounded text-editor-muted hover:text-editor-text hover:bg-editor-surface transition-colors"
                  title="Copy command"
                >
                  {copied ? <Check size={14} className="text-editor-success" /> : <Copy size={14} />}
                </button>
              </div>
              <code className="block text-sm font-mono text-editor-text bg-editor-surface px-3 py-2 rounded overflow-x-auto">
                {build.command}
              </code>
            </div>
            <BuildStatusBadge status={build.status} size="lg" />
          </div>
        </div>

        {/* Metadata section */}
        <div className="px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-6 text-sm">
            {/* Duration */}
            {build.durationMs !== undefined && (
              <div className="flex items-center gap-1.5 text-editor-muted">
                <Clock size={14} />
                <span>{formatDuration(build.durationMs)}</span>
              </div>
            )}

            {/* Started at */}
            <div className="flex items-center gap-1.5 text-editor-muted">
              <Calendar size={14} />
              <span>{formatDate(build.startedAt)}</span>
            </div>

            {/* Exit code if completed */}
            {build.exitCode !== undefined && build.status !== 'running' && build.status !== 'pending' && (
              <div className={`flex items-center gap-1.5 ${build.exitCode === 0 ? 'text-editor-success' : 'text-editor-error'}`}>
                <span className="text-editor-muted">Exit:</span>
                <span className="font-mono">{build.exitCode}</span>
              </div>
            )}
          </div>

          {/* Actions */}
          <div className="flex items-center gap-2">
            {isRunning && (
              <button
                onClick={onCancel}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-warning hover:bg-editor-warning/10 rounded transition-colors"
              >
                <XCircle size={14} />
                Cancel
              </button>
            )}
            <button
              onClick={() => setShowDeleteConfirm(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-error hover:bg-editor-error/10 rounded transition-colors"
            >
              <Trash2 size={14} />
              Delete
            </button>
          </div>
        </div>
      </div>

      <ConfirmDialog
        isOpen={showDeleteConfirm}
        title="Delete Build"
        message="Are you sure you want to delete this build? This action cannot be undone."
        confirmText="Delete"
        variant="danger"
        onConfirm={() => {
          setShowDeleteConfirm(false);
          onDelete();
        }}
        onCancel={() => setShowDeleteConfirm(false)}
      />
    </>
  );
}
