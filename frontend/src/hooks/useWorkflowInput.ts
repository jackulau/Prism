import { useState, useEffect, useCallback, useRef } from 'react';
import { useWorkflowExecutionStore } from '../store/workflowExecutionStore';
import { useProvideInput } from './useWorkflowControls';

interface InputRequest {
  stepId: string;
  promptText: string;
  outputKey?: string;
  timeout?: number;
  requestedAt: number;
}

interface UseWorkflowInputOptions {
  workflowId: string;
  onInputRequired?: (request: InputRequest) => void;
  onInputSubmitted?: () => void;
  onInputTimeout?: (request: InputRequest) => void;
  onInputError?: (error: Error) => void;
}

interface UseWorkflowInputReturn {
  pendingInput: InputRequest | null;
  isWaitingForInput: boolean;
  remainingTime: number | null;
  submitInput: (input: unknown) => void;
  cancelInput: () => void;
  isSubmitting: boolean;
  error: Error | null;
  clearError: () => void;
}

export function useWorkflowInput(options: UseWorkflowInputOptions): UseWorkflowInputReturn {
  const { workflowId, onInputRequired, onInputSubmitted, onInputTimeout, onInputError } = options;

  const [pendingInput, setPendingInput] = useState<InputRequest | null>(null);
  const [remainingTime, setRemainingTime] = useState<number | null>(null);
  const [error, setError] = useState<Error | null>(null);

  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const { waitingForInput, inputPrompt, waitingStepId, clearWaitingInput } = useWorkflowExecutionStore();
  const provideMutation = useProvideInput(workflowId, {
    onSuccess: () => {
      clearWaitingInput();
      setPendingInput(null);
      setRemainingTime(null);
      setError(null);
      onInputSubmitted?.();
    },
    onError: (err) => {
      setError(err);
      onInputError?.(err);
    },
  });

  // Watch for input requests from the store
  useEffect(() => {
    if (waitingForInput && waitingStepId && inputPrompt) {
      const request: InputRequest = {
        stepId: waitingStepId,
        promptText: inputPrompt,
        requestedAt: Date.now(),
      };
      setPendingInput(request);
      onInputRequired?.(request);
    } else if (!waitingForInput && pendingInput) {
      // Input was cleared externally
      setPendingInput(null);
      setRemainingTime(null);
    }
  }, [waitingForInput, waitingStepId, inputPrompt, pendingInput, onInputRequired]);

  // Handle timeout countdown
  useEffect(() => {
    if (!pendingInput?.timeout) {
      setRemainingTime(null);
      return;
    }

    const timeout = pendingInput.timeout;
    const elapsed = Date.now() - pendingInput.requestedAt;
    const remaining = Math.max(0, timeout - elapsed);
    setRemainingTime(remaining);

    // Clear existing timers
    if (timerRef.current) {
      clearInterval(timerRef.current);
    }
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    // Update countdown every second
    timerRef.current = setInterval(() => {
      setRemainingTime((prev) => {
        if (prev === null || prev <= 1000) {
          return 0;
        }
        return prev - 1000;
      });
    }, 1000);

    // Set timeout for auto-cancel
    timeoutRef.current = setTimeout(() => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
      setRemainingTime(0);
      onInputTimeout?.(pendingInput);
      // Don't clear pending input - let the UI show timeout state
    }, remaining);

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [pendingInput, onInputTimeout]);

  const submitInput = useCallback((input: unknown) => {
    if (!pendingInput) {
      setError(new Error('No pending input request'));
      return;
    }

    provideMutation.mutate({
      stepId: pendingInput.stepId,
      input,
    });
  }, [pendingInput, provideMutation]);

  const cancelInput = useCallback(() => {
    clearWaitingInput();
    setPendingInput(null);
    setRemainingTime(null);
    if (timerRef.current) {
      clearInterval(timerRef.current);
    }
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
  }, [clearWaitingInput]);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return {
    pendingInput,
    isWaitingForInput: waitingForInput,
    remainingTime,
    submitInput,
    cancelInput,
    isSubmitting: provideMutation.isPending,
    error,
    clearError,
  };
}

export type { InputRequest, UseWorkflowInputOptions, UseWorkflowInputReturn };
