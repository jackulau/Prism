import { useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWorkspaceStore } from '../store/workspaceStore';
import { useAppStore } from '../store';
import { useShortcutsStore, isMac } from '../store/shortcutsStore';

interface WorkspaceShortcutsOptions {
  onOpenCommandPalette?: () => void;
  onOpenFilePicker?: () => void;
  onToggleFileTree?: () => void;
  onSendMessage?: () => void;
  onClosePreview?: () => void;
  onFocusInput?: () => void;
  onNewConversation?: () => void;
  onNewWorker?: () => void;
  enabled?: boolean;
}

// Navigation paths mapped to number keys
const NAV_PATHS = [
  '/',           // 1 - Dashboard
  '/workspace',  // 2 - Workspaces
  '/workers',    // 3 - Workers
  '/integrations', // 4 - Integrations
  '/usage',      // 5 - Usage
  '/organization', // 6 - Organization
  '/settings',   // 7 - Settings
] as const;

export function useWorkspaceShortcuts(options: WorkspaceShortcutsOptions = {}) {
  const { enabled = true } = options;
  const navigate = useNavigate();

  const {
    toggleFileTree,
    togglePreview,
    clearContextFiles,
    closeAllFiles,
  } = useWorkspaceStore();

  const {
    toggleSidebar,
    toggleChatPanel,
    toggleMetricsPanel,
    toggleSettingsPanel,
  } = useAppStore();

  const {
    openHelpModal,
    shortcuts,
    getShortcut,
  } = useShortcutsStore();

  // Check if a shortcut is enabled
  const isEnabled = useCallback((id: string) => {
    const shortcut = getShortcut(id);
    return shortcut?.enabled ?? true;
  }, [getShortcut]);

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!enabled) return;

    // Don't trigger shortcuts when typing in input fields (except for specific shortcuts)
    const target = e.target as HTMLElement;
    const isInput = target.tagName === 'INPUT' ||
                    target.tagName === 'TEXTAREA' ||
                    target.isContentEditable;

    const isMod = isMac ? e.metaKey : e.ctrlKey;
    const isShift = e.shiftKey;
    const key = e.key.toLowerCase();

    // Escape - always works, even in inputs
    if (e.key === 'Escape' && isEnabled('close-modal')) {
      options.onClosePreview?.();
      return;
    }

    // Don't process most shortcuts when in input fields
    if (isInput) {
      // Allow Cmd/Ctrl + Enter for sending messages
      if (isMod && e.key === 'Enter' && isEnabled('send-message')) {
        e.preventDefault();
        options.onSendMessage?.();
        return;
      }
      // Allow Cmd/Ctrl + K for command palette (common pattern)
      if (isMod && key === 'k' && isEnabled('command-palette')) {
        e.preventDefault();
        options.onOpenCommandPalette?.();
        return;
      }
      return;
    }

    // Cmd/Ctrl + K - Open command palette
    if (isMod && key === 'k' && !isShift && isEnabled('command-palette')) {
      e.preventDefault();
      options.onOpenCommandPalette?.();
      return;
    }

    // Cmd/Ctrl + P - Open file picker
    if (isMod && key === 'p' && !isShift && isEnabled('file-picker')) {
      e.preventDefault();
      options.onOpenFilePicker?.();
      return;
    }

    // Cmd/Ctrl + Shift + E - Toggle file tree
    if (isMod && isShift && key === 'e' && isEnabled('toggle-explorer')) {
      e.preventDefault();
      toggleFileTree();
      options.onToggleFileTree?.();
      return;
    }

    // Cmd/Ctrl + Shift + P - Toggle preview panel
    if (isMod && isShift && key === 'p' && isEnabled('toggle-preview')) {
      e.preventDefault();
      togglePreview();
      return;
    }

    // Cmd/Ctrl + Enter - Send message
    if (isMod && e.key === 'Enter' && isEnabled('send-message')) {
      e.preventDefault();
      options.onSendMessage?.();
      return;
    }

    // Cmd/Ctrl + Shift + L - Clear context files
    if (isMod && isShift && key === 'l' && isEnabled('clear-context')) {
      e.preventDefault();
      clearContextFiles();
      return;
    }

    // Cmd/Ctrl + W - Close all open files
    if (isMod && key === 'w' && !isShift && isEnabled('close-files')) {
      e.preventDefault();
      closeAllFiles();
      return;
    }

    // Cmd/Ctrl + / - Focus chat input
    if (isMod && e.key === '/' && isEnabled('focus-input')) {
      e.preventDefault();
      options.onFocusInput?.();
      return;
    }

    // Cmd/Ctrl + N - New conversation
    if (isMod && key === 'n' && !isShift && isEnabled('new-workspace')) {
      e.preventDefault();
      options.onNewConversation?.();
      return;
    }

    // Cmd/Ctrl + Shift + N - New worker
    if (isMod && isShift && key === 'n' && isEnabled('new-worker')) {
      e.preventDefault();
      options.onNewWorker?.();
      return;
    }

    // Cmd/Ctrl + , - Open settings
    if (isMod && e.key === ',' && isEnabled('open-settings')) {
      e.preventDefault();
      toggleSettingsPanel();
      return;
    }

    // Cmd/Ctrl + B - Toggle sidebar
    if (isMod && key === 'b' && !isShift && isEnabled('toggle-sidebar')) {
      e.preventDefault();
      toggleSidebar();
      return;
    }

    // Cmd/Ctrl + J - Toggle chat panel
    if (isMod && key === 'j' && !isShift && isEnabled('toggle-chat')) {
      e.preventDefault();
      toggleChatPanel();
      return;
    }

    // Cmd/Ctrl + . - Toggle metrics panel
    if (isMod && e.key === '.' && isEnabled('toggle-metrics')) {
      e.preventDefault();
      toggleMetricsPanel();
      return;
    }

    // Cmd/Ctrl + Shift + ? - Show keyboard shortcuts help
    if (isMod && isShift && (e.key === '?' || (e.key === '/' && isShift)) && isEnabled('show-shortcuts')) {
      e.preventDefault();
      openHelpModal();
      return;
    }

    // Cmd/Ctrl + 1-7 - Quick navigation
    const numKey = parseInt(e.key, 10);
    if (isMod && !isShift && numKey >= 1 && numKey <= 7) {
      const navId = `nav-${numKey}`;
      if (isEnabled(navId) && NAV_PATHS[numKey - 1]) {
        e.preventDefault();
        navigate(NAV_PATHS[numKey - 1]);
        return;
      }
    }
  }, [
    enabled,
    options,
    isEnabled,
    navigate,
    toggleFileTree,
    togglePreview,
    clearContextFiles,
    closeAllFiles,
    toggleSidebar,
    toggleChatPanel,
    toggleMetricsPanel,
    toggleSettingsPanel,
    openHelpModal,
  ]);

  useEffect(() => {
    if (!enabled) return;
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown, enabled]);

  return {
    shortcuts,
  };
}

// Simple hook for a single hotkey
export function useHotkey(
  key: string,
  callback: () => void,
  options: { mod?: boolean; shift?: boolean; alt?: boolean; enabled?: boolean } = {}
) {
  const { enabled = true } = options;
  const callbackRef = useRef(callback);

  // Keep callback ref updated
  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      const modMatch = options.mod
        ? (isMac ? e.metaKey : e.ctrlKey)
        : !(e.metaKey || e.ctrlKey);
      const shiftMatch = options.shift ? e.shiftKey : !e.shiftKey;
      const altMatch = options.alt ? e.altKey : !e.altKey;

      if (e.key.toLowerCase() === key.toLowerCase() && modMatch && shiftMatch && altMatch) {
        e.preventDefault();
        callbackRef.current();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [key, options.mod, options.shift, options.alt, enabled]);
}
