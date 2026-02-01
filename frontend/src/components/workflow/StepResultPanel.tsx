import React, { useState } from 'react';
import type { WorkflowStepResult } from '../../types';

interface StepResultPanelProps {
  step: WorkflowStepResult;
  onClose?: () => void;
  className?: string;
}

const STATUS_BADGES: Record<WorkflowStepResult['status'], { color: string; label: string }> = {
  pending: { color: 'bg-gray-500/20 text-gray-400', label: 'Pending' },
  running: { color: 'bg-blue-500/20 text-blue-400', label: 'Running' },
  completed: { color: 'bg-green-500/20 text-green-400', label: 'Completed' },
  failed: { color: 'bg-red-500/20 text-red-400', label: 'Failed' },
  skipped: { color: 'bg-gray-500/20 text-gray-400', label: 'Skipped' },
  retrying: { color: 'bg-yellow-500/20 text-yellow-400', label: 'Retrying' },
};

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = ((ms % 60000) / 1000).toFixed(1);
  return `${minutes}m ${seconds}s`;
}

function formatTimestamp(timestamp: number | string): string {
  const date = typeof timestamp === 'string' ? new Date(timestamp) : new Date(timestamp);
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  });
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      console.error('Failed to copy to clipboard');
    }
  };

  return (
    <button
      onClick={handleCopy}
      className="p-1 rounded hover:bg-gray-700 text-gray-400 hover:text-gray-300 transition-colors"
      title={copied ? 'Copied!' : 'Copy to clipboard'}
    >
      {copied ? (
        <svg className="w-4 h-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
      ) : (
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
      )}
    </button>
  );
}

function CollapsibleSection({
  title,
  children,
  defaultOpen = false,
}: {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
}) {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  return (
    <div className="border border-gray-700 rounded-lg overflow-hidden">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-between px-3 py-2 bg-gray-800 hover:bg-gray-750 transition-colors text-left"
      >
        <span className="text-sm font-medium text-gray-300">{title}</span>
        <svg
          className={`w-4 h-4 text-gray-500 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {isOpen && (
        <div className="p-3 bg-gray-850">
          {children}
        </div>
      )}
    </div>
  );
}

export function StepResultPanel({ step, onClose, className = '' }: StepResultPanelProps) {
  const badge = STATUS_BADGES[step.status];
  const outputString = step.output !== undefined
    ? JSON.stringify(step.output, null, 2)
    : null;

  return (
    <div className={`bg-gray-900 border-l border-gray-800 flex flex-col ${className}`}>
      {/* Header */}
      <div className="flex-shrink-0 flex items-center justify-between p-4 border-b border-gray-800">
        <div className="flex-1 min-w-0">
          <h3 className="text-lg font-semibold text-gray-100 truncate">
            {step.stepName}
          </h3>
          <p className="text-xs text-gray-500 font-mono">{step.stepId}</p>
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-gray-800 text-gray-400 hover:text-gray-300 transition-colors ml-2"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Status and Type */}
        <div className="flex items-center gap-3">
          <span className={`px-2 py-1 rounded text-xs font-medium ${badge.color}`}>
            {badge.label}
          </span>
          <span className="px-2 py-1 rounded text-xs bg-gray-800 text-gray-400">
            {step.stepType}
          </span>
        </div>

        {/* Timing info */}
        <div className="grid grid-cols-2 gap-4 text-sm">
          {step.startedAt && (
            <div>
              <div className="text-gray-500 text-xs mb-1">Started</div>
              <div className="text-gray-300 font-mono text-xs">
                {formatTimestamp(step.startedAt)}
              </div>
            </div>
          )}
          {step.completedAt && (
            <div>
              <div className="text-gray-500 text-xs mb-1">Completed</div>
              <div className="text-gray-300 font-mono text-xs">
                {formatTimestamp(step.completedAt)}
              </div>
            </div>
          )}
          {step.duration !== undefined && (
            <div>
              <div className="text-gray-500 text-xs mb-1">Duration</div>
              <div className="text-gray-300 font-mono text-xs">
                {formatDuration(step.duration)}
              </div>
            </div>
          )}
          {step.retryCount !== undefined && step.retryCount > 0 && (
            <div>
              <div className="text-gray-500 text-xs mb-1">Retries</div>
              <div className="text-yellow-400 font-mono text-xs">
                {step.retryCount}
              </div>
            </div>
          )}
        </div>

        {/* Error message */}
        {step.error && (
          <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-3">
            <div className="flex items-center gap-2 mb-2">
              <svg className="w-4 h-4 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <span className="text-sm font-medium text-red-400">Error</span>
            </div>
            <pre className="text-xs text-red-300 whitespace-pre-wrap font-mono">
              {step.error}
            </pre>
          </div>
        )}

        {/* Output data */}
        {outputString && (
          <CollapsibleSection title="Output Data" defaultOpen>
            <div className="relative">
              <div className="absolute top-2 right-2">
                <CopyButton text={outputString} />
              </div>
              <pre className="bg-gray-800 rounded p-3 text-xs text-gray-300 font-mono overflow-x-auto max-h-64 overflow-y-auto">
                {outputString}
              </pre>
            </div>
          </CollapsibleSection>
        )}

        {/* Metadata section */}
        <CollapsibleSection title="Metadata">
          <div className="space-y-2 text-xs">
            <div className="flex justify-between">
              <span className="text-gray-500">Step ID</span>
              <span className="text-gray-300 font-mono">{step.stepId}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Step Type</span>
              <span className="text-gray-300">{step.stepType}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Status</span>
              <span className="text-gray-300">{step.status}</span>
            </div>
          </div>
        </CollapsibleSection>
      </div>
    </div>
  );
}

export default StepResultPanel;
