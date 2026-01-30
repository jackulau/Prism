import { useState } from 'react';
import {
  X,
  Bot,
  Play,
  Square,
  Trash2,
  Loader2,
  MessageSquare,
  Wrench,
  BarChart3,
  FileOutput,
} from 'lucide-react';
import { ConfirmDialog } from '../ConfirmDialog';
import { AgentMessages } from './AgentMessages';
import { AgentToolCalls } from './AgentToolCalls';
import { AgentMetrics } from './AgentMetrics';
import { AgentResults } from './AgentResults';
import type { Message, ToolCall } from '../../types';

export type AgentStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface AgentExecution {
  id: string;
  name: string;
  status: AgentStatus;
  model: string;
  provider: string;
  systemPrompt?: string;
  createdAt: Date;
  startedAt?: Date;
  completedAt?: Date;
  messages: Message[];
  toolCalls: ToolCall[];
  result?: unknown;
  error?: string;
  metrics?: AgentExecutionMetrics;
}

export interface AgentExecutionMetrics {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  inputCost?: number;
  outputCost?: number;
  totalCost?: number;
  duration?: number;
  iterationCount?: number;
}

interface AgentDetailProps {
  agent: AgentExecution;
  onClose: () => void;
  onCancel?: (id: string) => void;
  onRetry?: (id: string) => void;
  onDelete?: (id: string) => void;
  isCancelling?: boolean;
  isDeleting?: boolean;
}

type TabId = 'messages' | 'tools' | 'metrics' | 'results';

const tabs: { id: TabId; label: string; icon: React.ReactNode }[] = [
  { id: 'messages', label: 'Messages', icon: <MessageSquare size={16} /> },
  { id: 'tools', label: 'Tool Calls', icon: <Wrench size={16} /> },
  { id: 'metrics', label: 'Metrics', icon: <BarChart3 size={16} /> },
  { id: 'results', label: 'Results', icon: <FileOutput size={16} /> },
];

const statusConfig: Record<AgentStatus, { color: string; label: string }> = {
  pending: {
    color: 'bg-editor-muted/20 text-editor-muted border-editor-muted/30',
    label: 'Pending',
  },
  running: {
    color: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    label: 'Running',
  },
  completed: {
    color: 'bg-green-500/20 text-green-400 border-green-500/30',
    label: 'Completed',
  },
  failed: {
    color: 'bg-red-500/20 text-red-400 border-red-500/30',
    label: 'Failed',
  },
  cancelled: {
    color: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    label: 'Cancelled',
  },
};

export function AgentDetail({
  agent,
  onClose,
  onCancel,
  onRetry,
  onDelete,
  isCancelling = false,
  isDeleting = false,
}: AgentDetailProps) {
  const [activeTab, setActiveTab] = useState<TabId>('messages');
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const status = statusConfig[agent.status];
  const isRunning = agent.status === 'running';
  const isFailed = agent.status === 'failed';

  const handleDelete = () => {
    if (onDelete) {
      onDelete(agent.id);
    }
    setShowDeleteConfirm(false);
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case 'messages':
        return <AgentMessages messages={agent.messages} />;
      case 'tools':
        return <AgentToolCalls toolCalls={agent.toolCalls} />;
      case 'metrics':
        return (
          <AgentMetrics
            metrics={agent.metrics}
            model={agent.model}
            provider={agent.provider}
            startedAt={agent.startedAt}
            completedAt={agent.completedAt}
          />
        );
      case 'results':
        return (
          <AgentResults
            result={agent.result}
            error={agent.error}
            status={agent.status}
          />
        );
      default:
        return null;
    }
  };

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 z-40"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] max-w-[95vw] max-h-[90vh] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50 flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border shrink-0">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-editor-accent/10 rounded-lg">
              <Bot className="w-5 h-5 text-editor-accent" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-semibold text-editor-text">
                  {agent.name || 'Agent Execution'}
                </h2>
                <span
                  className={`px-2 py-0.5 text-xs rounded-full border ${status.color}`}
                >
                  {isRunning && (
                    <Loader2 size={10} className="inline mr-1 animate-spin" />
                  )}
                  {status.label}
                </span>
              </div>
              <p className="text-xs text-editor-muted mt-0.5">
                {agent.provider} / {agent.model}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {/* Cancel button (only for running agents) */}
            {isRunning && onCancel && (
              <button
                onClick={() => onCancel(agent.id)}
                disabled={isCancelling}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-yellow-500/20 text-yellow-400 border border-yellow-500/30 rounded-lg hover:bg-yellow-500/30 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {isCancelling ? (
                  <Loader2 size={14} className="animate-spin" />
                ) : (
                  <Square size={14} />
                )}
                Cancel
              </button>
            )}

            {/* Retry button (only for failed agents) */}
            {isFailed && onRetry && (
              <button
                onClick={() => onRetry(agent.id)}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-editor-accent/20 text-editor-accent border border-editor-accent/30 rounded-lg hover:bg-editor-accent/30 transition-colors"
              >
                <Play size={14} />
                Retry
              </button>
            )}

            {/* Delete button */}
            {onDelete && (
              <button
                onClick={() => setShowDeleteConfirm(true)}
                disabled={isDeleting || isRunning}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-red-500/20 text-red-400 border border-red-500/30 rounded-lg hover:bg-red-500/30 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {isDeleting ? (
                  <Loader2 size={14} className="animate-spin" />
                ) : (
                  <Trash2 size={14} />
                )}
                Delete
              </button>
            )}

            {/* Close button */}
            <button
              onClick={onClose}
              className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
            >
              <X size={20} />
            </button>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="flex items-center gap-1 px-6 py-2 border-b border-editor-border bg-editor-surface/30 shrink-0">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-4 py-2 text-sm rounded-lg transition-colors ${
                activeTab === tab.id
                  ? 'bg-editor-accent/20 text-editor-accent'
                  : 'text-editor-muted hover:text-editor-text hover:bg-editor-surface'
              }`}
            >
              {tab.icon}
              {tab.label}
              {tab.id === 'tools' && agent.toolCalls.length > 0 && (
                <span className="px-1.5 py-0.5 text-xs bg-editor-surface rounded-full">
                  {agent.toolCalls.length}
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Tab Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {renderTabContent()}
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        isOpen={showDeleteConfirm}
        title="Delete Agent Execution"
        message="Are you sure you want to delete this agent execution? This action cannot be undone."
        confirmText="Delete"
        variant="danger"
        onConfirm={handleDelete}
        onCancel={() => setShowDeleteConfirm(false)}
      />
    </>
  );
}

export default AgentDetail;
