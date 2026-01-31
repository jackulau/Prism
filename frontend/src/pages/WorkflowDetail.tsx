import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Play,
  Pause,
  StopCircle,
  RefreshCw,
  Copy,
  Download,
  Clock,
  Layers,
  Loader2,
  CheckCircle,
  XCircle,
  AlertCircle,
} from 'lucide-react';
import { StepTimeline } from '../components/workflow/StepTimeline';
import { StepLogViewer } from '../components/workflow/StepLogViewer';
import { ConfirmDialog } from '../components/ConfirmDialog';
import {
  useWorkflow,
  useStartWorkflow,
  usePauseWorkflow,
  useResumeWorkflow,
  useCancelWorkflow,
  getWorkflowDuration,
  formatDuration,
  formatRelativeTime,
  type WorkflowStatus,
  type Step,
} from '../hooks/useWorkflows';
import { toast } from '../store/toastStore';

const STATUS_CONFIG: Record<
  WorkflowStatus,
  { icon: typeof CheckCircle; color: string; bgColor: string; label: string }
> = {
  pending: {
    icon: Clock,
    color: 'text-editor-muted',
    bgColor: 'bg-editor-muted/10',
    label: 'Pending',
  },
  running: {
    icon: Loader2,
    color: 'text-editor-warning',
    bgColor: 'bg-editor-warning/10',
    label: 'Running',
  },
  paused: {
    icon: Pause,
    color: 'text-yellow-500',
    bgColor: 'bg-yellow-500/10',
    label: 'Paused',
  },
  completed: {
    icon: CheckCircle,
    color: 'text-editor-success',
    bgColor: 'bg-editor-success/10',
    label: 'Completed',
  },
  failed: {
    icon: XCircle,
    color: 'text-editor-error',
    bgColor: 'bg-editor-error/10',
    label: 'Failed',
  },
  cancelled: {
    icon: AlertCircle,
    color: 'text-editor-muted',
    bgColor: 'bg-editor-muted/10',
    label: 'Cancelled',
  },
};

