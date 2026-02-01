import { useAppStore } from '../store';
import { useSandboxStore } from '../store/sandboxStore';
import { useAgentProgressStore } from '../store/agentProgressStore';
import type {
  Message,
  SandboxFile,
  ToolCall,
  ChatMode,
  FileContext,
  AgentProgressMessage,
  AgentStepStartedMessage,
  AgentStepCompletedMessage,
  AgentThinkingMessage,
  AgentEstimateMessage,
  SwarmProgressMessage,
} from '../types';
import type { FileNode } from '../store/sandboxStore';

// Event data types matching backend SSE events
interface ChatChunkData {
  conversation_id: string;
  message_id: string;
  delta: string;
}

interface ChatCompleteData {
  conversation_id: string;
  message_id: string;
  finish_reason: string;
}

interface ToolStartedData {
  conversation_id: string;
  execution_id: string;
  tool_name: string;
  parameters: Record<string, unknown>;
  is_mcp_tool?: boolean;
  mcp_server_name?: string;
  is_stdio_mcp?: boolean;
}

interface ToolCompletedData {
  conversation_id: string;
  execution_id: string;
  result: unknown;
  status: string;
}

interface ToolConfirmData {
  conversation_id: string;
  execution_id: string;
  tool_name: string;
  parameters: Record<string, unknown>;
  is_mcp_tool?: boolean;
  mcp_server_name?: string;
  is_stdio_mcp?: boolean;
  iteration_count?: number;
}

interface ErrorData {
  code: string;
  message: string;
}

type EventHandler<T = unknown> = (data: T) => void;

interface SSEOptions {
  token: string;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event) => void;
}

export class SSEService {
  private eventSource: EventSource | null = null;
  private handlers: Map<string, EventHandler[]> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private token: string | null = null;
  private isConnecting = false;
  private intentionalDisconnect = false;
  private options: SSEOptions | null = null;

  /**
   * Connect to the SSE endpoint
   */
  connect(options: SSEOptions): void {
    if (this.isConnecting) {
      return;
    }

    // If already connected with same token, don't reconnect
    if (this.eventSource && this.eventSource.readyState === EventSource.OPEN && this.token === options.token) {
      return;
    }

    // Close existing connection if any
    if (this.eventSource) {
      this.intentionalDisconnect = true;
      this.eventSource.close();
      this.eventSource = null;
    }

    this.isConnecting = true;
    this.intentionalDisconnect = false;
    this.token = options.token;
    this.options = options;

    // Build SSE URL with auth token as query parameter
    // (EventSource doesn't support custom headers, so we use query param)
    const baseUrl = `${window.location.protocol}//${window.location.host}/api/v1/sse`;
    const url = `${baseUrl}?token=${encodeURIComponent(options.token)}`;

    useAppStore.getState().setConnectionStatus('connecting');

    this.eventSource = new EventSource(url);

    this.eventSource.onopen = () => {
      this.isConnecting = false;
      this.reconnectAttempts = 0;
      useAppStore.getState().setConnectionStatus('connected');
      options.onConnect?.();
    };

    this.eventSource.onerror = (error) => {
      this.isConnecting = false;
      useAppStore.getState().setConnectionStatus('error');
      options.onError?.(error);

      // Only attempt reconnect if this wasn't an intentional disconnect
      if (!this.intentionalDisconnect && this.eventSource?.readyState === EventSource.CLOSED) {
        this.attemptReconnect();
      }
    };

    // Register default event handlers
    this.registerDefaultHandlers();
  }

