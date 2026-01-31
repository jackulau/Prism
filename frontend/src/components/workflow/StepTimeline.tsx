import { useState } from 'react';
import {
  CheckCircle,
  XCircle,
  Circle,
  Loader2,
  SkipForward,
  ChevronDown,
  ChevronRight,
  Clock,
  Bot,
  Wrench,
  GitBranch,
  Layers,
  Timer,
  Sparkles,
} from 'lucide-react';
import type { Step, StepResult, StepStatus, StepType } from '../../hooks/useWorkflows';
import { formatDuration } from '../../hooks/useWorkflows';

interface StepTimelineProps {
  steps: Step[];
  stepResults?: StepResult[];
  currentStep: number;
  onStepClick?: (stepId: string) => void;
}

const STEP_STATUS_CONFIG: Record<
  StepStatus,
  { icon: typeof CheckCircle; color: string; bgColor: string }
> = {
  pending: {
    icon: Circle,
    color: 'text-editor-muted',
    bgColor: 'bg-editor-muted/10',
  },
  running: {
    icon: Loader2,
    color: 'text-editor-warning',
    bgColor: 'bg-editor-warning/10',
  },
  completed: {
    icon: CheckCircle,
    color: 'text-editor-success',
    bgColor: 'bg-editor-success/10',
  },
  failed: {
    icon: XCircle,
    color: 'text-editor-error',
    bgColor: 'bg-editor-error/10',
  },
  skipped: {
    icon: SkipForward,
    color: 'text-editor-muted',
    bgColor: 'bg-editor-muted/10',
  },
};

const STEP_TYPE_ICONS: Record<StepType, typeof Bot> = {
  agent: Bot,
  tool: Wrench,
  condition: GitBranch,
  parallel: Layers,
  wait: Timer,
  transform: Sparkles,
};

export function StepTimeline({
  steps,
  stepResults,
  currentStep,
  onStepClick,
}: StepTimelineProps) {
  const [expandedSteps, setExpandedSteps] = useState<Set<string>>(new Set());

  const getStepResult = (stepId: string): StepResult | undefined => {
    return stepResults?.find((r) => r.step_id === stepId);
  };

  const getStepStatus = (step: Step, index: number): StepStatus => {
    const result = getStepResult(step.id);
    if (result) return result.status;
    if (index === currentStep) return 'running';
    if (index < currentStep) return 'completed';
    return 'pending';
  };

  const toggleExpanded = (stepId: string) => {
    setExpandedSteps((prev) => {
      const next = new Set(prev);
      if (next.has(stepId)) {
        next.delete(stepId);
      } else {
        next.add(stepId);
      }
      return next;
    });
  };

  return (
    <div className="space-y-0">
      {steps.map((step, index) => {
        const status = getStepStatus(step, index);
        const result = getStepResult(step.id);
        const statusConfig = STEP_STATUS_CONFIG[status];
        const StatusIcon = statusConfig.icon;
        const TypeIcon = STEP_TYPE_ICONS[step.type] || Layers;
        const isExpanded = expandedSteps.has(step.id);
        const isLast = index === steps.length - 1;
        const hasDetails = result?.output || result?.error;

        return (
          <div key={step.id} className="relative">
            {/* Connector Line */}
            {!isLast && (
              <div
                className={`absolute left-5 top-10 w-0.5 h-full -ml-px ${
                  index < currentStep ? 'bg-editor-success' : 'bg-editor-border'
                }`}
              />
            )}

            {/* Step Item */}
            <div
              className={`relative flex items-start gap-4 p-3 rounded-lg transition-colors ${
                onStepClick || hasDetails
                  ? 'cursor-pointer hover:bg-editor-surface'
                  : ''
              } ${status === 'failed' ? 'bg-editor-error/5' : ''}`}
              onClick={() => {
                if (hasDetails) toggleExpanded(step.id);
                onStepClick?.(step.id);
              }}
            >
              {/* Status Icon */}
              <div
                className={`flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center ${statusConfig.bgColor} z-10`}
              >
                <StatusIcon
                  size={20}
                  className={`${statusConfig.color} ${
                    status === 'running' ? 'animate-spin' : ''
                  }`}
                />
              </div>

              {/* Content */}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  {hasDetails && (
                    <button className="p-0.5 -ml-0.5">
                      {isExpanded ? (
                        <ChevronDown size={16} className="text-editor-muted" />
                      ) : (
                        <ChevronRight size={16} className="text-editor-muted" />
                      )}
                    </button>
                  )}
                  <h4 className="font-medium text-editor-text truncate">
                    {step.name}
                  </h4>
                  <span
                    className={`inline-flex items-center gap-1 px-2 py-0.5 text-xs rounded ${statusConfig.bgColor} ${statusConfig.color}`}
                  >
                    {status}
                  </span>
                </div>

                <div className="flex items-center gap-3 mt-1 text-xs text-editor-muted">
                  <span className="inline-flex items-center gap-1">
                    <TypeIcon size={12} />
                    {step.type}
                  </span>
                  {result?.duration !== undefined && (
                    <span className="inline-flex items-center gap-1">
                      <Clock size={12} />
                      {formatDuration(result.duration)}
                    </span>
                  )}
                  {result?.retry_count && result.retry_count > 0 && (
                    <span className="text-editor-warning">
                      {result.retry_count} retries
                    </span>
                  )}
                </div>

                {step.description && !isExpanded && (
                  <p className="text-sm text-editor-muted mt-1 truncate">
                    {step.description}
                  </p>
                )}

                {/* Error Preview (always visible for failed steps) */}
                {status === 'failed' && result?.error && !isExpanded && (
                  <div className="mt-2 p-2 bg-editor-error/10 border border-editor-error/20 rounded text-xs text-editor-error truncate">
                    {result.error}
                  </div>
                )}

                {/* Expanded Details */}
                {isExpanded && hasDetails && (
                  <div className="mt-3 space-y-3">
                    {step.description && (
                      <p className="text-sm text-editor-muted">
                        {step.description}
                      </p>
                    )}

                    {result?.output !== undefined && result.output !== null && (
                      <div className="p-3 bg-editor-bg rounded-lg">
                        <p className="text-xs text-editor-muted mb-2 font-medium">
                          Output
                        </p>
                        <pre className="text-xs text-editor-text whitespace-pre-wrap overflow-x-auto max-h-48 overflow-y-auto">
                          {typeof result.output === 'string'
                            ? result.output
                            : String(JSON.stringify(result.output, null, 2))}
                        </pre>
                      </div>
                    )}

                    {result?.error && (
                      <div className="p-3 bg-editor-error/10 border border-editor-error/20 rounded-lg">
                        <p className="text-xs text-editor-error mb-2 font-medium">
                          Error
                        </p>
                        <pre className="text-xs text-editor-error whitespace-pre-wrap">
                          {result.error}
                        </pre>
                      </div>
                    )}

                    {result && (
                      <div className="flex items-center gap-4 text-xs text-editor-muted">
                        <span>
                          Started: {new Date(result.started_at).toLocaleString()}
                        </span>
                        <span>
                          Completed: {new Date(result.completed_at).toLocaleString()}
                        </span>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
