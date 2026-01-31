import type { LucideIcon } from 'lucide-react';
import {
  Home,
  MessageSquare,
  Bot,
  Plug,
  BarChart3,
  Building,
  Settings,
  Plus,
  Sun,
  Moon,
  Brain,
  FileText,
  X,
  Trash2,
  Download,
  PanelLeft,
  FolderTree,
} from 'lucide-react';
import type { Theme } from '../types';

export type CommandCategory = 'navigation' | 'action' | 'settings' | 'workspace';

export interface Command {
  id: string;
  title: string;
  description?: string;
  icon?: LucideIcon;
  keywords?: string[]; // For fuzzy matching
  category: CommandCategory;
  shortcut?: string; // Display shortcut if exists
  action: () => void;
}

export interface CommandGroup {
  category: CommandCategory;
  label: string;
  commands: Command[];
}

// Category labels for display
export const categoryLabels: Record<CommandCategory, string> = {
  navigation: 'Navigation',
  action: 'Actions',
  settings: 'Settings',
  workspace: 'Workspace',
};

// Check if theme is a dark theme
function isDarkTheme(theme: Theme): boolean {
  return theme !== 'catppuccin-latte';
}

// Factory function to create commands with navigation and store access
export function createCommands(deps: {
  navigate: (path: string) => void;
  toggleTheme: () => void;
  currentTheme: Theme;
  toggleSidebar: () => void;
  toggleSettingsPanel: () => void;
  toggleExtendedThinking: () => void;
  extendedThinkingEnabled: boolean;
  clearMessages: () => void;
  createNewConversation: () => Promise<string | null>;
  toggleFileTree: () => void;
  clearContextFiles: () => void;
  closeAllFiles: () => void;
}): Command[] {
  const {
    navigate,
    toggleTheme,
    currentTheme,
    toggleSidebar,
    toggleSettingsPanel,
    toggleExtendedThinking,
    extendedThinkingEnabled,
    clearMessages,
    createNewConversation,
    toggleFileTree,
    clearContextFiles,
    closeAllFiles,
  } = deps;

  const isDark = isDarkTheme(currentTheme);

  return [
    // Navigation Commands
    {
      id: 'nav-dashboard',
      title: 'Go to Dashboard',
      description: 'Navigate to the main dashboard',
      icon: Home,
      keywords: ['home', 'main', 'overview'],
      category: 'navigation',
      action: () => navigate('/'),
    },
    {
      id: 'nav-workspaces',
      title: 'Go to Workspaces',
      description: 'Navigate to workspaces/chat',
      icon: MessageSquare,
      keywords: ['chat', 'conversation', 'messages'],
      category: 'navigation',
      action: () => navigate('/workspace'),
    },
    {
      id: 'nav-workers',
      title: 'Go to Workers',
      description: 'Manage AI workers',
      icon: Bot,
      keywords: ['agents', 'ai', 'automation'],
      category: 'navigation',
      action: () => navigate('/workers'),
    },
    {
      id: 'nav-integrations',
      title: 'Go to Integrations',
      description: 'Configure integrations',
      icon: Plug,
      keywords: ['connect', 'plugins', 'apis'],
      category: 'navigation',
      action: () => navigate('/integrations'),
    },
    {
      id: 'nav-usage',
      title: 'Go to Usage',
      description: 'View usage statistics',
      icon: BarChart3,
      keywords: ['stats', 'billing', 'metrics'],
      category: 'navigation',
      action: () => navigate('/usage'),
    },
    {
      id: 'nav-organization',
      title: 'Go to Organization',
      description: 'Manage organization settings',
      icon: Building,
      keywords: ['team', 'org', 'company'],
      category: 'navigation',
      action: () => navigate('/organization'),
    },
    {
      id: 'nav-settings',
      title: 'Go to Settings',
      description: 'App settings and preferences',
      icon: Settings,
      keywords: ['preferences', 'config', 'options'],
      category: 'navigation',
      action: () => navigate('/settings'),
    },

    // Action Commands
    {
      id: 'action-new-workspace',
      title: 'New Workspace',
      description: 'Create a new conversation',
      icon: Plus,
      keywords: ['create', 'start', 'new chat'],
      category: 'action',
      action: async () => {
        const id = await createNewConversation();
        if (id) {
          navigate(`/workspace/${id}`);
        }
      },
    },
    {
      id: 'action-clear-chat',
      title: 'Clear Chat',
      description: 'Clear all messages in current chat',
      icon: Trash2,
      keywords: ['delete', 'reset', 'clear messages'],
      category: 'action',
      action: () => clearMessages(),
    },
    {
      id: 'action-export-conversation',
      title: 'Export Conversation',
      description: 'Download conversation as file',
      icon: Download,
      keywords: ['save', 'download', 'backup'],
      category: 'action',
      action: () => {
        // TODO: Implement export functionality
        console.log('Export conversation');
      },
    },

    // Settings Commands
    {
      id: 'settings-toggle-theme',
      title: isDark ? 'Switch to Light Theme' : 'Switch to Dark Theme',
      description: 'Toggle between light and dark mode',
      icon: isDark ? Sun : Moon,
      keywords: ['dark', 'light', 'mode', 'appearance'],
      category: 'settings',
      action: () => toggleTheme(),
    },
    {
      id: 'settings-toggle-sidebar',
      title: 'Toggle Sidebar',
      description: 'Show or hide the sidebar',
      icon: PanelLeft,
      keywords: ['collapse', 'expand', 'menu'],
      category: 'settings',
      shortcut: '⌘ \\',
      action: () => toggleSidebar(),
    },
    {
      id: 'settings-toggle-thinking',
      title: extendedThinkingEnabled ? 'Disable Extended Thinking' : 'Enable Extended Thinking',
      description: 'Toggle extended thinking mode for Claude',
      icon: Brain,
      keywords: ['reasoning', 'analysis', 'deep', 'claude'],
      category: 'settings',
      action: () => toggleExtendedThinking(),
    },
    {
      id: 'settings-open-panel',
      title: 'Open Settings Panel',
      description: 'Open the settings side panel',
      icon: Settings,
      keywords: ['preferences', 'configure'],
      category: 'settings',
      action: () => toggleSettingsPanel(),
    },

    // Workspace Commands
    {
      id: 'workspace-toggle-file-tree',
      title: 'Toggle File Explorer',
      description: 'Show or hide the file tree',
      icon: FolderTree,
      keywords: ['files', 'explorer', 'tree', 'folder'],
      category: 'workspace',
      shortcut: '⌘ ⇧ E',
      action: () => toggleFileTree(),
    },
    {
      id: 'workspace-clear-context',
      title: 'Clear Context Files',
      description: 'Remove all files from context',
      icon: FileText,
      keywords: ['remove', 'context', 'files'],
      category: 'workspace',
      shortcut: '⌘ ⇧ L',
      action: () => clearContextFiles(),
    },
    {
      id: 'workspace-close-files',
      title: 'Close All Files',
      description: 'Close all open file previews',
      icon: X,
      keywords: ['close', 'files', 'tabs'],
      category: 'workspace',
      shortcut: '⌘ W',
      action: () => closeAllFiles(),
    },
  ];
}

