import { useState } from 'react';
import { XCircle, Loader2, Shield } from 'lucide-react';
import type { Session } from '../../store/sessionStore';
import {
  getDeviceIcon,
  parseDeviceInfo,
  formatRelativeTime,
  formatSessionDate,
} from '../../utils/deviceHelpers';

interface SessionCardProps {
  session: Session;
  onTerminate?: () => Promise<void>;
}

export function SessionCard({ session, onTerminate }: SessionCardProps) {
  const [isTerminating, setIsTerminating] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  const deviceInfo = parseDeviceInfo(session.user_agent);
  const DeviceIcon = getDeviceIcon(session.device_name || session.user_agent);

  const handleTerminate = async () => {
    if (!onTerminate) return;
    setIsTerminating(true);
    try {
      await onTerminate();
    } finally {
      setIsTerminating(false);
      setShowConfirm(false);
    }
  };

  return (
    <div
      className={`p-4 rounded-lg border transition-colors ${
        session.is_current
          ? 'bg-primary/5 border-primary/30'
          : 'bg-editor-bg border-editor-border'
      }`}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-3 min-w-0">
          <div
            className={`p-2 rounded-lg ${
              session.is_current ? 'bg-primary/10 text-primary' : 'bg-editor-surface text-editor-muted'
            }`}
          >
            <DeviceIcon className="w-5 h-5" />
          </div>

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <span className="font-medium text-editor-text truncate">
                {deviceInfo.browser} on {deviceInfo.os}
              </span>
              {session.is_current && (
                <span className="flex items-center gap-1 px-2 py-0.5 text-xs bg-primary/20 text-primary rounded-full">
                  <Shield className="w-3 h-3" />
                  Current
                </span>
              )}
            </div>

            <div className="mt-1 space-y-0.5 text-sm text-editor-muted">
              <div className="flex items-center gap-2">
                <span>IP: {session.ip_address}</span>
              </div>
              <div>
                Active: {formatRelativeTime(session.last_activity)}
              </div>
              <div>
                Started: {formatSessionDate(session.created_at)}
              </div>
            </div>
          </div>
        </div>

        {!session.is_current && onTerminate && (
          <div className="flex-shrink-0">
            {showConfirm ? (
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setShowConfirm(false)}
                  className="px-2 py-1 text-xs text-editor-muted hover:text-editor-text transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleTerminate}
                  disabled={isTerminating}
                  className="px-2 py-1 text-xs bg-red-500/20 text-red-400 hover:bg-red-500/30 rounded transition-colors disabled:opacity-50"
                >
                  {isTerminating ? (
                    <Loader2 className="w-3 h-3 animate-spin" />
                  ) : (
                    'Confirm'
                  )}
                </button>
              </div>
            ) : (
              <button
                onClick={() => setShowConfirm(true)}
                className="p-2 text-editor-muted hover:text-red-400 hover:bg-red-400/10 rounded-lg transition-colors"
                title="End session"
              >
                <XCircle className="w-5 h-5" />
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
