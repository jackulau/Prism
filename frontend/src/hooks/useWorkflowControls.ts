import { useMutation, useQueryClient, type UseMutationOptions } from '@tanstack/react-query';
import { useCallback, useRef } from 'react';

interface WorkflowControlResponse {
  success: boolean;
  workflowId: string;
  status?: string;
  message?: string;
}

interface WorkflowInputPayload {
  stepId: string;
  input: unknown;
}

class WorkflowControlError extends Error {
  constructor(message: string, public statusCode?: number) {
    super(message);
    this.name = 'WorkflowControlError';
  }
}

async function controlWorkflow(
  workflowId: string,
  action: 'pause' | 'resume' | 'cancel'
): Promise<WorkflowControlResponse> {
  const response = await fetch(`/api/v1/workflows/${workflowId}/${action}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Unknown error' }));
    throw new WorkflowControlError(error.message || `Failed to ${action} workflow`, response.status);
  }

  return response.json();
}

async function provideWorkflowInput(
  workflowId: string,
  payload: WorkflowInputPayload
): Promise<WorkflowControlResponse> {
  const response = await fetch(`/api/v1/workflows/${workflowId}/input`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Unknown error' }));
    throw new WorkflowControlError(error.message || 'Failed to provide input', response.status);
  }

  return response.json();
}

export const workflowQueryKeys = {
  all: ['workflows'] as const,
  detail: (id: string) => [...workflowQueryKeys.all, 'detail', id] as const,
  execution: (id: string) => [...workflowQueryKeys.all, 'execution', id] as const,
};

export function usePauseWorkflow(
  workflowId: string,
  options?: UseMutationOptions<WorkflowControlResponse, WorkflowControlError, void>
) {
  const queryClient = useQueryClient();
  const lastCallRef = useRef<number>(0);
  const debounceMs = 300;

  return useMutation({
    mutationFn: async () => {
      const now = Date.now();
      if (now - lastCallRef.current < debounceMs) {
        throw new WorkflowControlError('Please wait before trying again');
      }
      lastCallRef.current = now;
      return controlWorkflow(workflowId, 'pause');
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowQueryKeys.execution(workflowId) });
    },
    ...options,
  });
}

export function useResumeWorkflow(
  workflowId: string,
  options?: UseMutationOptions<WorkflowControlResponse, WorkflowControlError, void>
) {
  const queryClient = useQueryClient();
  const lastCallRef = useRef<number>(0);
  const debounceMs = 300;

  return useMutation({
    mutationFn: async () => {
      const now = Date.now();
      if (now - lastCallRef.current < debounceMs) {
        throw new WorkflowControlError('Please wait before trying again');
      }
      lastCallRef.current = now;
      return controlWorkflow(workflowId, 'resume');
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowQueryKeys.execution(workflowId) });
    },
    ...options,
  });
}

export function useCancelWorkflow(
  workflowId: string,
  options?: UseMutationOptions<WorkflowControlResponse, WorkflowControlError, void>
) {
  const queryClient = useQueryClient();
  const lastCallRef = useRef<number>(0);
  const debounceMs = 300;

  return useMutation({
    mutationFn: async () => {
      const now = Date.now();
      if (now - lastCallRef.current < debounceMs) {
        throw new WorkflowControlError('Please wait before trying again');
      }
      lastCallRef.current = now;
      return controlWorkflow(workflowId, 'cancel');
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowQueryKeys.execution(workflowId) });
    },
    ...options,
  });
}

export function useProvideInput(
  workflowId: string,
  options?: UseMutationOptions<WorkflowControlResponse, WorkflowControlError, WorkflowInputPayload>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: WorkflowInputPayload) => {
      return provideWorkflowInput(workflowId, payload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowQueryKeys.execution(workflowId) });
    },
    ...options,
  });
}

export function useWorkflowControls(workflowId: string) {
  const pauseMutation = usePauseWorkflow(workflowId);
  const resumeMutation = useResumeWorkflow(workflowId);
  const cancelMutation = useCancelWorkflow(workflowId);
  const inputMutation = useProvideInput(workflowId);

  const pause = useCallback(() => {
    pauseMutation.mutate();
  }, [pauseMutation]);

  const resume = useCallback(() => {
    resumeMutation.mutate();
  }, [resumeMutation]);

  const cancel = useCallback(() => {
    cancelMutation.mutate();
  }, [cancelMutation]);

  const provideInput = useCallback((stepId: string, input: unknown) => {
    inputMutation.mutate({ stepId, input });
  }, [inputMutation]);

  return {
    pause,
    resume,
    cancel,
    provideInput,
    isPausing: pauseMutation.isPending,
    isResuming: resumeMutation.isPending,
    isCancelling: cancelMutation.isPending,
    isProvidingInput: inputMutation.isPending,
    isLoading: pauseMutation.isPending || resumeMutation.isPending || cancelMutation.isPending || inputMutation.isPending,
    pauseError: pauseMutation.error,
    resumeError: resumeMutation.error,
    cancelError: cancelMutation.error,
    inputError: inputMutation.error,
  };
}

export type { WorkflowControlResponse, WorkflowInputPayload, WorkflowControlError };
