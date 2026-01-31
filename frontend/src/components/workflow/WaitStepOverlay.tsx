import { useState, useEffect, useCallback } from 'react';
import { MessageSquare, Clock } from 'lucide-react';

interface WaitStepOverlayProps {
  isWaiting: boolean;
  stepId: string;
  promptText?: string;
  waitStartTime?: number;
  timeout?: number;
  onClick?: () => void;
  className?: string;
}

function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
}

export function WaitStepOverlay({
  isWaiting,
  stepId: _stepId,
  promptText,
  waitStartTime,
  timeout,
  onClick,
  className = '',
}: WaitStepOverlayProps) {
  // stepId is available for future use (e.g., data attributes, debugging)
  void _stepId;
  const [waitDuration, setWaitDuration] = useState(0);
  const [remainingTime, setRemainingTime] = useState<number | null>(null);

  // Update wait duration timer
  useEffect(() => {
    if (!isWaiting || !waitStartTime) {
      setWaitDuration(0);
      setRemainingTime(null);
      return;
    }

    const updateDuration = () => {
      const elapsed = Date.now() - waitStartTime;
      setWaitDuration(elapsed);

      if (timeout) {
        const remaining = Math.max(0, timeout - elapsed);
        setRemainingTime(remaining);
      }
    };

    updateDuration();
    const interval = setInterval(updateDuration, 1000);

    return () => clearInterval(interval);
  }, [isWaiting, waitStartTime, timeout]);

  const handleClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    onClick?.();
  }, [onClick]);

  if (!isWaiting) return null;

  const isTimeout = remainingTime === 0;
  const isWarning = remainingTime !== null && remainingTime > 0 && remainingTime <= 30000;

  return (
    <div
      className={`absolute inset-0 bg-purple-500/10 border-2 border-purple-500/50 rounded-lg flex flex-col items-center justify-center cursor-pointer transition-all hover:bg-purple-500/20 ${className}`}
      onClick={handleClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && handleClick(e as unknown as React.MouseEvent)}
      aria-label="Click to provide input"
    >
      {/* Pulsing attention indicator */}
      <div className="absolute -top-2 -right-2">
        <div className="relative">
          <div className="absolute inset-0 w-4 h-4 bg-purple-500 rounded-full animate-ping" />
          <div className="relative w-4 h-4 bg-purple-500 rounded-full flex items-center justify-center">
            <span className="text-white text-[10px] font-bold">!</span>
          </div>
        </div>
      </div>

      {/* Icon and label */}
      <div className="flex flex-col items-center gap-2 p-3">
        <div className="p-2 bg-purple-500/20 rounded-full">
          <MessageSquare size={24} className="text-purple-400 animate-pulse" />
        </div>
        <span className="text-xs font-medium text-purple-300">
          Waiting for input
        </span>

        {/* Wait duration */}
        <div className="flex items-center gap-1 text-[10px] text-gray-400">
          <Clock size={10} />
          <span className="font-mono">{formatDuration(waitDuration)}</span>
        </div>

        {/* Remaining time (if timeout) */}
        {remainingTime !== null && (
          <div className={`px-2 py-0.5 rounded text-[10px] font-mono ${
            isTimeout ? 'bg-red-500/20 text-red-400' :
            isWarning ? 'bg-yellow-500/20 text-yellow-400' :
            'bg-gray-800 text-gray-400'
          }`}>
            {isTimeout ? 'Timed out' : `${formatDuration(remainingTime)} left`}
          </div>
        )}

        {/* Prompt preview */}
        {promptText && (
          <p className="text-[10px] text-gray-500 text-center max-w-[120px] truncate">
            "{promptText}"
          </p>
        )}
      </div>

      {/* Click hint */}
      <div className="absolute bottom-2 left-1/2 -translate-x-1/2">
        <span className="text-[9px] text-purple-400/60">Click to respond</span>
      </div>
    </div>
  );
}

interface WaitStepBadgeProps {
  isWaiting: boolean;
  onClick?: () => void;
}

export function WaitStepBadge({ isWaiting, onClick }: WaitStepBadgeProps) {
  if (!isWaiting) return null;

  return (
    <button
      onClick={onClick}
      className="flex items-center gap-1 px-2 py-0.5 bg-purple-500/20 border border-purple-500/50 rounded-full text-xs text-purple-300 hover:bg-purple-500/30 transition-colors"
    >
      <MessageSquare size={12} className="animate-pulse" />
      <span>Input required</span>
    </button>
  );
}

export default WaitStepOverlay;