  /**
   * Register handlers for all SSE events
   */
  private registerDefaultHandlers(): void {
    if (!this.eventSource) return;

    // Connection event
    this.eventSource.addEventListener('connected', (e) => {
      const data = JSON.parse(e.data);
      console.log('SSE connected:', data);
    });

    // Chat events
    this.eventSource.addEventListener('chat.chunk', (e) => {
      const data: ChatChunkData = JSON.parse(e.data);
      this.handleChatChunk(data);
    });

    this.eventSource.addEventListener('chat.complete', (e) => {
      const data: ChatCompleteData = JSON.parse(e.data);
      this.handleChatComplete(data);
    });

    // Tool events
    this.eventSource.addEventListener('tool.started', (e) => {
      const data: ToolStartedData = JSON.parse(e.data);
      this.handleToolStarted(data);
    });

    this.eventSource.addEventListener('tool.completed', (e) => {
      const data: ToolCompletedData = JSON.parse(e.data);
      this.handleToolCompleted(data);
    });

    this.eventSource.addEventListener('tool.confirm', (e) => {
      const data: ToolConfirmData = JSON.parse(e.data);
      this.handleToolConfirm(data);
    });

    // Error event
    this.eventSource.addEventListener('error', (e) => {
      if (e instanceof MessageEvent) {
        const data: ErrorData = JSON.parse(e.data);
        this.handleError(data);
      }
    });

    // Heartbeat (for connection keep-alive)
    this.eventSource.addEventListener('heartbeat', () => {
      // Heartbeat received - connection is alive
    });

    // Agent events
    this.eventSource.addEventListener('agent.check_in', (e) => {
      const data = JSON.parse(e.data);
      this.handleAgentCheckIn(data);
    });

    // Sandbox events
    this.eventSource.addEventListener('build.started', (e) => {
      const data = JSON.parse(e.data);
      this.handleBuildStarted(data);
    });

    this.eventSource.addEventListener('build.output', (e) => {
      const data = JSON.parse(e.data);
      this.handleBuildOutput(data);
    });

    this.eventSource.addEventListener('build.completed', (e) => {
      const data = JSON.parse(e.data);
      this.handleBuildCompleted(data);
    });

    this.eventSource.addEventListener('files.updated', (e) => {
      const data = JSON.parse(e.data);
      this.handleFilesUpdated(data);
    });

    this.eventSource.addEventListener('file.content', (e) => {
      const data = JSON.parse(e.data);
      this.handleFileContent(data);
    });

    // Agent progress events
    this.eventSource.addEventListener('agent.progress', (e) => {
      const data: AgentProgressMessage = JSON.parse(e.data);
      this.handleAgentProgress(data);
    });

    this.eventSource.addEventListener('agent.step_started', (e) => {
      const data: AgentStepStartedMessage = JSON.parse(e.data);
      this.handleAgentStepStarted(data);
    });

    this.eventSource.addEventListener('agent.step_completed', (e) => {
      const data: AgentStepCompletedMessage = JSON.parse(e.data);
      this.handleAgentStepCompleted(data);
    });

    this.eventSource.addEventListener('agent.thinking_start', (e) => {
      const data: AgentThinkingMessage = JSON.parse(e.data);
      this.handleAgentThinking(data);
    });

    this.eventSource.addEventListener('agent.thinking_end', (e) => {
      const data: AgentThinkingMessage = JSON.parse(e.data);
      this.handleAgentThinking(data);
    });

    this.eventSource.addEventListener('agent.estimate', (e) => {
      const data: AgentEstimateMessage = JSON.parse(e.data);
      this.handleAgentEstimate(data);
    });

    this.eventSource.addEventListener('swarm.progress', (e) => {
      const data: SwarmProgressMessage = JSON.parse(e.data);
      this.handleSwarmProgress(data);
    });
  }

  /**
   * Handle chat chunk event
   */
  private handleChatChunk(data: ChatChunkData): void {
    const store = useAppStore.getState();

    if (data.delta && store.streamingMessageId) {
      // Record first token timing
      if (store.metrics.firstTokenTime === null) {
        store.recordFirstToken();
      }

      store.appendToMessage(store.streamingMessageId, data.delta);
      const estimatedTokens = this.estimateTokens(data.delta);
      store.incrementTokens(estimatedTokens, data.delta.length);
    }

    this.emit('chat.chunk', data);
  }

