import { create } from 'zustand';

export interface ShortcutDefinition {
  id: string;
  keys: string[];
  description: string;
  category: ShortcutCategory;
  enabled: boolean;
}

export type ShortcutCategory =
  | 'navigation'
  | 'workspace'
  | 'panels'
  | 'editing'
  | 'general';

export const SHORTCUT_CATEGORIES: Record<ShortcutCategory, string> = {
  navigation: 'Navigation',
  workspace: 'Workspace',
  panels: 'Panels',
  editing: 'Editing',
  general: 'General',
};

// Detect platform for correct modifier key display
export const isMac = typeof navigator !== 'undefined' && /Mac|iPod|iPhone|iPad/.test(navigator.platform);

// Get the display string for the modifier key
export const getModKey = () => isMac ? '⌘' : 'Ctrl';

// Format a key array for display
export function formatShortcut(keys: string[]): string {
  return keys.map(key => {
    if (key === 'Mod') return getModKey();
    if (key === 'Shift') return isMac ? '⇧' : 'Shift';
    if (key === 'Alt') return isMac ? '⌥' : 'Alt';
    if (key === 'Escape') return 'Esc';
    if (key === 'Enter') return '↵';
    if (key === '?') return '?';
    return key;
  }).join(isMac ? '' : '+');
}

// Default shortcut definitions
export const DEFAULT_SHORTCUTS: ShortcutDefinition[] = [
  // Navigation
  { id: 'nav-1', keys: ['Mod', '1'], description: 'Go to Dashboard', category: 'navigation', enabled: true },
  { id: 'nav-2', keys: ['Mod', '2'], description: 'Go to Workspaces', category: 'navigation', enabled: true },
  { id: 'nav-3', keys: ['Mod', '3'], description: 'Go to Workers', category: 'navigation', enabled: true },
  { id: 'nav-4', keys: ['Mod', '4'], description: 'Go to Integrations', category: 'navigation', enabled: true },
  { id: 'nav-5', keys: ['Mod', '5'], description: 'Go to Usage', category: 'navigation', enabled: true },
  { id: 'nav-6', keys: ['Mod', '6'], description: 'Go to Organization', category: 'navigation', enabled: true },
  { id: 'nav-7', keys: ['Mod', '7'], description: 'Go to Settings', category: 'navigation', enabled: true },

  // Workspace
  { id: 'new-workspace', keys: ['Mod', 'N'], description: 'New conversation', category: 'workspace', enabled: true },
  { id: 'new-worker', keys: ['Mod', 'Shift', 'N'], description: 'New worker', category: 'workspace', enabled: true },
  { id: 'command-palette', keys: ['Mod', 'K'], description: 'Open command palette', category: 'workspace', enabled: true },
  { id: 'file-picker', keys: ['Mod', 'P'], description: 'Open file picker', category: 'workspace', enabled: true },
  { id: 'send-message', keys: ['Mod', 'Enter'], description: 'Send message', category: 'workspace', enabled: true },
  { id: 'clear-context', keys: ['Mod', 'Shift', 'L'], description: 'Clear context files', category: 'workspace', enabled: true },
  { id: 'close-files', keys: ['Mod', 'W'], description: 'Close all files', category: 'workspace', enabled: true },
  { id: 'focus-input', keys: ['Mod', '/'], description: 'Focus chat input', category: 'workspace', enabled: true },

  // Panels
  { id: 'toggle-sidebar', keys: ['Mod', 'B'], description: 'Toggle sidebar', category: 'panels', enabled: true },
  { id: 'toggle-chat', keys: ['Mod', 'J'], description: 'Toggle chat panel', category: 'panels', enabled: true },
  { id: 'toggle-metrics', keys: ['Mod', '.'], description: 'Toggle metrics panel', category: 'panels', enabled: true },
  { id: 'toggle-explorer', keys: ['Mod', 'Shift', 'E'], description: 'Toggle file explorer', category: 'panels', enabled: true },
  { id: 'toggle-preview', keys: ['Mod', 'Shift', 'P'], description: 'Toggle preview panel', category: 'panels', enabled: true },

  // General
  { id: 'open-settings', keys: ['Mod', ','], description: 'Open settings', category: 'general', enabled: true },
  { id: 'show-shortcuts', keys: ['Mod', 'Shift', '?'], description: 'Show keyboard shortcuts', category: 'general', enabled: true },
  { id: 'close-modal', keys: ['Escape'], description: 'Close modal/panel', category: 'general', enabled: true },
];

interface ShortcutsState {
  shortcuts: ShortcutDefinition[];
  isHelpModalOpen: boolean;

  // Actions
  openHelpModal: () => void;
  closeHelpModal: () => void;
  toggleHelpModal: () => void;
  setShortcutEnabled: (id: string, enabled: boolean) => void;
  resetShortcuts: () => void;
  getShortcutsByCategory: (category: ShortcutCategory) => ShortcutDefinition[];
  getShortcut: (id: string) => ShortcutDefinition | undefined;
}

const STORAGE_KEY = 'keyboard-shortcuts';

const loadStoredShortcuts = (): ShortcutDefinition[] => {
  if (typeof window === 'undefined') return DEFAULT_SHORTCUTS;

  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored) as Record<string, boolean>;
      return DEFAULT_SHORTCUTS.map(shortcut => ({
        ...shortcut,
        enabled: parsed[shortcut.id] ?? shortcut.enabled,
      }));
    }
  } catch {
    // Ignore errors, use defaults
  }
  return DEFAULT_SHORTCUTS;
};

const saveShortcuts = (shortcuts: ShortcutDefinition[]) => {
  if (typeof window === 'undefined') return;

  const enabledState: Record<string, boolean> = {};
  shortcuts.forEach(s => {
    enabledState[s.id] = s.enabled;
  });
  localStorage.setItem(STORAGE_KEY, JSON.stringify(enabledState));
};

export const useShortcutsStore = create<ShortcutsState>((set, get) => ({
  shortcuts: loadStoredShortcuts(),
  isHelpModalOpen: false,

  openHelpModal: () => set({ isHelpModalOpen: true }),
  closeHelpModal: () => set({ isHelpModalOpen: false }),
  toggleHelpModal: () => set(state => ({ isHelpModalOpen: !state.isHelpModalOpen })),

  setShortcutEnabled: (id, enabled) => set(state => {
    const shortcuts = state.shortcuts.map(s =>
      s.id === id ? { ...s, enabled } : s
    );
    saveShortcuts(shortcuts);
    return { shortcuts };
  }),

  resetShortcuts: () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem(STORAGE_KEY);
    }
    set({ shortcuts: DEFAULT_SHORTCUTS });
  },

  getShortcutsByCategory: (category) => {
    return get().shortcuts.filter(s => s.category === category);
  },

  getShortcut: (id) => {
    return get().shortcuts.find(s => s.id === id);
  },
}));
