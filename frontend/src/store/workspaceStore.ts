import { create } from 'zustand';
import type { FileContext, EditSuggestion, Workspace, Range } from '../types/workspace';
import type { AttributionSummary, AgentInfo } from './sandboxStore';
import { wsService } from '../services/websocket';

interface WorkspaceState {
  // Workspace state
  currentWorkspace: Workspace | null;
  contextFiles: FileContext[];
  activeFile: FileContext | null;
  openFiles: FileContext[];
  pendingEdits: Map<string, EditSuggestion[]>;

  // UI state
  isFileTreeCollapsed: boolean;
  isPreviewCollapsed: boolean;
  fileTreeWidth: number;
  previewWidth: number;
  highlightedRanges: Map<string, Range[]>;

  // Actions - Workspace
  setWorkspace: (workspace: Workspace | null) => void;

  // Actions - File context
  addContextFile: (file: FileContext) => void;
  removeContextFile: (path: string) => void;
  clearContextFiles: () => void;
  updateContextFileContent: (path: string, content: string) => void;

  // Actions - Active file
  setActiveFile: (file: FileContext | null) => void;

  // Actions - Open files
  openFile: (file: FileContext) => void;
  closeFile: (path: string) => void;
  closeAllFiles: () => void;

  // Actions - Edit suggestions
  addEditSuggestion: (edit: EditSuggestion) => void;
  acceptEdit: (editId: string) => void;
  rejectEdit: (editId: string) => void;
  clearPendingEdits: (path: string) => void;

  // Actions - UI
  toggleFileTree: () => void;
  togglePreview: () => void;
  setFileTreeWidth: (width: number) => void;
  setPreviewWidth: (width: number) => void;

  // Actions - Highlights
  setHighlightedRanges: (path: string, ranges: Range[]) => void;
  clearHighlightedRanges: (path: string) => void;

  // Attribution state
  attributionSummary: AttributionSummary | null;
  selectedAgent: string | null;
  selectedTool: string | null;
  agentList: AgentInfo[];
  toolList: string[];

  // Attribution actions
  setAttributionSummary: (summary: AttributionSummary | null) => void;
  filterByAgent: (agentId: string | null) => void;
  filterByTool: (toolName: string | null) => void;
  setAgentList: (agents: AgentInfo[]) => void;
  setToolList: (tools: string[]) => void;
  loadAttributionSummary: () => Promise<void>;
  navigateToMessage: (messageId: string, conversationId: string) => void;
}

const STORAGE_KEYS = {
  fileTreeWidth: 'workspace-file-tree-width',
  previewWidth: 'workspace-preview-width',
  isFileTreeCollapsed: 'workspace-file-tree-collapsed',
  isPreviewCollapsed: 'workspace-preview-collapsed',
};

const getStoredNumber = (key: string, defaultValue: number): number => {
  if (typeof window === 'undefined') return defaultValue;
  const stored = localStorage.getItem(key);
  return stored ? parseInt(stored, 10) : defaultValue;
};

const getStoredBoolean = (key: string, defaultValue: boolean): boolean => {
  if (typeof window === 'undefined') return defaultValue;
  const stored = localStorage.getItem(key);
  return stored ? stored === 'true' : defaultValue;
};