  /**
   * Handle chat complete event
   */
  private handleChatComplete(data: ChatCompleteData): void {
    const store = useAppStore.getState();

    if (store.streamingMessageId) {
      store.updateMessage(store.streamingMessageId, {
        isStreaming: false,
        metrics: {
          promptTokens: 0,
          completionTokens: store.metrics.tokenCount,
          totalTokens: store.metrics.tokenCount,
          startTime: store.metrics.startTime ?? 0,
          endTime: performance.now(),
          tokensPerSecond: store.metrics.tokensPerSecond,
          timeToFirstToken: store.metrics.timeToFirstToken ?? undefined,
        },
      });
      store.setStreamingMessageId(null);
      store.endGeneration();

      // Process next queued message
      queueMicrotask(() => {
        const currentStore = useAppStore.getState();
        const nextMessage = currentStore.processNextInQueue();
        if (nextMessage && currentStore.currentConversationId) {
          this.sendChatMessage(currentStore.currentConversationId, nextMessage.content);
        }
      });
    }

    this.emit('chat.complete', data);
  }

  /**
   * Handle tool started event
   */
  private handleToolStarted(data: ToolStartedData): void {
    const store = useAppStore.getState();

    if (store.streamingMessageId) {
      const newToolCall: ToolCall = {
        id: data.execution_id,
        name: data.tool_name,
        parameters: data.parameters,
        status: 'running',
        isMCP: data.is_mcp_tool,
        serverName: data.mcp_server_name,
        isStdioMCP: data.is_stdio_mcp,
      };
      store.addToolCallToMessage(store.streamingMessageId, newToolCall);

      const toolInfo = `\n\n**Using tool:** \`${data.tool_name}\`\n`;
      store.appendToMessage(store.streamingMessageId, toolInfo);
    }

    this.emit('tool.started', data);
  }

  /**
   * Handle tool completed event
   */
  private handleToolCompleted(data: ToolCompletedData): void {
    const store = useAppStore.getState();

    if (store.streamingMessageId) {
      const status: ToolCall['status'] = data.status === 'completed' ? 'completed' :
                                         data.status === 'rejected' ? 'rejected' : 'failed';
      store.updateToolCallStatus(store.streamingMessageId, data.execution_id, status, data.result);
    }

    this.emit('tool.completed', data);
  }

  /**
   * Handle tool confirmation request
   */
  private handleToolConfirm(data: ToolConfirmData): void {
    const store = useAppStore.getState();

    if (store.streamingMessageId) {
      const pendingToolCall: ToolCall = {
        id: data.execution_id,
        name: data.tool_name,
        parameters: data.parameters,
        status: 'pending',
        isMCP: data.is_mcp_tool,
        serverName: data.mcp_server_name,
        isStdioMCP: data.is_stdio_mcp,
      };
      store.addToolCallToMessage(store.streamingMessageId, pendingToolCall);
      store.updateMessage(store.streamingMessageId, { isStreaming: false });
    }

    this.emit('tool.confirm', data);
  }

  /**
   * Handle error event
   */
  private handleError(data: ErrorData): void {
    const store = useAppStore.getState();

    if (store.streamingMessageId) {
      const currentContent = store.messages.find(m => m.id === store.streamingMessageId)?.content ?? '';
      store.updateMessage(store.streamingMessageId, {
        isStreaming: false,
        content: currentContent + `\n\n**Error:** ${data.message}`,
      });
      store.setStreamingMessageId(null);
      store.endGeneration();
    }

    this.emit('error', data);
  }

