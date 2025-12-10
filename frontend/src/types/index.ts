// WebSocket message types
export type MessageType =
  | 'chat.message'
  | 'chat.chunk'
  | 'chat.complete'
  | 'tool.started'
  | 'tool.completed'
  | 'tool.confirm'
  | 'chat.stop'
  | 'error'
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
  | 'file.history_content';

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
}

export interface OutgoingWSMessage {
  type: MessageType;
  conversation_id: string;
  message_id?: string;
  delta?: string;
  finish_reason?: string;
  execution_id?: string;
  tool_name?: string;
  parameters?: unknown;
  result?: unknown;
  error?: string;
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
}

// Chat types
export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system' | 'tool';
  content: string;
  timestamp: Date;
  toolCalls?: ToolCall[];
  isStreaming?: boolean;
  metrics?: MessageMetrics;
}

export interface ToolCall {
  id: string;
  name: string;
  parameters: Record<string, unknown>;
  result?: unknown;
  status: 'pending' | 'running' | 'completed' | 'failed';
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
