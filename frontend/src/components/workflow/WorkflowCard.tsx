import { useNavigate } from 'react-router-dom';
import {
  Play,
  Pause,
  StopCircle,
  MoreVertical,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2,
  Layers,
} from 'lucide-react';
import { useState } from 'react';
import type { Workflow, WorkflowStatus } from '../../hooks/useWorkflows';
import {
  useStartWorkflow,
  usePauseWorkflow,
  useResumeWorkflow,
  useCancelWorkflow,
  getWorkflowDuration,
  formatDuration,
  formatRelativeTime,
} from '../../hooks/useWorkflows';
import { ConfirmDialog } from '../ConfirmDialog';
import { toast } from '../../store/toastStore';

interface WorkflowCardProps {
  workflow: Workflow;
}

const STATUS_CONFIG: Record<
  WorkflowStatus,
  { icon: typeof CheckCircle; color: string; bgColor: string }
> = {
  pending: {
    icon: Clock,
    color: 'text-editor-muted',
    bgColor: 'bg-editor-muted/10',
  },
  running: {
    icon: Loader2,
    color: 'text-editor-warning',
    bgColor: 'bg-editor-warning/10',
  },
  paused: {
    icon: Pause,
    color: 'text-yellow-500',
    bgColor: 'bg-yellow-500/10',
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
  cancelled: {
    icon: AlertCircle,
    color: 'text-editor-muted',
    bgColor: 'bg-editor-muted/10',
  },
};

export function WorkflowCard({ workflow }: WorkflowCardProps) {
  const navigate = useNavigate();
  const [showMenu, setShowMenu] = useState(false);
  const [showCancelConfirm, setShowCancelConfirm] = useState(false);

  const startMutation = useStartWorkflow();
  const pauseMutation = usePauseWorkflow();
  const resumeMutation = useResumeWorkflow();
  const cancelMutation = useCancelWorkflow();

  const statusConfig = STATUS_CONFIG[workflow.status];
  const StatusIcon = statusConfig.icon;
  const duration = getWorkflowDuration(workflow);
  const completedSteps = workflow.step_results?.filter(
    (r) => r.status === 'completed'
  ).length ?? 0;
  const totalSteps = workflow.steps.length;

  const handleStart = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await startMutation.mutateAsync({ id: workflow.id });
      toast.success('Workflow started');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to start workflow');
    }
  };

  const handlePause = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await pauseMutation.mutateAsync(workflow.id);
      toast.success('Workflow paused');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to pause workflow');
    }
  };

  const handleResume = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await resumeMutation.mutateAsync(workflow.id);
      toast.success('Workflow resumed');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to resume workflow');
    }
  };

  const handleCancel = async () => {
    try {
      await cancelMutation.mutateAsync(workflow.id);
      toast.success('Workflow cancelled');
      setShowCancelConfirm(false);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to cancel workflow');
    }
  };

  const isPending =
    startMutation.isPending ||
    pauseMutation.isPending ||
    resumeMutation.isPending ||
    cancelMutation.isPending;

  return (
    <>
      <div
        onClick={() => navigate(`/workflows/${workflow.id}`)}
        className="bg-editor-surface border border-editor-border rounded-lg p-4 hover:border-editor-accent/30 transition-colors cursor-pointer"
      >
        <div className="flex items-start justify-between">
          {/* Main Info */}
          <div className="flex items-start gap-4 flex-1 min-w-0">
            {/* Icon */}
            <div className={`p-2 rounded-lg ${statusConfig.bgColor}`}>
              <Layers size={20} className={statusConfig.color} />
            </div>

            {/* Content */}
            <div className="flex-1 min-w-0 space-y-1">
              <div className="flex items-center gap-2">
                <h3 className="font-medium text-editor-text truncate">
                  {workflow.name}
                </h3>
                <span
                  className={`inline-flex items-center gap-1 px-2 py-0.5 text-xs rounded-full ${statusConfig.bgColor} ${statusConfig.color}`}
                >
                  <StatusIcon
                    size={12}
                    className={workflow.status === 'running' ? 'animate-spin' : ''}
                  />
                  {workflow.status}
                </span>
              </div>

              {workflow.description && (
                <p className="text-sm text-editor-muted truncate">
                  {workflow.description}
                </p>
              )}

              {/* Meta Info */}
              <div className="flex items-center gap-4 text-xs text-editor-muted">
                <span className="flex items-center gap-1">
                  <Layers size={12} />
                  {completedSteps}/{totalSteps} steps
                </span>
                <span>{formatRelativeTime(workflow.created_at)}</span>
                {duration !== null && (
                  <span className="flex items-center gap-1">
                    <Clock size={12} />
                    {formatDuration(duration)}
                  </span>
                )}
              </div>

              {/* Error Message */}
              {workflow.error && (
                <div className="mt-2 p-2 bg-editor-error/10 border border-editor-error/20 rounded text-xs text-editor-error">
                  {workflow.error}
                </div>
              )}
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center gap-1 ml-4">
            {workflow.status === 'pending' && (
              <button
                onClick={handleStart}
                disabled={isPending}
                className="p-2 text-editor-muted hover:text-editor-success hover:bg-editor-success/10 rounded-lg transition-colors"
                title="Start workflow"
              >
                {startMutation.isPending ? (
                  <Loader2 size={18} className="animate-spin" />
                ) : (
                  <Play size={18} />
                )}
              </button>
            )}

            {workflow.status === 'running' && (
              <button
                onClick={handlePause}
                disabled={isPending}
                className="p-2 text-editor-muted hover:text-yellow-500 hover:bg-yellow-500/10 rounded-lg transition-colors"
                title="Pause workflow"
              >
                {pauseMutation.isPending ? (
                  <Loader2 size={18} className="animate-spin" />
                ) : (
                  <Pause size={18} />
                )}
              </button>
            )}

            {workflow.status === 'paused' && (
              <button
                onClick={handleResume}
                disabled={isPending}
                className="p-2 text-editor-muted hover:text-editor-success hover:bg-editor-success/10 rounded-lg transition-colors"
                title="Resume workflow"
              >
                {resumeMutation.isPending ? (
                  <Loader2 size={18} className="animate-spin" />
                ) : (
                  <Play size={18} />
                )}
              </button>
            )}

            {(workflow.status === 'running' || workflow.status === 'paused') && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  setShowCancelConfirm(true);
                }}
                disabled={isPending}
                className="p-2 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
                title="Cancel workflow"
              >
                <StopCircle size={18} />
              </button>
            )}

            {/* More Menu */}
            <div className="relative">
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  setShowMenu(!showMenu);
                }}
                className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-hover rounded-lg transition-colors"
              >
                <MoreVertical size={18} />
              </button>

              {showMenu && (
                <>
                  <div
                    className="fixed inset-0 z-10"
                    onClick={(e) => {
                      e.stopPropagation();
                      setShowMenu(false);
                    }}
                  />
                  <div className="absolute right-0 mt-1 w-40 bg-editor-surface border border-editor-border rounded-lg shadow-lg z-20 overflow-hidden">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate(`/workflows/${workflow.id}`);
                      }}
                      className="w-full px-4 py-2 text-sm text-left text-editor-muted hover:text-editor-text hover:bg-editor-hover transition-colors"
                    >
                      View Details
                    </button>
                    {workflow.status !== 'running' && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleStart(e);
                          setShowMenu(false);
                        }}
                        className="w-full px-4 py-2 text-sm text-left text-editor-muted hover:text-editor-text hover:bg-editor-hover transition-colors"
                      >
                        Run Again
                      </button>
                    )}
                  </div>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Progress Bar for running workflows */}
        {workflow.status === 'running' && totalSteps > 0 && (
          <div className="mt-4">
            <div className="h-1 bg-editor-border rounded-full overflow-hidden">
              <div
                className="h-full bg-editor-accent transition-all duration-500"
                style={{ width: `${(completedSteps / totalSteps) * 100}%` }}
              />
            </div>
          </div>
        )}
      </div>

      <ConfirmDialog
        isOpen={showCancelConfirm}
        title="Cancel Workflow"
        message="Are you sure you want to cancel this workflow? This action cannot be undone."
        confirmText="Cancel Workflow"
        variant="danger"
        onConfirm={handleCancel}
        onCancel={() => setShowCancelConfirm(false)}
      />
    </>
  );
}
