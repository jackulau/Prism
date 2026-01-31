// Agent execution state types

export type AgentExecutionStatus = 'idle' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface AgentConfig {
  id?: string;
  name: string;
  model: string;
  provider: string;
  systemPrompt?: string;
  tools?: string[];
  maxIterations?: number;
  metadata?: Record<string, unknown>;
}

export interface AgentExecutionState {
  agentId: string | null;
  config: AgentConfig | null;
  status: AgentExecutionStatus;
  output: string[];
  error: string | null;
  startedAt: Date | null;
  completedAt: Date | null;
  iterationCount: number;
  lastToolCall?: AgentToolCall;
}

export interface AgentToolCall {
  id: string;
  name: string;
  parameters: Record<string, unknown>;
  result?: unknown;
  status: 'pending' | 'running' | 'completed' | 'failed';
  startedAt?: Date;
  completedAt?: Date;
}

export interface AgentResult {
  agentId: string;
  status: 'completed' | 'failed' | 'cancelled';
  output: string[];
  error?: string;
  iterationCount: number;
  startedAt: Date;
  completedAt: Date;
  toolCalls: AgentToolCall[];
}

// WebSocket message types for agent execution
export type AgentMessageType =
  | 'agent.run'
  | 'agent.started'
  | 'agent.stream_chunk'
  | 'agent.tool_started'
  | 'agent.tool_completed'
  | 'agent.completed'
  | 'agent.failed'
  | 'agent.cancelled'
  | 'agent.stop'
  | 'agent.status_request'
  | 'agent.status_response'
  | 'agent.list_request'
  | 'agent.list_response';

// Outgoing messages (from client to server)
export interface AgentRunMessage {
  type: 'agent.run';
  config: AgentConfig;
  conversationId?: string;
}

export interface AgentStopMessage {
  type: 'agent.stop';
  agentId: string;
}

export interface AgentStatusRequestMessage {
  type: 'agent.status_request';
  agentId: string;
}

export interface AgentListRequestMessage {
  type: 'agent.list_request';
}

export type OutgoingAgentMessage =
  | AgentRunMessage
  | AgentStopMessage
  | AgentStatusRequestMessage
  | AgentListRequestMessage;

// Incoming messages (from server to client)
export interface AgentStartedMessage {
  type: 'agent.started';
  agentId: string;
  config: AgentConfig;
  startedAt: string;
}

export interface AgentStreamChunkMessage {
  type: 'agent.stream_chunk';
  agentId: string;
  chunk: string;
  iterationCount?: number;
}

export interface AgentToolStartedMessage {
  type: 'agent.tool_started';
  agentId: string;
  toolCall: {
    id: string;
    name: string;
    parameters: Record<string, unknown>;
  };
}

export interface AgentToolCompletedMessage {
  type: 'agent.tool_completed';
  agentId: string;
  toolCallId: string;
  result?: unknown;
  status: 'completed' | 'failed';
  error?: string;
}

export interface AgentCompletedMessage {
  type: 'agent.completed';
  agentId: string;
  result: AgentResult;
}

export interface AgentFailedMessage {
  type: 'agent.failed';
  agentId: string;
  error: string;
  iterationCount?: number;
}

export interface AgentCancelledMessage {
  type: 'agent.cancelled';
  agentId: string;
  iterationCount?: number;
}

export interface AgentStatusResponseMessage {
  type: 'agent.status_response';
  agentId: string;
  status: AgentExecutionStatus;
  iterationCount: number;
  lastActivity?: string;
}

export interface AgentListResponseMessage {
  type: 'agent.list_response';
  agents: Array<{
    id: string;
    config: AgentConfig;
    status: AgentExecutionStatus;
    startedAt: string;
  }>;
}

export type IncomingAgentMessage =
  | AgentStartedMessage
  | AgentStreamChunkMessage
  | AgentToolStartedMessage
  | AgentToolCompletedMessage
  | AgentCompletedMessage
  | AgentFailedMessage
  | AgentCancelledMessage
  | AgentStatusResponseMessage
  | AgentListResponseMessage;
