import React, { useState, useEffect, useCallback } from 'react';
import {
  ChevronRight,
  ChevronDown,
  File,
  Folder,
  FolderOpen,
  FileCode,
  FileJson,
  FileText,
  Image,
  Package,
  Settings,
  Database,
  Terminal,
  RefreshCw,
  Plus,
  Search,
  FolderInput,
  Github,
  Loader2,
} from 'lucide-react';
import { useAppStore } from '../store';
import { useSandboxStore, type FileNode as SandboxFileNode } from '../store/sandboxStore';
import { apiService } from '../services/api';
import { wsService } from '../services/websocket';
import type { FileNode } from '../types';

const getFileIcon = (name: string, type: 'file' | 'directory', isExpanded?: boolean) => {
  if (type === 'directory') {
    return isExpanded ? (
      <FolderOpen className="w-4 h-4 text-yellow-400" />
    ) : (
      <Folder className="w-4 h-4 text-yellow-400" />
    );
  }

  const ext = name.split('.').pop()?.toLowerCase();

  switch (ext) {
    case 'ts':
    case 'tsx':
      return <FileCode className="w-4 h-4 text-blue-400" />;
    case 'js':
    case 'jsx':
      return <FileCode className="w-4 h-4 text-yellow-300" />;
    case 'py':
      return <FileCode className="w-4 h-4 text-green-400" />;
    case 'go':
      return <FileCode className="w-4 h-4 text-cyan-400" />;
    case 'rs':
      return <FileCode className="w-4 h-4 text-orange-400" />;
    case 'json':
      return <FileJson className="w-4 h-4 text-yellow-500" />;
    case 'md':
    case 'txt':
      return <FileText className="w-4 h-4 text-gray-400" />;
    case 'png':
    case 'jpg':
    case 'jpeg':
    case 'gif':
    case 'svg':
      return <Image className="w-4 h-4 text-purple-400" />;
    case 'yml':
    case 'yaml':
    case 'toml':
      return <Settings className="w-4 h-4 text-red-400" />;
    case 'sql':
    case 'db':
      return <Database className="w-4 h-4 text-pink-400" />;
    case 'sh':
    case 'bash':
      return <Terminal className="w-4 h-4 text-green-300" />;
    case 'lock':
      return <Package className="w-4 h-4 text-orange-300" />;
    default:
      return <File className="w-4 h-4 text-gray-400" />;
  }
};

interface FileTreeNodeProps {
  node: FileNode;
  depth: number;
}

