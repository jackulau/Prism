import { useWorkflowExecutionStore } from '../../store/workflowExecutionStore';
import { useWorkflowExecution } from '../../hooks/useWorkflowExecution';
import type { WorkflowExecutionStatus } from '../../types';

interface ExecutionControlsProps {
  onViewLogs?: () => void;
  className?: string;
}

const STATUS_COLORS: Record<WorkflowExecutionStatus, string> = {
  idle: 'bg-gray-500',
  running: 'bg-blue-500',
  paused: 'bg-yellow-500',
  completed: 'bg-green-500',
  failed: 'bg-red-500',
  cancelled: 'bg-gray-500',
  waiting_input: 'bg-purple-500',
};

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${Math.floor(ms / 1000)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}:${seconds.toString().padStart(2, '0')}`;
}

export function ExecutionControls({ onViewLogs, className = '' }: ExecutionControlsProps) {
  const {
    status,
    currentStepIndex,
    totalSteps,
    startedAt,
  } = useWorkflowExecutionStore();

  const {
    pauseWorkflow,
    resumeWorkflow,
    stopWorkflow,
  } = useWorkflowExecution();

  const isRunning = status === 'running';
  const isPaused = status === 'paused';
  const isActive = isRunning || isPaused || status === 'waiting_input';
  const isIdle = status === 'idle';
  const isFinished = status === 'completed' || status === 'failed' || status === 'cancelled';

  const progressPercent = totalSteps > 0
    ? Math.round((currentStepIndex / totalSteps) * 100)
    : 0;

  const elapsed = startedAt ? Date.now() - startedAt : 0;

  return (
    <div className={`flex items-center gap-3 px-4 py-2 bg-gray-900 border-b border-gray-800 ${className}`}>
      {/* Status indicator */}
      <div className="flex items-center gap-2">
        <div className={`w-2 h-2 rounded-full ${STATUS_COLORS[status]} ${isRunning ? 'animate-pulse' : ''}`} />
        <span className="text-xs text-gray-400 capitalize">{status.replace('_', ' ')}</span>
      </div>

      {/* Divider */}
      <div className="h-4 w-px bg-gray-700" />

      {/* Control buttons */}
      <div className="flex items-center gap-1">
        {/* Play/Resume button */}
        {(isPaused || status === 'waiting_input') && (
          <button
            onClick={resumeWorkflow}
            className="p-1.5 rounded hover:bg-gray-800 text-green-400 hover:text-green-300 transition-colors"
            title="Resume"
          >
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <path d="M8 5v14l11-7z" />
            </svg>
          </button>
        )}

        {/* Pause button */}
        {isRunning && (
          <button
            onClick={pauseWorkflow}
            className="p-1.5 rounded hover:bg-gray-800 text-yellow-400 hover:text-yellow-300 transition-colors"
            title="Pause"
          >
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z" />
            </svg>
          </button>
        )}

        {/* Stop button */}
        {isActive && (
          <button
            onClick={stopWorkflow}
            className="p-1.5 rounded hover:bg-gray-800 text-red-400 hover:text-red-300 transition-colors"
            title="Stop"
          >
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <path d="M6 6h12v12H6z" />
            </svg>
          </button>
        )}
      </div>

      {/* Divider */}
      {isActive && <div className="h-4 w-px bg-gray-700" />}

      {/* Progress info */}
      {!isIdle && (
        <div className="flex items-center gap-3 flex-1">
          {/* Progress bar (small) */}
          <div className="flex-1 max-w-32">
            <div className="h-1.5 bg-gray-800 rounded-full overflow-hidden">
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

          {/* Step counter */}
          <span className="text-xs text-gray-400 whitespace-nowrap">
            {currentStepIndex}/{totalSteps}
          </span>

          {/* Progress percentage */}
          <span className="text-xs text-gray-400 font-mono">
            {progressPercent}%
          </span>

          {/* Elapsed time */}
          {isActive && (
            <span className="text-xs text-gray-500 font-mono">
              {formatDuration(elapsed)}
            </span>
          )}
        </div>
      )}

      {/* Idle state message */}
      {isIdle && (
        <span className="text-xs text-gray-500">No workflow running</span>
      )}

      {/* Finished state info */}
      {isFinished && (
        <span className={`text-xs ${
          status === 'completed' ? 'text-green-400' :
          status === 'failed' ? 'text-red-400' :
          'text-gray-400'
        }`}>
          {status === 'completed' ? 'Completed successfully' :
           status === 'failed' ? 'Execution failed' :
           'Cancelled'}
        </span>
      )}

      {/* Spacer */}
      <div className="flex-1" />

      {/* View logs button */}
      {onViewLogs && !isIdle && (
        <button
          onClick={onViewLogs}
          className="flex items-center gap-1 px-2 py-1 text-xs text-gray-400 hover:text-gray-300 hover:bg-gray-800 rounded transition-colors"
        >
          <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          Logs
        </button>
      )}
    </div>
  );
}

export default ExecutionControls;
