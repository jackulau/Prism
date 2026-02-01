import { useEffect, useCallback, useRef, useState } from 'react';
import { MessageSquare, Play, Pause, Square, FileText, Loader2 } from 'lucide-react';
import { useWorkflowExecutionStore } from '../../store/workflowExecutionStore';
import { useWorkflowExecution } from '../../hooks/useWorkflowExecution';
import type { WorkflowExecutionStatus } from '../../types';

interface ExecutionControlsProps {
  onViewLogs?: () => void;
  onInputRequired?: () => void;
  className?: string;
  enableKeyboardShortcuts?: boolean;
}

const STATUS_COLORS: Record<WorkflowExecutionStatus, string> = {
  idle: 'bg-gray-500',
  pending: 'bg-gray-400',
  running: 'bg-blue-500',
  paused: 'bg-yellow-500',
  completed: 'bg-green-500',
  failed: 'bg-red-500',
  cancelled: 'bg-gray-500',
  waiting_input: 'bg-purple-500',
};

const STATUS_LABELS: Record<WorkflowExecutionStatus, string> = {
  idle: 'Idle',
  pending: 'Pending',
  running: 'Running',
  paused: 'Paused',
  completed: 'Completed',
  failed: 'Failed',
  cancelled: 'Cancelled',
  waiting_input: 'Waiting for input',
};

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${Math.floor(ms / 1000)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}:${seconds.toString().padStart(2, '0')}`;
}

export function ExecutionControls({
  onViewLogs,
  onInputRequired,
  className = '',
  enableKeyboardShortcuts = true,
}: ExecutionControlsProps) {
  const {
    status,
    currentStepIndex,
    totalSteps,
    startedAt,
    waitingForInput,
  } = useWorkflowExecutionStore();

  const {
    pauseWorkflow,
    resumeWorkflow,
    stopWorkflow,
  } = useWorkflowExecution();

  const [isPauseLoading, setIsPauseLoading] = useState(false);
  const [isResumeLoading, setIsResumeLoading] = useState(false);
  const [isStopLoading, setIsStopLoading] = useState(false);

  // Debounce refs
  const pauseTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const resumeTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const stopTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const isRunning = status === 'running';
  const isPaused = status === 'paused';
  const isWaitingInput = status === 'waiting_input';
  const isActive = isRunning || isPaused || isWaitingInput;
  const isIdle = status === 'idle';
  const isFinished = status === 'completed' || status === 'failed' || status === 'cancelled';

  const progressPercent = totalSteps > 0
    ? Math.round((currentStepIndex / totalSteps) * 100)
    : 0;

  const elapsed = startedAt ? Date.now() - startedAt : 0;

  // Debounced pause handler
  const handlePause = useCallback(() => {
    if (pauseTimeoutRef.current || isPauseLoading) return;

    setIsPauseLoading(true);
    pauseWorkflow();

    pauseTimeoutRef.current = setTimeout(() => {
      setIsPauseLoading(false);
      pauseTimeoutRef.current = null;
    }, 500);
  }, [pauseWorkflow, isPauseLoading]);

  // Debounced resume handler
  const handleResume = useCallback(() => {
    if (resumeTimeoutRef.current || isResumeLoading) return;

    setIsResumeLoading(true);
    resumeWorkflow();

    resumeTimeoutRef.current = setTimeout(() => {
      setIsResumeLoading(false);
      resumeTimeoutRef.current = null;
    }, 500);
  }, [resumeWorkflow, isResumeLoading]);

  // Debounced stop handler
  const handleStop = useCallback(() => {
    if (stopTimeoutRef.current || isStopLoading) return;

    setIsStopLoading(true);
    stopWorkflow();

    stopTimeoutRef.current = setTimeout(() => {
      setIsStopLoading(false);
      stopTimeoutRef.current = null;
    }, 500);
  }, [stopWorkflow, isStopLoading]);

  // Handle input click
  const handleInputClick = useCallback(() => {
    onInputRequired?.();
  }, [onInputRequired]);

  // Keyboard shortcuts
  useEffect(() => {
    if (!enableKeyboardShortcuts) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't trigger if user is typing in an input
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement ||
        e.target instanceof HTMLSelectElement
      ) {
        return;
      }

      // Space - Pause/Resume toggle
      if (e.code === 'Space' && !e.metaKey && !e.ctrlKey && !e.altKey) {
        e.preventDefault();
        if (isRunning) {
          handlePause();
        } else if (isPaused) {
          handleResume();
        }
      }

      // Escape - Stop
      if (e.code === 'Escape' && isActive) {
        e.preventDefault();
        handleStop();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [enableKeyboardShortcuts, isRunning, isPaused, isActive, handlePause, handleResume, handleStop]);

  // Cleanup timeouts on unmount
  useEffect(() => {
    return () => {
      if (pauseTimeoutRef.current) clearTimeout(pauseTimeoutRef.current);
      if (resumeTimeoutRef.current) clearTimeout(resumeTimeoutRef.current);
      if (stopTimeoutRef.current) clearTimeout(stopTimeoutRef.current);
    };
  }, []);

  return (
    <div className={`flex items-center gap-3 px-4 py-2 bg-gray-900 border-b border-gray-800 ${className}`}>
      {/* Status indicator */}
      <div className="flex items-center gap-2">
        <div className={`w-2 h-2 rounded-full ${STATUS_COLORS[status]} ${isRunning ? 'animate-pulse' : ''}`} />
        <span className="text-xs text-gray-400">{STATUS_LABELS[status]}</span>
      </div>

      {/* Input required badge */}
      {waitingForInput && (
        <button
          onClick={handleInputClick}
          className="flex items-center gap-1.5 px-2 py-1 bg-purple-500/20 border border-purple-500/50 rounded-full text-xs text-purple-300 hover:bg-purple-500/30 transition-colors animate-pulse"
        >
          <MessageSquare size={12} />
          <span>Input required</span>
        </button>
      )}

      {/* Divider */}
      <div className="h-4 w-px bg-gray-700" />

      {/* Control buttons */}
      <div className="flex items-center gap-1">
        {/* Play/Resume button */}
        {(isPaused || isWaitingInput) && (
          <button
            onClick={handleResume}
            disabled={isResumeLoading}
            className="p-1.5 rounded hover:bg-gray-800 text-green-400 hover:text-green-300 transition-colors disabled:opacity-50"
            title="Resume (Space)"
          >
            {isResumeLoading ? (
              <Loader2 size={16} className="animate-spin" />
            ) : (
              <Play size={16} fill="currentColor" />
            )}
          </button>
        )}

        {/* Pause button */}
        {isRunning && (
          <button
            onClick={handlePause}
            disabled={isPauseLoading}
            className="p-1.5 rounded hover:bg-gray-800 text-yellow-400 hover:text-yellow-300 transition-colors disabled:opacity-50"
            title="Pause (Space)"
          >
            {isPauseLoading ? (
              <Loader2 size={16} className="animate-spin" />
            ) : (
              <Pause size={16} fill="currentColor" />
            )}
          </button>
        )}

        {/* Stop button */}
        {isActive && (
          <button
            onClick={handleStop}
            disabled={isStopLoading}
            className="p-1.5 rounded hover:bg-gray-800 text-red-400 hover:text-red-300 transition-colors disabled:opacity-50"
            title="Stop (Esc)"
          >
            {isStopLoading ? (
              <Loader2 size={16} className="animate-spin" />
            ) : (
              <Square size={16} fill="currentColor" />
            )}
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
                  isWaitingInput ? 'bg-purple-500' :
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

      {/* Keyboard shortcuts hint */}
      {enableKeyboardShortcuts && isActive && (
        <div className="hidden lg:flex items-center gap-2 text-[10px] text-gray-600">
          <span>
            <kbd className="px-1 py-0.5 bg-gray-800 rounded">Space</kbd> {isRunning ? 'Pause' : 'Resume'}
          </span>
          <span>
            <kbd className="px-1 py-0.5 bg-gray-800 rounded">Esc</kbd> Stop
          </span>
        </div>
      )}

      {/* View logs button */}
      {onViewLogs && !isIdle && (
        <button
          onClick={onViewLogs}
          className="flex items-center gap-1 px-2 py-1 text-xs text-gray-400 hover:text-gray-300 hover:bg-gray-800 rounded transition-colors"
        >
          <FileText size={14} />
          Logs
        </button>
      )}
    </div>
  );
}

export default ExecutionControls;
