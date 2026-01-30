import { useState } from 'react';
import {
  FileOutput,
  Copy,
  Check,
  AlertCircle,
  CheckCircle,
  Clock,
  Loader2,
  XCircle,
} from 'lucide-react';
import type { AgentStatus } from './types';

interface AgentResultsProps {
  result?: unknown;
  error?: string;
  status: AgentStatus;
}

const statusMessages: Record<
  AgentStatus,
  { icon: React.ReactNode; message: string; color: string }
> = {
  pending: {
    icon: <Clock size={48} className="text-editor-muted" />,
    message: 'Agent execution is pending...',
    color: 'text-editor-muted',
  },
  running: {
    icon: <Loader2 size={48} className="text-blue-400 animate-spin" />,
    message: 'Agent is currently running...',
    color: 'text-blue-400',
  },
  completed: {
    icon: <CheckCircle size={48} className="text-green-400" />,
    message: 'Agent completed successfully',
    color: 'text-green-400',
  },
  failed: {
    icon: <XCircle size={48} className="text-red-400" />,
    message: 'Agent execution failed',
    color: 'text-red-400',
  },
  cancelled: {
    icon: <AlertCircle size={48} className="text-yellow-400" />,
    message: 'Agent execution was cancelled',
    color: 'text-yellow-400',
  },
};

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={handleCopy}
      className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-editor-surface border border-editor-border rounded-lg hover:bg-editor-surface/80 transition-colors"
    >
      {copied ? (
        <>
          <Check size={14} className="text-editor-success" />
          <span className="text-editor-success">Copied!</span>
        </>
      ) : (
        <>
          <Copy size={14} className="text-editor-muted" />
          <span className="text-editor-text">Copy</span>
        </>
      )}
    </button>
  );
}

function ResultDisplay({ result }: { result: unknown }) {
  const isString = typeof result === 'string';
  const displayContent = isString
    ? result
    : JSON.stringify(result, null, 2);

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-editor-text flex items-center gap-2">
          <FileOutput size={16} className="text-editor-accent" />
          Final Result
        </h3>
        <CopyButton text={displayContent} />
      </div>
      <div className="bg-editor-surface rounded-lg border border-editor-border overflow-hidden">
        <pre className="p-4 text-sm font-mono text-editor-text whitespace-pre-wrap break-words overflow-x-auto max-h-96">
          {displayContent}
        </pre>
      </div>
    </div>
  );
}

function ErrorDisplay({ error }: { error: string }) {
  return (
    <div className="space-y-3">
      <h3 className="text-sm font-medium text-red-400 flex items-center gap-2">
        <AlertCircle size={16} />
        Error Details
      </h3>
      <div className="bg-red-500/10 border border-red-500/30 rounded-lg overflow-hidden">
        <div className="flex items-center justify-between px-4 py-2 bg-red-500/10 border-b border-red-500/20">
          <span className="text-xs text-red-400 font-medium">Error Message</span>
          <CopyButton text={error} />
        </div>
        <pre className="p-4 text-sm font-mono text-red-400 whitespace-pre-wrap break-words overflow-x-auto">
          {error}
        </pre>
      </div>
    </div>
  );
}

function StatusPlaceholder({ status }: { status: AgentStatus }) {
  const config = statusMessages[status];

  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      {config.icon}
      <h3 className={`text-lg font-medium mt-4 mb-2 ${config.color}`}>
        {config.message}
      </h3>
      {status === 'pending' && (
        <p className="text-sm text-editor-muted max-w-md">
          The agent execution has not started yet. Results will appear here once
          the agent begins processing.
        </p>
      )}
      {status === 'running' && (
        <p className="text-sm text-editor-muted max-w-md">
          The agent is actively processing your request. Results will appear
          here when complete.
        </p>
      )}
      {status === 'cancelled' && (
        <p className="text-sm text-editor-muted max-w-md">
          The agent execution was cancelled before completion. No final result
          is available.
        </p>
      )}
    </div>
  );
}

export function AgentResults({ result, error, status }: AgentResultsProps) {
  const hasResult = result !== undefined && result !== null;
  const hasError = !!error;

  // Show placeholder for non-completed statuses without results
  if (!hasResult && !hasError && status !== 'completed' && status !== 'failed') {
    return <StatusPlaceholder status={status} />;
  }

  return (
    <div className="space-y-6">
      {/* Status indicator */}
      <div
        className={`flex items-center gap-3 p-4 rounded-lg border ${
          status === 'completed'
            ? 'bg-green-500/10 border-green-500/30'
            : status === 'failed'
            ? 'bg-red-500/10 border-red-500/30'
            : 'bg-editor-surface border-editor-border'
        }`}
      >
        {statusMessages[status].icon}
        <div>
          <h3 className={`font-medium ${statusMessages[status].color}`}>
            {statusMessages[status].message}
          </h3>
          {status === 'completed' && hasResult && (
            <p className="text-sm text-editor-muted mt-0.5">
              View the final output below
            </p>
          )}
        </div>
      </div>

      {/* Error display */}
      {hasError && <ErrorDisplay error={error} />}

      {/* Result display */}
      {hasResult && <ResultDisplay result={result} />}

      {/* No result for completed/failed without data */}
      {!hasResult && !hasError && (status === 'completed' || status === 'failed') && (
        <div className="bg-editor-surface/50 rounded-lg p-6 text-center">
          <FileOutput className="w-12 h-12 text-editor-muted mx-auto mb-4" />
          <h3 className="text-lg font-medium text-editor-text mb-2">
            No result data
          </h3>
          <p className="text-sm text-editor-muted max-w-md mx-auto">
            {status === 'completed'
              ? 'The agent completed but did not produce a final result.'
              : 'The agent failed without providing detailed error information.'}
          </p>
        </div>
      )}
    </div>
  );
}

export default AgentResults;
