import { AlertTriangle, LogOut, RefreshCw } from 'lucide-react';

interface IdleWarningModalProps {
  remainingTime: number | null;
  onStayLoggedIn: () => void;
  onLogout: () => void;
}

/**
 * Modal that displays when the user session is about to expire due to inactivity.
 * Shows a countdown timer and options to stay logged in or log out.
 */
export function IdleWarningModal({
  remainingTime,
  onStayLoggedIn,
  onLogout,
}: IdleWarningModalProps) {
  // Format remaining time as MM:SS
  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <>
      {/* Backdrop */}
      <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50" />

      {/* Modal */}
      <div className="fixed inset-0 flex items-center justify-center z-50 p-4">
        <div className="w-full max-w-md bg-editor-surface border border-editor-border rounded-xl shadow-2xl">
          {/* Header */}
          <div className="flex items-center gap-3 p-6 border-b border-editor-border">
            <div className="p-3 rounded-full bg-amber-500/10">
              <AlertTriangle className="w-6 h-6 text-amber-500" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-editor-text">
                Session Expiring Soon
              </h2>
              <p className="text-sm text-editor-muted">
                You&apos;ve been inactive for a while
              </p>
            </div>
          </div>

          {/* Content */}
          <div className="p-6 space-y-6">
            {/* Countdown */}
            <div className="text-center">
              <div className="inline-flex items-center justify-center w-24 h-24 rounded-full bg-amber-500/10 border-4 border-amber-500/30">
                <span className="text-3xl font-bold text-amber-500 font-mono">
                  {remainingTime !== null ? formatTime(remainingTime) : '--:--'}
                </span>
              </div>
            </div>

            {/* Message */}
            <p className="text-center text-editor-muted">
              Your session will expire in{' '}
              <span className="font-medium text-editor-text">
                {remainingTime !== null ? formatTime(remainingTime) : 'a few seconds'}
              </span>{' '}
              due to inactivity. Click &quot;Stay Logged In&quot; to continue your session, or
              you will be logged out automatically.
            </p>

            {/* Actions */}
            <div className="flex flex-col sm:flex-row gap-3">
              <button
                onClick={onLogout}
                className="flex-1 flex items-center justify-center gap-2 px-4 py-3 text-editor-muted hover:text-editor-text border border-editor-border rounded-lg hover:bg-editor-bg transition-colors"
              >
                <LogOut className="w-4 h-4" />
                Log Out Now
              </button>
              <button
                onClick={onStayLoggedIn}
                className="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors font-medium"
              >
                <RefreshCw className="w-4 h-4" />
                Stay Logged In
              </button>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
