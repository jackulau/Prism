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

export type AgentType = 'assistant' | 'autonomous' | 'workflow'

export interface FileHistoryEntry {
  id: string
  file_path: string
  operation: string
  size: number
  created_at: string
  // Attribution fields
  agent_id?: string
  agent_name?: string
  agent_type?: AgentType
  tool_name?: string
  tool_slug?: string
  message_id?: string
  conversation_id?: string
  workflow_id?: string
  description?: string
}

export interface AttributionSummary {
  total_changes: number
  by_agent: Record<string, number>
  by_tool: Record<string, number>
  by_operation: Record<string, number>
  timeline_by_day: Record<string, number>
  most_active_agent: string
  most_used_tool: string
}

export interface AgentInfo {
  id: string
  name: string
  type: AgentType
}

// Restore types
export interface RestoreConflict {
  file_path: string
  history_timestamp: string
  current_modified: string
  has_newer_version: boolean
}

export interface RestoreResult {
  history_id: string
  file_path: string
  success: boolean
  error?: string
  backup_id?: string
}

// Vercel/Provider sandbox types
export type SandboxProvider = 'docker' | 'vercel'
export type SandboxStatus = 'creating' | 'ready' | 'deploying' | 'deployed' | 'failed' | 'deleted'
export type DeploymentStatus = 'queued' | 'building' | 'ready' | 'error' | 'cancelled'
export type Framework = 'nextjs' | 'react' | 'vue' | 'vite' | 'static'

export interface SandboxInfo {
  id: string
  provider: SandboxProvider
  status: SandboxStatus
  previewUrl?: string
  framework?: Framework
  createdAt?: number
  updatedAt?: number
}

export interface DeploymentInfo {
  id: string
  status: DeploymentStatus
  previewUrl?: string
  error?: string
  createdAt?: number
  readyAt?: number
}

export interface SandboxLogEntry {
  timestamp: number
  message: string
  level: 'info' | 'warn' | 'error'
  source?: string
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
  fileLoadError: string | null

  // Terminal state
  terminalOutput: TerminalLine[]

  // File history state
  fileHistory: FileHistoryEntry[]
  selectedHistoryEntry: FileHistoryEntry | null
  historyContent: string | null
  isLoadingHistory: boolean
  showHistoryPanel: boolean

  // Restore state
  pendingRestore: FileHistoryEntry | null
  isRestoring: boolean
  lastRestoreEntry: FileHistoryEntry | null
  lastRestoreBackupId: string | null
  restoreError: string | null
  restoreConflicts: RestoreConflict[]
  showRestorePreview: boolean
  batchRestoreEntries: FileHistoryEntry[]
  showBatchRestoreDialog: boolean
  batchRestoreResults: RestoreResult[]

  // Vercel/Provider sandbox state
  provider: SandboxProvider
  currentSandbox: SandboxInfo | null
  currentDeployment: DeploymentInfo | null
  sandboxLogs: SandboxLogEntry[]
  availableProviders: SandboxProvider[]

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
  setFileLoadError: (error: string | null) => void
  clearFileLoadError: () => void
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

  // Restore actions
  initiateRestore: (entry: FileHistoryEntry) => void
  confirmRestore: (createBackup?: boolean) => void
  cancelRestore: () => void
  undoLastRestore: () => void
  setRestoreError: (error: string | null) => void
  setRestoreConflicts: (conflicts: RestoreConflict[]) => void
  setIsRestoring: (restoring: boolean) => void
  setShowRestorePreview: (show: boolean) => void
  setBatchRestoreEntries: (entries: FileHistoryEntry[]) => void
  setShowBatchRestoreDialog: (show: boolean) => void
  initiateBatchRestore: (entries: FileHistoryEntry[]) => void
  confirmBatchRestore: (historyIds: string[], createBackup?: boolean) => void
  setBatchRestoreResults: (results: RestoreResult[]) => void
  setLastRestoreEntry: (entry: FileHistoryEntry | null) => void
  setLastRestoreBackupId: (backupId: string | null) => void

  // Vercel/Provider sandbox actions
  setProvider: (provider: SandboxProvider) => void
  setCurrentSandbox: (sandbox: SandboxInfo | null) => void
  setCurrentDeployment: (deployment: DeploymentInfo | null) => void
  addSandboxLog: (log: SandboxLogEntry) => void
  clearSandboxLogs: () => void
  setAvailableProviders: (providers: SandboxProvider[]) => void
  updateSandboxStatus: (status: SandboxStatus) => void
  updateDeploymentStatus: (status: DeploymentStatus, previewUrl?: string, error?: string) => void
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
  fileLoadError: null as string | null,
  terminalOutput: [] as TerminalLine[],
  // File history state
  fileHistory: [] as FileHistoryEntry[],
  selectedHistoryEntry: null as FileHistoryEntry | null,
  historyContent: null as string | null,
  isLoadingHistory: false,
  showHistoryPanel: false,
  // Restore state
  pendingRestore: null as FileHistoryEntry | null,
  isRestoring: false,
  lastRestoreEntry: null as FileHistoryEntry | null,
  lastRestoreBackupId: null as string | null,
  restoreError: null as string | null,
  restoreConflicts: [] as RestoreConflict[],
  showRestorePreview: false,
  batchRestoreEntries: [] as FileHistoryEntry[],
  showBatchRestoreDialog: false,
  batchRestoreResults: [] as RestoreResult[],
  // Vercel/Provider sandbox state
  provider: 'docker' as SandboxProvider,
  currentSandbox: null as SandboxInfo | null,
  currentDeployment: null as DeploymentInfo | null,
  sandboxLogs: [] as SandboxLogEntry[],
  availableProviders: ['docker'] as SandboxProvider[],
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

