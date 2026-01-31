import { useState } from 'react';
import {
  Copy,
  Check,
  Clock,
  RefreshCw,
  ChevronDown,
  ChevronRight,
  Bot,
  Wrench,
  GitBranch,
  Layers,
  Timer,
  Sparkles,
} from 'lucide-react';
import type { Step, StepResult, StepType } from '../../hooks/useWorkflows';
import { formatDuration } from '../../hooks/useWorkflows';

interface StepLogViewerProps {
  step: Step;
  result?: StepResult;
  inputData?: Record<string, unknown>;
  onClose?: () => void;
}

const STEP_TYPE_ICONS: Record<StepType, typeof Bot> = {
  agent: Bot,
  tool: Wrench,
  condition: GitBranch,
  parallel: Layers,
  wait: Timer,
  transform: Sparkles,
};

export function StepLogViewer({ step, result, inputData, onClose }: StepLogViewerProps) {
  const [copiedField, setCopiedField] = useState<string | null>(null);
  const [expandedSections, setExpandedSections] = useState<Set<string>>(
    new Set(['config', 'output'])
  );

  const TypeIcon = STEP_TYPE_ICONS[step.type] || Layers;

  const copyToClipboard = async (text: string, field: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch {
      // Failed to copy
    }
  };

  const toggleSection = (section: string) => {
    setExpandedSections((prev) => {
      const next = new Set(prev);
      if (next.has(section)) {
        next.delete(section);
      } else {
        next.add(section);
      }
      return next;
    });
  };

  const formatJson = (data: unknown): string => {
    if (typeof data === 'string') return data;
    return JSON.stringify(data, null, 2);
  };

  const getStatusColor = () => {
    if (!result) return 'text-editor-muted';
    switch (result.status) {
      case 'completed':
        return 'text-editor-success';
      case 'failed':
        return 'text-editor-error';
      case 'running':
        return 'text-editor-warning';
      case 'skipped':
        return 'text-editor-muted';
      default:
        return 'text-editor-muted';
    }
  };

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-editor-border">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-editor-accent/10 rounded-lg">
            <TypeIcon size={20} className="text-editor-accent" />
          </div>
          <div>
            <h3 className="font-medium text-editor-text">{step.name}</h3>
            <div className="flex items-center gap-2 text-sm">
              <span className="text-editor-muted">{step.type}</span>
              {result && (
                <>
                  <span className="text-editor-muted">\u2022</span>
                  <span className={getStatusColor()}>{result.status}</span>
                </>
              )}
            </div>
          </div>
        </div>

        {onClose && (
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text rounded-lg hover:bg-editor-hover transition-colors"
          >
            \u2715
          </button>
        )}
      </div>

      {/* Content */}
      <div className="p-4 space-y-4">
        {/* Execution Info */}
        {result && (
          <div className="flex flex-wrap items-center gap-4 text-sm text-editor-muted">
            <span className="inline-flex items-center gap-1">
              <Clock size={14} />
              Duration: {formatDuration(result.duration)}
            </span>
            {result.retry_count !== undefined && result.retry_count > 0 && (
              <span className="inline-flex items-center gap-1 text-editor-warning">
                <RefreshCw size={14} />
                {result.retry_count} retries
              </span>
            )}
            <span>Started: {new Date(result.started_at).toLocaleString()}</span>
            <span>Completed: {new Date(result.completed_at).toLocaleString()}</span>
          </div>
        )}

        {/* Description */}
        {step.description && (
          <p className="text-sm text-editor-muted">{step.description}</p>
        )}

        {/* Step Configuration */}
        <CollapsibleSection
          title="Configuration"
          isExpanded={expandedSections.has('config')}
          onToggle={() => toggleSection('config')}
        >
          <div className="relative">
            <pre className="text-xs text-editor-text bg-editor-bg p-3 rounded-lg overflow-x-auto max-h-48 overflow-y-auto">
              {String(JSON.stringify(step.config, null, 2))}
            </pre>
            <CopyButton
              onClick={() => copyToClipboard(String(JSON.stringify(step.config, null, 2)), 'config')}
              copied={copiedField === 'config'}
            />
          </div>
        </CollapsibleSection>

        {/* Input Data */}
        {inputData && Object.keys(inputData).length > 0 && (
          <CollapsibleSection
            title="Input Data"
            isExpanded={expandedSections.has('input')}
            onToggle={() => toggleSection('input')}
          >
            <div className="relative">
              <pre className="text-xs text-editor-text bg-editor-bg p-3 rounded-lg overflow-x-auto max-h-48 overflow-y-auto">
                {formatJson(inputData)}
              </pre>
              <CopyButton
                onClick={() => copyToClipboard(formatJson(inputData), 'input')}
                copied={copiedField === 'input'}
              />
            </div>
          </CollapsibleSection>
        )}

        {/* Output Data */}
        {result?.output !== undefined && result.output !== null && (
          <CollapsibleSection
            title="Output"
            isExpanded={expandedSections.has('output')}
            onToggle={() => toggleSection('output')}
          >
            <div className="relative">
              <pre className="text-xs text-editor-text bg-editor-bg p-3 rounded-lg overflow-x-auto max-h-64 overflow-y-auto">
                {formatJson(result.output)}
              </pre>
              <CopyButton
                onClick={() => copyToClipboard(formatJson(result.output), 'output')}
                copied={copiedField === 'output'}
              />
            </div>
          </CollapsibleSection>
        )}

        {/* Error */}
        {result?.error && (
          <CollapsibleSection
            title="Error"
            isExpanded={expandedSections.has('error')}
            onToggle={() => toggleSection('error')}
            variant="error"
          >
            <div className="relative">
              <pre className="text-xs text-editor-error bg-editor-error/5 p-3 rounded-lg overflow-x-auto whitespace-pre-wrap">
                {result.error}
              </pre>
              <CopyButton
                onClick={() => copyToClipboard(result.error!, 'error')}
                copied={copiedField === 'error'}
              />
            </div>
          </CollapsibleSection>
        )}

        {/* Retry Policy */}
        {step.retry_policy && (
          <CollapsibleSection
            title="Retry Policy"
            isExpanded={expandedSections.has('retry')}
            onToggle={() => toggleSection('retry')}
          >
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div className="text-editor-muted">Max Retries</div>
              <div className="text-editor-text">{step.retry_policy.max_retries}</div>
              <div className="text-editor-muted">Delay</div>
              <div className="text-editor-text">{formatDuration(step.retry_policy.delay)}</div>
              {step.retry_policy.backoff_type && (
                <>
                  <div className="text-editor-muted">Backoff Type</div>
                  <div className="text-editor-text">{step.retry_policy.backoff_type}</div>
                </>
              )}
              {step.retry_policy.max_delay && (
                <>
                  <div className="text-editor-muted">Max Delay</div>
                  <div className="text-editor-text">
                    {formatDuration(step.retry_policy.max_delay)}
                  </div>
                </>
              )}
            </div>
          </CollapsibleSection>
        )}

        {/* Metadata */}
        {result?.metadata && Object.keys(result.metadata).length > 0 && (
          <CollapsibleSection
            title="Metadata"
            isExpanded={expandedSections.has('metadata')}
            onToggle={() => toggleSection('metadata')}
          >
            <pre className="text-xs text-editor-text bg-editor-bg p-3 rounded-lg overflow-x-auto">
              {formatJson(result.metadata)}
            </pre>
          </CollapsibleSection>
        )}
      </div>
    </div>
  );
}

