import { useEffect, useCallback } from 'react';
import { useWorkspaceStore } from '../store/workspaceStore';

interface WorkspaceShortcutsOptions {
  onOpenCommandPalette?: () => void;
  onOpenFilePicker?: () => void;
  onToggleFileTree?: () => void;
  onSendMessage?: () => void;
  onClosePreview?: () => void;
  onFocusInput?: () => void;
}

export function useWorkspaceShortcuts(options: WorkspaceShortcutsOptions = {}) {
  const {
    toggleFileTree,
    togglePreview,
    clearContextFiles,
    closeAllFiles,
  } = useWorkspaceStore();

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    const isMod = e.metaKey || e.ctrlKey;
    const isShift = e.shiftKey;

    // Cmd/Ctrl + K - Open command palette
    if (isMod && e.key === 'k') {
      e.preventDefault();
      options.onOpenCommandPalette?.();
      return;
    }

    // Cmd/Ctrl + P - Open file picker
    if (isMod && e.key === 'p' && !isShift) {
      e.preventDefault();
      options.onOpenFilePicker?.();
      return;
    }

    // Cmd/Ctrl + Shift + E - Toggle file tree
    if (isMod && isShift && e.key === 'e') {
      e.preventDefault();
      toggleFileTree();
      options.onToggleFileTree?.();
      return;
    }

    // Cmd/Ctrl + Shift + P - Toggle preview panel
    if (isMod && isShift && e.key === 'p') {
      e.preventDefault();
      togglePreview();
      return;
    }

    // Cmd/Ctrl + Enter - Send message (if focused in input)
    if (isMod && e.key === 'Enter') {
      e.preventDefault();
      options.onSendMessage?.();
      return;
    }

    // Escape - Close preview or clear context
    if (e.key === 'Escape') {
      options.onClosePreview?.();
      return;
    }

    // Cmd/Ctrl + Shift + L - Clear context files
    if (isMod && isShift && e.key === 'l') {
      e.preventDefault();
      clearContextFiles();
      return;
    }

    // Cmd/Ctrl + W - Close all open files
    if (isMod && e.key === 'w' && !isShift) {
      e.preventDefault();
      closeAllFiles();
      return;
    }

    // Cmd/Ctrl + / - Focus chat input
    if (isMod && e.key === '/') {
      e.preventDefault();
      options.onFocusInput?.();
      return;
    }
  }, [
    options,
    toggleFileTree,
    togglePreview,
    clearContextFiles,
    closeAllFiles,
  ]);

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return {
    shortcuts: [
      { keys: ['Cmd/Ctrl', 'K'], description: 'Open command palette' },
      { keys: ['Cmd/Ctrl', 'P'], description: 'Open file picker' },
      { keys: ['Cmd/Ctrl', 'Shift', 'E'], description: 'Toggle explorer' },
      { keys: ['Cmd/Ctrl', 'Shift', 'P'], description: 'Toggle preview' },
      { keys: ['Cmd/Ctrl', 'Enter'], description: 'Send message' },
      { keys: ['Cmd/Ctrl', 'Shift', 'L'], description: 'Clear context' },
      { keys: ['Cmd/Ctrl', 'W'], description: 'Close all files' },
      { keys: ['Cmd/Ctrl', '/'], description: 'Focus chat input' },
      { keys: ['Escape'], description: 'Close preview' },
    ],
  };
}

// Simple hook for a single hotkey
export function useHotkey(
  key: string,
  callback: () => void,
  options: { mod?: boolean; shift?: boolean; alt?: boolean } = {}
) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const modMatch = options.mod ? (e.metaKey || e.ctrlKey) : !(e.metaKey || e.ctrlKey);
      const shiftMatch = options.shift ? e.shiftKey : !e.shiftKey;
      const altMatch = options.alt ? e.altKey : !e.altKey;

      if (e.key.toLowerCase() === key.toLowerCase() && modMatch && shiftMatch && altMatch) {
        e.preventDefault();
        callback();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [key, callback, options.mod, options.shift, options.alt]);
}
