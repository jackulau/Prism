import { useCallback } from 'react';
import { Pause, Play, MessageSquare, X } from 'lucide-react';

type PauseReason = 'manual' | 'waiting_input' | 'breakpoint';

interface PausedBannerProps {
  isVisible: boolean;
  reason: PauseReason;
  onResume: () => void;
  onDismiss?: () => void;
  onOpenInputModal?: () => void;
  isResuming?: boolean;
  className?: string;
}

const REASON_MESSAGES: Record<PauseReason, { title: string; description: string }> = {
  manual: {
    title: 'Workflow paused',
    description: 'The workflow has been paused. Click Resume to continue execution.',
  },
  waiting_input: {
    title: 'Awaiting input',
    description: 'The workflow is waiting for your input to continue.',
  },
  breakpoint: {
    title: 'Breakpoint reached',
    description: 'Execution stopped at a breakpoint. Resume when ready.',
  },
};

export function PausedBanner({
  isVisible,
  reason,
  onResume,
  onDismiss,
  onOpenInputModal,
  isResuming = false,
  className = '',
}: PausedBannerProps) {
  const handleResumeClick = useCallback(() => {
    if (!isResuming) {
      onResume();
    }
  }, [isResuming, onResume]);

  const handleInputClick = useCallback(() => {
    onOpenInputModal?.();
  }, [onOpenInputModal]);

  if (!isVisible) return null;

  const { title, description } = REASON_MESSAGES[reason];
  const isWaitingInput = reason === 'waiting_input';

  return (
    <div
      className={`flex items-center gap-3 px-4 py-3 bg-gradient-to-r from-yellow-500/10 to-orange-500/10 border-b border-yellow-500/30 ${className}`}
      role="alert"
      aria-live="polite"
    >
      {/* Icon */}
      <div className={`p-2 rounded-lg ${isWaitingInput ? 'bg-purple-500/20' : 'bg-yellow-500/20'}`}>
        {isWaitingInput ? (
          <MessageSquare size={18} className="text-purple-400" />
        ) : (
          <Pause size={18} className="text-yellow-400" />
        )}
      </div>

      {/* Message */}
      <div className="flex-1 min-w-0">
        <p className={`text-sm font-medium ${isWaitingInput ? 'text-purple-300' : 'text-yellow-300'}`}>
          {title}
        </p>
        <p className="text-xs text-gray-400 truncate">
          {description}
        </p>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2">
        {isWaitingInput && onOpenInputModal && (
          <button
            onClick={handleInputClick}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-purple-600 hover:bg-purple-500 rounded-lg transition-colors"
          >
            <MessageSquare size={14} />
            Provide Input
          </button>
        )}

        {!isWaitingInput && (
          <button
            onClick={handleResumeClick}
            disabled={isResuming}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-gray-900 bg-yellow-400 hover:bg-yellow-300 rounded-lg transition-colors disabled:opacity-50"
          >
            {isResuming ? (
              <>
                <div className="w-3 h-3 border-2 border-gray-900/30 border-t-gray-900 rounded-full animate-spin" />
                Resuming...
              </>
            ) : (
              <>
                <Play size={14} />
                Resume
              </>
            )}
          </button>
        )}

        {onDismiss && (
          <button
            onClick={onDismiss}
            className="p-1.5 text-gray-400 hover:text-white hover:bg-gray-700/50 rounded-lg transition-colors"
            aria-label="Dismiss"
          >
            <X size={16} />
          </button>
        )}
      </div>
    </div>
  );
}

export default PausedBanner;
