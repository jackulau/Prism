import { useEffect, useState, useCallback } from 'react';
import { useRemoteAccessStore } from '../../store/remoteAccessStore';
import { useAuthStore } from '../../store/authStore';
import {
  Users,
  RefreshCw,
  XCircle,
  Clock,
  ArrowDownUp,
  Loader2,
} from 'lucide-react';

export function RemoteSessions() {
  const { accessToken } = useAuthStore();
  const {
    enabled,
    sessions,
    sessionsLoading,
    setToken,
    fetchSessions,
    kickSession,
  } = useRemoteAccessStore();

  const [kickingSession, setKickingSession] = useState<string | null>(null);

  const refreshSessions = useCallback(() => {
    if (accessToken && enabled) {
      setToken(accessToken);
      fetchSessions();
    }
  }, [accessToken, enabled, setToken, fetchSessions]);

  useEffect(() => {
    refreshSessions();

    // Auto-refresh every 30 seconds when enabled
    if (enabled) {
      const interval = setInterval(refreshSessions, 30000);
      return () => clearInterval(interval);
    }
  }, [enabled, refreshSessions]);

  const handleKick = async (sessionId: string) => {
    setKickingSession(sessionId);
    await kickSession(sessionId);
    setKickingSession(null);
  };

  const formatDuration = (connectedAt: string) => {
    const start = new Date(connectedAt);
    const now = new Date();
    const diff = Math.floor((now.getTime() - start.getTime()) / 1000);

    if (diff < 60) return `${diff}s`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ${Math.floor((diff % 3600) / 60)}m`;
    return `${Math.floor(diff / 86400)}d ${Math.floor((diff % 86400) / 3600)}h`;
  };

  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  };

  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString(undefined, {
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (!enabled) {
    return null;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Users className="w-4 h-4 text-editor-muted" />
          <h4 className="text-sm font-medium text-editor-muted">Active Sessions</h4>
          {sessions.length > 0 && (
            <span className="px-2 py-0.5 text-xs bg-primary/20 text-primary rounded-full">
              {sessions.length}
            </span>
          )}
        </div>
        <button
          onClick={refreshSessions}
          disabled={sessionsLoading}
          className="p-1.5 hover:bg-editor-surface rounded text-editor-muted hover:text-editor-text transition-colors disabled:opacity-50"
          title="Refresh sessions"
        >
          <RefreshCw className={`w-4 h-4 ${sessionsLoading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {sessionsLoading && sessions.length === 0 ? (
        <div className="flex items-center justify-center p-8 text-editor-muted">
          <Loader2 className="w-5 h-5 animate-spin" />
        </div>
      ) : sessions.length === 0 ? (
        <div className="text-center p-8 text-editor-muted text-sm">
          <Users className="w-8 h-8 mx-auto mb-2 opacity-50" />
          No active sessions
        </div>
      ) : (
        <div className="space-y-2">
          {sessions.map((session) => (
            <div
              key={session.id}
              className="flex items-center justify-between p-3 bg-editor-bg rounded-lg"
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-sm truncate">{session.clientIP}</span>
                </div>
                <div className="flex items-center gap-4 mt-1 text-xs text-editor-muted">
                  <span className="flex items-center gap-1">
                    <Clock className="w-3 h-3" />
                    {formatDuration(session.connectedAt)}
                  </span>
                  <span className="flex items-center gap-1">
                    <ArrowDownUp className="w-3 h-3" />
                    {formatBytes(session.bytesIn)} / {formatBytes(session.bytesOut)}
                  </span>
                  <span>
                    Last active: {formatTime(session.lastActivity)}
                  </span>
                </div>
              </div>
              <button
                onClick={() => handleKick(session.id)}
                disabled={kickingSession === session.id}
                className="p-2 text-red-400 hover:bg-red-400/10 rounded transition-colors disabled:opacity-50"
                title="Disconnect session"
              >
                {kickingSession === session.id ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <XCircle className="w-4 h-4" />
                )}
              </button>
            </div>
          ))}
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