const FileTreeNode: React.FC<FileTreeNodeProps> = ({ node, depth }) => {
  const { toggleNodeExpanded, selectedFile, setSelectedFile } = useAppStore();
  const { setSelectedFile: setSandboxSelectedFile } = useSandboxStore();
  const isSelected = selectedFile?.id === node.id;

  const handleClick = () => {
    if (node.type === 'directory') {
      toggleNodeExpanded(node.id);
    } else {
      // Set in main store
      setSelectedFile(node);
      // Also set in sandbox store and trigger file load
      setSandboxSelectedFile(node.path);
      // Request file content via WebSocket
      wsService.requestFile(node.path);
    }
  };

  return (
    <div>
      <div
        className={`flex items-center gap-1 py-1 px-2 cursor-pointer transition-smooth hover:bg-sidebar-hover group ${
          isSelected ? 'bg-editor-accent/20 text-editor-accent' : 'text-editor-text'
        }`}
        style={{ paddingLeft: `${depth * 12 + 8}px` }}
        onClick={handleClick}
      >
        {node.type === 'directory' && (
          <span className="w-4 h-4 flex items-center justify-center">
            {node.isExpanded ? (
              <ChevronDown className="w-3 h-3 text-editor-muted" />
            ) : (
              <ChevronRight className="w-3 h-3 text-editor-muted" />
            )}
          </span>
        )}
        {node.type === 'file' && <span className="w-4" />}
        {getFileIcon(node.name, node.type, node.isExpanded)}
        <span className="text-sm truncate flex-1">{node.name}</span>
        {node.type === 'file' && node.size !== undefined && (
          <span className="text-xs text-editor-muted opacity-0 group-hover:opacity-100">
            {formatFileSize(node.size)}
          </span>
        )}
      </div>
      {node.type === 'directory' && node.isExpanded && node.children && (
        <div>
          {node.children.map((child) => (
            <FileTreeNode key={child.id} node={child} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  );
};

const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes}B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)}MB`;
};

// Convert API file response to FileNode format
interface ApiFileNode {
  name: string;
  path: string;
  is_directory: boolean;
  children?: ApiFileNode[];
  size?: number;
  modified?: number;
}

const convertApiFilesToFileNodes = (files: ApiFileNode[], parentPath = ''): FileNode[] => {
  return files.map((file, index) => {
    const id = `${parentPath}-${index}-${file.name}`;
    const node: FileNode = {
      id,
      name: file.name,
      type: file.is_directory ? 'directory' : 'file',
      path: file.path,
      size: file.size,
      isExpanded: false,
    };
    if (file.is_directory && file.children) {
      node.children = convertApiFilesToFileNodes(file.children, id);
    }
    return node;
  });
};

// Convert FileNode to sandbox FileNode format
const convertToSandboxFileNode = (node: FileNode): SandboxFileNode => {
  return {
    name: node.name,
    path: node.path,
    isDirectory: node.type === 'directory',
    children: node.children?.map(convertToSandboxFileNode),
  };
};

export const FileTree: React.FC = () => {
  const { fileTree, setFileTree } = useAppStore();
  const { setFiles: setSandboxFiles } = useSandboxStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreatingFile, setIsCreatingFile] = useState(false);
  const [newFileName, setNewFileName] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentWorkspacePath, setCurrentWorkspacePath] = useState<string>('');
  const [showGitHubPanel, setShowGitHubPanel] = useState(false);
  const [gitHubStatus, setGitHubStatus] = useState<{ connected: boolean; username?: string } | null>(null);
  const [gitHubRepos, setGitHubRepos] = useState<Array<{ id: number; name: string; full_name: string; clone_url: string }>>([]);
  const [isCloning, setIsCloning] = useState(false);
  const [visibleReposCount, setVisibleReposCount] = useState(10);

  // Load files from backend API
  const loadFiles = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      // Load files and workspace path in parallel
      const [filesResponse, workspaceResponse] = await Promise.all([
        apiService.getSandboxFiles(),
        apiService.getWorkspaceDirectory(),
      ]);

      if (filesResponse.data?.files) {
        const nodes = convertApiFilesToFileNodes(filesResponse.data.files);
        setFileTree(nodes);
        // Also sync to sandbox store for CodeViewer
        const sandboxNodes = nodes.map(convertToSandboxFileNode);
        setSandboxFiles(sandboxNodes);
      } else if (filesResponse.error) {
        setError(filesResponse.error);
      }

      if (workspaceResponse.data?.path) {
        setCurrentWorkspacePath(workspaceResponse.data.path);
      }
    } catch (err) {
      setError('Failed to load files');
    } finally {
      setIsLoading(false);
    }
  }, [setFileTree, setSandboxFiles]);

  // Load GitHub status
  const loadGitHubStatus = useCallback(async () => {
    try {
      const response = await apiService.getGitHubStatus();
      if (response.data) {
        setGitHubStatus({ connected: response.data.connected, username: response.data.username });
      }
    } catch {
      // GitHub not configured or not connected
      setGitHubStatus({ connected: false });
    }
  }, []);

  // Initialize - load files on mount
  useEffect(() => {
    loadFiles();
    loadGitHubStatus();
  }, [loadFiles, loadGitHubStatus]);

  const handleRefresh = () => {
    loadFiles();
  };

  // Open native folder picker
  const handleOpenFolderPicker = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await apiService.openFolderPicker();
      if (response.data?.success && response.data.path) {
        setCurrentWorkspacePath(response.data.path);
        loadFiles();
      } else if (response.data?.cancelled) {
        // User cancelled, do nothing
      } else if (response.error) {
        setError(response.error);
      }
    } catch {
      setError('Failed to open folder picker');
    } finally {
      setIsLoading(false);
    }
  };

  // GitHub OAuth
  const handleGitHubConnect = async () => {
    try {
      const response = await apiService.getGitHubAuthUrl();
      if (response.data?.url) {
        window.location.href = response.data.url;
      }
    } catch {
      setError('Failed to initiate GitHub connection');
    }
  };

  // Load GitHub repos
  const handleLoadRepos = async () => {
    try {
      const response = await apiService.getGitHubRepos();
      if (response.data?.repos) {
        setGitHubRepos(response.data.repos);
      }
    } catch {
      setError('Failed to load repositories');
    }
  };

  // Clone GitHub repo
  const handleCloneRepo = async (cloneUrl: string) => {
    setIsCloning(true);
    try {
      const response = await apiService.cloneGitHubRepo(cloneUrl);
      if (response.data?.success) {
        setShowGitHubPanel(false);
        loadFiles();
      } else if (response.error) {
        setError(response.error);
      }
    } catch {
      setError('Failed to clone repository');
    } finally {
      setIsCloning(false);
    }
  };

  const handleCreateFile = async () => {
    if (!newFileName.trim()) return;

    const fileName = newFileName.trim();
    const filePath = `/${fileName}`;

    setIsLoading(true);
    setError(null);

    try {
      // Persist file to backend first
      const response = await apiService.writeSandboxFile(filePath, '');

      if (response.error) {
        setError(response.error);
        return;
      }

      // Refresh file tree from server to ensure consistency
      await loadFiles();

      setNewFileName('');
      setIsCreatingFile(false);
    } catch (err) {
      setError('Failed to create file');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancelCreate = () => {
    setNewFileName('');
    setIsCreatingFile(false);
  };

  const filterTree = (nodes: FileNode[], query: string): FileNode[] => {
    if (!query) return nodes;
    return nodes
      .map((node) => {
        if (node.type === 'directory' && node.children) {
          const filteredChildren = filterTree(node.children, query);
          if (filteredChildren.length > 0) {
            return { ...node, children: filteredChildren, isExpanded: true };
          }
        }
        if (node.name.toLowerCase().includes(query.toLowerCase())) {
          return node;
        }
        return null;
      })
      .filter((node): node is FileNode => node !== null);
  };

  const displayTree = filterTree(fileTree, searchQuery);

  return (
    <div className="h-full flex flex-col bg-sidebar-bg">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-editor-border">
        <span className="text-xs font-semibold text-editor-muted uppercase tracking-wider">
          Explorer
        </span>
        <div className="flex items-center gap-1">
          <button
            onClick={handleOpenFolderPicker}
            disabled={isLoading}
            className="p-1 hover:bg-sidebar-hover rounded transition-smooth disabled:opacity-50"
            title="Open Folder"
          >
            <FolderInput className="w-3.5 h-3.5 text-editor-muted" />
          </button>
          <button
            onClick={() => {
              setShowGitHubPanel(!showGitHubPanel);
              if (!showGitHubPanel && gitHubStatus?.connected) {
                handleLoadRepos();
              }
            }}
            className={`p-1 hover:bg-sidebar-hover rounded transition-smooth ${gitHubStatus?.connected ? 'text-green-400' : 'text-editor-muted'}`}
            title={gitHubStatus?.connected ? `GitHub: ${gitHubStatus.username}` : 'Connect GitHub'}
          >
            <Github className="w-3.5 h-3.5" />
          </button>
          <button
            onClick={handleRefresh}
            className="p-1 hover:bg-sidebar-hover rounded transition-smooth"
            title="Refresh"
            disabled={isLoading}
          >
            {isLoading ? (
              <Loader2 className="w-3.5 h-3.5 text-editor-muted animate-spin" />
            ) : (
              <RefreshCw className="w-3.5 h-3.5 text-editor-muted" />
            )}
          </button>
          <button
            onClick={() => setIsCreatingFile(true)}
            className="p-1 hover:bg-sidebar-hover rounded transition-smooth"
            title="New File"
          >
            <Plus className="w-3.5 h-3.5 text-editor-muted" />
          </button>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="px-2 py-2 bg-red-500/10 border-b border-red-500/20 text-red-400 text-xs">
          {error}
          <button onClick={() => setError(null)} className="ml-2 underline">dismiss</button>
        </div>
      )}

      {/* Current Workspace Path Display */}
      {currentWorkspacePath && (
        <div className="px-2 py-1.5 border-b border-editor-border bg-editor-surface/30">
          <div className="text-xs text-editor-muted truncate font-mono" title={currentWorkspacePath}>
            {currentWorkspacePath}
          </div>
        </div>
      )}

      {/* GitHub Panel */}
      {showGitHubPanel && (
        <div className="px-2 py-2 border-b border-editor-border bg-editor-surface/50 max-h-64 overflow-y-auto">
          <div className="text-xs text-editor-muted mb-2">GitHub Integration</div>
          {!gitHubStatus?.connected ? (
            <button
              onClick={handleGitHubConnect}
              className="w-full px-3 py-2 bg-gray-800 text-white text-sm rounded flex items-center justify-center gap-2 hover:bg-gray-700"
            >
              <Github className="w-4 h-4" />
              Connect GitHub Account
            </button>
          ) : (
            <div>
              <div className="text-xs text-green-400 mb-2">
                Connected as @{gitHubStatus.username}
              </div>
              {gitHubRepos.length > 0 ? (
                <div className="space-y-1">
                  {gitHubRepos.slice(0, visibleReposCount).map((repo) => (
                    <div
                      key={repo.id}
                      className="flex items-center justify-between p-2 bg-editor-surface rounded hover:bg-sidebar-hover cursor-pointer"
                      onClick={() => !isCloning && handleCloneRepo(repo.clone_url)}
                    >
                      <span className="text-xs text-editor-text truncate flex-1">{repo.full_name}</span>
                      {isCloning ? (
                        <Loader2 className="w-3 h-3 text-editor-muted animate-spin" />
                      ) : (
                        <span className="text-xs text-editor-accent">Clone</span>
                      )}
                    </div>
                  ))}
                  {gitHubRepos.length > visibleReposCount && (
                    <button
                      onClick={() => setVisibleReposCount(prev => prev + 10)}
                      className="w-full px-3 py-1.5 text-xs text-editor-accent hover:text-editor-text hover:bg-editor-surface rounded transition-colors"
                    >
                      Load more ({gitHubRepos.length - visibleReposCount} remaining)
                    </button>
                  )}
                </div>
              ) : (
                <button
                  onClick={handleLoadRepos}
                  className="w-full px-3 py-2 bg-editor-surface text-editor-text text-sm rounded hover:bg-sidebar-hover"
                >
                  Load Repositories
                </button>
              )}
            </div>
          )}
        </div>
      )}

      {/* New File Input */}
      {isCreatingFile && (
        <div className="px-2 py-2 border-b border-editor-border">
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={newFileName}
              onChange={(e) => setNewFileName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleCreateFile();
                if (e.key === 'Escape') handleCancelCreate();
              }}
              placeholder="filename.ext"
              className="flex-1 bg-editor-surface text-editor-text text-sm px-2 py-1.5 rounded border border-editor-border focus:border-editor-accent focus:outline-none"
              autoFocus
            />
            <button
              onClick={handleCreateFile}
              disabled={!newFileName.trim()}
              className="p-1.5 bg-editor-accent text-white rounded hover:bg-editor-accent/80 disabled:opacity-50 disabled:cursor-not-allowed"
              title="Create"
            >
              <Plus className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      )}

      {/* Search */}
      <div className="px-2 py-2 border-b border-editor-border">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-editor-muted" />
          <input
            type="text"
            placeholder="Search files..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-editor-surface text-editor-text text-sm pl-7 pr-2 py-1.5 rounded border border-editor-border focus:border-editor-accent focus:outline-none"
          />
        </div>
      </div>

      {/* Tree */}
      <div className="flex-1 overflow-y-auto py-1">
        {displayTree.length > 0 ? (
          displayTree.map((node) => (
            <FileTreeNode key={node.id} node={node} depth={0} />
          ))
        ) : (
          <div className="px-4 py-8 text-center text-editor-muted text-sm">
            No files found
          </div>
        )}
      </div>

      {/* Footer stats */}
      <div className="px-3 py-2 border-t border-editor-border text-xs text-editor-muted">
        {fileTree.length} items
      </div>
    </div>
  );
};
