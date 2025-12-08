import { useAppStore } from '../store';
import { useSandboxStore } from '../store/sandboxStore';
import type { OutgoingWSMessage, IncomingWSMessage, Message, SandboxFile } from '../types';
import type { FileNode } from '../store/sandboxStore';

class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private token: string | null = null;
  private messageQueue: IncomingWSMessage[] = [];

  connect(token?: string) {
    this.token = token || null;
    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/ws`;

    useAppStore.getState().setConnectionStatus('connecting');

    // If token provided, pass via Sec-WebSocket-Protocol header for auth
    // Otherwise connect without auth for anonymous/development mode
    this.ws = token ? new WebSocket(wsUrl, ['auth', token]) : new WebSocket(wsUrl);

    this.ws.onopen = () => {
      useAppStore.getState().setConnectionStatus('connected');
      this.reconnectAttempts = 0;

      // Send queued messages
      while (this.messageQueue.length > 0) {
        const msg = this.messageQueue.shift();
        if (msg) this.send(msg);
      }
    };

    this.ws.onmessage = (event) => {
      try {
        const message: OutgoingWSMessage = JSON.parse(event.data);
        this.handleMessage(message);
      } catch {
        // Failed to parse WebSocket message - ignore malformed data
      }
    };

    this.ws.onclose = () => {
      useAppStore.getState().setConnectionStatus('disconnected');
      this.attemptReconnect();
    };

    this.ws.onerror = () => {
      useAppStore.getState().setConnectionStatus('error');
    };
  }

  private handleMessage(message: OutgoingWSMessage) {
    const store = useAppStore.getState();

    switch (message.type) {
      case 'chat.chunk':
        if (message.delta) {
          // Record first token timing
          if (store.streamingMessageId && store.metrics.firstTokenTime === null) {
            store.recordFirstToken();
          }

          // Append to streaming message
          if (store.streamingMessageId) {
            store.appendToMessage(store.streamingMessageId, message.delta);
            // Better token estimation: ~1.3 tokens per word + punctuation
            const estimatedTokens = this.estimateTokens(message.delta);
            store.incrementTokens(estimatedTokens, message.delta.length);
          }
        }
        break;

      case 'chat.complete':
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

          // Process next queued message if any - use queueMicrotask for immediate processing
          // after current state updates are complete, avoiding race conditions
          queueMicrotask(() => {
            const currentStore = useAppStore.getState();
            const nextMessage = currentStore.processNextInQueue();
            if (nextMessage && currentStore.currentConversationId) {
              this.sendChatMessage(currentStore.currentConversationId, nextMessage.content);
            }
          });
        }
        break;

      case 'tool.started':
        if (store.streamingMessageId) {
          const toolInfo = `\n\n**Using tool:** \`${message.tool_name}\`\n`;
          store.appendToMessage(store.streamingMessageId, toolInfo);
        }
        break;

      case 'tool.completed':
        if (store.streamingMessageId && message.result) {
          const resultStr = typeof message.result === 'string'
            ? message.result
            : JSON.stringify(message.result, null, 2);
          const toolResult = `\n\`\`\`\n${resultStr}\n\`\`\`\n`;
          store.appendToMessage(store.streamingMessageId, toolResult);
        }
        break;

      case 'error':
        if (store.streamingMessageId) {
          const currentContent = store.messages.find(m => m.id === store.streamingMessageId)?.content ?? '';
          store.updateMessage(store.streamingMessageId, {
            isStreaming: false,
            content: currentContent + `\n\n**Error:** ${message.error}`,
          });
          store.setStreamingMessageId(null);
          store.endGeneration();
        }
        break;

      // Sandbox/Preview message handlers
      case 'build.started':
        this.handleBuildStarted(message);
        break;

      case 'build.output':
        this.handleBuildOutput(message);
        break;

      case 'build.completed':
        this.handleBuildCompleted(message);
        break;

      case 'files.updated':
        this.handleFilesUpdated(message);
        break;

      case 'file.content':
        this.handleFileContent(message);
        break;

      case 'preview.ready':
        this.handlePreviewReady(message);
        break;

      case 'preview.content':
        this.handlePreviewContent(message);
        break;

      case 'preview.error':
        this.handlePreviewError(message);
        break;

      case 'file.history_list':
        this.handleFileHistoryList(message);
        break;

      case 'file.history_content':
        this.handleFileHistoryContent(message);
        break;
    }
  }

  // Sandbox message handlers
  private handleBuildStarted(message: OutgoingWSMessage) {
    const sandboxStore = useSandboxStore.getState();
    sandboxStore.setBuildStatus('building');
    sandboxStore.setIsRunning(true);
    if (message.build_id) {
      sandboxStore.setCurrentBuildId(message.build_id);
    }
    sandboxStore.addTerminalLine('Build started...', 'info');
  }

  private handleBuildOutput(message: OutgoingWSMessage) {
    const sandboxStore = useSandboxStore.getState();
    const content = message.content || '';
    const stream = message.stream || 'stdout';
    sandboxStore.addTerminalLine(content, stream === 'stderr' ? 'error' : 'stdout');
  }

  private handleBuildCompleted(message: OutgoingWSMessage) {
    const sandboxStore = useSandboxStore.getState();
    const success = message.success ?? false;

    sandboxStore.setBuildStatus(success ? 'success' : 'error');
    sandboxStore.setIsRunning(false);
    sandboxStore.setCurrentBuildId(null);
    sandboxStore.addTerminalLine(
      success ? 'Build completed successfully!' : 'Build failed',
      success ? 'success' : 'error'
    );

    if (message.preview_url) {
      sandboxStore.setPreviewUrl(message.preview_url);
    }
  }

  private handleFilesUpdated(message: OutgoingWSMessage) {
    const sandboxStore = useSandboxStore.getState();
    const files = message.files || [];
    sandboxStore.setFiles(this.convertSandboxFilesToNodes(files));
  }

  private handleFileContent(message: OutgoingWSMessage) {
    const sandboxStore = useSandboxStore.getState();
    if (message.file_path) {
      sandboxStore.setFileContent(message.file_path, message.content || '');
    }
  }

  private handlePreviewReady(message: OutgoingWSMessage) {
    const sandboxStore = useSandboxStore.getState();
    if (message.url) {
      sandboxStore.setPreviewUrl(message.url);
    }
    sandboxStore.setLoading(false);
  }

  private handlePreviewContent(message: OutgoingWSMessage) {
    const sandboxStore = useSandboxStore.getState();
    sandboxStore.setPreviewContent(message.content || '');
    sandboxStore.setLoading(false);
  }

  private handlePreviewError(message: OutgoingWSMessage) {
    const sandboxStore = useSandboxStore.getState();
    sandboxStore.setError(message.error || 'Preview error');
  }

  private handleFileHistoryList(message: OutgoingWSMessage) {
    const sandboxStore = useSandboxStore.getState();
    const metadata = message.metadata as { entries?: Array<{ id: string; file_path: string; operation: string; size: number; created_at: string }> } | undefined;
    const entries = metadata?.entries || [];
    sandboxStore.setFileHistory(entries);
    sandboxStore.setIsLoadingHistory(false);
  }

  private handleFileHistoryContent(message: OutgoingWSMessage) {
    const sandboxStore = useSandboxStore.getState();
    sandboxStore.setHistoryContent(message.content || '');
    sandboxStore.setIsLoadingHistory(false);
  }

  // Helper to convert SandboxFile[] to FileNode[]
  private convertSandboxFilesToNodes(files: SandboxFile[]): FileNode[] {
    return files.map(f => ({
      name: f.name,
      path: f.path,
      isDirectory: f.is_directory,
      children: f.children ? this.convertSandboxFilesToNodes(f.children) : undefined,
    }));
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1), 30000);
      useAppStore.getState().setConnectionStatus('connecting');
      setTimeout(() => {
        this.connect(this.token || undefined);
      }, delay);
    } else {
      // Max reconnection attempts reached - set failed status for UI feedback
      useAppStore.getState().setConnectionStatus('error');
    }
  }

  // Better token estimation based on word count and punctuation
  private estimateTokens(text: string): number {
    // Count words (split by whitespace)
    const words = text.split(/\s+/).filter(Boolean).length;
    // Count punctuation (these often become separate tokens)
    const punctuation = (text.match(/[.,!?;:'"()[\]{}]/g) || []).length;
    // Count numbers (often tokenized differently)
    const numbers = (text.match(/\d+/g) || []).length;

    // Rough estimate: ~1.3 tokens per word + punctuation + numbers
    // This is more accurate than simple character division
    const estimate = Math.ceil(words * 1.3 + punctuation * 0.5 + numbers * 0.5);

    // Minimum of 1 token if there's any content
    return Math.max(estimate, text.length > 0 ? 1 : 0);
  }

  // Manual reconnect method for when automatic reconnection fails
  manualReconnect() {
    this.reconnectAttempts = 0;
    this.connect(this.token || undefined);
  }

  send(message: IncomingWSMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      this.messageQueue.push(message);
    }
  }

  sendChatMessage(conversationId: string, content: string) {
    const store = useAppStore.getState();

    // Create user message
    const userMessage: Message = {
      id: `user-${Date.now()}`,
      role: 'user',
      content,
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

    // Start tracking metrics
    store.startGeneration();

    // Send message
    this.send({
      type: 'chat.message',
      conversation_id: conversationId,
      content,
    });
  }

  stopGeneration(conversationId: string) {
    this.send({
      type: 'chat.stop',
      conversation_id: conversationId,
    });

    const store = useAppStore.getState();
    if (store.streamingMessageId) {
      store.updateMessage(store.streamingMessageId, { isStreaming: false });
      store.setStreamingMessageId(null);
      store.endGeneration();
    }
  }

  // Sandbox methods
  startBuild(command?: string, args?: string[]) {
    const sandboxStore = useSandboxStore.getState();
    sandboxStore.clearTerminal();
    sandboxStore.setIsRunning(true);
    sandboxStore.setBuildStatus('building');

    this.send({
      type: 'build.start',
      conversation_id: '',
      params: {
        command: command || 'npm',
        args: args || ['run', 'dev'],
      },
    } as IncomingWSMessage);
  }

  stopBuild(buildId: string) {
    const sandboxStore = useSandboxStore.getState();

    this.send({
      type: 'build.stop',
      conversation_id: '',
      params: {
        build_id: buildId,
      },
    } as IncomingWSMessage);

    sandboxStore.setIsRunning(false);
    sandboxStore.setBuildStatus('idle');
    sandboxStore.addTerminalLine('Build stopped by user', 'warning');
  }

  requestFile(path: string) {
    this.send({
      type: 'file.request',
      conversation_id: '',
      params: {
        path,
      },
    } as IncomingWSMessage);
  }

  requestFileHistory(path?: string) {
    this.send({
      type: 'file.history_request',
      conversation_id: '',
      params: {
        action: 'list',
        path: path || '',
        limit: 50,
      },
    } as IncomingWSMessage);
  }

  requestHistoryContent(historyId: string) {
    this.send({
      type: 'file.history_request',
      conversation_id: '',
      params: {
        action: 'get',
        history_id: historyId,
      },
    } as IncomingWSMessage);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

export const wsService = new WebSocketService();
