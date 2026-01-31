import { useCallback, useEffect, useRef } from 'react';
import { useAgentStore } from '../store/agentStore';
import type {
  AgentConfig,
  IncomingAgentMessage,
  OutgoingAgentMessage,
} from '../types/agent';

interface UseAgentExecutionOptions {
  onStarted?: (agentId: string) => void;
  onChunk?: (chunk: string) => void;
  onCompleted?: (agentId: string) => void;
  onFailed?: (error: string) => void;
  onCancelled?: () => void;
}

export function useAgentExecution(options: UseAgentExecutionOptions = {}) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;

  const {
    currentExecution,
    startAgent,
    stopAgent,
    resetAgent,
    appendOutput,
    clearOutput,
    updateStatus,
    setError,
    addToolCall,
    updateToolCallStatus,
    incrementIteration,
    completeExecution,
    failExecution,
    isRunning,
  } = useAgentStore();

  // Handle incoming WebSocket messages
  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message: IncomingAgentMessage = JSON.parse(event.data);

        switch (message.type) {
          case 'agent.started':
            updateStatus('running', message.agentId);
            options.onStarted?.(message.agentId);
            break;

          case 'agent.stream_chunk':
            appendOutput(message.chunk, message.agentId);
            if (message.iterationCount !== undefined) {
              const currentIteration = useAgentStore.getState().currentExecution.iterationCount;
              if (message.iterationCount > currentIteration) {
                incrementIteration(message.agentId);
              }
            }
            options.onChunk?.(message.chunk);
            break;

          case 'agent.tool_started':
            addToolCall(
              {
                id: message.toolCall.id,
                name: message.toolCall.name,
                parameters: message.toolCall.parameters,
              },
              message.agentId
            );
            break;

          case 'agent.tool_completed':
            updateToolCallStatus(
              message.toolCallId,
              message.status,
              message.result,
              message.agentId
            );
            break;

          case 'agent.completed':
            completeExecution(message.result);
            options.onCompleted?.(message.agentId);
            break;

          case 'agent.failed':
            setError(message.error, message.agentId);
            failExecution(message.error, message.agentId);
            options.onFailed?.(message.error);
            break;

          case 'agent.cancelled':
            updateStatus('cancelled', message.agentId);
            options.onCancelled?.();
            break;

          default:
            // Ignore unhandled message types
            break;
        }
      } catch {
        // Failed to parse message - ignore malformed data
      }
    },
    [
      appendOutput,
      addToolCall,
      updateToolCallStatus,
      updateStatus,
      setError,
      completeExecution,
      failExecution,
      incrementIteration,
      options,
    ]
  );

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/ws/agent`;
    wsRef.current = new WebSocket(wsUrl);

    wsRef.current.onopen = () => {
      reconnectAttemptsRef.current = 0;
    };

    wsRef.current.onmessage = handleMessage;

    wsRef.current.onclose = () => {
      // Attempt reconnect if we have an active execution
      if (isRunning() && reconnectAttemptsRef.current < maxReconnectAttempts) {
        reconnectAttemptsRef.current++;
        const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current - 1), 30000);
        reconnectTimeoutRef.current = setTimeout(connect, delay);
      }
    };

    wsRef.current.onerror = () => {
      // Error handling - connection will be closed and onclose will handle reconnection
    };
  }, [handleMessage, isRunning]);

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  // Send message to WebSocket
  const send = useCallback((message: OutgoingAgentMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  // Run an agent with the given configuration
  const runAgent = useCallback(
    (config: AgentConfig, conversationId?: string) => {
      // Generate a unique agent ID
      const agentId = `agent-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;

      // Initialize the agent in the store
      startAgent(config, agentId);

      // Ensure WebSocket is connected
      connect();

      // Send the run message
      send({
        type: 'agent.run',
        config: { ...config, id: agentId },
        conversationId,
      });

      return agentId;
    },
    [connect, send, startAgent]
  );

  // Stop the currently running agent
  const stop = useCallback(
    (agentId?: string) => {
      const targetId = agentId || currentExecution.agentId;
      if (!targetId) return;

      send({
        type: 'agent.stop',
        agentId: targetId,
      });

      stopAgent(targetId);
    },
    [currentExecution.agentId, send, stopAgent]
  );

  // Reset the agent state
  const reset = useCallback(() => {
    resetAgent();
  }, [resetAgent]);

  // Clear the output buffer
  const clear = useCallback(
    (agentId?: string) => {
      clearOutput(agentId);
    },
    [clearOutput]
  );

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    // State
    execution: currentExecution,
    isRunning: isRunning(),
    output: currentExecution.output,
    error: currentExecution.error,
    status: currentExecution.status,
    iterationCount: currentExecution.iterationCount,
    lastToolCall: currentExecution.lastToolCall,

    // Actions
    runAgent,
    stop,
    reset,
    clear,

    // WebSocket control
    connect,
    disconnect,
  };
}
