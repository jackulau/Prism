import { useState, useEffect } from 'react';
import { X, Plus, Save } from 'lucide-react';
import type { z } from 'zod';
import type { BatchTask, BatchTaskFormData, BatchTaskPriority } from '../../types/batch';
import { batchTaskFormSchema } from '../../schemas/batchTask';

interface AddTaskModalProps {
  isOpen: boolean;
  editingTask?: BatchTask | null;
  onClose: () => void;
  onSave: (data: BatchTaskFormData) => void;
}

const priorityOptions: { value: BatchTaskPriority; label: string }[] = [
  { value: 'low', label: 'Low' },
  { value: 'normal', label: 'Normal' },
  { value: 'high', label: 'High' },
];

export function AddTaskModal({ isOpen, editingTask, onClose, onSave }: AddTaskModalProps) {
  const [prompt, setPrompt] = useState('');
  const [context, setContext] = useState('');
  const [priority, setPriority] = useState<BatchTaskPriority>('normal');
  const [errors, setErrors] = useState<{ prompt?: string; context?: string }>({});

  const isEditing = !!editingTask;

  useEffect(() => {
    if (editingTask) {
      setPrompt(editingTask.prompt);
      setContext(editingTask.context || '');
      setPriority(editingTask.priority);
    } else {
      setPrompt('');
      setContext('');
      setPriority('normal');
    }
    setErrors({});
  }, [editingTask, isOpen]);

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const result = batchTaskFormSchema.safeParse({
      prompt,
      context: context || undefined,
      priority,
    });

    if (!result.success) {
      const fieldErrors: { prompt?: string; context?: string } = {};
      result.error.errors.forEach((err: z.ZodIssue) => {
        const field = err.path[0] as string;
        if (field === 'prompt' || field === 'context') {
          fieldErrors[field] = err.message;
        }
      });
      setErrors(fieldErrors);
      return;
    }

    onSave(result.data);
    onClose();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    }
  };

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 z-40"
        onClick={onClose}
        onKeyDown={handleKeyDown}
        role="button"
        tabIndex={-1}
        aria-label="Close modal"
      />

      {/* Modal */}
      <div
        className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[520px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50"
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border">
          <h2 id="modal-title" className="text-lg font-semibold text-editor-text">
            {isEditing ? 'Edit Task' : 'Add Task'}
          </h2>
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
            aria-label="Close"
          >
            <X size={20} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Prompt */}
          <div className="space-y-2">
            <label htmlFor="prompt" className="block text-sm font-medium text-editor-text">
              Prompt <span className="text-red-400">*</span>
            </label>
            <textarea
              id="prompt"
              value={prompt}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => {
                setPrompt(e.target.value);
                if (errors.prompt) setErrors((prev: typeof errors) => ({ ...prev, prompt: undefined }));
              }}
              placeholder="Enter the task prompt..."
              rows={4}
              className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none ${
                errors.prompt ? 'border-red-400' : 'border-editor-border'
              }`}
              autoFocus
            />
            {errors.prompt && (
              <p className="text-sm text-red-400">{errors.prompt}</p>
            )}
          </div>

          {/* Context */}
          <div className="space-y-2">
            <label htmlFor="context" className="block text-sm font-medium text-editor-text">
              Context <span className="text-editor-muted">(optional)</span>
            </label>
            <textarea
              id="context"
              value={context}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => {
                setContext(e.target.value);
                if (errors.context) setErrors((prev: typeof errors) => ({ ...prev, context: undefined }));
              }}
              placeholder="Additional context or instructions..."
              rows={2}
              className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none ${
                errors.context ? 'border-red-400' : 'border-editor-border'
              }`}
            />
            {errors.context && (
              <p className="text-sm text-red-400">{errors.context}</p>
            )}
          </div>

          {/* Priority */}
          <div className="space-y-2">
            <label htmlFor="priority" className="block text-sm font-medium text-editor-text">
              Priority
            </label>
            <select
              id="priority"
              value={priority}
              onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setPriority(e.target.value as BatchTaskPriority)}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
            >
              {priorityOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
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
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            {isEditing ? (
              <>
                <Save size={16} />
                Save Changes
              </>
            ) : (
              <>
                <Plus size={16} />
                Add Task
              </>
            )}
          </button>
        </div>
      </div>
    </>
  );
}
