import { useEffect, useState, useRef } from 'react';
import { useWorkflowExecutionStore } from '../../store/workflowExecutionStore';
import type { WorkflowStepResult, WorkflowExecutionStatus } from '../../types';

interface ExecutionMonitorProps {
  onStepClick?: (stepId: string) => void;
  className?: string;
}

const STATUS_COLORS: Record<WorkflowExecutionStatus, string> = {
  idle: 'text-gray-400 bg-gray-400/10',
  pending: 'text-gray-400 bg-gray-400/10',
  running: 'text-blue-400 bg-blue-400/10',
  paused: 'text-yellow-400 bg-yellow-400/10',
  completed: 'text-green-400 bg-green-400/10',
  failed: 'text-red-400 bg-red-400/10',
  cancelled: 'text-gray-400 bg-gray-400/10',
  waiting_input: 'text-purple-400 bg-purple-400/10',
};

const STATUS_LABELS: Record<WorkflowExecutionStatus, string> = {
  idle: 'Idle',
  pending: 'Pending',
  running: 'Running',
  paused: 'Paused',
  completed: 'Completed',
  failed: 'Failed',
  cancelled: 'Cancelled',
  waiting_input: 'Waiting for Input',
};

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

function StepStatusIcon({ status }: { status: WorkflowStepResult['status'] }) {
  switch (status) {
    case 'pending':
      return (
        <div className="w-4 h-4 rounded-full border-2 border-gray-500" />
      );
    case 'running':
      return (
        <div className="w-4 h-4 relative">
          <div className="absolute inset-0 rounded-full border-2 border-blue-400 border-t-transparent animate-spin" />
        </div>
      );
    case 'completed':
      return (
        <svg className="w-4 h-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
      );
    case 'failed':
      return (
        <svg className="w-4 h-4 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
        </svg>
      );
    case 'skipped':
      return (
        <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M5 5l7 7-7 7" />
        </svg>
      );
    case 'retrying':
      return (
        <svg className="w-4 h-4 text-yellow-400 animate-spin" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
      );
    default:
      return <div className="w-4 h-4 rounded-full bg-gray-600" />;
  }
}

