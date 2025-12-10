import { create } from 'zustand';
import type {
  Message,
  FileNode,
  GenerationMetrics,
  ConnectionStatus,
  Conversation,
  QueuedMessage,
  Theme,
  Provider,
  ToolCall
} from '../types';
import { apiService } from '../services/api';
import { applyTheme, getStoredTheme } from '../config/themes';

interface AppState {
  // Connection
  connectionStatus: ConnectionStatus;
  setConnectionStatus: (status: ConnectionStatus) => void;

  // Messages
  messages: Message[];
  addMessage: (message: Message) => void;
  updateMessage: (id: string, updates: Partial<Message>) => void;
  appendToMessage: (id: string, delta: string) => void;
  clearMessages: () => void;

  // Tool Call Management
  addToolCallToMessage: (messageId: string, toolCall: ToolCall) => void;
  updateToolCallStatus: (messageId: string, toolCallId: string, status: ToolCall['status'], result?: unknown) => void;

  // Current streaming message
  streamingMessageId: string | null;
  setStreamingMessageId: (id: string | null) => void;

  // Conversations
  conversations: Conversation[];
  currentConversationId: string | null;
  setCurrentConversationId: (id: string | null) => void;
  setConversations: (conversations: Conversation[]) => void;

  // File tree
  fileTree: FileNode[];
  setFileTree: (tree: FileNode[]) => void;
  toggleNodeExpanded: (id: string) => void;
  selectedFile: FileNode | null;
  setSelectedFile: (file: FileNode | null) => void;

  // Generation metrics
  metrics: GenerationMetrics;
  updateMetrics: (updates: Partial<GenerationMetrics>) => void;
  resetMetrics: () => void;
  startGeneration: () => void;
  recordFirstToken: () => void;
  incrementTokens: (count: number, chars: number) => void;
  endGeneration: () => void;

  // UI State
  isSidebarOpen: boolean;
  toggleSidebar: () => void;
  isFileTreeOpen: boolean;
  toggleFileTree: () => void;
  isMetricsPanelOpen: boolean;
  toggleMetricsPanel: () => void;

  // Input
  inputValue: string;
  setInputValue: (value: string) => void;

  // Message Queue
  messageQueue: QueuedMessage[];
  addToQueue: (content: string) => void;
  removeFromQueue: (id: string) => void;
  clearQueue: () => void;
  processNextInQueue: () => QueuedMessage | null;
  rollbackToMessage: (messageId: string) => void;

  // Chat Panel Toggle
  isChatPanelOpen: boolean;
  toggleChatPanel: () => void;

  // Settings Panel
  isSettingsPanelOpen: boolean;
  toggleSettingsPanel: () => void;

  // Theme
  theme: Theme;
  setTheme: (theme: Theme) => void;

  // Project Folder
  currentProjectFolder: string;
  setCurrentProjectFolder: (folder: string) => void;

  // Providers
  providers: Provider[];
  selectedProvider: string;
  selectedModel: string;
  loadProviders: () => Promise<void>;
  setSelectedProvider: (provider: string) => void;
  setSelectedModel: (model: string) => void;

  // Conversation Actions
  loadConversations: () => Promise<void>;
  loadMessages: (conversationId: string) => Promise<void>;
  createNewConversation: () => Promise<string | null>;
  deleteConversation: (id: string) => Promise<void>;
}

const initialMetrics: GenerationMetrics = {
  isGenerating: false,
  startTime: null,
  endTime: null,
  firstTokenTime: null,
  tokenCount: 0,
  charCount: 0,
  tokensPerSecond: 0,
  timeToFirstToken: null,
  elapsedTime: 0,
  estimatedTokensRemaining: null,
};

