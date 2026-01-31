import { useState, useEffect, useRef, useCallback } from 'react';
import { X, Loader2, AlertCircle, Clock } from 'lucide-react';
import type { InputRequest } from '../../hooks/useWorkflowInput';

interface UserInputModalProps {
  isOpen: boolean;
  request: InputRequest | null;
  remainingTime: number | null;
  isSubmitting: boolean;
  error: Error | null;
  onSubmit: (input: string) => void;
  onCancel: () => void;
  onClearError: () => void;
}

function formatTime(ms: number): string {
  const seconds = Math.ceil(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
}

export function UserInputModal({
  isOpen,
  request,
  remainingTime,
  isSubmitting,
  error,
  onSubmit,
  onCancel,
  onClearError,
}: UserInputModalProps) {
  const [inputValue, setInputValue] = useState('');
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // Auto-focus input when modal opens
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen]);

  // Reset input value when modal opens with new request
  useEffect(() => {
    if (isOpen && request) {
      setInputValue('');
    }
  }, [isOpen, request?.stepId]);

  const handleSubmit = useCallback((e: React.FormEvent) => {
    e.preventDefault();
    if (inputValue.trim() && !isSubmitting) {
      onSubmit(inputValue.trim());
    }
  }, [inputValue, isSubmitting, onSubmit]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && e.metaKey) {
      e.preventDefault();
      if (inputValue.trim() && !isSubmitting) {
        onSubmit(inputValue.trim());
      }
    }
    if (e.key === 'Escape') {
      e.preventDefault();
      onCancel();
    }
  }, [inputValue, isSubmitting, onSubmit, onCancel]);

  if (!isOpen || !request) return null;

  const isTimeout = remainingTime === 0;
  const isWarning = remainingTime !== null && remainingTime > 0 && remainingTime <= 30000;

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/60 z-50 animate-fade-in"
        onClick={onCancel}
        aria-hidden="true"
      />

      {/* Modal */}
      <div
        className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] max-w-[90vw] bg-gray-900 border border-gray-700 rounded-xl shadow-2xl z-50"
        role="dialog"
        aria-modal="true"
        aria-labelledby="input-modal-title"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-800">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-500/20 rounded-lg">
              <svg className="w-5 h-5 text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
              </svg>
            </div>
            <h2 id="input-modal-title" className="text-lg font-semibold text-white">
              Input Required
            </h2>
          </div>
          <div className="flex items-center gap-3">
            {/* Timeout countdown */}
            {remainingTime !== null && remainingTime > 0 && (
              <div className={`flex items-center gap-1.5 px-2 py-1 rounded text-xs ${
                isWarning ? 'bg-yellow-500/20 text-yellow-400' : 'bg-gray-800 text-gray-400'
              }`}>
                <Clock size={12} />
                <span className="font-mono">{formatTime(remainingTime)}</span>
              </div>
            )}
            <button
              onClick={onCancel}
              className="p-1.5 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg transition-colors"
              aria-label="Close"
            >
              <X size={18} />
            </button>
          </div>
        </div>

        {/* Content */}
        <form onSubmit={handleSubmit}>
          <div className="p-6 space-y-4">
            {/* Prompt text */}
            <div className="p-4 bg-gray-800/50 rounded-lg border border-gray-700">
              <p className="text-sm text-gray-200 whitespace-pre-wrap">
                {request.promptText}
              </p>
            </div>

            {/* Timeout message */}
            {isTimeout && (
              <div className="flex items-center gap-2 p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
                <AlertCircle size={16} className="text-red-400 shrink-0" />
                <p className="text-sm text-red-300">
                  This input request has timed out. You can try submitting anyway or cancel.
                </p>
              </div>
            )}

            {/* Error message */}
            {error && (
              <div className="flex items-center gap-2 p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
                <AlertCircle size={16} className="text-red-400 shrink-0" />
                <p className="text-sm text-red-300 flex-1">{error.message}</p>
                <button
                  type="button"
                  onClick={onClearError}
                  className="text-xs text-red-400 hover:text-red-300"
                >
                  Dismiss
                </button>
              </div>
            )}

            {/* Input field */}
            <div className="space-y-2">
              <label htmlFor="workflow-input" className="block text-sm font-medium text-gray-300">
                Your response
              </label>
              <textarea
                ref={inputRef}
                id="workflow-input"
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Type your response here..."
                rows={4}
                disabled={isSubmitting}
                className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-purple-500 focus:ring-1 focus:ring-purple-500 resize-none disabled:opacity-50"
              />
              <p className="text-xs text-gray-500">
                Press <kbd className="px-1.5 py-0.5 bg-gray-800 rounded text-gray-400">Cmd+Enter</kbd> to submit
              </p>
            </div>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-800 bg-gray-900/50">
            <button
              type="button"
              onClick={onCancel}
              disabled={isSubmitting}
              className="px-4 py-2 text-sm text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!inputValue.trim() || isSubmitting}
              className="flex items-center gap-2 px-4 py-2 text-sm text-white bg-purple-600 hover:bg-purple-500 rounded-lg transition-colors disabled:opacity-50 disabled:hover:bg-purple-600"
            >
              {isSubmitting ? (
                <>
                  <Loader2 size={14} className="animate-spin" />
                  Submitting...
                </>
              ) : (
                'Submit'
              )}
            </button>
          </div>
        </form>
      </div>
    </>
  );
}

export default UserInputModal;
