import { Clock, CheckCircle, XCircle, AlertCircle, Trash2, ChevronDown, ChevronUp } from 'lucide-react';
import { useState } from 'react';
import { useAgentStore } from '../../store/agentStore';
import type { AgentResult, AgentExecutionStatus } from '../../types/agent';

export function RecentExecutions() {
  const { executionHistory, clearHistory } = useAgentStore();
  const [isExpanded, setIsExpanded] = useState(true);
  const [selectedExecution, setSelectedExecution] = useState<string | null>(null);

  // Get recent executions (last 10, reversed for most recent first)
  const recentExecutions = executionHistory.slice(-10).reverse();

  if (recentExecutions.length === 0) {
    return null;
  }

  return (
    <div className="border-t border-editor-border">
      {/* Header */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between px-4 py-3 hover:bg-editor-surface/50 transition-colors"
      >
        <div className="flex items-center gap-2">
          <Clock size={14} className="text-editor-muted" />
          <span className="text-sm font-medium text-editor-text">Recent Executions</span>
          <span className="text-xs text-editor-muted">({recentExecutions.length})</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={(e) => {
              e.stopPropagation();
              clearHistory();
            }}
            className="p-1 text-editor-muted hover:text-red-400 rounded transition-colors"
            title="Clear history"
          >
            <Trash2 size={14} />
          </button>
          {isExpanded ? (
            <ChevronUp size={14} className="text-editor-muted" />
          ) : (
            <ChevronDown size={14} className="text-editor-muted" />
          )}
        </div>
      </button>

      {/* Execution List */}
      {isExpanded && (
        <div className="px-2 pb-2 space-y-1 max-h-64 overflow-y-auto">
          {recentExecutions.map((execution) => (
            <ExecutionItem
              key={execution.agentId}
              execution={execution}
              isSelected={selectedExecution === execution.agentId}
              onSelect={() => setSelectedExecution(
                selectedExecution === execution.agentId ? null : execution.agentId
              )}
            />
          ))}
        </div>
      )}
    </div>
  );
}

interface ExecutionItemProps {
  execution: AgentResult;
  isSelected: boolean;
  onSelect: () => void;
}

function ExecutionItem({ execution, isSelected, onSelect }: ExecutionItemProps) {
  return (
    <div>
      <button
        onClick={onSelect}
        className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-left transition-colors ${
          isSelected
            ? 'bg-editor-accent/10 border border-editor-accent/20'
            : 'hover:bg-editor-surface border border-transparent'
        }`}
      >
        <StatusIcon status={execution.status} />
        <div className="flex-1 min-w-0">
          <div className="text-sm text-editor-text truncate">
            Agent {execution.agentId.slice(0, 8)}...
          </div>
          <div className="text-xs text-editor-muted flex items-center gap-2">
            <span>{execution.startedAt.toLocaleString()}</span>
            {execution.completedAt && (
              <span>
                ({formatDuration(execution.startedAt, execution.completedAt)})
              </span>
            )}
          </div>
        </div>
        <div className="text-xs text-editor-muted">
          {execution.output.length} outputs
        </div>
      </button>

      {/* Expanded Details */}
      {isSelected && (
        <div className="mt-1 ml-9 mr-2 p-3 bg-editor-surface/50 rounded-lg border border-editor-border">
          <div className="space-y-2 text-xs">
            <div className="flex items-center justify-between">
              <span className="text-editor-muted">Iterations:</span>
              <span className="text-editor-text">{execution.iterationCount}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-editor-muted">Tool Calls:</span>
              <span className="text-editor-text">{execution.toolCalls.length}</span>
            </div>
            {execution.error && (
              <div className="mt-2 p-2 bg-red-500/10 border border-red-500/20 rounded text-red-400">
                Error: {execution.error}
              </div>
            )}
            {execution.output.length > 0 && (
              <div className="mt-2">
                <span className="text-editor-muted block mb-1">Output:</span>
                <p className="text-editor-text bg-editor-bg/50 p-2 rounded max-h-20 overflow-y-auto whitespace-pre-wrap">
                  {execution.output.join('\n').slice(0, 200)}
                  {execution.output.join('\n').length > 200 ? '...' : ''}
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function StatusIcon({ status }: { status: AgentResult['status'] }) {
  switch (status) {
    case 'completed':
      return <CheckCircle size={14} className="text-green-400 flex-shrink-0" />;
    case 'failed':
      return <XCircle size={14} className="text-red-400 flex-shrink-0" />;
    case 'cancelled':
      return <AlertCircle size={14} className="text-yellow-400 flex-shrink-0" />;
    default:
      return <Clock size={14} className="text-editor-muted flex-shrink-0" />;
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
