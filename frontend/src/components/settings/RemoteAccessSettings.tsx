import { useState, useEffect } from 'react';
import { useRemoteAccessStore } from '../../store/remoteAccessStore';
import { useAuthStore } from '../../store/authStore';
import {
  Wifi,
  WifiOff,
  Eye,
  EyeOff,
  Copy,
  RefreshCw,
  AlertTriangle,
  CheckCircle,
  Loader2,
} from 'lucide-react';

export function RemoteAccessSettings() {
  const { accessToken } = useAuthStore();
  const {
    enabled,
    port,
    password,
    loading,
    error,
    setToken,
    fetchStatus,
    enable,
    disable,
    regeneratePassword,
    clearError,
  } = useRemoteAccessStore();

  const [showPassword, setShowPassword] = useState(false);
  const [portInput, setPortInput] = useState(String(port));
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (accessToken) {
      setToken(accessToken);
      fetchStatus();
    }
  }, [accessToken, setToken, fetchStatus]);

  useEffect(() => {
    setPortInput(String(port));
  }, [port]);

  const handleToggle = async () => {
    clearError();
    if (enabled) {
      await disable();
    } else {
      await enable({ port: parseInt(portInput, 10) });
    }
  };

  const handleRegeneratePassword = async () => {
    clearError();
    await regeneratePassword();
  };

  const handleCopyPassword = async () => {
    if (password) {
      await navigator.clipboard.writeText(password);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className="space-y-4">
      {/* Security Warning */}
      <div className="flex items-start gap-3 p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
        <AlertTriangle className="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" />
        <div className="text-sm">
          <p className="font-medium text-yellow-500">Security Notice</p>
          <p className="text-editor-muted mt-1">
            Enabling remote access allows connections from other devices on your network.
            Keep your password secure and only share it with trusted users.
          </p>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="flex items-center gap-2 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
          <AlertTriangle className="w-4 h-4 flex-shrink-0" />
          {error}
        </div>
      )}

      {/* Enable/Disable Toggle */}
      <div className="flex items-center justify-between p-4 bg-editor-bg rounded-lg">
        <div className="flex items-center gap-3">
          {enabled ? (
            <Wifi className="w-5 h-5 text-green-500" />
          ) : (
            <WifiOff className="w-5 h-5 text-editor-muted" />
          )}
          <div>
            <p className="font-medium">Remote Access</p>
            <p className="text-sm text-editor-muted">
              {enabled ? 'Remote connections are enabled' : 'Remote connections are disabled'}
            </p>
          </div>
        </div>
        <button
          onClick={handleToggle}
          disabled={loading}
          className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
            enabled ? 'bg-green-500' : 'bg-editor-border'
          } ${loading ? 'opacity-50 cursor-wait' : ''}`}
        >
          <span
            className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
              enabled ? 'translate-x-6' : 'translate-x-1'
            }`}
          />
          {loading && (
            <Loader2 className="absolute inset-0 m-auto w-4 h-4 animate-spin text-white" />
          )}
        </button>
      </div>

      {/* Port Configuration */}
      <div className="p-4 bg-editor-bg rounded-lg">
        <label className="block text-sm font-medium mb-2">Port</label>
        <div className="flex items-center gap-2">
          <input
            type="number"
            value={portInput}
            onChange={(e) => setPortInput(e.target.value)}
            disabled={enabled}
            min={1024}
            max={65535}
            className="w-32 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm disabled:opacity-50"
          />
          <span className="text-sm text-editor-muted">
            {enabled ? 'Disable remote access to change port' : 'Port range: 1024-65535'}
          </span>
        </div>
      </div>

      {/* Password Section - Only show when enabled */}
      {enabled && (
        <div className="p-4 bg-editor-bg rounded-lg space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Access Password</label>
            <div className="flex items-center gap-2">
              <div className="flex-1 relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password || ''}
                  readOnly
                  className="w-full px-3 py-2 pr-20 bg-editor-surface border border-editor-border rounded-lg text-sm font-mono"
                />
                <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
                  <button
                    onClick={() => setShowPassword(!showPassword)}
                    className="p-1 hover:bg-editor-border rounded text-editor-muted hover:text-editor-text"
                    title={showPassword ? 'Hide password' : 'Show password'}
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                  <button
                    onClick={handleCopyPassword}
                    className="p-1 hover:bg-editor-border rounded text-editor-muted hover:text-editor-text"
                    title="Copy password"
                  >
                    {copied ? (
                      <CheckCircle className="w-4 h-4 text-green-500" />
                    ) : (
                      <Copy className="w-4 h-4" />
                    )}
                  </button>
                </div>
              </div>
              <button
                onClick={handleRegeneratePassword}
                disabled={loading}
                className="px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm hover:bg-editor-border transition-colors disabled:opacity-50 flex items-center gap-2"
                title="Generate new password"
              >
                <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
                Regenerate
              </button>
            </div>
            <p className="text-xs text-editor-muted mt-2">
              Share this password with users who need to connect remotely.
              Regenerating the password will disconnect all current sessions.
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
