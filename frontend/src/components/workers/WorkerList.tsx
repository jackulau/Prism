import { useState } from 'react';
import { Bot, Pause } from 'lucide-react';
import { trpc } from '../../lib/trpc';
import { ConfirmDialog } from '../ConfirmDialog';
import { toast } from '../../store/toastStore';

export function WorkerList() {
  const [deleteId, setDeleteId] = useState<string | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: executions, isLoading, refetch } = (trpc as any).workers.listExecutions.useQuery();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const cancelMutation = (trpc as any).workers.cancelExecution.useMutation({
    onSuccess: () => {
      toast.success('Execution cancelled');
      refetch();
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      toast.error(error.message);
    },
  });

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse"
          >
            <div className="flex items-center gap-4">
              <div className="w-10 h-10 bg-editor-border rounded-lg" />
              <div className="flex-1 space-y-2">
                <div className="h-4 bg-editor-border rounded w-1/4" />
                <div className="h-3 bg-editor-border rounded w-1/2" />
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  const executionList = executions || [];

  if (executionList.length === 0) {
    return (
      <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
        <Bot className="w-12 h-12 text-editor-muted mx-auto mb-4" />
        <h3 className="text-lg font-medium text-editor-text mb-2">No worker executions</h3>
        <p className="text-editor-muted">
          Run a task to see worker executions here
        </p>
      </div>
    );
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'bg-editor-warning/10 text-editor-warning';
      case 'completed':
        return 'bg-editor-success/10 text-editor-success';
      case 'failed':
      case 'cancelled':
        return 'bg-editor-error/10 text-editor-error';
      default:
        return 'bg-editor-muted/10 text-editor-muted';
    }
  };

  return (
    <>
      <div className="space-y-4">
        {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
        {executionList.map((execution: any) => (
          <div
            key={execution.id}
            className="bg-editor-surface border border-editor-border rounded-lg p-4 hover:border-editor-accent/30 transition-colors"
          >
            <div className="flex items-start justify-between">
              <div className="flex items-start gap-4">
                <div className="p-2 bg-editor-accent/10 rounded-lg">
                  <Bot className="w-5 h-5 text-editor-accent" />
                </div>
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <h3 className="font-medium text-editor-text">
                      {execution.task || 'Task Execution'}
                    </h3>
                    <span
                      className={`px-2 py-0.5 text-xs rounded-full ${getStatusColor(
                        execution.status
                      )}`}
                    >
                      {execution.status}
                    </span>
                  </div>
                  <p className="text-sm text-editor-muted">
                    ID: {execution.id}
                  </p>
                  {execution.startedAt && (
                    <p className="text-xs text-editor-muted">
                      Started: {new Date(execution.startedAt).toLocaleString()}
                    </p>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-2">
                {execution.status === 'running' && (
                  <button
                    onClick={() => cancelMutation.mutate({ executionId: execution.id })}
                    disabled={cancelMutation.isPending}
                    className="p-2 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
                    title="Cancel execution"
                  >
                    <Pause size={18} />
                  </button>
                )}
              </div>
            </div>

            {/* Result/Error display */}
            {execution.result && (
              <div className="mt-4 p-3 bg-editor-bg rounded-lg">
                <p className="text-xs text-editor-muted mb-1">Result:</p>
                <pre className="text-sm text-editor-text whitespace-pre-wrap overflow-x-auto">
                  {typeof execution.result === 'string'
                    ? execution.result
                    : JSON.stringify(execution.result, null, 2)}
                </pre>
              </div>
            )}

            {execution.error && (
              <div className="mt-4 p-3 bg-editor-error/10 rounded-lg border border-editor-error/20">
                <p className="text-xs text-editor-error mb-1">Error:</p>
                <pre className="text-sm text-editor-error whitespace-pre-wrap">
                  {execution.error}
                </pre>
              </div>
            )}
          </div>
        ))}
      </div>

      <ConfirmDialog
        isOpen={deleteId !== null}
        title="Cancel Execution"
        message="Are you sure you want to cancel this execution?"
        confirmText="Cancel Execution"
        variant="danger"
        onConfirm={() => {
          if (deleteId) {
            cancelMutation.mutate({ executionId: deleteId });
          }
          setDeleteId(null);
        }}
        onCancel={() => setDeleteId(null)}
      />
    </>
  );
}