  setFileLoadError: (error) => set({ fileLoadError: error }),

  clearFileLoadError: () => set({ fileLoadError: null }),

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
    set({ isLoadingHistory: true, error: null })
    wsService.requestFileHistory(path)
    // Timeout fallback in case response never arrives
    setTimeout(() => {
      if (get().isLoadingHistory) {
        set({ isLoadingHistory: false, error: 'History request timed out' })
      }
    }, 10000)
  },

  requestHistoryContent: (historyId: string) => {
    set({ isLoadingHistory: true, error: null })
    wsService.requestHistoryContent(historyId)
    // Timeout fallback in case response never arrives
    setTimeout(() => {
      if (get().isLoadingHistory) {
        set({ isLoadingHistory: false, error: 'Content request timed out' })
      }
    }, 10000)
  },

  // Restore actions
  initiateRestore: (entry: FileHistoryEntry) => {
    set({
      pendingRestore: entry,
      showRestorePreview: true,
      restoreError: null,
      restoreConflicts: [],
    })
  },

  confirmRestore: (createBackup = true) => {
    const pending = get().pendingRestore
    if (!pending) return

    set({ isRestoring: true, restoreError: null })
    wsService.restoreFromHistory(pending.id, createBackup)

    // Timeout fallback
    setTimeout(() => {
      if (get().isRestoring) {
        set({ isRestoring: false, restoreError: 'Restore request timed out' })
      }
    }, 15000)
  },

  cancelRestore: () => {
    set({
      pendingRestore: null,
      showRestorePreview: false,
      restoreError: null,
      restoreConflicts: [],
    })
  },

  undoLastRestore: () => {
    const backupId = get().lastRestoreBackupId
    if (!backupId) return

    set({ isRestoring: true, restoreError: null })
    wsService.restoreFromHistory(backupId, false)

    // Timeout fallback
    setTimeout(() => {
      if (get().isRestoring) {
        set({ isRestoring: false, restoreError: 'Undo restore request timed out' })
      }
    }, 15000)
  },

  setRestoreError: (error) => set({ restoreError: error }),

  setRestoreConflicts: (conflicts) => set({ restoreConflicts: conflicts }),

  setIsRestoring: (restoring) => set({ isRestoring: restoring }),

  setShowRestorePreview: (show) => set({ showRestorePreview: show }),

  setBatchRestoreEntries: (entries) => set({ batchRestoreEntries: entries }),

  setShowBatchRestoreDialog: (show) => set({ showBatchRestoreDialog: show }),

  initiateBatchRestore: (entries: FileHistoryEntry[]) => {
    set({
      batchRestoreEntries: entries,
      showBatchRestoreDialog: true,
      restoreError: null,
      restoreConflicts: [],
      batchRestoreResults: [],
    })
  },

  confirmBatchRestore: (historyIds: string[], createBackup = true) => {
    if (historyIds.length === 0) return

    set({ isRestoring: true, restoreError: null })
    wsService.batchRestoreFromHistory(historyIds, createBackup)

    // Timeout fallback
    setTimeout(() => {
      if (get().isRestoring) {
        set({ isRestoring: false, restoreError: 'Batch restore request timed out' })
      }
    }, 30000)
  },

  setBatchRestoreResults: (results) => set({ batchRestoreResults: results }),

  setLastRestoreEntry: (entry) => set({ lastRestoreEntry: entry }),

  setLastRestoreBackupId: (backupId) => set({ lastRestoreBackupId: backupId }),

  // Vercel/Provider sandbox actions
  setProvider: (provider) => set({ provider }),

  setCurrentSandbox: (sandbox) => set({ currentSandbox: sandbox }),

  setCurrentDeployment: (deployment) => set({ currentDeployment: deployment }),

  addSandboxLog: (log) => set((state) => ({
    sandboxLogs: [...state.sandboxLogs, log]
  })),

  clearSandboxLogs: () => set({ sandboxLogs: [] }),

  setAvailableProviders: (providers) => set({ availableProviders: providers }),

  updateSandboxStatus: (status) => set((state) => ({
    currentSandbox: state.currentSandbox
      ? { ...state.currentSandbox, status, updatedAt: Date.now() }
      : null
  })),

  updateDeploymentStatus: (status, previewUrl, error) => set((state) => {
    const deployment = state.currentDeployment
    if (!deployment) return state

    const updated: DeploymentInfo = {
      ...deployment,
      status,
      ...(previewUrl && { previewUrl }),
      ...(error && { error }),
    }

    if (status === 'ready') {
      updated.readyAt = Date.now()
    }

    // Also update preview URL if provided
    const newState: Partial<SandboxState> = { currentDeployment: updated }
    if (previewUrl) {
      newState.previewUrl = previewUrl
    }

    return newState
  }),
}))

// Note: Sandbox WebSocket messages are now handled by the main WebSocket service
// in services/websocket.ts to avoid having two separate connections
