import { wsService } from './websocket';
import { sseService, SSEService } from './sse';
import type { ChatMode, FileContext } from '../types';

export type StreamingTransport = 'websocket' | 'sse';

interface StreamingServiceOptions {
  transport?: StreamingTransport;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event | Error) => void;
}

/**
 * StreamingService provides a unified interface for both WebSocket and SSE transports.
 * This allows the application to easily switch between streaming protocols based on
 * preference or network conditions.
 */
export class StreamingService {
  private transport: StreamingTransport;
  private token: string | null = null;
  private options: StreamingServiceOptions;

  constructor(options: StreamingServiceOptions = {}) {
    this.transport = options.transport || 'websocket';
    this.options = options;
  }

  /**
   * Set the transport type
   */
  setTransport(transport: StreamingTransport): void {
    if (this.transport !== transport) {
      // Disconnect current transport
      this.disconnect();
      this.transport = transport;
      // Reconnect with new transport if we have a token
      if (this.token) {
        this.connect(this.token);
      }
    }
  }

  /**
   * Get current transport type
   */
  getTransport(): StreamingTransport {
    return this.transport;
  }

  /**
   * Connect to the streaming service
   */
  connect(token: string): void {
    this.token = token;

    if (this.transport === 'sse') {
      sseService.connect({
        token,
        onConnect: this.options.onConnect,
        onDisconnect: this.options.onDisconnect,
        onError: this.options.onError,
      });
    } else {
      wsService.connect(token);
    }
  }

  /**
   * Disconnect from the streaming service
   */
  disconnect(): void {
    if (this.transport === 'sse') {
      sseService.disconnect();
    } else {
      wsService.disconnect();
    }
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    if (this.transport === 'sse') {
      return sseService.isConnected();
    }
    // WebSocket service doesn't expose isConnected, but we can check via store
    return true; // Simplified - actual implementation would check WS state
  }

  /**
   * Send a chat message
   */
  sendChatMessage(
    conversationId: string,
    content: string,
    attachments?: Array<{ name: string; type: string; data: string }>,
    options?: {
      mode?: ChatMode;
      extendedThinking?: boolean;
      fileContext?: FileContext | null;
    }
  ): void {
    if (this.transport === 'sse') {
      sseService.sendChatMessage(conversationId, content, attachments, options);
    } else {
      wsService.sendChatMessage(conversationId, content, attachments, options);
    }
  }

  /**
   * Stop ongoing generation
   */
  stopGeneration(conversationId: string): void {
    if (this.transport === 'sse') {
      sseService.stopGeneration(conversationId);
    } else {
      wsService.stopGeneration(conversationId);
    }
  }

  /**
   * Manual reconnect
   */
  manualReconnect(): void {
    if (this.transport === 'sse') {
      sseService.manualReconnect();
    } else {
      wsService.manualReconnect();
    }
  }

  /**
   * Register an event handler (SSE only - WS uses store directly)
   */
  on<T = unknown>(event: string, handler: (data: T) => void): void {
    if (this.transport === 'sse') {
      sseService.on(event, handler);
    }
  }

  /**
   * Unregister an event handler (SSE only)
   */
  off<T = unknown>(event: string, handler: (data: T) => void): void {
    if (this.transport === 'sse') {
      sseService.off(event, handler);
    }
  }

  /**
   * Confirm or reject a tool execution (SSE only - WS handles via messages)
   */
  async confirmTool(executionId: string, approved: boolean): Promise<void> {
    if (this.transport === 'sse') {
      await sseService.confirmTool(executionId, approved);
    }
    // For WebSocket, tool confirmation is handled via wsService.send()
  }

  // Sandbox methods - delegate to appropriate service
  startBuild(command?: string, args?: string[]): void {
    if (this.transport === 'websocket') {
      wsService.startBuild(command, args);
    }
    // SSE build handling would need a REST endpoint
  }

  stopBuild(buildId: string): void {
    if (this.transport === 'websocket') {
      wsService.stopBuild(buildId);
    }
  }

  requestFile(path: string): Promise<string> {
    if (this.transport === 'websocket') {
      return wsService.requestFile(path);
    }
    // SSE file request would need a REST endpoint
    return Promise.reject(new Error('File request not supported via SSE'));
  }

  requestFileHistory(path?: string): void {
    if (this.transport === 'websocket') {
      wsService.requestFileHistory(path);
    }
  }

  requestHistoryContent(historyId: string): void {
    if (this.transport === 'websocket') {
      wsService.requestHistoryContent(historyId);
    }
  }
}

// Default instance using WebSocket (existing behavior)
export const streamingService = new StreamingService({ transport: 'websocket' });

// Helper to get the appropriate service based on preference
export function getStreamingService(preferSSE: boolean = false): typeof wsService | SSEService {
  return preferSSE ? sseService : wsService;
}

/**
 * Detect the best transport based on environment
 * SSE is better for:
 * - Environments with aggressive proxy timeouts for WebSocket
 * - HTTP/2 connections (SSE performs better)
 * - Simpler debugging needs (SSE is just HTTP)
 *
 * WebSocket is better for:
 * - Bidirectional communication (sandbox, file requests)
 * - Lower latency requirements
 * - Full-duplex needs
 */
export function detectBestTransport(): StreamingTransport {
  // Check if we're behind a proxy that might not support WebSocket well
  // This is a heuristic and could be expanded based on actual usage patterns

  // For now, default to WebSocket as it supports more features
  // Users can opt into SSE via settings
  return 'websocket';
}