export default function WorkflowDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [selectedStep, setSelectedStep] = useState<Step | null>(null);
  const [showCancelConfirm, setShowCancelConfirm] = useState(false);
  const [showStateViewer, setShowStateViewer] = useState(false);

  const { data: workflow, isLoading, error } = useWorkflow(id);
  const startMutation = useStartWorkflow();
  const pauseMutation = usePauseWorkflow();
  const resumeMutation = useResumeWorkflow();
  const cancelMutation = useCancelWorkflow();

  const handleStart = async () => {
    if (!workflow) return;
    try {
      await startMutation.mutateAsync({ id: workflow.id });
      toast.success('Workflow started');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to start workflow');
    }
  };

  const handlePause = async () => {
    if (!workflow) return;
    try {
      await pauseMutation.mutateAsync(workflow.id);
      toast.success('Workflow paused');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to pause workflow');
    }
  };

  const handleResume = async () => {
    if (!workflow) return;
    try {
      await resumeMutation.mutateAsync(workflow.id);
      toast.success('Workflow resumed');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to resume workflow');
    }
  };

  const handleCancel = async () => {
    if (!workflow) return;
    try {
      await cancelMutation.mutateAsync(workflow.id);
      toast.success('Workflow cancelled');
      setShowCancelConfirm(false);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to cancel workflow');
    }
  };

  const handleExportJson = () => {
    if (!workflow) return;
    const data = JSON.stringify(workflow, null, 2);
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `workflow-${workflow.id}.json`;
    a.click();
    URL.revokeObjectURL(url);
    toast.success('Workflow exported');
  };

  const handleCopyDefinition = async () => {
    if (!workflow) return;
    try {
      await navigator.clipboard.writeText(JSON.stringify(workflow, null, 2));
      toast.success('Workflow copied to clipboard');
    } catch {
      toast.error('Failed to copy workflow');
    }
  };

  const handleStepClick = (stepId: string) => {
    const step = workflow?.steps.find((s) => s.id === stepId);
    if (step) {
      setSelectedStep(step);
    }
  };

  const isPending =
    startMutation.isPending ||
    pauseMutation.isPending ||
    resumeMutation.isPending ||
    cancelMutation.isPending;

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-editor-accent animate-spin" />
      </div>
    );
  }

  if (error || !workflow) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <XCircle className="w-12 h-12 text-editor-error mx-auto mb-4" />
          <h2 className="text-lg font-medium text-editor-text mb-2">
            Workflow not found
          </h2>
          <p className="text-editor-muted mb-4">
            {error instanceof Error ? error.message : 'The workflow could not be loaded.'}
          </p>
          <button
            onClick={() => navigate('/workflows')}
            className="px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            Back to Workflows
          </button>
        </div>
      </div>
    );
  }

  const statusConfig = STATUS_CONFIG[workflow.status];
  const StatusIcon = statusConfig.icon;
  const duration = getWorkflowDuration(workflow);
  const completedSteps = workflow.step_results?.filter(
    (r) => r.status === 'completed'
  ).length ?? 0;

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-start justify-between">
          <div className="flex items-start gap-4">
            <button
              onClick={() => navigate('/workflows')}
              className="p-2 text-editor-muted hover:text-editor-text rounded-lg hover:bg-editor-surface transition-colors"
            >
              <ArrowLeft size={20} />
            </button>

            <div className="space-y-2">
              <div className="flex items-center gap-3">
                <h1 className="text-2xl font-bold text-editor-text">
                  {workflow.name}
                </h1>
                <span
                  className={`inline-flex items-center gap-1.5 px-3 py-1 text-sm rounded-full ${statusConfig.bgColor} ${statusConfig.color}`}
                >
                  <StatusIcon
                    size={16}
                    className={workflow.status === 'running' ? 'animate-spin' : ''}
                  />
                  {statusConfig.label}
                </span>
              </div>

              {workflow.description && (
                <p className="text-editor-muted">{workflow.description}</p>
              )}

              {/* Meta Info */}
              <div className="flex items-center gap-4 text-sm text-editor-muted">
                <span className="flex items-center gap-1">
                  <Layers size={14} />
                  {completedSteps}/{workflow.steps.length} steps
                </span>
                <span>Created {formatRelativeTime(workflow.created_at)}</span>
                {workflow.started_at && (
                  <span>Started {formatRelativeTime(workflow.started_at)}</span>
                )}
                {duration !== null && (
                  <span className="flex items-center gap-1">
                    <Clock size={14} />
                    {formatDuration(duration)}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center gap-2">
            {workflow.status === 'pending' && (
              <button
                onClick={handleStart}
                disabled={isPending}
                className="flex items-center gap-2 px-4 py-2 bg-editor-success text-white rounded-lg hover:bg-editor-success/90 disabled:opacity-50 transition-colors"
              >
                {startMutation.isPending ? (
                  <Loader2 size={18} className="animate-spin" />
                ) : (
                  <Play size={18} />
                )}
                Start
              </button>
            )}

            {workflow.status === 'running' && (
              <button
                onClick={handlePause}
                disabled={isPending}
                className="flex items-center gap-2 px-4 py-2 bg-yellow-500 text-white rounded-lg hover:bg-yellow-500/90 disabled:opacity-50 transition-colors"
              >
                {pauseMutation.isPending ? (
                  <Loader2 size={18} className="animate-spin" />
                ) : (
                  <Pause size={18} />
                )}
                Pause
              </button>
            )}

            {workflow.status === 'paused' && (
              <button
                onClick={handleResume}
                disabled={isPending}
                className="flex items-center gap-2 px-4 py-2 bg-editor-success text-white rounded-lg hover:bg-editor-success/90 disabled:opacity-50 transition-colors"
              >
                {resumeMutation.isPending ? (
                  <Loader2 size={18} className="animate-spin" />
                ) : (
                  <Play size={18} />
                )}
                Resume
              </button>
            )}

            {(workflow.status === 'running' || workflow.status === 'paused') && (
              <button
                onClick={() => setShowCancelConfirm(true)}
                disabled={isPending}
                className="flex items-center gap-2 px-4 py-2 bg-editor-error text-white rounded-lg hover:bg-editor-error/90 disabled:opacity-50 transition-colors"
              >
                <StopCircle size={18} />
                Cancel
              </button>
            )}

            {(workflow.status === 'completed' || workflow.status === 'failed' || workflow.status === 'cancelled') && (
              <button
                onClick={handleStart}
                disabled={isPending}
                className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 transition-colors"
              >
                {startMutation.isPending ? (
                  <Loader2 size={18} className="animate-spin" />
                ) : (
                  <RefreshCw size={18} />
                )}
                Run Again
              </button>
            )}

            <button
              onClick={handleCopyDefinition}
              className="p-2 text-editor-muted hover:text-editor-text rounded-lg hover:bg-editor-surface transition-colors"
              title="Copy workflow definition"
            >
              <Copy size={18} />
            </button>

            <button
              onClick={handleExportJson}
              className="p-2 text-editor-muted hover:text-editor-text rounded-lg hover:bg-editor-surface transition-colors"
              title="Export as JSON"
            >
              <Download size={18} />
            </button>
          </div>
        </div>

        {/* Error Banner */}
        {workflow.error && (
          <div className="p-4 bg-editor-error/10 border border-editor-error/20 rounded-lg">
            <h3 className="text-sm font-medium text-editor-error mb-1">
              Workflow Error
            </h3>
            <p className="text-sm text-editor-error/80">{workflow.error}</p>
          </div>
        )}

        {/* Main Content */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Step Timeline */}
          <div className="lg:col-span-2">
            <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-medium text-editor-text">
                  Execution Timeline
                </h2>
                {workflow.status === 'running' && (
                  <span className="text-xs text-editor-muted animate-pulse">
                    Live updating...
                  </span>
                )}
              </div>

              {workflow.steps.length === 0 ? (
                <div className="text-center py-8 text-editor-muted">
                  No steps defined
                </div>
              ) : (
                <StepTimeline
                  steps={workflow.steps}
                  stepResults={workflow.step_results}
                  currentStep={workflow.current_step}
                  onStepClick={handleStepClick}
                />
              )}
            </div>
          </div>

          {/* Side Panel */}
          <div className="space-y-6">
            {/* Step Details */}
            {selectedStep ? (
              <StepLogViewer
                step={selectedStep}
                result={workflow.step_results?.find(
                  (r) => r.step_id === selectedStep.id
                )}
                inputData={workflow.state}
                onClose={() => setSelectedStep(null)}
              />
            ) : (
              <div className="bg-editor-surface border border-editor-border rounded-lg p-6 text-center">
                <Layers className="w-10 h-10 text-editor-muted mx-auto mb-3" />
                <p className="text-sm text-editor-muted">
                  Click a step to view details
                </p>
              </div>
            )}

            {/* State Viewer */}
            <div className="bg-editor-surface border border-editor-border rounded-lg overflow-hidden">
              <button
                onClick={() => setShowStateViewer(!showStateViewer)}
                className="w-full flex items-center justify-between p-4 hover:bg-editor-hover/50 transition-colors"
              >
                <h3 className="text-sm font-medium text-editor-text">
                  Workflow State
                </h3>
                <span className="text-xs text-editor-muted">
                  {showStateViewer ? 'Hide' : 'Show'}
                </span>
              </button>

              {showStateViewer && (
                <div className="px-4 pb-4">
                  {workflow.state && Object.keys(workflow.state).length > 0 ? (
                    <pre className="text-xs text-editor-text bg-editor-bg p-3 rounded-lg overflow-x-auto max-h-48 overflow-y-auto">
                      {JSON.stringify(workflow.state, null, 2)}
                    </pre>
                  ) : (
                    <p className="text-sm text-editor-muted text-center py-4">
                      No state data
                    </p>
                  )}
                </div>
              )}
            </div>

            {/* Workflow Info */}
            <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
              <h3 className="text-sm font-medium text-editor-text mb-3">
                Workflow Info
              </h3>
              <dl className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <dt className="text-editor-muted">ID</dt>
                  <dd className="text-editor-text font-mono text-xs truncate ml-4 max-w-[200px]">
                    {workflow.id}
                  </dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-editor-muted">Created</dt>
                  <dd className="text-editor-text">
                    {new Date(workflow.created_at).toLocaleString()}
                  </dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-editor-muted">Updated</dt>
                  <dd className="text-editor-text">
                    {new Date(workflow.updated_at).toLocaleString()}
                  </dd>
                </div>
                {workflow.started_at && (
                  <div className="flex justify-between">
                    <dt className="text-editor-muted">Started</dt>
                    <dd className="text-editor-text">
                      {new Date(workflow.started_at).toLocaleString()}
                    </dd>
                  </div>
                )}
                {workflow.completed_at && (
                  <div className="flex justify-between">
                    <dt className="text-editor-muted">Completed</dt>
                    <dd className="text-editor-text">
                      {new Date(workflow.completed_at).toLocaleString()}
                    </dd>
                  </div>
                )}
              </dl>
            </div>
          </div>
        </div>
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
    </div>
  );
}
