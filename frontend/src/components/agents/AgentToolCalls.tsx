import { useState, memo } from 'react';
import {
  ChevronRight,
  ChevronDown,
  Wrench,
  Clock,
  Loader2,
  CheckCircle,
  XCircle,
  Copy,
  Check,
} from 'lucide-react';
import type { ToolCall } from '../../types';

interface AgentToolCallsProps {
  toolCalls: ToolCall[];
}

const statusConfig = {
  pending: {
    color: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    icon: Clock,
    spin: false,
    label: 'Pending',
  },
  running: {
    color: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    icon: Loader2,
    spin: true,
    label: 'Running',
  },
  completed: {
    color: 'bg-green-500/20 text-green-400 border-green-500/30',
    icon: CheckCircle,
    spin: false,
    label: 'Completed',
  },
  failed: {
    color: 'bg-red-500/20 text-red-400 border-red-500/30',
    icon: XCircle,
    spin: false,
    label: 'Failed',
  },
  rejected: {
    color: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
    icon: XCircle,
    spin: false,
    label: 'Rejected',
  },
};

function JsonDisplay({ data, label }: { data: unknown; label: string }) {
  const [copied, setCopied] = useState(false);

  const jsonString = JSON.stringify(data, null, 2);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(jsonString);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="border-t border-editor-border">
      <div className="flex items-center justify-between px-4 py-2 bg-editor-bg/30">
        <span className="text-xs text-editor-muted font-medium">{label}</span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 text-xs text-editor-muted hover:text-editor-text transition-colors"
        >
          {copied ? (
            <>
              <Check size={10} className="text-editor-success" />
              <span className="text-editor-success">Copied</span>
            </>
          ) : (
            <>
              <Copy size={10} />
              <span>Copy</span>
            </>
          )}
        </button>
      </div>
      <pre className="px-4 py-3 text-xs font-mono text-editor-text overflow-x-auto max-h-64 bg-editor-bg/20">
        {jsonString}
      </pre>
    </div>
  );
}

const ToolCallCard = memo(function ToolCallCard({
  toolCall,
  index,
}: {
  toolCall: ToolCall;
  index: number;
}) {
  const [isExpanded, setIsExpanded] = useState(false);

  const config = statusConfig[toolCall.status] || statusConfig.pending;
  const StatusIcon = config.icon;

  const displayName = toolCall.name.replace(/^mcp_[^_]+_/, '');
  const hasParams = Boolean(toolCall.parameters && Object.keys(toolCall.parameters).length > 0);
  const hasResult = Boolean(toolCall.result !== undefined && toolCall.result !== null);
  const hasError = Boolean(
    toolCall.status === 'failed' &&
    toolCall.result &&
    typeof toolCall.result === 'object' &&
    'error' in (toolCall.result as Record<string, unknown>)
  );

  return (
    <div className="rounded-lg border border-editor-border bg-editor-surface/50 overflow-hidden transition-all hover:border-editor-border/80">
      {/* Header */}
      <div
        className="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-editor-surface/70 transition-colors"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        {/* Index number */}
        <span className="text-xs text-editor-muted font-mono w-6">
          #{index + 1}
        </span>

        {/* Expand/Collapse icon */}
        <span className="text-editor-muted">
          {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
        </span>

        {/* Tool icon */}
        <div className="p-1.5 rounded-lg bg-orange-500/20 text-orange-400">
          <Wrench size={14} />
        </div>

        {/* Tool name and MCP badge */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-mono text-sm text-editor-text truncate">
              {displayName}
            </span>
            {toolCall.isMCP && (
              <span className="px-1.5 py-0.5 text-[10px] bg-purple-500/20 text-purple-400 rounded border border-purple-500/30">
                MCP
              </span>
            )}
            {toolCall.serverName && (
              <span className="text-xs text-editor-muted truncate">
                ({toolCall.serverName})
              </span>
            )}
          </div>
        </div>

        {/* Status badge */}
        <span
          className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border ${config.color}`}
        >
          <StatusIcon size={12} className={config.spin ? 'animate-spin' : ''} />
          {config.label}
        </span>
      </div>

      {/* Expandable content */}
      {isExpanded && (
        <div>
          {/* Parameters section */}
          {hasParams && (
            <JsonDisplay data={toolCall.parameters} label="Parameters" />
          )}

          {/* Result/Output section */}
          {hasResult && !hasError && (
            <JsonDisplay data={toolCall.result} label="Result" />
          )}

          {/* Error section */}
          {hasError && (
            <div className="border-t border-editor-border">
              <div className="px-4 py-2 bg-red-500/10">
                <span className="text-xs text-red-400 font-medium">Error</span>
              </div>
              <pre className="px-4 py-3 text-xs font-mono text-red-400 overflow-x-auto bg-red-500/5">
                {JSON.stringify((toolCall.result as Record<string, unknown>).error, null, 2)}
              </pre>
            </div>
          )}

          {/* No content message */}
          {!hasParams && !hasResult && toolCall.status === 'pending' && (
            <div className="px-4 py-6 text-center text-sm text-editor-muted border-t border-editor-border">
              Waiting for execution...
            </div>
          )}

          {/* Running indicator */}
          {toolCall.status === 'running' && (
            <div className="px-4 py-4 text-center text-sm text-blue-400 border-t border-editor-border bg-blue-500/5">
              <Loader2 size={16} className="inline animate-spin mr-2" />
              Tool is currently executing...
            </div>
          )}
        </div>
      )}
    </div>
  );
});

export function AgentToolCalls({ toolCalls }: AgentToolCallsProps) {
  if (!toolCalls || toolCalls.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <Wrench className="w-12 h-12 text-editor-muted mb-4" />
        <h3 className="text-lg font-medium text-editor-text mb-2">
          No tool calls
        </h3>
        <p className="text-sm text-editor-muted max-w-md">
          This agent execution did not make any tool calls.
        </p>
      </div>
    );
  }

  const completedCount = toolCalls.filter((t) => t.status === 'completed').length;
  const failedCount = toolCalls.filter((t) => t.status === 'failed').length;
  const runningCount = toolCalls.filter((t) => t.status === 'running').length;

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="flex items-center gap-4 text-sm">
        <span className="text-editor-muted">
          Total: <span className="text-editor-text font-medium">{toolCalls.length}</span>
        </span>
        {completedCount > 0 && (
          <span className="text-green-400">
            <CheckCircle size={14} className="inline mr-1" />
            {completedCount} completed
          </span>
        )}
        {failedCount > 0 && (
          <span className="text-red-400">
            <XCircle size={14} className="inline mr-1" />
            {failedCount} failed
          </span>
        )}
        {runningCount > 0 && (
          <span className="text-blue-400">
            <Loader2 size={14} className="inline mr-1 animate-spin" />
            {runningCount} running
          </span>
        )}
      </div>

      {/* Tool call list */}
      <div className="space-y-3">
        {toolCalls.map((toolCall, index) => (
          <ToolCallCard key={toolCall.id} toolCall={toolCall} index={index} />
        ))}
      </div>
    </div>
  );
}

export default AgentToolCalls;