interface CollapsibleSectionProps {
  title: string;
  isExpanded: boolean;
  onToggle: () => void;
  variant?: 'default' | 'error';
  children: React.ReactNode;
}

function CollapsibleSection({
  title,
  isExpanded,
  onToggle,
  variant = 'default',
  children,
}: CollapsibleSectionProps) {
  return (
    <div
      className={`border rounded-lg ${
        variant === 'error'
          ? 'border-editor-error/20 bg-editor-error/5'
          : 'border-editor-border'
      }`}
    >
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-2 p-3 text-left hover:bg-editor-hover/50 transition-colors"
      >
        {isExpanded ? (
          <ChevronDown size={16} className="text-editor-muted" />
        ) : (
          <ChevronRight size={16} className="text-editor-muted" />
        )}
        <span
          className={`text-sm font-medium ${
            variant === 'error' ? 'text-editor-error' : 'text-editor-text'
          }`}
        >
          {title}
        </span>
      </button>
      {isExpanded && <div className="px-3 pb-3">{children}</div>}
    </div>
  );
}

interface CopyButtonProps {
  onClick: () => void;
  copied: boolean;
}

function CopyButton({ onClick, copied }: CopyButtonProps) {
  return (
    <button
      onClick={onClick}
      className="absolute top-2 right-2 p-1.5 bg-editor-surface border border-editor-border rounded text-editor-muted hover:text-editor-text transition-colors"
      title="Copy to clipboard"
    >
      {copied ? <Check size={14} /> : <Copy size={14} />}
    </button>
  );
}