export const useWorkspaceStore = create<WorkspaceState>((set, get) => ({
  // Initial state
  currentWorkspace: null,
  contextFiles: [],
  activeFile: null,
  openFiles: [],
  pendingEdits: new Map(),
  isFileTreeCollapsed: getStoredBoolean(STORAGE_KEYS.isFileTreeCollapsed, false),
  isPreviewCollapsed: getStoredBoolean(STORAGE_KEYS.isPreviewCollapsed, false),
  fileTreeWidth: getStoredNumber(STORAGE_KEYS.fileTreeWidth, 250),
  previewWidth: getStoredNumber(STORAGE_KEYS.previewWidth, 400),
  highlightedRanges: new Map(),

  // Attribution state
  attributionSummary: null,
  selectedAgent: null,
  selectedTool: null,
  agentList: [],
  toolList: [],

  // Workspace actions
  setWorkspace: (workspace) => set({ currentWorkspace: workspace }),

  // File context actions
  addContextFile: (file) => set((state) => {
    // Don't add duplicate files
    if (state.contextFiles.some(f => f.path === file.path)) {
      return state;
    }
    return { contextFiles: [...state.contextFiles, file] };
  }),

  removeContextFile: (path) => set((state) => ({
    contextFiles: state.contextFiles.filter(f => f.path !== path),
  })),

  clearContextFiles: () => set({ contextFiles: [] }),

  updateContextFileContent: (path, content) => set((state) => ({
    contextFiles: state.contextFiles.map(f =>
      f.path === path ? { ...f, content } : f
    ),
  })),

  // Active file actions
  setActiveFile: (file) => set({ activeFile: file }),

  // Open files actions
  openFile: (file) => set((state) => {
    // If file is already open, just set it as active
    if (state.openFiles.some(f => f.path === file.path)) {
      return { activeFile: file };
    }
    return {
      openFiles: [...state.openFiles, file],
      activeFile: file,
    };
  }),

  closeFile: (path) => set((state) => {
    const newOpenFiles = state.openFiles.filter(f => f.path !== path);
    let newActiveFile = state.activeFile;

    // If we're closing the active file, switch to another open file
    if (state.activeFile?.path === path) {
      newActiveFile = newOpenFiles.length > 0 ? newOpenFiles[newOpenFiles.length - 1] : null;
    }

    return {
      openFiles: newOpenFiles,
      activeFile: newActiveFile,
    };
  }),

  closeAllFiles: () => set({
    openFiles: [],
    activeFile: null,
  }),

  // Edit suggestion actions
  addEditSuggestion: (edit) => set((state) => {
    const newPendingEdits = new Map(state.pendingEdits);
    const existingEdits = newPendingEdits.get(edit.path) || [];
    newPendingEdits.set(edit.path, [...existingEdits, edit]);
    return { pendingEdits: newPendingEdits };
  }),

  acceptEdit: (editId) => set((state) => {
    const newPendingEdits = new Map(state.pendingEdits);

    for (const [path, edits] of newPendingEdits) {
      const editIndex = edits.findIndex(e => e.id === editId);
      if (editIndex !== -1) {
        const edit = edits[editIndex];
        // Mark as accepted and remove from pending
        newPendingEdits.set(path, edits.filter(e => e.id !== editId));
        if (newPendingEdits.get(path)?.length === 0) {
          newPendingEdits.delete(path);
        }

        // Update file content in context if it exists
        const contextFile = get().contextFiles.find(f => f.path === path);
        if (contextFile) {
          get().updateContextFileContent(path, edit.modified);
        }

        break;
      }
    }

    return { pendingEdits: newPendingEdits };
  }),

  rejectEdit: (editId) => set((state) => {
    const newPendingEdits = new Map(state.pendingEdits);

    for (const [path, edits] of newPendingEdits) {
      const editIndex = edits.findIndex(e => e.id === editId);
      if (editIndex !== -1) {
        newPendingEdits.set(path, edits.filter(e => e.id !== editId));
        if (newPendingEdits.get(path)?.length === 0) {
          newPendingEdits.delete(path);
        }
        break;
      }
    }

    return { pendingEdits: newPendingEdits };
  }),

  clearPendingEdits: (path) => set((state) => {
    const newPendingEdits = new Map(state.pendingEdits);
    newPendingEdits.delete(path);
    return { pendingEdits: newPendingEdits };
  }),

  // UI actions
  toggleFileTree: () => set((state) => {
    const newValue = !state.isFileTreeCollapsed;
    if (typeof window !== 'undefined') {
      localStorage.setItem(STORAGE_KEYS.isFileTreeCollapsed, String(newValue));
    }
    return { isFileTreeCollapsed: newValue };
  }),

  togglePreview: () => set((state) => {
    const newValue = !state.isPreviewCollapsed;
    if (typeof window !== 'undefined') {
      localStorage.setItem(STORAGE_KEYS.isPreviewCollapsed, String(newValue));
    }
    return { isPreviewCollapsed: newValue };
  }),

  setFileTreeWidth: (width) => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(STORAGE_KEYS.fileTreeWidth, String(width));
    }
    set({ fileTreeWidth: width });
  },

  setPreviewWidth: (width) => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(STORAGE_KEYS.previewWidth, String(width));
    }
    set({ previewWidth: width });
  },

  // Highlight actions
  setHighlightedRanges: (path, ranges) => set((state) => {
    const newHighlights = new Map(state.highlightedRanges);
    newHighlights.set(path, ranges);
    return { highlightedRanges: newHighlights };
  }),

  clearHighlightedRanges: (path) => set((state) => {
    const newHighlights = new Map(state.highlightedRanges);
    newHighlights.delete(path);
    return { highlightedRanges: newHighlights };
  }),

  // Attribution actions
  setAttributionSummary: (summary) => set({ attributionSummary: summary }),

  filterByAgent: (agentId) => set({ selectedAgent: agentId }),

  filterByTool: (toolName) => set({ selectedTool: toolName }),

  setAgentList: (agents) => set({ agentList: agents }),

  setToolList: (tools) => set({ toolList: tools }),

  loadAttributionSummary: async () => {
    // Request attribution summary via WebSocket
    wsService.requestAttributionSummary();
  },

  navigateToMessage: (messageId, conversationId) => {
    // Navigate to conversation page and scroll to message
    // This will be handled by the router/navigation system
    const url = `/conversation/${conversationId}?message=${messageId}`;
    window.location.href = url;
  },
}));
