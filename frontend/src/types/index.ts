// Re-export agent progress types (except AgentStatus which is defined locally)
export type {
  SwarmStatus,
  ProgressEvent,
  ProgressMetrics,
  AgentProgress,
  SwarmProgress,
} from './agentProgress';

// WebSocket message types
export type MessageType =
  | 'chat.message'
  | 'chat.chunk'
  | 'chat.thinking_chunk'
  | 'chat.complete'
  | 'tool.started'
  | 'tool.completed'
  | 'tool.confirm'
  | 'chat.stop'
  | 'error'
  | 'agent.check_in'
  | 'agent.continue'
  // Agent lifecycle message types
  | 'agent.started'
  | 'agent.completed'
  | 'agent.failed'
  | 'agent.cancelled'
  // Swarm message types
  | 'swarm.started'
  | 'swarm.agent_started'
  | 'swarm.agent_completed'
  | 'swarm.completed'
  | 'swarm.failed'
  // Heartbeat message types
  | 'heartbeat'
  | 'heartbeat.ack'
  // Preview/Sandbox message types
  | 'preview.ready'
  | 'preview.content'
  | 'preview.error'
  | 'build.start'
  | 'build.started'
  | 'build.output'
  | 'build.completed'
  | 'build.stop'
  | 'files.updated'
  | 'file.content'
  | 'file.request'
  // File history message types
  | 'file.history_request'
  | 'file.history_list'
  | 'file.history_content'
  | 'file.history_restore'
  | 'file.history_restored'
  | 'file.history_batch_restore'
  | 'file.history_batch_restored'
  | 'file.history_conflicts'
  // Attribution message types
  | 'attribution.summary_request'
  | 'attribution.summary'
  | 'attribution.by_agent_request'
  | 'attribution.by_agent'
  // CloudProvider message types
  | 'cloud_agent.created'
  | 'cloud_agent.message'
  | 'cloud_agent.chunk'
  | 'cloud_agent.complete'
  | 'cloud_agent.error';

export interface Attachment {
  name: string;
  type: string;
  data: string;
}

export interface IncomingWSMessage {
  type: MessageType;
  conversation_id: string;
  content?: string;
  attachments?: Attachment[];
  execution_id?: string;
  params?: Record<string, unknown>;
  approved?: boolean;
  // Chat options
  mode?: ChatMode;
  extended_thinking?: boolean;
  file_context?: FileContext | null;
}

export interface OutgoingWSMessage {
  type: MessageType;
  conversation_id: string;
  message_id?: string;
  delta?: string;
  thinking_delta?: string;
  finish_reason?: string;
  execution_id?: string;
  tool_name?: string;
  parameters?: unknown;
  result?: unknown;
  error?: string;
  // Enhanced message fields
  provider?: string;
  model?: string;
  input_tokens?: number;
  output_tokens?: number;
  // Sandbox/Preview fields
  build_id?: string;
  content?: string;
  stream?: 'stdout' | 'stderr';
  success?: boolean;
  preview_url?: string;
  url?: string;
  file_path?: string;
  files?: SandboxFile[];
  duration?: number;
  status?: string;
  // File history fields
  metadata?: Record<string, unknown>;
  // MCP-related fields (sent by backend for tool.confirm/tool.started)
  is_mcp_tool?: boolean;
  mcp_server_name?: string;
  mcp_server_id?: string;
  is_stdio_mcp?: boolean;
  iteration_count?: number;
}

// Message status
export type MessageStatus = 'streaming' | 'complete' | 'error';

// Chat types
export interface Message {
  id: string;
  conversation_id?: string;
  parent_id?: string;
  role: 'user' | 'assistant' | 'system' | 'tool';
  content: string;
  thinking_content?: string;
  timestamp: Date;
  toolCalls?: ToolCall[];
  provider?: string;
  model?: string;
  status?: MessageStatus;
  input_tokens?: number;
  output_tokens?: number;
  finish_reason?: string;
  isStreaming?: boolean;
  metrics?: MessageMetrics;
}

export interface ToolCall {
  id: string;
  name: string;
  parameters: Record<string, unknown>;
  result?: unknown;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'rejected';
  isMCP?: boolean;
  serverName?: string;
  isStdioMCP?: boolean;
}

export interface MessageMetrics {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  startTime: number;
  endTime?: number;
  tokensPerSecond?: number;
  firstTokenTime?: number;
  timeToFirstToken?: number;
}

// File tree types
export interface FileNode {
  id: string;
  name: string;
  type: 'file' | 'directory';
  path: string;
  children?: FileNode[];
  isExpanded?: boolean;
  language?: string;
  size?: number;
  modified?: Date;
}

// Conversation types
export interface Conversation {
  id: string;
  title: string;
  createdAt: Date;
  updatedAt: Date;
  messageCount: number;
  provider: string;
  model: string;
}

// Generation metrics
export interface GenerationMetrics {
  isGenerating: boolean;
  startTime: number | null;
  endTime: number | null;
  firstTokenTime: number | null;
  tokenCount: number;
  charCount: number;
  tokensPerSecond: number;
  timeToFirstToken: number | null;
  elapsedTime: number;
  estimatedTokensRemaining: number | null;
}

// Connection status
export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

// Chat Mode
export type ChatMode = 'plan' | 'ask-before-edits' | 'edit-automatically';

// File Context for chat
export interface FileContext {
  path: string;
  content: string;
  language?: string;
}

// Message Queue
export interface QueuedMessage {
  id: string;
  content: string;
  createdAt: Date;
  status: 'queued' | 'sending' | 'cancelled';
  model: string;
  provider: string;
  projectFolder: string;
}

