import React from 'react';
import { useAppStore } from '../store';
import { wsService } from '../services/websocket';
import { Wifi, WifiOff, RefreshCw, AlertTriangle, Loader2 } from 'lucide-react';

interface ConnectionStatusBarProps {
  className?: string;
  compact?: boolean;
}

export const ConnectionStatusBar: React.FC<ConnectionStatusBarProps> = ({
  className,
  compact = false,
}) => {
  const connectionStatus = useAppStore((state) => state.connectionStatus);
  const [isReconnecting, setIsReconnecting] = React.useState(false);

  const handleManualReconnect = () => {
    setIsReconnecting(true);
    wsService.manualReconnect();
    // Reset after attempt
    setTimeout(() => setIsReconnecting(false), 2000);
  };

  const renderStatusContent = () => {
    switch (connectionStatus) {
      case 'connected':
        return (
          <div
            className={`flex items-center gap-1.5 ${
              compact
                ? 'px-2.5 py-1 rounded-full bg-editor-success/20 text-editor-success text-xs'
                : 'gap-2 text-editor-success'
            }`}
          >
            <Wifi className={compact ? 'w-3.5 h-3.5' : 'w-4 h-4'} />
            <span>{compact ? 'Connected' : 'Connected'}</span>
          </div>
        );

      case 'connecting':
        return (
          <div
            className={`flex items-center gap-1.5 ${
              compact
                ? 'px-2.5 py-1 rounded-full bg-editor-accent/20 text-editor-accent text-xs'
                : 'gap-2 text-editor-accent'
            }`}
          >
            <Loader2 className={`${compact ? 'w-3.5 h-3.5' : 'w-4 h-4'} animate-spin`} />
            <span>{compact ? 'Connecting' : 'Connecting...'}</span>
          </div>
        );

      case 'disconnected':
        return (
          <div
            className={`flex items-center gap-1.5 ${
              compact
                ? 'px-2.5 py-1 rounded-full bg-editor-warning/20 text-editor-warning text-xs'
                : 'gap-2 text-editor-warning'
            }`}
          >
            <WifiOff className={compact ? 'w-3.5 h-3.5' : 'w-4 h-4'} />
            <span>Disconnected</span>
            {!compact && (
              <button
                onClick={handleManualReconnect}
                disabled={isReconnecting}
                className="ml-2 px-2 py-1 text-xs rounded bg-editor-warning/20 hover:bg-editor-warning/30 transition-colors flex items-center gap-1 disabled:opacity-50"
              >
                <RefreshCw className={`w-3 h-3 ${isReconnecting ? 'animate-spin' : ''}`} />
                Reconnect
              </button>
            )}
          </div>
        );

      case 'error':
        return (
          <div
            className={`flex items-center gap-1.5 ${
              compact
                ? 'px-2.5 py-1 rounded-full bg-editor-error/20 text-editor-error text-xs'
                : 'gap-2 text-editor-error'
            }`}
          >
            <AlertTriangle className={compact ? 'w-3.5 h-3.5' : 'w-4 h-4'} />
            <span>{compact ? 'Error' : 'Connection Error'}</span>
            {!compact && (
              <button
                onClick={handleManualReconnect}
                disabled={isReconnecting}
                className="ml-2 px-2 py-1 text-xs rounded bg-editor-error/20 hover:bg-editor-error/30 transition-colors flex items-center gap-1 disabled:opacity-50"
              >
                <RefreshCw className={`w-3 h-3 ${isReconnecting ? 'animate-spin' : ''}`} />
                Retry
              </button>
            )}
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className={`flex items-center ${className || ''}`}>
      {renderStatusContent()}
    </div>
  );
};
