import { Clock, CheckCircle, XCircle, AlertCircle, Trash2, ChevronDown, ChevronUp } from 'lucide-react';
import { useState } from 'react';
import { useAgentStore, AgentExecution, AgentExecutionStatus } from '../../store/agentStore';

export function RecentExecutions() {
  const { recentExecutions, clearHistory } = useAgentStore();
  const [isExpanded, setIsExpanded] = useState(true);
  const [selectedExecution, setSelectedExecution] = useState<string | null>(null);

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
              key={execution.id}
              execution={execution}
              isSelected={selectedExecution === execution.id}
              onSelect={() => setSelectedExecution(
                selectedExecution === execution.id ? null : execution.id
              )}
            />
          ))}
        </div>
      )}
    </div>
  );
}

interface ExecutionItemProps {
  execution: AgentExecution;
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
            {execution.config.name}
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
          {execution.outputs.length} outputs
        </div>
      </button>

      {/* Expanded Details */}
      {isSelected && (
        <div className="mt-1 ml-9 mr-2 p-3 bg-editor-surface/50 rounded-lg border border-editor-border">
          <div className="space-y-2 text-xs">
            <div className="flex items-center justify-between">
              <span className="text-editor-muted">Provider:</span>
              <span className="text-editor-text">{execution.config.provider}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-editor-muted">Model:</span>
              <span className="text-editor-text truncate ml-2 max-w-[150px]">
                {execution.config.model}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-editor-muted">Temperature:</span>
              <span className="text-editor-text">{execution.config.temperature}</span>
            </div>
            {execution.error && (
              <div className="mt-2 p-2 bg-red-500/10 border border-red-500/20 rounded text-red-400">
                Error: {execution.error}
              </div>
            )}
            {execution.config.systemPrompt && (
              <div className="mt-2">
                <span className="text-editor-muted block mb-1">System Prompt:</span>
                <p className="text-editor-text bg-editor-bg/50 p-2 rounded max-h-20 overflow-y-auto">
                  {execution.config.systemPrompt}
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function StatusIcon({ status }: { status: AgentExecutionStatus }) {
  switch (status) {
    case 'complete':
      return <CheckCircle size={14} className="text-green-400 flex-shrink-0" />;
    case 'error':
      return <XCircle size={14} className="text-red-400 flex-shrink-0" />;
    case 'stopped':
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