  /**
   * Handle agent check-in event
   */
  private handleAgentCheckIn(data: { conversation_id: string; iteration_count: number; message: string }): void {
    const store = useAppStore.getState();

    if (store.streamingMessageId) {
      const checkInMsg = data.message || `Agent has executed ${data.iteration_count || 'many'} tool calls. Would you like to continue?`;
      store.appendToMessage(store.streamingMessageId, `\n\n---\n**Agent Check-in:** ${checkInMsg}\n`);
      store.updateMessage(store.streamingMessageId, { isStreaming: false });
      store.setStreamingMessageId(null);
      store.endGeneration();
    }

    this.emit('agent.check_in', data);
  }

  // Sandbox event handlers
  private handleBuildStarted(data: { build_id?: string }): void {
    const sandboxStore = useSandboxStore.getState();
    sandboxStore.setBuildStatus('building');
    sandboxStore.setIsRunning(true);
    if (data.build_id) {
      sandboxStore.setCurrentBuildId(data.build_id);
    }
    sandboxStore.addTerminalLine('Build started...', 'info');
  }

  private handleBuildOutput(data: { content?: string; stream?: string }): void {
    const sandboxStore = useSandboxStore.getState();
    const content = data.content || '';
    const stream = data.stream || 'stdout';
    sandboxStore.addTerminalLine(content, stream === 'stderr' ? 'error' : 'stdout');
  }

  private handleBuildCompleted(data: { success?: boolean; preview_url?: string }): void {
    const sandboxStore = useSandboxStore.getState();
    const success = data.success ?? false;

    sandboxStore.setBuildStatus(success ? 'success' : 'error');
    sandboxStore.setIsRunning(false);
    sandboxStore.setCurrentBuildId(null);
    sandboxStore.addTerminalLine(
      success ? 'Build completed successfully!' : 'Build failed',
      success ? 'success' : 'error'
    );

    if (data.preview_url) {
      sandboxStore.setPreviewUrl(data.preview_url);
    }
  }

  private handleFilesUpdated(data: { files?: SandboxFile[] }): void {
    const sandboxStore = useSandboxStore.getState();
    const files = data.files || [];
    sandboxStore.setFiles(this.convertSandboxFilesToNodes(files));
  }

  private handleFileContent(data: { file_path?: string; content?: string; error?: string }): void {
    const sandboxStore = useSandboxStore.getState();
    const filePath = data.file_path;

    if (!filePath) return;

    if (data.error) {
      console.error('File content error:', data.error);
      return;
    }

    sandboxStore.setFileContent(filePath, data.content || '');
  }

  // Agent Progress handlers
  private handleAgentProgress(data: AgentProgressMessage): void {
    const progressStore = useAgentProgressStore.getState();

    // Start agent if not already tracked
    const existing = progressStore.getAgentProgress(data.agent_id);
    if (!existing) {
      progressStore.startAgent(data.agent_id, data.step_name || 'Agent', data.total_steps);
    }

    progressStore.updateProgress(data.agent_id, {
      currentStep: data.current_step,
      totalSteps: data.total_steps,
      percentComplete: data.percent_complete,
      stepName: data.step_name,
      message: data.message,
      status: 'running',
    });

    this.emit('agent.progress', data);

    if (process.env.NODE_ENV === 'development') {
      console.log('[SSE] Agent progress:', data.agent_id, data.percent_complete + '%');
    }
  }

  private handleAgentStepStarted(data: AgentStepStartedMessage): void {
    const progressStore = useAgentProgressStore.getState();

    progressStore.updateProgress(data.agent_id, {
      currentStep: data.step_number,
      stepName: data.step_name,
      status: 'running',
    });

    progressStore.addProgressEvent(data.agent_id, {
      type: 'step_started',
      data: { step: data.step_number, name: data.step_name },
    });

    this.emit('agent.step_started', data);

    if (process.env.NODE_ENV === 'development') {
      console.log('[SSE] Agent step started:', data.agent_id, data.step_name);
    }
  }

