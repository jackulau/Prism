import { useEffect, useCallback, useState } from 'react';
import { useSessionStore } from '../../store/sessionStore';
import { useAuthStore } from '../../store/authStore';
import { apiService } from '../../services/api';
import { Monitor, RefreshCw, Loader2, LogOut } from 'lucide-react';
import { SessionCard } from './SessionCard';

export function ActiveSessions() {
  const { accessToken } = useAuthStore();
  const { sessions, isLoading, error, fetchSessions, terminateSession, terminateOthers } =
    useSessionStore();

  const [terminatingOthers, setTerminatingOthers] = useState(false);
  const [showTerminateOthersConfirm, setShowTerminateOthersConfirm] = useState(false);

  const refreshSessions = useCallback(() => {
    if (accessToken) {
      apiService.setToken(accessToken);
      fetchSessions();
    }
  }, [accessToken, fetchSessions]);

  useEffect(() => {
    refreshSessions();

    // Auto-refresh every 30 seconds
    const interval = setInterval(refreshSessions, 30000);
    return () => clearInterval(interval);
  }, [refreshSessions]);

  const handleTerminateSession = async (sessionId: number) => {
    await terminateSession(sessionId);
  };

  const handleTerminateOthers = async () => {
    setTerminatingOthers(true);
    try {
      await terminateOthers();
      setShowTerminateOthersConfirm(false);
    } finally {
      setTerminatingOthers(false);
    }
  };

  const currentSession = sessions.find((s) => s.is_current);
  const otherSessions = sessions.filter((s) => !s.is_current);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Monitor className="w-4 h-4 text-editor-muted" />
          <h4 className="text-sm font-medium text-editor-muted">Active Sessions</h4>
          {sessions.length > 0 && (
            <span className="px-2 py-0.5 text-xs bg-primary/20 text-primary rounded-full">
              {sessions.length}
            </span>
          )}
        </div>
        <button
          onClick={refreshSessions}
          disabled={isLoading}
          className="p-1.5 hover:bg-editor-surface rounded text-editor-muted hover:text-editor-text transition-colors disabled:opacity-50"
          title="Refresh sessions"
        >
          <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-sm text-red-400">
          {error}
        </div>
      )}

      {isLoading && sessions.length === 0 ? (
        <div className="flex items-center justify-center p-8 text-editor-muted">
          <Loader2 className="w-5 h-5 animate-spin" />
        </div>
      ) : sessions.length === 0 ? (
        <div className="text-center p-8 text-editor-muted text-sm">
          <Monitor className="w-8 h-8 mx-auto mb-2 opacity-50" />
          No active sessions
        </div>
      ) : (
        <div className="space-y-4">
          {/* Current Session */}
          {currentSession && (
            <div>
              <h5 className="text-xs font-medium text-editor-muted uppercase tracking-wide mb-2">
                Current Session
              </h5>
              <SessionCard session={currentSession} />
            </div>
          )}

          {/* Other Sessions */}
          {otherSessions.length > 0 && (
            <div>
              <h5 className="text-xs font-medium text-editor-muted uppercase tracking-wide mb-2">
                Other Sessions ({otherSessions.length})
              </h5>
              <div className="space-y-2">
                {otherSessions.map((session) => (
                  <SessionCard
                    key={session.id}
                    session={session}
                    onTerminate={() => handleTerminateSession(session.id)}
                  />
                ))}
              </div>
            </div>
          )}

          {/* End All Other Sessions */}
          {otherSessions.length > 0 && (
            <div className="pt-2 border-t border-editor-border">
              {showTerminateOthersConfirm ? (
                <div className="flex items-center justify-between p-3 bg-red-500/5 border border-red-500/20 rounded-lg">
                  <span className="text-sm text-editor-text">
                    End all {otherSessions.length} other session{otherSessions.length !== 1 ? 's' : ''}?
                  </span>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => setShowTerminateOthersConfirm(false)}
                      className="px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text transition-colors"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={handleTerminateOthers}
                      disabled={terminatingOthers}
                      className="flex items-center gap-2 px-3 py-1.5 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50"
                    >
                      {terminatingOthers ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        <LogOut className="w-4 h-4" />
                      )}
                      Confirm
                    </button>
                  </div>
                </div>
              ) : (
                <button
                  onClick={() => setShowTerminateOthersConfirm(true)}
                  className="w-full flex items-center justify-center gap-2 py-2 text-sm text-red-400 hover:bg-red-400/10 rounded-lg transition-colors"
                >
                  <LogOut className="w-4 h-4" />
                  End All Other Sessions
                </button>
              )}
            </div>
          )}
        </div>
      )}

      {sessions.length > 0 && (
        <p className="text-xs text-editor-muted text-center">
          Sessions auto-refresh every 30 seconds
        </p>
      )}
    </div>
  );
}
