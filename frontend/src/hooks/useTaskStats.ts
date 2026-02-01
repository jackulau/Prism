/**
 * Custom hook for fetching task queue statistics.
 * Supports polling for live updates with configurable interval.
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import { apiService } from '../services/api';
import type { TaskStats } from './useTasks';

interface UseTaskStatsOptions {
  /** Enable polling for live updates (default: false) */
  polling?: boolean;
  /** Polling interval in milliseconds (default: 5000) */
  pollingInterval?: number;
  /** Initial enabled state (default: true) */
  enabled?: boolean;
}

interface UseTaskStatsResult {
  /** Task statistics data */
  stats: TaskStats | null;
  /** Loading state */
  isLoading: boolean;
  /** Error message if fetch failed */
  error: string | null;
  /** Refetch the stats manually */
  refetch: () => Promise<void>;
  /** Whether polling is currently active */
  isPolling: boolean;
  /** Start polling */
  startPolling: () => void;
  /** Stop polling */
  stopPolling: () => void;
}

const DEFAULT_POLLING_INTERVAL = 5000;

export function useTaskStats(options: UseTaskStatsOptions = {}): UseTaskStatsResult {
  const {
    polling = false,
    pollingInterval = DEFAULT_POLLING_INTERVAL,
    enabled = true,
  } = options;

  const [stats, setStats] = useState<TaskStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isPolling, setIsPolling] = useState(polling);

  const intervalRef = useRef<number | null>(null);
  const mountedRef = useRef(true);

  const fetchStats = useCallback(async () => {
    if (!enabled) return;

    try {
      const response = await apiService.getTaskStats();

      if (!mountedRef.current) return;

      if (response.error) {
        setError(response.error);
        setStats(null);
      } else if (response.data) {
        setStats(response.data);
        setError(null);
      }
    } catch (err) {
      if (!mountedRef.current) return;
      setError(err instanceof Error ? err.message : 'Failed to fetch task stats');
      setStats(null);
    } finally {
      if (mountedRef.current) {
        setIsLoading(false);
      }
    }
  }, [enabled]);

  const refetch = useCallback(async () => {
    setIsLoading(true);
    await fetchStats();
  }, [fetchStats]);

  const startPolling = useCallback(() => {
    setIsPolling(true);
  }, []);

  const stopPolling = useCallback(() => {
    setIsPolling(false);
    if (intervalRef.current !== null) {
      window.clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, []);

  // Initial fetch
  useEffect(() => {
    mountedRef.current = true;
    fetchStats();

    return () => {
      mountedRef.current = false;
    };
  }, [fetchStats]);

  // Polling effect
  useEffect(() => {
    if (isPolling && enabled) {
      intervalRef.current = window.setInterval(fetchStats, pollingInterval);
    } else if (intervalRef.current !== null) {
      window.clearInterval(intervalRef.current);
      intervalRef.current = null;
    }

    return () => {
      if (intervalRef.current !== null) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [isPolling, enabled, pollingInterval, fetchStats]);

  return {
    stats,
    isLoading,
    error,
    refetch,
    isPolling,
    startPolling,
    stopPolling,
  };
}
