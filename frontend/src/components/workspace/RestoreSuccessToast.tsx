import { useEffect, useState } from 'react';
import { Check, Undo2, X } from 'lucide-react';
import { useSandboxStore } from '../../store/sandboxStore';

interface RestoreSuccessToastProps {
  filePath: string;
  restoredFromTime: string;
  onDismiss: () => void;
}

const AUTO_DISMISS_DELAY = 10000; // 10 seconds

export function RestoreSuccessToast({
  filePath,
  restoredFromTime,
  onDismiss,
}: RestoreSuccessToastProps) {
  const [timeRemaining, setTimeRemaining] = useState(AUTO_DISMISS_DELAY / 1000);
  const [isVisible, setIsVisible] = useState(true);
  const { undoLastRestore, lastRestoreBackupId, isRestoring } = useSandboxStore();

  useEffect(() => {
    // Auto-dismiss countdown
    const interval = setInterval(() => {
      setTimeRemaining((prev) => {
        if (prev <= 1) {
          handleDismiss();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  const handleDismiss = () => {
    setIsVisible(false);
    setTimeout(onDismiss, 200); // Wait for animation
  };

  const handleUndo = () => {
    if (lastRestoreBackupId) {
      undoLastRestore();
      handleDismiss();
    }
  };

  const fileName = filePath.split('/').pop() || filePath;

  if (!isVisible) return null;

  return (
    <div
      className={`fixed bottom-6 right-6 z-50 flex items-center gap-4 bg-editor-surface border border-editor-border rounded-lg shadow-xl p-4 transition-all duration-200 ${
        isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
      }`}
    >
      {/* Success Icon */}
      <div className="flex-shrink-0 w-10 h-10 rounded-full bg-green-500/20 flex items-center justify-center">
        <Check className="w-5 h-5 text-green-500" />
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-editor-text">File restored</p>
        <p className="text-xs text-editor-muted truncate max-w-[200px]">
          {fileName} from {restoredFromTime}
        </p>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2 flex-shrink-0">
        {lastRestoreBackupId && (
          <button
            onClick={handleUndo}
            disabled={isRestoring}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm text-blue-400 hover:bg-blue-500/10 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <Undo2 className="w-4 h-4" />
            <span>Undo</span>
          </button>
        )}
        <button
          onClick={handleDismiss}
          className="p-1.5 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      {/* Progress bar */}
      <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-editor-border rounded-b-lg overflow-hidden">
        <div
          className="h-full bg-green-500 transition-all duration-1000 ease-linear"
          style={{ width: `${(timeRemaining / (AUTO_DISMISS_DELAY / 1000)) * 100}%` }}
        />
      </div>
    </div>
  );
}

// Hook to manage restore success toast state
export function useRestoreSuccessToast() {
  const [toastData, setToastData] = useState<{
    filePath: string;
    restoredFromTime: string;
  } | null>(null);

  const showToast = (filePath: string, restoredFromTime: string) => {
    setToastData({ filePath, restoredFromTime });
  };

  const dismissToast = () => {
    setToastData(null);
  };

  return {
    toastData,
    showToast,
    dismissToast,
    ToastComponent: toastData ? (
      <RestoreSuccessToast
        filePath={toastData.filePath}
        restoredFromTime={toastData.restoredFromTime}
        onDismiss={dismissToast}
      />
    ) : null,
  };
}
