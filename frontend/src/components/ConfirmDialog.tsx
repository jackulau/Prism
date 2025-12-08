import { AlertTriangle, X } from 'lucide-react';

interface ConfirmDialogProps {
  isOpen: boolean;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'danger' | 'warning';
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmDialog({
  isOpen,
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  variant = 'danger',
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 z-50 animate-fade-in"
        onClick={onCancel}
      />

      {/* Dialog */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 p-6 animate-slide-in-right">
        <div className="flex items-start gap-3 mb-4">
          <div
            className={`p-2 rounded-lg ${
              variant === 'danger' ? 'bg-red-500/20' : 'bg-yellow-500/20'
            }`}
          >
            <AlertTriangle
              className={variant === 'danger' ? 'text-red-400' : 'text-yellow-400'}
              size={24}
            />
          </div>
          <div className="flex-1">
            <h3 className="text-lg font-semibold text-editor-text">{title}</h3>
            <p className="text-sm text-editor-muted mt-1">{message}</p>
          </div>
          <button
            onClick={onCancel}
            className="p-1 hover:bg-editor-surface rounded transition-colors"
          >
            <X size={20} className="text-editor-muted" />
          </button>
        </div>

        <div className="flex justify-end gap-2">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            {cancelText}
          </button>
          <button
            onClick={onConfirm}
            className={`px-4 py-2 text-sm text-white rounded-lg transition-colors ${
              variant === 'danger'
                ? 'bg-red-500 hover:bg-red-600'
                : 'bg-yellow-500 hover:bg-yellow-600'
            }`}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </>
  );
}
