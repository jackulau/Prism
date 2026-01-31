import { useRef, useEffect } from 'react';
import { Square, Loader2, CheckCircle, XCircle, AlertCircle, Brain, Wrench, MessageSquare, Trash2 } from 'lucide-react';
import { useAgentStore, AgentOutput, AgentExecutionStatus } from '../../store/agentStore';

interface AgentExecutionPanelProps {
  onStop?: () => void;
}

export function AgentExecutionPanel({ onStop }: AgentExecutionPanelProps) {
  const { executionStatus, outputs, currentExecution, clearOutputs } = useAgentStore();
  const outputsEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new outputs arrive
  useEffect(() => {
    outputsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [outputs]);

  const isRunning = executionStatus === 'running';

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border bg-editor-surface/50">
        <div className="flex items-center gap-3">
          <StatusIndicator status={executionStatus} />
          <div>
            <h3 className="text-sm font-medium text-editor-text">
              {currentExecution?.config.name || 'Agent Execution'}
            </h3>
            <p className="text-xs text-editor-muted">
              {getStatusText(executionStatus)}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {isRunning && onStop && (
            <button
              onClick={onStop}
              className="flex items-center gap-2 px-3 py-1.5 bg-red-500/10 text-red-400 border border-red-500/20 rounded-lg hover:bg-red-500/20 transition-colors text-sm"
            >
              <Square size={14} />
              Stop
            </button>
          )}
          {!isRunning && outputs.length > 0 && (
            <button
              onClick={clearOutputs}
              className="flex items-center gap-2 px-3 py-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors text-sm"
            >
              <Trash2 size={14} />
              Clear
            </button>
          )}
        </div>
      </div>

      {/* Output Area */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {outputs.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-editor-muted">
            <MessageSquare size={48} className="mb-3 opacity-30" />
            <p className="text-sm">No output yet</p>
            <p className="text-xs mt-1">Configure and run an agent to see results</p>
          </div>
        ) : (
          <>
            {outputs.map((output) => (
              <OutputItem key={output.id} output={output} />
            ))}
            <div ref={outputsEndRef} />
          </>
        )}

        {/* Loading indicator when running */}
        {isRunning && (
          <div className="flex items-center gap-2 text-editor-muted">
            <Loader2 size={14} className="animate-spin" />
            <span className="text-sm">Processing...</span>
          </div>
        )}
      </div>

      {/* Footer with metrics */}
      {currentExecution && (
        <div className="px-4 py-2 border-t border-editor-border bg-editor-surface/30 text-xs text-editor-muted">
          <div className="flex items-center gap-4">
            <span>
              Started: {currentExecution.startedAt.toLocaleTimeString()}
            </span>
            {currentExecution.completedAt && (
              <span>
                Duration: {formatDuration(currentExecution.startedAt, currentExecution.completedAt)}
              </span>
            )}
            <span>
              Outputs: {outputs.length}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}

function StatusIndicator({ status }: { status: AgentExecutionStatus }) {
  switch (status) {
    case 'running':
      return (
        <div className="w-8 h-8 rounded-full bg-blue-500/10 flex items-center justify-center">
          <Loader2 size={16} className="text-blue-400 animate-spin" />
        </div>
      );
    case 'complete':
      return (
        <div className="w-8 h-8 rounded-full bg-green-500/10 flex items-center justify-center">
          <CheckCircle size={16} className="text-green-400" />
        </div>
      );
    case 'error':
      return (
        <div className="w-8 h-8 rounded-full bg-red-500/10 flex items-center justify-center">
          <XCircle size={16} className="text-red-400" />
        </div>
      );
    case 'stopped':
      return (
        <div className="w-8 h-8 rounded-full bg-yellow-500/10 flex items-center justify-center">
          <AlertCircle size={16} className="text-yellow-400" />
        </div>
      );
    default:
      return (
        <div className="w-8 h-8 rounded-full bg-editor-surface flex items-center justify-center">
          <MessageSquare size={16} className="text-editor-muted" />
        </div>
      );
  }
}

function OutputItem({ output }: { output: AgentOutput }) {
  const getIcon = () => {
    switch (output.type) {
      case 'thinking':
        return <Brain size={14} className="text-purple-400" />;
      case 'tool_call':
      case 'tool_result':
        return <Wrench size={14} className="text-yellow-400" />;
      case 'error':
        return <XCircle size={14} className="text-red-400" />;
      default:
        return <MessageSquare size={14} className="text-editor-muted" />;
    }
  };

  const getBackgroundClass = () => {
    switch (output.type) {
      case 'thinking':
        return 'bg-purple-500/5 border-purple-500/20';
      case 'tool_call':
        return 'bg-yellow-500/5 border-yellow-500/20';
      case 'tool_result':
        return 'bg-blue-500/5 border-blue-500/20';
      case 'error':
        return 'bg-red-500/5 border-red-500/20';
      default:
        return 'bg-editor-surface border-editor-border';
    }
  };

  return (
    <div className={`p-3 rounded-lg border ${getBackgroundClass()}`}>
      <div className="flex items-start gap-2">
        <div className="mt-0.5">{getIcon()}</div>
        <div className="flex-1 min-w-0">
          {output.type === 'tool_call' && output.toolName && (
            <div className="text-xs font-medium text-yellow-400 mb-1">
              Tool: {output.toolName}
            </div>
          )}
          <div className="text-sm text-editor-text whitespace-pre-wrap break-words">
            {output.content}
          </div>
          {output.toolParams && (
            <pre className="mt-2 text-xs text-editor-muted bg-editor-bg/50 p-2 rounded overflow-x-auto">
              {JSON.stringify(output.toolParams, null, 2)}
            </pre>
          )}
          {output.toolResult !== undefined && (
            <div className="mt-2">
              <span className="text-xs text-editor-muted">Result:</span>
              <pre className="mt-1 text-xs text-editor-text bg-editor-bg/50 p-2 rounded overflow-x-auto">
                {typeof output.toolResult === 'string'
                  ? output.toolResult
                  : JSON.stringify(output.toolResult, null, 2)}
              </pre>
            </div>
          )}
          <div className="text-xs text-editor-muted mt-2">
            {output.timestamp.toLocaleTimeString()}
          </div>
        </div>
      </div>
    </div>
  );
}

function getStatusText(status: AgentExecutionStatus): string {
  switch (status) {
    case 'running':
      return 'Agent is running...';
    case 'complete':
      return 'Execution completed';
    case 'error':
      return 'Execution failed';
    case 'stopped':
      return 'Execution stopped by user';
    default:
      return 'Ready to run';
  }
}

function formatDuration(start: Date, end: Date): string {
  const ms = end.getTime() - start.getTime();
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;

  if (minutes > 0) {
    return `${minutes}m ${remainingSeconds}s`;
  }
  return `${seconds}s`;
}
