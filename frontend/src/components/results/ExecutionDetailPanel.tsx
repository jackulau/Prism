import React, { useState, useEffect } from 'react';
import {
  X,
  RotateCcw,
  Download,
  Trash2,
  LayoutGrid,
  Users,
  Clock,
  FileText,
} from 'lucide-react';
import { ExecutionDetail, type ExecutionDetailData } from './ExecutionDetail';
import { AgentResultCard, type AgentResult } from './AgentResultCard';
import { ExecutionTimeline, type TimelineItem } from './ExecutionTimeline';
import { ExecutionMetrics, type ExecutionMetricsData } from './ExecutionMetrics';

type TabId = 'overview' | 'agents' | 'timeline' | 'logs';

interface Tab {
  id: TabId;
  label: string;
  icon: React.ReactNode;
}

const tabs: Tab[] = [
  { id: 'overview', label: 'Overview', icon: <LayoutGrid className="w-4 h-4" /> },
  { id: 'agents', label: 'Agents', icon: <Users className="w-4 h-4" /> },
  { id: 'timeline', label: 'Timeline', icon: <Clock className="w-4 h-4" /> },
  { id: 'logs', label: 'Logs', icon: <FileText className="w-4 h-4" /> },
];

export interface ExecutionPanelData {
  execution: ExecutionDetailData;
  agentResults: AgentResult[];
  timelineItems: TimelineItem[];
  metrics: ExecutionMetricsData;
  logs?: string[];
}

interface ExecutionDetailPanelProps {
  isOpen: boolean;
  onClose: () => void;
  data: ExecutionPanelData | null;
  onRetry?: (executionId: string) => void;
  onExport?: (executionId: string) => void;
  onDelete?: (executionId: string) => void;
}

export const ExecutionDetailPanel: React.FC<ExecutionDetailPanelProps> = ({
  isOpen,
  onClose,
  data,
  onRetry,
  onExport,
  onDelete,
}) => {
  const [activeTab, setActiveTab] = useState<TabId>('overview');
  const [isVisible, setIsVisible] = useState(false);
  const [isClosing, setIsClosing] = useState(false);

  // Handle open/close animations
  useEffect(() => {
    if (isOpen) {
      setIsVisible(true);
      setIsClosing(false);
    } else if (isVisible) {
      setIsClosing(true);
      const timer = setTimeout(() => {
        setIsVisible(false);
        setIsClosing(false);
      }, 300);
      return () => clearTimeout(timer);
    }
  }, [isOpen, isVisible]);

  // Reset to overview tab when opening with new data
  useEffect(() => {
    if (isOpen && data) {
      setActiveTab('overview');
    }
  }, [isOpen, data?.execution.id]);

  // Handle escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleEscape);
    return () => window.removeEventListener('keydown', handleEscape);
  }, [isOpen, onClose]);

  if (!isVisible || !data) return null;

  const renderTabContent = () => {
    switch (activeTab) {
      case 'overview':
        return (
          <div className="space-y-6">
            <ExecutionDetail execution={data.execution} onBack={onClose} />
            <ExecutionMetrics metrics={data.metrics} />
          </div>
        );

      case 'agents':
        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-medium text-editor-text">
                Agent Results ({data.agentResults.length})
              </h3>
            </div>
            {data.agentResults.length === 0 ? (
              <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
                <Users className="w-8 h-8 text-editor-muted mx-auto mb-2" />
                <p className="text-editor-muted">No agent results available</p>
              </div>
            ) : (
              <div className="space-y-3">
                {data.agentResults.map((result) => (
                  <AgentResultCard key={result.id} result={result} />
                ))}
              </div>
            )}
          </div>
        );

      case 'timeline':
        return (
          <div className="space-y-4">
            <h3 className="text-lg font-medium text-editor-text">Execution Timeline</h3>
            <ExecutionTimeline
              items={data.timelineItems}
              executionStartedAt={data.execution.startedAt || undefined}
              executionCompletedAt={data.execution.completedAt || undefined}
            />
          </div>
        );

      case 'logs':
        return (
          <div className="space-y-4">
            <h3 className="text-lg font-medium text-editor-text">Execution Logs</h3>
            {!data.logs || data.logs.length === 0 ? (
              <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
                <FileText className="w-8 h-8 text-editor-muted mx-auto mb-2" />
                <p className="text-editor-muted">No logs available</p>
              </div>
            ) : (
              <div className="bg-editor-bg rounded-lg border border-editor-border overflow-hidden">
                <div className="p-3 border-b border-editor-border bg-editor-surface">
                  <span className="text-xs text-editor-muted">
                    {data.logs.length} log entries
                  </span>
                </div>
                <pre className="p-4 text-sm text-editor-text font-mono overflow-x-auto max-h-96 overflow-y-auto">
                  {data.logs.join('\n')}
                </pre>
              </div>
            )}
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <>
      {/* Backdrop */}
      <div
        className={`fixed inset-0 bg-black/50 z-40 transition-opacity duration-300 ${
          isClosing ? 'opacity-0' : 'opacity-100'
        }`}
        onClick={onClose}
      />

      {/* Panel */}
      <div
        className={`fixed inset-y-0 right-0 w-full max-w-3xl bg-editor-bg border-l border-editor-border z-50 flex flex-col transition-transform duration-300 ease-out ${
          isClosing ? 'translate-x-full' : 'translate-x-0'
        }`}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border bg-editor-surface">
          <h2 className="text-lg font-semibold text-editor-text">Execution Details</h2>
          <div className="flex items-center gap-2">
            {/* Action buttons */}
            {onRetry && (
              <button
                onClick={() => onRetry(data.execution.id)}
                className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded-lg transition-colors"
                title="Retry execution"
              >
                <RotateCcw className="w-4 h-4" />
              </button>
            )}
            {onExport && (
              <button
                onClick={() => onExport(data.execution.id)}
                className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded-lg transition-colors"
                title="Export results"
              >
                <Download className="w-4 h-4" />
              </button>
            )}
            {onDelete && (
              <button
                onClick={() => onDelete(data.execution.id)}
                className="p-2 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
                title="Delete execution"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            )}
            <div className="w-px h-6 bg-editor-border mx-2" />
            <button
              onClick={onClose}
              className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded-lg transition-colors"
              title="Close panel (Esc)"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-editor-border bg-editor-surface px-4">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors relative ${
                activeTab === tab.id
                  ? 'text-editor-accent'
                  : 'text-editor-muted hover:text-editor-text'
              }`}
            >
              {tab.icon}
              {tab.label}
              {activeTab === tab.id && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-editor-accent" />
              )}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {renderTabContent()}
        </div>
      </div>
    </>
  );
};