  private handleAgentStepCompleted(data: AgentStepCompletedMessage): void {
    const progressStore = useAgentProgressStore.getState();

    progressStore.addProgressEvent(data.agent_id, {
      type: 'step_completed',
      data: { step: data.step_number, name: data.step_name, result: data.result },
    });

    this.emit('agent.step_completed', data);

    if (process.env.NODE_ENV === 'development') {
      console.log('[SSE] Agent step completed:', data.agent_id, data.step_name);
    }
  }

  private handleAgentThinking(data: AgentThinkingMessage): void {
    const progressStore = useAgentProgressStore.getState();
    const isThinking = data.type === 'agent.thinking_start';

    progressStore.setThinking(data.agent_id, isThinking);

    this.emit(data.type, data);

    if (process.env.NODE_ENV === 'development') {
      console.log('[SSE] Agent thinking:', data.agent_id, isThinking ? 'started' : 'ended');
    }
  }

  private handleAgentEstimate(data: AgentEstimateMessage): void {
    const progressStore = useAgentProgressStore.getState();

    progressStore.updateProgress(data.agent_id, {
      estimatedTokensRemaining: data.estimated_tokens_remaining,
      estimatedTimeRemaining: data.estimated_time_ms,
    });

    this.emit('agent.estimate', data);

    if (process.env.NODE_ENV === 'development') {
      console.log('[SSE] Agent estimate:', data.agent_id,
        `${data.estimated_time_ms}ms remaining, confidence: ${data.confidence}`);
    }
  }

  private handleSwarmProgress(data: SwarmProgressMessage): void {
    const progressStore = useAgentProgressStore.getState();

    // Ensure swarm exists
    const existingSwarm = progressStore.getSwarmProgress(data.swarm_id);
    if (!existingSwarm) {
      // Create swarm with known agent IDs
      const agentIds = data.agent_progress ? Object.keys(data.agent_progress) : [];
      progressStore.startSwarm(data.swarm_id, agentIds);
    }

    // Update individual agent progress from swarm data
    if (data.agent_progress) {
      for (const [agentId, percent] of Object.entries(data.agent_progress)) {
        const agent = progressStore.getAgentProgress(agentId);
        if (agent) {
          progressStore.updateProgress(agentId, { percentComplete: percent });
        }
      }
    }

    // Update swarm progress
    progressStore.updateSwarmProgress(data.swarm_id);

    // Handle completion
    if (data.status === 'completed' || data.status === 'failed') {
      progressStore.completeSwarm(data.swarm_id, data.status as 'completed' | 'failed');
    }

    this.emit('swarm.progress', data);

    if (process.env.NODE_ENV === 'development') {
      console.log('[SSE] Swarm progress:', data.swarm_id,
        `${data.overall_percent}% (${data.completed_agents}/${data.total_agents} agents)`);
    }
  }

  private convertSandboxFilesToNodes(files: SandboxFile[]): FileNode[] {
    return files.map(f => ({
      name: f.name,
      path: f.path,
      isDirectory: f.is_directory,
      children: f.children ? this.convertSandboxFilesToNodes(f.children) : undefined,
    }));
  }

