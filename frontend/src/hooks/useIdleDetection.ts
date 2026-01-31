import { useEffect, useRef, useState, useCallback } from 'react';

export interface UseIdleDetectionOptions {
  idleTimeout: number;      // ms before showing warning
  warningDuration: number;  // ms for warning countdown
  onIdle: () => void;       // Called when idle timeout reached (after warning)
  onWarning: () => void;    // Called when warning should show
  onActive: () => void;     // Called when user becomes active
  enabled?: boolean;        // Whether detection is enabled
}

export interface UseIdleDetectionReturn {
  isIdle: boolean;
  isWarning: boolean;
  remainingTime: number | null;
  resetIdleTimer: () => void;
}

/**
 * Hook to detect user inactivity and trigger warnings before session timeout.
 * Tracks mouse movement, keyboard input, clicks, scroll, and touch events.
 */
export function useIdleDetection(options: UseIdleDetectionOptions): UseIdleDetectionReturn {
  const {
    idleTimeout,
    warningDuration,
    onIdle,
    onWarning,
    onActive,
    enabled = true,
  } = options;

  const [isIdle, setIsIdle] = useState(false);
  const [isWarning, setIsWarning] = useState(false);
  const [remainingTime, setRemainingTime] = useState<number | null>(null);

  const idleTimerRef = useRef<number | null>(null);
  const warningTimerRef = useRef<number | null>(null);
  const countdownIntervalRef = useRef<number | null>(null);
  const lastActivityRef = useRef<number>(Date.now());

  const clearTimers = useCallback(() => {
    if (idleTimerRef.current) {
      window.clearTimeout(idleTimerRef.current);
      idleTimerRef.current = null;
    }
    if (warningTimerRef.current) {
      window.clearTimeout(warningTimerRef.current);
      warningTimerRef.current = null;
    }
    if (countdownIntervalRef.current) {
      window.clearInterval(countdownIntervalRef.current);
      countdownIntervalRef.current = null;
    }
  }, []);

  const startWarningCountdown = useCallback(() => {
    setIsWarning(true);
    setRemainingTime(Math.ceil(warningDuration / 1000));
    onWarning();

    // Update countdown every second
    countdownIntervalRef.current = window.setInterval(() => {
      setRemainingTime((prev) => {
        if (prev === null || prev <= 1) {
          return null;
        }
        return prev - 1;
      });
    }, 1000);

    // Set timeout for when warning period ends
    warningTimerRef.current = window.setTimeout(() => {
      clearTimers();
      setIsIdle(true);
      setIsWarning(false);
      setRemainingTime(null);
      onIdle();
    }, warningDuration);
  }, [warningDuration, onWarning, onIdle, clearTimers]);

  const resetIdleTimer = useCallback(() => {
    if (!enabled) return;

    clearTimers();
    lastActivityRef.current = Date.now();

    // If we were in warning or idle state, reset to active
    if (isWarning || isIdle) {
      setIsWarning(false);
      setIsIdle(false);
      setRemainingTime(null);
      onActive();
    }

    // Start new idle timer
    idleTimerRef.current = window.setTimeout(() => {
      startWarningCountdown();
    }, idleTimeout);
  }, [enabled, idleTimeout, isWarning, isIdle, onActive, clearTimers, startWarningCountdown]);

  // Set up activity listeners
  useEffect(() => {
    if (!enabled) {
      clearTimers();
      setIsIdle(false);
      setIsWarning(false);
      setRemainingTime(null);
      return;
    }

    const activityEvents = [
      'mousedown',
      'mousemove',
      'keydown',
      'keypress',
      'touchstart',
      'touchmove',
      'scroll',
      'wheel',
      'click',
    ];

    // Throttle activity handler to avoid excessive resets
    let lastReset = 0;
    const throttleMs = 1000; // Only reset once per second

    const handleActivity = () => {
      const now = Date.now();
      if (now - lastReset > throttleMs) {
        lastReset = now;
        resetIdleTimer();
      }
    };

    // Add event listeners
    activityEvents.forEach((event) => {
      window.addEventListener(event, handleActivity, { passive: true });
    });

    // Also listen on document for events that might not bubble to window
    activityEvents.forEach((event) => {
      document.addEventListener(event, handleActivity, { passive: true });
    });

    // Start the initial idle timer
    resetIdleTimer();

    // Cleanup
    return () => {
      activityEvents.forEach((event) => {
        window.removeEventListener(event, handleActivity);
        document.removeEventListener(event, handleActivity);
      });
      clearTimers();
    };
  }, [enabled, resetIdleTimer, clearTimers]);

  // Handle visibility change (tab becomes visible again)
  useEffect(() => {
    if (!enabled) return;

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        // Tab became visible, check if we need to show warning or have timed out
        const elapsed = Date.now() - lastActivityRef.current;

        if (elapsed >= idleTimeout + warningDuration) {
          // User has been away longer than idle + warning, trigger logout
          clearTimers();
          setIsIdle(true);
          setIsWarning(false);
          setRemainingTime(null);
          onIdle();
        } else if (elapsed >= idleTimeout) {
          // User was away past idle timeout but still in warning period
          const warningElapsed = elapsed - idleTimeout;
          const warningRemaining = warningDuration - warningElapsed;

          if (warningRemaining > 0) {
            clearTimers();
            setIsWarning(true);
            setRemainingTime(Math.ceil(warningRemaining / 1000));
            onWarning();

            // Continue countdown
            countdownIntervalRef.current = window.setInterval(() => {
              setRemainingTime((prev) => {
                if (prev === null || prev <= 1) {
                  return null;
                }
                return prev - 1;
              });
            }, 1000);

            warningTimerRef.current = window.setTimeout(() => {
              clearTimers();
              setIsIdle(true);
              setIsWarning(false);
              setRemainingTime(null);
              onIdle();
            }, warningRemaining);
          }
        }
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [enabled, idleTimeout, warningDuration, onIdle, onWarning, clearTimers]);

  return {
    isIdle,
    isWarning,
    remainingTime,
    resetIdleTimer,
  };
}