// Simple fuzzy match scoring
export function fuzzyMatch(text: string, query: string): { match: boolean; score: number; indices: number[] } {
  if (!query) return { match: true, score: 0, indices: [] };

  const textLower = text.toLowerCase();
  const queryLower = query.toLowerCase();

  // Exact substring match gets highest score
  if (textLower.includes(queryLower)) {
    const startIndex = textLower.indexOf(queryLower);
    const indices = Array.from({ length: queryLower.length }, (_, i) => startIndex + i);
    return { match: true, score: 100, indices };
  }

  // Fuzzy character matching
  let queryIndex = 0;
  let score = 0;
  const indices: number[] = [];
  let lastMatchIndex = -1;

  for (let i = 0; i < textLower.length && queryIndex < queryLower.length; i++) {
    if (textLower[i] === queryLower[queryIndex]) {
      indices.push(i);
      // Consecutive matches score higher
      if (lastMatchIndex === i - 1) {
        score += 10;
      } else {
        score += 5;
      }
      // Start of word matches score higher
      if (i === 0 || textLower[i - 1] === ' ') {
        score += 5;
      }
      lastMatchIndex = i;
      queryIndex++;
    }
  }

  const match = queryIndex === queryLower.length;
  return { match, score: match ? score : 0, indices };
}

// Filter and sort commands by search query
export function filterCommands(commands: Command[], query: string): Command[] {
  if (!query.trim()) {
    return commands;
  }

  const results = commands
    .map((cmd) => {
      // Match against title, description, and keywords
      const titleMatch = fuzzyMatch(cmd.title, query);
      const descMatch = cmd.description ? fuzzyMatch(cmd.description, query) : { match: false, score: 0, indices: [] };
      const keywordMatches = (cmd.keywords || []).map((kw) => fuzzyMatch(kw, query));
      const bestKeywordMatch = keywordMatches.reduce(
        (best, m) => (m.score > best.score ? m : best),
        { match: false, score: 0, indices: [] }
      );

      const bestMatch = [titleMatch, descMatch, bestKeywordMatch].reduce(
        (best, m) => (m.score > best.score ? m : best),
        { match: false, score: 0, indices: [] }
      );

      return { command: cmd, score: bestMatch.score, match: bestMatch.match };
    })
    .filter((r) => r.match)
    .sort((a, b) => b.score - a.score);

  return results.map((r) => r.command);
}

// Group commands by category
export function groupCommands(commands: Command[]): CommandGroup[] {
  const groups: Record<CommandCategory, Command[]> = {
    navigation: [],
    action: [],
    settings: [],
    workspace: [],
  };

  for (const cmd of commands) {
    groups[cmd.category].push(cmd);
  }

  return (Object.keys(groups) as CommandCategory[])
    .filter((cat) => groups[cat].length > 0)
    .map((cat) => ({
      category: cat,
      label: categoryLabels[cat],
      commands: groups[cat],
    }));
}
