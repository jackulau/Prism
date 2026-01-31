import { useEffect, useCallback, useRef, useState } from 'react';
import { useWorkflowExecutionStore } from '../store/workflowExecutionStore';
import type { WorkflowWSMessage, MessageType, WorkflowInfo } from '../types';

interface UseWorkflowExecutionOptions {
  workflowId?: string;
  autoConnect?: boolean;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: string) => void;
}

interface UseWorkflowExecutionReturn {
  isConnected: boolean;
  isConnecting: boolean;
  connect: (workflowId: string) => void;
  disconnect: () => void;
  runWorkflow: (templateId: string, initialState?: Record<string, unknown>) => void;
  pauseWorkflow: () => void;
  resumeWorkflow: () => void;
  stopWorkflow: () => void;
  provideInput: (stepId: string, input: unknown) => void;
}

export function useWorkflowExecution(
  options: UseWorkflowExecutionOptions = {}
): UseWorkflowExecutionReturn {
  const { autoConnect = false, onConnect, onDisconnect, onError } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const workflowIdRef = useRef<string | null>(options.workflowId ?? null);

  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);

  const store = useWorkflowExecutionStore();

  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const message: WorkflowWSMessage = JSON.parse(event.data);
      const type = message.type as MessageType;

      switch (type) {
        case 'workflow.started': {
          const info: WorkflowInfo = {
            id: message.workflow_id || '',
            name: message.workflow_info?.name,
            status: 'running',
            currentStep: 0,
            totalSteps: message.total_steps || message.workflow_info?.totalSteps || 0,
            startedAt: Date.now(),
          };
          store.startExecution(message.workflow_id || '', info);
          break;
        }

        case 'workflow.paused':
          store.pause();
          store.updateProgress(
            message.current_step || store.currentStepIndex,
            store.totalSteps
          );
          break;

        case 'workflow.resumed':
          store.resume();
          store.updateProgress(
            message.current_step || store.currentStepIndex,
            store.totalSteps
          );
          break;

        case 'workflow.cancelled':
          store.cancel();
          break;

        case 'workflow.progress':
          store.updateProgress(
            message.current_step || 0,
            message.total_steps || store.totalSteps
          );
          if (message.state) {
            store.updateState(message.state);
          }
          break;

        case 'workflow.status':
          if (message.workflow_info) {
            store.updateWorkflowInfo(message.workflow_info);
          }
          if (message.state) {
            store.updateState(message.state);
          }
          break;

        case 'workflow.step_started':
          store.stepStarted(
            message.step_id || '',
            message.step_name || '',
            message.step_type || '',
            message.current_step || store.currentStepIndex
          );
          break;

        case 'workflow.step_completed':
          store.stepCompleted(
            message.step_id || '',
            message.result,
            message.duration
          );
          break;

        case 'workflow.step_failed':
          store.stepFailed(
            message.step_id || '',
            message.error || 'Unknown error'
          );
          break;

        case 'workflow.step_skipped':
          store.stepSkipped(message.step_id || '');
          break;

        case 'workflow.completed':
          store.complete(message.state, message.duration);
          break;

        case 'workflow.failed':
          store.fail(message.error || 'Workflow execution failed');
          break;

        case 'workflow.waiting_input':
          store.setWaitingInput(
            message.step_id || '',
            message.message || 'Input required'
          );
          break;
      }
    } catch (err) {
      console.error('Failed to parse workflow message:', err);
    }
  }, [store]);

  const connect = useCallback((workflowId: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    setIsConnecting(true);
    workflowIdRef.current = workflowId;

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/api/v1/ws`;

    try {
      wsRef.current = new WebSocket(wsUrl);

      wsRef.current.onopen = () => {
        setIsConnected(true);
        setIsConnecting(false);
        reconnectAttemptsRef.current = 0;
        onConnect?.();
      };

      wsRef.current.onmessage = handleMessage;

      wsRef.current.onclose = () => {
        setIsConnected(false);
        setIsConnecting(false);
        onDisconnect?.();

        // Attempt reconnection with exponential backoff
        if (reconnectAttemptsRef.current < 5) {
          const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttemptsRef.current++;
            if (workflowIdRef.current) {
              connect(workflowIdRef.current);
            }
          }, delay);
        }
      };

      wsRef.current.onerror = () => {
        setIsConnecting(false);
        onError?.('WebSocket connection error');
      };
    } catch (err) {
      setIsConnecting(false);
      onError?.('Failed to create WebSocket connection');
    }
  }, [handleMessage, onConnect, onDisconnect, onError]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    reconnectAttemptsRef.current = 0;
    workflowIdRef.current = null;

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const send = useCallback((message: Record<string, unknown>) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket not connected, cannot send message');
    }
  }, []);

  const runWorkflow = useCallback((templateId: string, initialState?: Record<string, unknown>) => {
    send({
      type: 'workflow.run',
      template_id: templateId,
      state: initialState,
    });
  }, [send]);

  const pauseWorkflow = useCallback(() => {
    if (workflowIdRef.current) {
      send({
        type: 'workflow.pause',
        workflow_id: workflowIdRef.current,
      });
    }
  }, [send]);

  const resumeWorkflow = useCallback(() => {
    if (workflowIdRef.current) {
      send({
        type: 'workflow.resume',
        workflow_id: workflowIdRef.current,
      });
    }
  }, [send]);

  const stopWorkflow = useCallback(() => {
    if (workflowIdRef.current) {
      send({
        type: 'workflow.stop',
        workflow_id: workflowIdRef.current,
      });
    }
  }, [send]);

  const provideInput = useCallback((stepId: string, input: unknown) => {
    if (workflowIdRef.current) {
      send({
        type: 'workflow.provide_input',
        workflow_id: workflowIdRef.current,
        step_id: stepId,
        input,
      });
      store.clearWaitingInput();
      store.setStatus('running');
    }
  }, [send, store]);

  // Auto-connect effect
  useEffect(() => {
    if (autoConnect && options.workflowId) {
      connect(options.workflowId);
    }

    return () => {
      disconnect();
    };
  }, [autoConnect, options.workflowId, connect, disconnect]);

  return {
    isConnected,
    isConnecting,
    connect,
    disconnect,
    runWorkflow,
    pauseWorkflow,
    resumeWorkflow,
    stopWorkflow,
    provideInput,
  };
}