function StepItem({
  step,
  isActive,
  onClick,
}: {
  step: WorkflowStepResult;
  isActive: boolean;
  onClick?: () => void;
}) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div
      className={`
        border-l-2 pl-4 py-2 cursor-pointer transition-colors
        ${isActive ? 'border-blue-400 bg-blue-400/5' : 'border-gray-700 hover:bg-gray-800/50'}
        ${step.status === 'failed' ? 'border-red-400' : ''}
        ${step.status === 'completed' ? 'border-green-400' : ''}
      `}
      onClick={() => {
        setIsExpanded(!isExpanded);
        onClick?.();
      }}
    >
      <div className="flex items-center gap-3">
        <StepStatusIcon status={step.status} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-sm text-gray-200 truncate">
              {step.stepName}
            </span>
            <span className="text-xs text-gray-500 px-1.5 py-0.5 bg-gray-800 rounded">
              {step.stepType}
            </span>
          </div>
          {step.duration !== undefined && step.status !== 'running' && (
            <span className="text-xs text-gray-500">
              {formatDuration(step.duration)}
            </span>
          )}
        </div>
        {step.status === 'running' && (
          <div className="w-2 h-2 rounded-full bg-blue-400 animate-pulse" />
        )}
        <svg
          className={`w-4 h-4 text-gray-500 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </div>

      {isExpanded && (
        <div className="mt-2 pl-7 text-sm">
          {step.error && (
            <div className="text-red-400 bg-red-400/10 px-2 py-1 rounded text-xs mb-2">
              {step.error}
            </div>
          )}
          {step.output !== undefined && (
            <div className="bg-gray-800 rounded p-2 text-xs text-gray-300 font-mono overflow-x-auto">
              <pre>{JSON.stringify(step.output, null, 2)}</pre>
            </div>
          )}
          {step.retryCount !== undefined && step.retryCount > 0 && (
            <div className="text-yellow-400 text-xs mt-1">
              Retry count: {step.retryCount}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export function ExecutionMonitor({ onStepClick, className = '' }: ExecutionMonitorProps) {
  const {
    workflowId,
    workflowInfo,
    status,
    currentStepIndex,
    totalSteps,
    stepResults,
    startedAt,
    error,
    waitingForInput,
    inputPrompt,
  } = useWorkflowExecutionStore();

  const [elapsedTime, setElapsedTime] = useState(0);
  const stepListRef = useRef<HTMLDivElement>(null);

  // Update elapsed time every second while running
  useEffect(() => {
    if (status !== 'running' || !startedAt) return;

    const interval = setInterval(() => {
      setElapsedTime(Date.now() - startedAt);
    }, 1000);

    return () => clearInterval(interval);
  }, [status, startedAt]);

  // Auto-scroll to current step
  useEffect(() => {
    if (stepListRef.current && status === 'running') {
      const activeStep = stepListRef.current.querySelector('[data-active="true"]');
      if (activeStep) {
        activeStep.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      }
    }
  }, [currentStepIndex, status]);

  const progressPercent = totalSteps > 0
    ? Math.round((currentStepIndex / totalSteps) * 100)
    : 0;

  const steps = Array.from(stepResults.values());

  if (!workflowId && status === 'idle') {
    return (
      <div className={`flex items-center justify-center p-8 text-gray-500 ${className}`}>
        No workflow execution in progress
      </div>
    );
  }

  return (
    <div className={`flex flex-col h-full bg-gray-900 ${className}`}>
      {/* Header */}
      <div className="flex-shrink-0 p-4 border-b border-gray-800">
        <div className="flex items-center justify-between mb-3">
          <div>
            <h3 className="text-lg font-semibold text-gray-100">
              {workflowInfo?.name || 'Workflow Execution'}
            </h3>
            {workflowId && (
              <p className="text-xs text-gray-500 font-mono">{workflowId}</p>
            )}
          </div>
          <span className={`px-2 py-1 rounded text-xs font-medium ${STATUS_COLORS[status]}`}>
            {STATUS_LABELS[status]}
          </span>
        </div>

        {/* Progress bar */}
        <div className="mb-3">
          <div className="flex items-center justify-between text-xs text-gray-400 mb-1">
            <span>Step {currentStepIndex} of {totalSteps}</span>
            <span>{progressPercent}%</span>
          </div>
          <div className="h-2 bg-gray-800 rounded-full overflow-hidden">
            <div
              className={`h-full transition-all duration-300 ${
                status === 'failed' ? 'bg-red-500' :
                status === 'completed' ? 'bg-green-500' :
                'bg-blue-500'
              }`}
              style={{ width: `${progressPercent}%` }}
            />
          </div>
        </div>

        {/* Elapsed time */}
        <div className="flex items-center gap-4 text-sm text-gray-400">
          <div className="flex items-center gap-1">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>{formatDuration(elapsedTime)}</span>
          </div>
          {steps.filter(s => s.status === 'completed').length > 0 && (
            <span className="text-green-400">
              {steps.filter(s => s.status === 'completed').length} completed
            </span>
          )}
          {steps.filter(s => s.status === 'failed').length > 0 && (
            <span className="text-red-400">
              {steps.filter(s => s.status === 'failed').length} failed
            </span>
          )}
        </div>
      </div>

      {/* Waiting for input alert */}
      {waitingForInput && inputPrompt && (
        <div className="flex-shrink-0 p-3 bg-purple-500/10 border-b border-purple-500/20">
          <div className="flex items-center gap-2 text-purple-400">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="text-sm">{inputPrompt}</span>
          </div>
        </div>
      )}

      {/* Error display */}
      {error && (
        <div className="flex-shrink-0 p-3 bg-red-500/10 border-b border-red-500/20">
          <div className="flex items-center gap-2 text-red-400">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <span className="text-sm">{error}</span>
          </div>
        </div>
      )}

      {/* Step list */}
      <div ref={stepListRef} className="flex-1 overflow-y-auto p-4">
        {steps.length === 0 ? (
          <div className="text-center text-gray-500 py-8">
            {status === 'running' ? 'Waiting for steps...' : 'No steps executed'}
          </div>
        ) : (
          <div className="space-y-1">
            {steps.map((step) => (
              <div key={step.stepId} data-active={step.status === 'running'}>
                <StepItem
                  step={step}
                  isActive={step.status === 'running'}
                  onClick={() => onStepClick?.(step.stepId)}
                />
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default ExecutionMonitor;
