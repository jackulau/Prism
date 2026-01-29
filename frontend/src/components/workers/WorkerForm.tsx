import { useState } from 'react';
import { X, Bot, Play, Loader2 } from 'lucide-react';
import { trpc } from '../../lib/trpc';
import { toast } from '../../store/toastStore';

interface WorkerFormProps {
  workerId?: string | null;
  onClose: () => void;
}

export function WorkerForm({ workerId, onClose }: WorkerFormProps) {
  const [task, setTask] = useState('');
  const [model, setModel] = useState('claude-3-sonnet-20240229');
  const [maxTokens, setMaxTokens] = useState(4096);
  const [temperature, setTemperature] = useState(0.7);

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const utils = (trpc as any).useUtils();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const runTaskMutation = (trpc as any).workers.runTask.useMutation({
    onSuccess: () => {
      toast.success('Task started successfully');
      utils.workers.listExecutions.invalidate();
      onClose();
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      toast.error(error.message);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!task.trim()) {
      toast.error('Please enter a task');
      return;
    }

    runTaskMutation.mutate({
      task,
      config: {
        model,
        maxTokens,
        temperature,
      },
    });
  };

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 z-40"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] max-w-[90vw] max-h-[90vh] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50 overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-editor-accent/10 rounded-lg">
              <Bot className="w-5 h-5 text-editor-accent" />
            </div>
            <h2 className="text-lg font-semibold text-editor-text">
              {workerId ? 'Edit Worker' : 'Run New Task'}
            </h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-6 overflow-y-auto max-h-[calc(90vh-140px)]">
          {/* Task */}
          <div className="space-y-2">
            <label className="block text-sm font-medium text-editor-text">
              Task Description
            </label>
            <textarea
              value={task}
              onChange={(e) => setTask(e.target.value)}
              placeholder="Describe the task you want the worker to perform..."
              rows={4}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none"
            />
          </div>

          {/* Model Selection */}
          <div className="space-y-2">
            <label className="block text-sm font-medium text-editor-text">
              Model
            </label>
            <select
              value={model}
              onChange={(e) => setModel(e.target.value)}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
            >
              <option value="claude-3-opus-20240229">Claude 3 Opus</option>
              <option value="claude-3-sonnet-20240229">Claude 3 Sonnet</option>
              <option value="claude-3-haiku-20240307">Claude 3 Haiku</option>
              <option value="gpt-4-turbo-preview">GPT-4 Turbo</option>
              <option value="gpt-4">GPT-4</option>
              <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
            </select>
          </div>

          {/* Advanced Settings */}
          <div className="space-y-4">
            <h3 className="text-sm font-medium text-editor-text">Advanced Settings</h3>

            <div className="grid grid-cols-2 gap-4">
              {/* Max Tokens */}
              <div className="space-y-2">
                <label className="block text-xs text-editor-muted">
                  Max Tokens
                </label>
                <input
                  type="number"
                  value={maxTokens}
                  onChange={(e) => setMaxTokens(Number(e.target.value))}
                  min={1}
                  max={100000}
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
                />
              </div>

              {/* Temperature */}
              <div className="space-y-2">
                <label className="block text-xs text-editor-muted">
                  Temperature ({temperature.toFixed(1)})
                </label>
                <input
                  type="range"
                  value={temperature}
                  onChange={(e) => setTemperature(Number(e.target.value))}
                  min={0}
                  max={2}
                  step={0.1}
                  className="w-full"
                />
              </div>
            </div>
          </div>
        </form>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-editor-border bg-editor-surface/50">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={runTaskMutation.isPending || !task.trim()}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {runTaskMutation.isPending ? (
              <>
                <Loader2 size={18} className="animate-spin" />
                Running...
              </>
            ) : (
              <>
                <Play size={18} />
                Run Task
              </>
            )}
          </button>
        </div>
      </div>
    </>
  );
}
