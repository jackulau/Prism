import { useCallback, useEffect, useRef, useState } from 'react';
import { useAgentStore } from '../store/agentStore';
import type {
  AgentExecutionStatus,
  AgentListResponseMessage,
  AgentStatusResponseMessage,
  OutgoingAgentMessage,
} from '../types/agent';

interface AgentInfo {
  id: string;
  name: string;
  model: string;
  provider: string;
  status: AgentExecutionStatus;
  startedAt: Date;
}

interface UseAgentStatusOptions {
  pollInterval?: number;
  enablePolling?: boolean;
}

export function useAgentStatus(
  agentId?: string,
  options: UseAgentStatusOptions = {}
) {
  const { pollInterval = 5000, enablePolling = false } = options;

  const [status, setStatus] = useState<AgentExecutionStatus>('idle');
  const [iterationCount, setIterationCount] = useState(0);
  const [lastActivity, setLastActivity] = useState<Date | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const { getExecution, currentExecution } = useAgentStore();

  // Get status from store first
  const storeExecution = agentId ? getExecution(agentId) : currentExecution;

  // Handle incoming status response
  const handleStatusResponse = useCallback((message: AgentStatusResponseMessage) => {
    setStatus(message.status);
    setIterationCount(message.iterationCount);
    if (message.lastActivity) {
      setLastActivity(new Date(message.lastActivity));
    }
    setIsLoading(false);
    setError(null);
  }, []);

  // Handle WebSocket messages
  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message = JSON.parse(event.data);

        if (message.type === 'agent.status_response') {
          if (!agentId || message.agentId === agentId) {
            handleStatusResponse(message as AgentStatusResponseMessage);
          }
        }
      } catch {
        // Failed to parse message - ignore malformed data
      }
    },
    [agentId, handleStatusResponse]
  );

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/ws/agent`;
    wsRef.current = new WebSocket(wsUrl);

    wsRef.current.onmessage = handleMessage;

    wsRef.current.onerror = () => {
      setError('WebSocket connection error');
    };
  }, [handleMessage]);

  // Send message to WebSocket
  const send = useCallback((message: OutgoingAgentMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  // Request status update
  const requestStatus = useCallback(
    (targetAgentId?: string) => {
      const id = targetAgentId || agentId;
      if (!id) return;

      setIsLoading(true);
      connect();
      send({
        type: 'agent.status_request',
        agentId: id,
      });
    },
    [agentId, connect, send]
  );

  // Set up polling if enabled
  useEffect(() => {
    if (enablePolling && agentId) {
      requestStatus();
      pollIntervalRef.current = setInterval(() => {
        requestStatus();
      }, pollInterval);
    }

    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current);
        pollIntervalRef.current = null;
      }
    };
  }, [enablePolling, agentId, pollInterval, requestStatus]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

  // Prefer store data if available
  const effectiveStatus = storeExecution?.status || status;
  const effectiveIterationCount = storeExecution?.iterationCount ?? iterationCount;

  return {
    status: effectiveStatus,
    iterationCount: effectiveIterationCount,
    lastActivity: storeExecution?.startedAt || lastActivity,
    isLoading,
    error,
    refresh: requestStatus,
  };
}

export function useActiveAgents(options: UseAgentStatusOptions = {}) {
  const { pollInterval = 5000, enablePolling = false } = options;

  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const { activeAgents } = useAgentStore();

  // Handle incoming list response
  const handleListResponse = useCallback((message: AgentListResponseMessage) => {
    setAgents(
      message.agents.map((agent) => ({
        id: agent.id,
        name: agent.config.name,
        model: agent.config.model,
        provider: agent.config.provider,
        status: agent.status,
        startedAt: new Date(agent.startedAt),
      }))
    );
    setIsLoading(false);
    setError(null);
  }, []);

  // Handle WebSocket messages
  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message = JSON.parse(event.data);

        if (message.type === 'agent.list_response') {
          handleListResponse(message as AgentListResponseMessage);
        }
      } catch {
        // Failed to parse message - ignore malformed data
      }
    },
    [handleListResponse]
  );

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/ws/agent`;
    wsRef.current = new WebSocket(wsUrl);

    wsRef.current.onmessage = handleMessage;

    wsRef.current.onerror = () => {
      setError('WebSocket connection error');
    };
  }, [handleMessage]);

  // Send message to WebSocket
  const send = useCallback((message: OutgoingAgentMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  // Request list of active agents
  const requestList = useCallback(() => {
    setIsLoading(true);
    connect();
    send({
      type: 'agent.list_request',
    });
  }, [connect, send]);

  // Set up polling if enabled
  useEffect(() => {
    if (enablePolling) {
      requestList();
      pollIntervalRef.current = setInterval(() => {
        requestList();
      }, pollInterval);
    }

    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current);
        pollIntervalRef.current = null;
      }
    };
  }, [enablePolling, pollInterval, requestList]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

  // Merge store data with server data
  const storeAgents: AgentInfo[] = Array.from(activeAgents.entries()).map(([id, execution]) => ({
    id,
    name: execution.config?.name || 'Unknown',
    model: execution.config?.model || 'Unknown',
    provider: execution.config?.provider || 'Unknown',
    status: execution.status,
    startedAt: execution.startedAt || new Date(),
  }));

  // Prefer store data, but include server-only agents
  const mergedAgents = [...storeAgents];
  for (const serverAgent of agents) {
    if (!storeAgents.some((a) => a.id === serverAgent.id)) {
      mergedAgents.push(serverAgent);
    }
  }

  return {
    agents: mergedAgents,
    isLoading,
    error,
    refresh: requestList,
  };
}