// Theme
export type Theme =
  | 'catppuccin-mocha'   // Default dark theme (current)
  | 'catppuccin-latte'   // Light theme
  | 'dracula'            // Popular dark theme
  | 'nord'               // Cool, bluish theme
  | 'github-dark'        // GitHub's dark theme
  | 'solarized-dark'     // Classic dark theme
  | 'one-dark';          // Atom One Dark

export interface ThemeColors {
  bg: string;
  surface: string;
  border: string;
  text: string;
  muted: string;
  accent: string;
  success: string;
  warning: string;
  error: string;
  sidebarBg: string;
  sidebarHover: string;
}

// Preview/Sandbox types
export interface PreviewWSMessage {
  type: MessageType;
  url?: string;
  content?: string;
  error?: string;
  build_id?: string;
  stream?: 'stdout' | 'stderr';
  success?: boolean;
  preview_url?: string;
  duration?: number;
  status?: string;
  file_path?: string;
  files?: SandboxFile[];
}

export interface SandboxFile {
  name: string;
  path: string;
  is_directory: boolean;
  children?: SandboxFile[];
  size?: number;
  modified?: number;
}

export type BuildStatus = 'idle' | 'building' | 'success' | 'error';

export interface BuildInfo {
  id: string;
  status: BuildStatus;
  startTime?: number;
  endTime?: number;
  error?: string;
  previewUrl?: string;
}

// Provider types
export interface ProviderModel {
  id: string;
  name: string;
  context_window: number;
  supports_tools: boolean;
  supports_vision: boolean;
}

export interface Provider {
  name: string;
  models: ProviderModel[];
  supports_tools: boolean;
  supports_vision: boolean;
}

// CloudProvider types
export interface CloudProvider {
  name: string;
  hasCredentials: boolean;
}

export interface CloudAgent {
  id: string;
  providerId: string;
  providerName: string;
  name: string;
  status: CloudAgentStatus;
  createdAt: Date;
  updatedAt?: Date;
  model?: string;
  systemPrompt?: string;
}

export type CloudAgentStatus = 'active' | 'idle' | 'terminated' | 'error';

export interface CreateCloudAgentParams {
  provider: string;
  name?: string;
  systemPrompt?: string;
  model?: string;
  tools?: string[];
  metadata?: Record<string, string>;
}

export interface CloudProviderMessage {
  id: string;
  role: 'user' | 'assistant' | 'system' | 'tool';
  content: string;
  timestamp: Date;
  toolCalls?: ToolCall[];
  images?: CloudImageData[];
}

export interface CloudImageData {
  url?: string;
  base64?: string;
  mimeType?: string;
}

export interface CloudMessageChunk {
  delta?: string;
  toolCalls?: ToolCall[];
  finishReason?: string;
  error?: string;
}

// ============================================================================
// Agent Types
// ============================================================================

/** Agent status values */
export type AgentStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

/** Agent record from the backend */
export interface Agent {
  id: string;
  conversationId?: string;
  name: string;
  description?: string;
  provider: string;
  model: string;
  systemPrompt?: string;
  status: AgentStatus;
  error?: string;
  createdAt: Date;
  startedAt?: Date;
  completedAt?: Date;
}

/** Agent execution record */
export interface AgentExecution {
  id: string;
  userId: string;
  provider: string;
  llmProvider: string;
  model: string;
  agentName: string;
  status: AgentStatus;
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  inputCost: number;
  outputCost: number;
  totalCost: number;
  currency: string;
  error?: string;
  startedAt?: Date;
  completedAt?: Date;
  createdAt: Date;
  metadata?: Record<string, unknown>;
}

/** Agent message in an execution */
export interface AgentMessage {
  id: string;
  executionId: string;
  role: 'user' | 'assistant' | 'system' | 'tool';
  content: string;
  toolCalls?: AgentToolCallInfo[];
  toolCallId?: string;
  promptTokens: number;
  completionTokens: number;
  createdAt: Date;
}

/** Tool call info within an agent message */
export interface AgentToolCallInfo {
  id: string;
  name: string;
  parameters: Record<string, unknown>;
}

/** Agent tool call record */
export interface AgentToolCall {
  id: string;
  executionId: string;
  messageId?: string;
  toolName: string;
  parameters: Record<string, unknown>;
  output?: string;
  error?: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  durationMs: number;
  createdAt: Date;
  completedAt?: Date;
}

/** Agent result record */
export interface AgentResult {
  id: string;
  agentId: string;
  taskId?: string;
  success: boolean;
  output?: string;
  error?: string;
  usage?: AgentUsage;
  metadata?: Record<string, unknown>;
  durationMs: number;
  createdAt: Date;
}

/** Agent token usage */
export interface AgentUsage {
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
}

// ============================================================================
// Task Types
// ============================================================================

/** Task status values */
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

/** Task priority levels */
export type TaskPriority = 0 | 1 | 2 | 3;

/** Task record from the backend */
export interface Task {
  id: string;
  userId: string;
  prompt: string;
  context?: string;
  priority: TaskPriority;
  status: TaskStatus;
  agentConfig?: TaskAgentConfig;
  metadata?: Record<string, unknown>;
  result?: Record<string, unknown>;
  error?: string;
  callbackUrl?: string;
  createdAt: number;
  startedAt?: number;
  completedAt?: number;
}

/** Task agent configuration */
export interface TaskAgentConfig {
  provider?: string;
  model?: string;
  temperature?: number;
  maxTokens?: number;
}

/** Task statistics */
export interface TaskStats {
  total: number;
  pending: number;
  running: number;
  completed: number;
  failed: number;
}

// Re-export monitoring types
export * from './monitoring';
