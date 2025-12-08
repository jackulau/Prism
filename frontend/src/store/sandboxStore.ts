import { create } from 'zustand'
import { wsService } from '../services/websocket'

export interface FileNode {
  name: string
  path: string
  isDirectory: boolean
  children?: FileNode[]
  content?: string
}

export interface TerminalLine {
  content: string
  type: 'stdout' | 'stderr' | 'error' | 'warning' | 'success' | 'info'
  timestamp: number
}

export interface FileHistoryEntry {
  id: string
  file_path: string
  operation: string
  size: number
  created_at: string
}

interface SandboxState {
  // Preview state
  previewUrl: string | null
  previewContent: string | null
  isLoading: boolean
  error: string | null
  refreshKey: number

  // Build state
  isRunning: boolean
  buildStatus: 'idle' | 'building' | 'success' | 'error' | null
  currentBuildId: string | null

  // Files state
  files: FileNode[]
  selectedFile: string | null
  fileContents: Record<string, string>

  // Terminal state
  terminalOutput: TerminalLine[]

  // File history state
  fileHistory: FileHistoryEntry[]
  selectedHistoryEntry: FileHistoryEntry | null
  historyContent: string | null
  isLoadingHistory: boolean
  showHistoryPanel: boolean

  // Actions
  setPreviewUrl: (url: string | null) => void
  setPreviewContent: (content: string | null) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  refreshPreview: () => void
  startBuild: (command?: string, args?: string[]) => void
  stopBuild: () => void
  setBuildStatus: (status: 'idle' | 'building' | 'success' | 'error' | null) => void
  setIsRunning: (running: boolean) => void
  setCurrentBuildId: (id: string | null) => void
  setFiles: (files: FileNode[]) => void
  setSelectedFile: (path: string | null) => void
  setFileContent: (path: string, content: string) => void
  getFileContent: (path: string) => string | null
  addTerminalLine: (content: string, type: TerminalLine['type']) => void
  clearTerminal: () => void
  reset: () => void

  // File history actions
  setFileHistory: (history: FileHistoryEntry[]) => void
  setSelectedHistoryEntry: (entry: FileHistoryEntry | null) => void
  setHistoryContent: (content: string | null) => void
  setIsLoadingHistory: (loading: boolean) => void
  setShowHistoryPanel: (show: boolean) => void
  requestFileHistory: (path?: string) => void
  requestHistoryContent: (historyId: string) => void
}

const initialState = {
  previewUrl: null,
  previewContent: null,
  isLoading: false,
  error: null,
  refreshKey: 0,
  isRunning: false,
  buildStatus: null as 'idle' | 'building' | 'success' | 'error' | null,
  currentBuildId: null as string | null,
  files: [] as FileNode[],
  selectedFile: null as string | null,
  fileContents: {} as Record<string, string>,
  terminalOutput: [] as TerminalLine[],
  // File history state
  fileHistory: [] as FileHistoryEntry[],
  selectedHistoryEntry: null as FileHistoryEntry | null,
  historyContent: null as string | null,
  isLoadingHistory: false,
  showHistoryPanel: false,
}

export const useSandboxStore = create<SandboxState>((set, get) => ({
  ...initialState,

  setPreviewUrl: (url) => set({ previewUrl: url, error: null }),

  setPreviewContent: (content) => set({ previewContent: content, error: null }),

  setLoading: (loading) => set({ isLoading: loading }),

  setError: (error) => set({ error, isLoading: false }),

  refreshPreview: () => set((state) => ({ refreshKey: state.refreshKey + 1 })),

  startBuild: (command?: string, args?: string[]) => {
    // Delegate to wsService which will send the WebSocket message
    // and update state when response arrives
    wsService.startBuild(command, args)
  },

  stopBuild: () => {
    const buildId = get().currentBuildId
    if (buildId) {
      wsService.stopBuild(buildId)
    } else {
      // No build ID, just reset local state
      set({ isRunning: false, buildStatus: 'idle' })
      get().addTerminalLine('Build stopped', 'warning')
    }
  },

  setBuildStatus: (status) => set({ buildStatus: status }),

  setIsRunning: (running) => set({ isRunning: running }),

  setCurrentBuildId: (id) => set({ currentBuildId: id }),

  setFiles: (files) => set({ files }),

  setSelectedFile: (path) => set({ selectedFile: path }),

  setFileContent: (path, content) => set((state) => ({
    fileContents: { ...state.fileContents, [path]: content }
  })),

  getFileContent: (path) => get().fileContents[path] || null,

  addTerminalLine: (content, type) => set((state) => ({
    terminalOutput: [
      ...state.terminalOutput,
      { content, type, timestamp: Date.now() }
    ]
  })),

  clearTerminal: () => set({ terminalOutput: [] }),

  reset: () => set(initialState),

  // File history actions
  setFileHistory: (history) => set({ fileHistory: history }),

  setSelectedHistoryEntry: (entry) => set({ selectedHistoryEntry: entry }),

  setHistoryContent: (content) => set({ historyContent: content }),

  setIsLoadingHistory: (loading) => set({ isLoadingHistory: loading }),

  setShowHistoryPanel: (show) => set({ showHistoryPanel: show }),

  requestFileHistory: (path?: string) => {
    set({ isLoadingHistory: true })
    wsService.requestFileHistory(path)
  },

  requestHistoryContent: (historyId: string) => {
    set({ isLoadingHistory: true })
    wsService.requestHistoryContent(historyId)
  },
}))

// Note: Sandbox WebSocket messages are now handled by the main WebSocket service
// in services/websocket.ts to avoid having two separate connections
