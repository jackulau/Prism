import { create } from 'zustand';
import type { WorkflowExecutionStatus, WorkflowStepResult, WorkflowStepStatus, WorkflowInfo } from '../types';

interface WorkflowExecutionState {
  // Execution state
  workflowId: string | null;
  workflowInfo: WorkflowInfo | null;
  status: WorkflowExecutionStatus;
  currentStepIndex: number;
  totalSteps: number;
  stepResults: Map<string, WorkflowStepResult>;
  workflowState: Record<string, unknown>;
  startedAt: number | null;
  completedAt: number | null;
  duration: number;
  error: string | null;

  // Input waiting state
  waitingForInput: boolean;
  inputPrompt: string | null;
  waitingStepId: string | null;

  // Actions
  startExecution: (workflowId: string, info?: WorkflowInfo) => void;
  updateWorkflowInfo: (info: Partial<WorkflowInfo>) => void;
  setStatus: (status: WorkflowExecutionStatus) => void;
  updateProgress: (currentStep: number, totalSteps: number) => void;

  // Step actions
  stepStarted: (stepId: string, stepName: string, stepType: string, stepIndex: number) => void;
  stepCompleted: (stepId: string, output?: unknown, duration?: number) => void;
  stepFailed: (stepId: string, error: string) => void;
  stepSkipped: (stepId: string) => void;
  stepRetrying: (stepId: string, retryCount: number) => void;

  // State actions
  updateState: (state: Record<string, unknown>) => void;
  setError: (error: string | null) => void;

  // Input actions
  setWaitingInput: (stepId: string, prompt: string) => void;
  clearWaitingInput: () => void;

  // Completion actions
  complete: (finalState?: Record<string, unknown>, duration?: number) => void;
  fail: (error: string) => void;
  cancel: () => void;
  pause: () => void;
  resume: () => void;

  // Utility actions
  reset: () => void;
  getStepResult: (stepId: string) => WorkflowStepResult | undefined;
  getStepsByStatus: (status: WorkflowStepStatus) => WorkflowStepResult[];
}

const initialState = {
  workflowId: null,
  workflowInfo: null,
  status: 'idle' as WorkflowExecutionStatus,
  currentStepIndex: 0,
  totalSteps: 0,
  stepResults: new Map<string, WorkflowStepResult>(),
  workflowState: {},
  startedAt: null,
  completedAt: null,
  duration: 0,
  error: null,
  waitingForInput: false,
  inputPrompt: null,
  waitingStepId: null,
};

export const useWorkflowExecutionStore = create<WorkflowExecutionState>((set, get) => ({
  ...initialState,

  startExecution: (workflowId, info) => set({
    workflowId,
    workflowInfo: info || null,
    status: 'running',
    currentStepIndex: 0,
    totalSteps: info?.totalSteps || 0,
    stepResults: new Map(),
    workflowState: {},
    startedAt: Date.now(),
    completedAt: null,
    duration: 0,
    error: null,
    waitingForInput: false,
    inputPrompt: null,
    waitingStepId: null,
  }),

  updateWorkflowInfo: (info) => set((state) => ({
    workflowInfo: state.workflowInfo
      ? { ...state.workflowInfo, ...info }
      : info as WorkflowInfo,
    totalSteps: info.totalSteps ?? state.totalSteps,
  })),

  setStatus: (status) => set({ status }),

  updateProgress: (currentStep, totalSteps) => set({
    currentStepIndex: currentStep,
    totalSteps,
  }),

  stepStarted: (stepId, stepName, stepType, stepIndex) => set((state) => {
    const newResults = new Map(state.stepResults);
    newResults.set(stepId, {
      stepId,
      stepName,
      stepType,
      status: 'running',
      startedAt: Date.now(),
    });
    return {
      stepResults: newResults,
      currentStepIndex: stepIndex,
    };
  }),

  stepCompleted: (stepId, output, duration) => set((state) => {
    const newResults = new Map(state.stepResults);
    const existing = newResults.get(stepId);
    if (existing) {
      const startTime = typeof existing.startedAt === 'number' ? existing.startedAt : 0;
      newResults.set(stepId, {
        ...existing,
        status: 'completed',
        completedAt: Date.now(),
        duration: duration ?? (startTime ? Date.now() - startTime : 0),
        output,
      });
    }
    return { stepResults: newResults };
  }),

  stepFailed: (stepId, error) => set((state) => {
    const newResults = new Map(state.stepResults);
    const existing = newResults.get(stepId);
    if (existing) {
      const startTime = typeof existing.startedAt === 'number' ? existing.startedAt : 0;
      newResults.set(stepId, {
        ...existing,
        status: 'failed',
        completedAt: Date.now(),
        duration: startTime ? Date.now() - startTime : 0,
        error,
      });
    }
    return { stepResults: newResults };
  }),

  stepSkipped: (stepId) => set((state) => {
    const newResults = new Map(state.stepResults);
    const existing = newResults.get(stepId);
    if (existing) {
      newResults.set(stepId, {
        ...existing,
        status: 'skipped',
        completedAt: Date.now(),
      });
    }
    return { stepResults: newResults };
  }),

  stepRetrying: (stepId, retryCount) => set((state) => {
    const newResults = new Map(state.stepResults);
    const existing = newResults.get(stepId);
    if (existing) {
      newResults.set(stepId, {
        ...existing,
        status: 'retrying',
        retryCount,
      });
    }
    return { stepResults: newResults };
  }),

  updateState: (workflowState) => set({ workflowState }),

  setError: (error) => set({ error }),

  setWaitingInput: (stepId, prompt) => set({
    status: 'waiting_input',
    waitingForInput: true,
    inputPrompt: prompt,
    waitingStepId: stepId,
  }),

  clearWaitingInput: () => set({
    waitingForInput: false,
    inputPrompt: null,
    waitingStepId: null,
  }),

  complete: (finalState, duration) => set((state) => ({
    status: 'completed',
    workflowState: finalState ?? state.workflowState,
    completedAt: Date.now(),
    duration: duration ?? (state.startedAt ? Date.now() - state.startedAt : 0),
    error: null,
  })),

  fail: (error) => set((state) => ({
    status: 'failed',
    error,
    completedAt: Date.now(),
    duration: state.startedAt ? Date.now() - state.startedAt : 0,
  })),

  cancel: () => set((state) => ({
    status: 'cancelled',
    completedAt: Date.now(),
    duration: state.startedAt ? Date.now() - state.startedAt : 0,
  })),

  pause: () => set({ status: 'paused' }),

  resume: () => set({ status: 'running' }),

  reset: () => set(initialState),

  getStepResult: (stepId) => get().stepResults.get(stepId),

  getStepsByStatus: (status) => {
    const results: WorkflowStepResult[] = [];
    get().stepResults.forEach((result) => {
      if (result.status === status) {
        results.push(result);
      }
    });
    return results;
  },
}));
