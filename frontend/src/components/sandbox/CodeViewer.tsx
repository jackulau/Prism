import { useState, useEffect, useRef, useCallback } from 'react'
import Editor, { OnMount, OnChange, DiffEditor } from '@monaco-editor/react'
import { ChevronRight, ChevronDown, File, Folder, Copy, Check, Loader2, Save, History, X, RotateCcw } from 'lucide-react'
import { useSandboxStore } from '../../store/sandboxStore'
import { wsService } from '../../services/websocket'
import { apiService } from '../../services/api'

import type { FileNode, FileHistoryEntry } from '../../store/sandboxStore'
import type { editor } from 'monaco-editor'

interface FileTreeItemProps extends FileNode {
  level: number
  selectedFile: string | null
  onSelect: (path: string | null) => void
  dirtyFiles: Set<string>
}

function FileTreeItem({ name, path, isDirectory, children, level, selectedFile, onSelect, dirtyFiles }: FileTreeItemProps) {
  const [isExpanded, setIsExpanded] = useState(level < 2)

  const handleClick = () => {
    if (isDirectory) {
      setIsExpanded(!isExpanded)
    } else {
      onSelect(path)
    }
  }

  const isDirty = dirtyFiles.has(path)

  return (
    <div>
      <button
        onClick={handleClick}
        className={`w-full flex items-center gap-1 px-2 py-1 text-sm hover:bg-editor-border/50 transition-colors ${
          selectedFile === path ? 'bg-editor-accent/20 text-editor-accent' : 'text-editor-text'
        }`}
        style={{ paddingLeft: `${level * 12 + 8}px` }}
      >
        {isDirectory ? (
          <>
            {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
            <Folder size={14} className="text-editor-warning" />
          </>
        ) : (
          <>
            <span className="w-3.5" />
            <File size={14} className="text-editor-muted" />
          </>
        )}
        <span className="truncate">{name}</span>
        {isDirty && <span className="ml-1 text-editor-warning">●</span>}
      </button>
      {isDirectory && isExpanded && children?.map((child) => (
        <FileTreeItem
          key={child.path}
          name={child.name}
          path={child.path}
          isDirectory={child.isDirectory}
          children={child.children}
          content={child.content}
          level={level + 1}
          selectedFile={selectedFile}
          onSelect={onSelect}
          dirtyFiles={dirtyFiles}
        />
      ))}
    </div>
  )
}

export function CodeViewer() {
  const {
    files,
    selectedFile,
    setSelectedFile,
    getFileContent,
    setFileContent,
    fileHistory,
    selectedHistoryEntry,
    historyContent,
    isLoadingHistory,
    showHistoryPanel,
    setShowHistoryPanel,
    setSelectedHistoryEntry,
    requestFileHistory,
    requestHistoryContent,
    setHistoryContent,
  } = useSandboxStore()
  const [copied, setCopied] = useState(false)
  const [isLoadingFile, setIsLoadingFile] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [dirtyFiles, setDirtyFiles] = useState<Set<string>>(new Set())
  const [localContent, setLocalContent] = useState<string | null>(null)
  const [showDiff, setShowDiff] = useState(false)
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null)

  const content = selectedFile ? getFileContent(selectedFile) : null
  const language = selectedFile ? getLanguageFromPath(selectedFile) : 'plaintext'

  // Sync local content when file selection changes or content is loaded
  useEffect(() => {
    if (content !== null) {
      setLocalContent(content)
    }
  }, [content, selectedFile])

  // Request file content when a file is selected and not cached
  useEffect(() => {
    if (selectedFile && !getFileContent(selectedFile)) {
      setIsLoadingFile(true)
      setLocalContent(null)
      wsService.requestFile(selectedFile)
    } else {
      setIsLoadingFile(false)
    }
  }, [selectedFile, getFileContent])

  // Clear loading state when content arrives
  useEffect(() => {
    if (content) {
      setIsLoadingFile(false)
    }
  }, [content])

  const handleEditorDidMount: OnMount = (editor) => {
    editorRef.current = editor
    // Focus the editor when mounted
    editor.focus()
  }

  const handleEditorChange: OnChange = (value) => {
    if (selectedFile && value !== undefined) {
      setLocalContent(value)
      // Mark file as dirty if content differs from original
      const originalContent = getFileContent(selectedFile)
      if (value !== originalContent) {
        setDirtyFiles(prev => new Set(prev).add(selectedFile))
      } else {
        setDirtyFiles(prev => {
          const next = new Set(prev)
          next.delete(selectedFile)
          return next
        })
      }
      setSaveError(null)
    }
  }

  const handleSave = useCallback(async () => {
    if (!selectedFile || localContent === null) return

    setIsSaving(true)
    setSaveError(null)

    try {
      const response = await apiService.writeSandboxFile(selectedFile, localContent)
      if (response.error) {
        setSaveError(response.error)
      } else {
        // Update the store with the saved content
        setFileContent(selectedFile, localContent)
        // Remove from dirty files
        setDirtyFiles(prev => {
          const next = new Set(prev)
          next.delete(selectedFile)
          return next
        })
      }
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to save file')
    } finally {
      setIsSaving(false)
    }
  }, [selectedFile, localContent, setFileContent])

  // Keyboard shortcut for save (Cmd/Ctrl + S)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault()
        if (selectedFile && dirtyFiles.has(selectedFile)) {
          handleSave()
        }
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [selectedFile, dirtyFiles, handleSave])

  const handleCopy = async () => {
    if (localContent) {
      await navigator.clipboard.writeText(localContent)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const isCurrentFileDirty = selectedFile ? dirtyFiles.has(selectedFile) : false

  const handleOpenHistory = () => {
    if (selectedFile) {
      setShowHistoryPanel(true)
      requestFileHistory(selectedFile)
    }
  }

  const handleSelectHistoryEntry = (entry: FileHistoryEntry) => {
    setSelectedHistoryEntry(entry)
    requestHistoryContent(entry.id)
    setShowDiff(true)
  }

  const handleCloseHistory = () => {
    setShowHistoryPanel(false)
    setSelectedHistoryEntry(null)
    setHistoryContent(null)
    setShowDiff(false)
  }

  const handleRestoreFromHistory = async () => {
    if (!selectedHistoryEntry || !historyContent || !selectedFile) return

    setIsSaving(true)
    setSaveError(null)

    try {
      const response = await apiService.writeSandboxFile(selectedFile, historyContent)
      if (response.error) {
        setSaveError(response.error)
      } else {
        setFileContent(selectedFile, historyContent)
        setLocalContent(historyContent)
        setDirtyFiles(prev => {
          const next = new Set(prev)
          next.delete(selectedFile)
          return next
        })
        handleCloseHistory()
      }
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to restore file')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="h-full flex">
      {/* File Tree */}
      <div className="w-64 border-r border-editor-border bg-editor-bg overflow-y-auto flex-shrink-0">
        <div className="p-2 text-xs font-semibold text-editor-muted uppercase tracking-wider border-b border-editor-border">
          Files
        </div>
        <div className="py-1">
          {files.length > 0 ? (
            files.map((file) => (
              <FileTreeItem
                key={file.path}
                name={file.name}
                path={file.path}
                isDirectory={file.isDirectory}
                children={file.children}
                content={file.content}
                level={0}
                selectedFile={selectedFile}
                onSelect={setSelectedFile}
                dirtyFiles={dirtyFiles}
              />
            ))
          ) : (
            <div className="p-4 text-sm text-editor-muted text-center">
              No files available
            </div>
          )}
        </div>
      </div>

      {/* Code Editor */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {selectedFile ? (
          <>
            {/* File Header */}
            <div className="flex items-center justify-between px-4 py-2 border-b border-editor-border bg-editor-bg">
              <div className="flex items-center gap-2 text-sm">
                <File size={14} className="text-editor-muted" />
                <span className="text-editor-text">{selectedFile}</span>
                {isCurrentFileDirty && (
                  <span className="text-editor-warning text-xs">(unsaved)</span>
                )}
              </div>
              <div className="flex items-center gap-2">
                {saveError && (
                  <span className="text-editor-error text-xs">{saveError}</span>
                )}
                <button
                  onClick={handleSave}
                  disabled={!isCurrentFileDirty || isSaving}
                  className={`p-1.5 rounded transition-colors flex items-center gap-1 text-xs ${
                    isCurrentFileDirty
                      ? 'text-editor-accent hover:bg-editor-accent/20'
                      : 'text-editor-muted cursor-not-allowed'
                  }`}
                  title="Save (Cmd/Ctrl + S)"
                >
                  {isSaving ? (
                    <Loader2 size={14} className="animate-spin" />
                  ) : (
                    <Save size={14} />
                  )}
                  <span>Save</span>
                </button>
                <button
                  onClick={handleOpenHistory}
                  className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors flex items-center gap-1 text-xs"
                  title="View file history"
                >
                  <History size={14} />
                  <span>History</span>
                </button>
                <button
                  onClick={handleCopy}
                  className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
                  title="Copy code"
                >
                  {copied ? <Check size={14} className="text-editor-success" /> : <Copy size={14} />}
                </button>
              </div>
            </div>

            {/* Editor Content with optional History Panel */}
            <div className="flex-1 flex overflow-hidden">
              {/* Main Editor */}
              <div className={`flex-1 overflow-hidden ${showHistoryPanel ? 'border-r border-editor-border' : ''}`}>
                {isLoadingFile ? (
                  <div className="flex items-center justify-center h-full">
                    <Loader2 size={24} className="animate-spin text-editor-accent" />
                    <span className="ml-2 text-editor-muted">Loading file...</span>
                  </div>
                ) : showDiff && historyContent !== null && localContent !== null ? (
                  <DiffEditor
                    height="100%"
                    language={language}
                    original={historyContent}
                    modified={localContent}
                    theme="vs-dark"
                    options={{
                      fontSize: 14,
                      fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
                      readOnly: true,
                      renderSideBySide: true,
                      automaticLayout: true,
                    }}
                  />
                ) : localContent !== null ? (
                  <Editor
                    height="100%"
                    language={language}
                    value={localContent}
                    onChange={handleEditorChange}
                    onMount={handleEditorDidMount}
                    theme="vs-dark"
                    options={{
                      fontSize: 14,
                      fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
                      minimap: { enabled: true },
                      scrollBeyondLastLine: false,
                      wordWrap: 'on',
                      lineNumbers: 'on',
                      renderLineHighlight: 'line',
                      tabSize: 2,
                      insertSpaces: true,
                      automaticLayout: true,
                      padding: { top: 16, bottom: 16 },
                      scrollbar: {
                        verticalScrollbarSize: 10,
                        horizontalScrollbarSize: 10,
                      },
                    }}
                  />
                ) : (
                  <div className="p-4 text-editor-muted">Unable to load file content</div>
                )}
              </div>

              {/* History Panel */}
              {showHistoryPanel && (
                <div className="w-80 bg-editor-bg border-l border-editor-border flex flex-col">
                  <div className="p-3 border-b border-editor-border flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <History size={16} className="text-editor-accent" />
                      <span className="text-sm font-medium text-editor-text">File History</span>
                    </div>
                    <button
                      onClick={handleCloseHistory}
                      className="p-1 rounded hover:bg-editor-border/50 text-editor-muted hover:text-editor-text transition-colors"
                    >
                      <X size={16} />
                    </button>
                  </div>

                  {isLoadingHistory ? (
                    <div className="flex-1 flex items-center justify-center">
                      <Loader2 size={20} className="animate-spin text-editor-accent" />
                    </div>
                  ) : fileHistory.length > 0 ? (
                    <div className="flex-1 overflow-y-auto">
                      {fileHistory.map((entry) => (
                        <button
                          key={entry.id}
                          onClick={() => handleSelectHistoryEntry(entry)}
                          className={`w-full p-3 text-left border-b border-editor-border/50 hover:bg-editor-border/30 transition-colors ${
                            selectedHistoryEntry?.id === entry.id ? 'bg-editor-accent/10' : ''
                          }`}
                        >
                          <div className="flex items-center gap-2 mb-1">
                            <span className={`text-xs px-1.5 py-0.5 rounded ${
                              entry.operation === 'create' ? 'bg-green-500/20 text-green-400' :
                              entry.operation === 'update' ? 'bg-blue-500/20 text-blue-400' :
                              entry.operation === 'delete' ? 'bg-red-500/20 text-red-400' :
                              'bg-gray-500/20 text-gray-400'
                            }`}>
                              {entry.operation}
                            </span>
                            <span className="text-xs text-editor-muted">{entry.size} bytes</span>
                          </div>
                          <div className="text-xs text-editor-muted">{entry.created_at}</div>
                        </button>
                      ))}
                    </div>
                  ) : (
                    <div className="flex-1 flex items-center justify-center text-editor-muted text-sm">
                      No history available
                    </div>
                  )}

                  {selectedHistoryEntry && historyContent !== null && (
                    <div className="p-3 border-t border-editor-border">
                      <button
                        onClick={handleRestoreFromHistory}
                        disabled={isSaving}
                        className="w-full flex items-center justify-center gap-2 px-3 py-2 bg-editor-accent text-white rounded hover:bg-editor-accent/80 transition-colors disabled:opacity-50"
                      >
                        {isSaving ? (
                          <Loader2 size={14} className="animate-spin" />
                        ) : (
                          <RotateCcw size={14} />
                        )}
                        <span>Restore this version</span>
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-editor-muted">
            <div className="text-center">
              <File size={48} className="mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">Select a file to edit</p>
              <p className="text-sm mt-1">Choose a file from the tree on the left</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function getLanguageFromPath(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase()
  const languageMap: Record<string, string> = {
    js: 'javascript',
    jsx: 'javascript',
    ts: 'typescript',
    tsx: 'typescript',
    html: 'html',
    css: 'css',
    scss: 'scss',
    less: 'less',
    json: 'json',
    md: 'markdown',
    py: 'python',
    go: 'go',
    rs: 'rust',
    yaml: 'yaml',
    yml: 'yaml',
    sh: 'shell',
    bash: 'shell',
    sql: 'sql',
    xml: 'xml',
    svg: 'xml',
    java: 'java',
    c: 'c',
    cpp: 'cpp',
    h: 'c',
    hpp: 'cpp',
    rb: 'ruby',
    php: 'php',
    swift: 'swift',
    kt: 'kotlin',
    dart: 'dart',
    vue: 'vue',
    svelte: 'svelte',
  }
  return languageMap[ext || ''] || 'plaintext'
}