export const useAppStore = create<AppState>((set, get) => ({
  // Connection
  connectionStatus: 'disconnected',
  setConnectionStatus: (status) => set({ connectionStatus: status }),

  // Messages
  messages: [],
  addMessage: (message) => set((state) => ({
    messages: [...state.messages, message]
  })),
  updateMessage: (id, updates) => set((state) => ({
    messages: state.messages.map((msg) =>
      msg.id === id ? { ...msg, ...updates } : msg
    ),
  })),
  appendToMessage: (id, delta) => set((state) => ({
    messages: state.messages.map((msg) =>
      msg.id === id ? { ...msg, content: msg.content + delta } : msg
    ),
  })),
  clearMessages: () => set({ messages: [] }),

  // Tool Call Management
  addToolCallToMessage: (messageId, toolCall) => set((state) => ({
    messages: state.messages.map((msg) =>
      msg.id === messageId
        ? { ...msg, toolCalls: [...(msg.toolCalls || []), toolCall] }
        : msg
    ),
  })),
  updateToolCallStatus: (messageId, toolCallId, status, result) => set((state) => ({
    messages: state.messages.map((msg) =>
      msg.id === messageId
        ? {
            ...msg,
            toolCalls: msg.toolCalls?.map((tc) =>
              tc.id === toolCallId ? { ...tc, status, ...(result !== undefined && { result }) } : tc
            ),
          }
        : msg
    ),
  })),

  // Streaming
  streamingMessageId: null,
  setStreamingMessageId: (id) => set({ streamingMessageId: id }),

  // Conversations
  conversations: [],
  currentConversationId: null,
  setCurrentConversationId: (id) => set({ currentConversationId: id }),
  setConversations: (conversations) => set({ conversations }),

  // File tree
  fileTree: [],
  setFileTree: (tree) => set({ fileTree: tree }),
  toggleNodeExpanded: (id) => set((state) => {
    const toggleNode = (nodes: FileNode[]): FileNode[] =>
      nodes.map((node) => {
        if (node.id === id) {
          return { ...node, isExpanded: !node.isExpanded };
        }
        if (node.children) {
          return { ...node, children: toggleNode(node.children) };
        }
        return node;
      });
    return { fileTree: toggleNode(state.fileTree) };
  }),
  selectedFile: null,
  setSelectedFile: (file) => set({ selectedFile: file }),

  // Generation metrics
  metrics: initialMetrics,
  updateMetrics: (updates) => set((state) => ({
    metrics: { ...state.metrics, ...updates },
  })),
  resetMetrics: () => set({ metrics: initialMetrics }),
  startGeneration: () => {
    const now = performance.now();
    set({
      metrics: {
        ...initialMetrics,
        isGenerating: true,
        startTime: now,
      },
    });
  },
  recordFirstToken: () => {
    const state = get();
    if (state.metrics.firstTokenTime === null && state.metrics.startTime !== null) {
      const now = performance.now();
      set({
        metrics: {
          ...state.metrics,
          firstTokenTime: now,
          timeToFirstToken: now - state.metrics.startTime,
        },
      });
    }
  },
  incrementTokens: (count, chars) => {
    const state = get();
    const now = performance.now();
    const elapsed = state.metrics.startTime ? (now - state.metrics.startTime) / 1000 : 0;
    const newTokenCount = state.metrics.tokenCount + count;
    const tokensPerSecond = elapsed > 0 ? newTokenCount / elapsed : 0;

    set({
      metrics: {
        ...state.metrics,
        tokenCount: newTokenCount,
        charCount: state.metrics.charCount + chars,
        elapsedTime: elapsed * 1000,
        tokensPerSecond: Math.round(tokensPerSecond * 10) / 10,
      },
    });
  },
  endGeneration: () => {
    const state = get();
    const now = performance.now();
    const elapsed = state.metrics.startTime ? (now - state.metrics.startTime) / 1000 : 0;
    const finalTps = elapsed > 0 ? state.metrics.tokenCount / elapsed : 0;

    set({
      metrics: {
        ...state.metrics,
        isGenerating: false,
        endTime: now,
        elapsedTime: elapsed * 1000,
        tokensPerSecond: Math.round(finalTps * 10) / 10,
      },
    });
  },

  // UI State
  isSidebarOpen: true,
  toggleSidebar: () => set((state) => ({ isSidebarOpen: !state.isSidebarOpen })),
  isFileTreeOpen: true,
  toggleFileTree: () => set((state) => ({ isFileTreeOpen: !state.isFileTreeOpen })),
  isMetricsPanelOpen: true,
  toggleMetricsPanel: () => set((state) => ({ isMetricsPanelOpen: !state.isMetricsPanelOpen })),

  // Input
  inputValue: '',
  setInputValue: (value) => set({ inputValue: value }),

  // Message Queue
  messageQueue: [],
  addToQueue: (content) => {
    const state = get();
    const conversation = state.conversations.find(c => c.id === state.currentConversationId);
    set((s) => ({
      messageQueue: [...s.messageQueue, {
        id: `queue-${crypto.randomUUID()}`,
        content,
        createdAt: new Date(),
        status: 'queued',
        model: conversation?.model || 'unknown',
        provider: conversation?.provider || 'unknown',
        projectFolder: state.currentProjectFolder || 'default',
      }],
    }));
  },
  removeFromQueue: (id) => set((state) => ({
    messageQueue: state.messageQueue.filter((m) => m.id !== id),
  })),
  clearQueue: () => set({ messageQueue: [] }),
  processNextInQueue: () => {
    const state = get();
    if (state.messageQueue.length === 0) return null;
    const [next, ...rest] = state.messageQueue;
    set({ messageQueue: rest });
    return next;
  },
  rollbackToMessage: (messageId) => set((state) => {
    const messageIndex = state.messages.findIndex((m) => m.id === messageId);
    if (messageIndex === -1) return state;
    return {
      messages: state.messages.slice(0, messageIndex + 1),
      messageQueue: [],
    };
  }),

  // Chat Panel Toggle
  isChatPanelOpen: true,
  toggleChatPanel: () => set((state) => ({ isChatPanelOpen: !state.isChatPanelOpen })),

  // Settings Panel
  isSettingsPanelOpen: false,
  toggleSettingsPanel: () => set((state) => ({ isSettingsPanelOpen: !state.isSettingsPanelOpen })),

  // Theme
  theme: getStoredTheme(),
  setTheme: (theme) => {
    applyTheme(theme);
    set({ theme });
  },

  // Project Folder
  currentProjectFolder: 'default',
  setCurrentProjectFolder: (folder) => set({ currentProjectFolder: folder }),

  // Providers
  providers: [],
  selectedProvider: 'ollama',
  selectedModel: '',
  loadProviders: async () => {
    const response = await apiService.listProviders();
    if (response.data?.providers) {
      set({ providers: response.data.providers });
      // If Ollama is available and has models, set as default for demo
      const ollama = response.data.providers.find((p: Provider) => p.name === 'ollama');
      if (ollama && ollama.models.length > 0) {
        set({
          selectedProvider: 'ollama',
          selectedModel: ollama.models[0].id,
        });
      } else {
        // Fallback to first available provider with models
        const firstWithModels = response.data.providers.find((p: Provider) => p.models.length > 0);
        if (firstWithModels) {
          set({
            selectedProvider: firstWithModels.name,
            selectedModel: firstWithModels.models[0].id,
          });
        }
      }
    }
  },
  setSelectedProvider: (provider) => set({ selectedProvider: provider }),
  setSelectedModel: (model) => set({ selectedModel: model }),

  // Conversation Actions
  loadConversations: async () => {
    const response = await apiService.listConversations();
    if (response.data?.conversations) {
      set({
        conversations: response.data.conversations.map((c) => ({
          id: c.id,
          title: c.title,
          createdAt: new Date(c.created_at),
          updatedAt: new Date(c.updated_at),
          messageCount: 0,
          provider: c.provider,
          model: c.model,
        })),
      });
    }
  },
  loadMessages: async (conversationId) => {
    const response = await apiService.getMessages(conversationId);
    if (response.data?.messages) {
      set({
        messages: response.data.messages.map((m) => ({
          id: m.id,
          role: m.role as 'user' | 'assistant' | 'system' | 'tool',
          content: m.content,
          timestamp: new Date(m.created_at),
          toolCalls: m.tool_calls as Message['toolCalls'],
        })),
        currentConversationId: conversationId,
      });
    }
  },
  createNewConversation: async () => {
    let { selectedProvider, selectedModel } = get();

    // Ensure we have a valid model selected
    if (!selectedModel) {
      // Try loading providers first
      await get().loadProviders();
      const state = get();
      selectedProvider = state.selectedProvider;
      selectedModel = state.selectedModel;

      // If still no model, cannot create conversation
      if (!selectedModel) {
        console.error('No model selected - cannot create conversation');
        return null;
      }
    }

    const response = await apiService.createConversation(selectedProvider, selectedModel);
    if (response.data) {
      const newConv = response.data;
      set((state) => ({
        conversations: [{
          id: newConv.id,
          title: newConv.title,
          createdAt: new Date(newConv.created_at),
          updatedAt: new Date(newConv.updated_at),
          messageCount: 0,
          provider: newConv.provider,
          model: newConv.model,
        }, ...state.conversations],
        currentConversationId: newConv.id,
        messages: [],
        messageQueue: [],
      }));
      get().resetMetrics();
      return newConv.id;
    }
    return null;
  },
  deleteConversation: async (id) => {
    const response = await apiService.deleteConversation(id);
    if (!response.error) {
      set((state) => ({
        conversations: state.conversations.filter((c) => c.id !== id),
        currentConversationId: state.currentConversationId === id ? null : state.currentConversationId,
        messages: state.currentConversationId === id ? [] : state.messages,
        messageQueue: state.currentConversationId === id ? [] : state.messageQueue,
      }));
    }
  },
}));

// Alias for convenience
export const useStore = useAppStore;
