import { useState, useEffect, useCallback } from 'react';
import { useAppStore } from '../store';
import { wsService } from '../services/websocket';

export function useConnectionStatus() {
  const connectionStatus = useAppStore((state) => state.connectionStatus);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);
  const [lastConnectedAt, setLastConnectedAt] = useState<Date | null>(null);

  useEffect(() => {
    if (connectionStatus === 'connected') {
      setLastConnectedAt(new Date());
      setReconnectAttempts(0);
    } else if (connectionStatus === 'connecting') {
      // Increment reconnect attempts when trying to reconnect
      setReconnectAttempts((prev) => prev + 1);
    }
  }, [connectionStatus]);

  const reconnect = useCallback(() => {
    wsService.manualReconnect();
  }, []);

  return {
    status: connectionStatus,
    isConnected: connectionStatus === 'connected',
    isConnecting: connectionStatus === 'connecting',
    isDisconnected: connectionStatus === 'disconnected',
    hasError: connectionStatus === 'error',
    reconnectAttempts,
    lastConnectedAt,
    reconnect,
  };
}