  /**
   * Estimate token count from text
   */
  private estimateTokens(text: string): number {
    const words = text.split(/\s+/).filter(Boolean).length;
    const punctuation = (text.match(/[.,!?;:'"()[\]{}]/g) || []).length;
    const numbers = (text.match(/\d+/g) || []).length;
    const estimate = Math.ceil(words * 1.3 + punctuation * 0.5 + numbers * 0.5);
    return Math.max(estimate, text.length > 0 ? 1 : 0);
  }

  /**
   * Attempt to reconnect with exponential backoff
   */
  private attemptReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts && this.options) {
      this.reconnectAttempts++;
      const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1), 30000);
      useAppStore.getState().setConnectionStatus('connecting');

      setTimeout(() => {
        if (this.options) {
          this.connect(this.options);
        }
      }, delay);
    } else {
      useAppStore.getState().setConnectionStatus('error');
      this.options?.onDisconnect?.();
    }
  }

  /**
   * Register an event handler
   */
  on<T = unknown>(event: string, handler: EventHandler<T>): void {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, []);
    }
    this.handlers.get(event)!.push(handler as EventHandler);
  }

  /**
   * Unregister an event handler
   */
  off<T = unknown>(event: string, handler: EventHandler<T>): void {
    const handlers = this.handlers.get(event);
    if (handlers) {
      const index = handlers.indexOf(handler as EventHandler);
      if (index !== -1) {
        handlers.splice(index, 1);
      }
    }
  }

  /**
   * Emit an event to registered handlers
   */
  private emit(event: string, data: unknown): void {
    const handlers = this.handlers.get(event);
    if (handlers) {
      handlers.forEach(handler => handler(data));
    }
  }

  /**
   * Send a chat message via REST API (SSE is receive-only)
   */
  async sendChatMessage(
    conversationId: string,
    content: string,
    attachments?: Array<{ name: string; type: string; data: string }>,
    options?: {
      mode?: ChatMode;
      extendedThinking?: boolean;
      fileContext?: FileContext | null;
    }
  ): Promise<void> {
    const store = useAppStore.getState();

    // Create user message with attachment info if present
    const attachmentInfo = attachments && attachments.length > 0
      ? `\n\n*Attached: ${attachments.map(a => a.name).join(', ')}*`
      : '';

    const fileContextInfo = options?.fileContext
      ? `\n\n*Context: ${options.fileContext.path}*`
      : '';

    const userMessage: Message = {
      id: `user-${Date.now()}`,
      role: 'user',
      content: content + attachmentInfo + fileContextInfo,
      timestamp: new Date(),
    };
    store.addMessage(userMessage);

    // Create placeholder assistant message
    const assistantMessage: Message = {
      id: `assistant-${Date.now()}`,
      role: 'assistant',
      content: '',
      timestamp: new Date(),
      isStreaming: true,
    };
    store.addMessage(assistantMessage);
    store.setStreamingMessageId(assistantMessage.id);
    store.startGeneration();

    // Send via REST API
    try {
      const response = await fetch('/api/v1/sse/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.token}`,
        },
        body: JSON.stringify({
          conversation_id: conversationId,
          content,
          attachments,
          mode: options?.mode,
          extended_thinking: options?.extendedThinking,
          file_context: options?.fileContext,
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to send message');
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.handleError({ code: 'send_error', message: errorMessage });
    }
  }

  /**
   * Stop ongoing generation
   */
  async stopGeneration(conversationId: string): Promise<void> {
    const store = useAppStore.getState();

    try {
      await fetch('/api/v1/sse/chat/stop', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.token}`,
        },
        body: JSON.stringify({ conversation_id: conversationId }),
      });
    } catch (error) {
      console.error('Failed to stop generation:', error);
    }

    if (store.streamingMessageId) {
      store.updateMessage(store.streamingMessageId, { isStreaming: false });
      store.setStreamingMessageId(null);
      store.endGeneration();
    }
  }

  /**
   * Confirm or reject a tool execution
   */
  async confirmTool(executionId: string, approved: boolean): Promise<void> {
    try {
      await fetch('/api/v1/sse/tool/confirm', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.token}`,
        },
        body: JSON.stringify({
          execution_id: executionId,
          approved,
        }),
      });
    } catch (error) {
      console.error('Failed to confirm tool:', error);
    }
  }

  /**
   * Manual reconnect
   */
  manualReconnect(): void {
    this.intentionalDisconnect = false;
    this.reconnectAttempts = 0;
    if (this.options) {
      this.connect(this.options);
    }
  }

  /**
   * Disconnect from SSE
   */
  disconnect(): void {
    this.intentionalDisconnect = true;
    this.isConnecting = false;
    this.reconnectAttempts = 0;
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    this.options?.onDisconnect?.();
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.eventSource?.readyState === EventSource.OPEN;
  }
}

// Export singleton instance
export const sseService = new SSEService();
