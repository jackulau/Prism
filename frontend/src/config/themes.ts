import type { Theme, ThemeColors } from '../types';

export interface ThemeConfig {
  id: Theme;
  name: string;
  description: string;
  isDark: boolean;
  colors: ThemeColors;
}

export const themes: Record<Theme, ThemeConfig> = {
  'catppuccin-mocha': {
    id: 'catppuccin-mocha',
    name: 'Catppuccin Mocha',
    description: 'A warm, cozy dark theme',
    isDark: true,
    colors: {
      bg: '#1e1e2e',
      surface: '#181825',
      border: '#313244',
      text: '#cdd6f4',
      muted: '#6c7086',
      accent: '#89b4fa',
      success: '#a6e3a1',
      warning: '#f9e2af',
      error: '#f38ba8',
      sidebarBg: '#11111b',
      sidebarHover: '#1e1e2e',
    },
  },
  'catppuccin-latte': {
    id: 'catppuccin-latte',
    name: 'Catppuccin Latte',
    description: 'A soft, elegant light theme',
    isDark: false,
    colors: {
      bg: '#eff1f5',
      surface: '#e6e9ef',
      border: '#ccd0da',
      text: '#4c4f69',
      muted: '#8c8fa1',
      accent: '#1e66f5',
      success: '#40a02b',
      warning: '#df8e1d',
      error: '#d20f39',
      sidebarBg: '#dce0e8',
      sidebarHover: '#e6e9ef',
    },
  },
  'dracula': {
    id: 'dracula',
    name: 'Dracula',
    description: 'A dark theme with vibrant colors',
    isDark: true,
    colors: {
      bg: '#282a36',
      surface: '#21222c',
      border: '#44475a',
      text: '#f8f8f2',
      muted: '#6272a4',
      accent: '#bd93f9',
      success: '#50fa7b',
      warning: '#f1fa8c',
      error: '#ff5555',
      sidebarBg: '#1e1f29',
      sidebarHover: '#282a36',
    },
  },
  'nord': {
    id: 'nord',
    name: 'Nord',
    description: 'An arctic, north-bluish clean theme',
    isDark: true,
    colors: {
      bg: '#2e3440',
      surface: '#3b4252',
      border: '#4c566a',
      text: '#eceff4',
      muted: '#d8dee9',
      accent: '#88c0d0',
      success: '#a3be8c',
      warning: '#ebcb8b',
      error: '#bf616a',
      sidebarBg: '#242933',
      sidebarHover: '#2e3440',
    },
  },
  'github-dark': {
    id: 'github-dark',
    name: 'GitHub Dark',
    description: "GitHub's official dark theme",
    isDark: true,
    colors: {
      bg: '#0d1117',
      surface: '#161b22',
      border: '#30363d',
      text: '#e6edf3',
      muted: '#8b949e',
      accent: '#58a6ff',
      success: '#3fb950',
      warning: '#d29922',
      error: '#f85149',
      sidebarBg: '#010409',
      sidebarHover: '#0d1117',
    },
  },
  'solarized-dark': {
    id: 'solarized-dark',
    name: 'Solarized Dark',
    description: 'Classic precision color scheme',
    isDark: true,
    colors: {
      bg: '#002b36',
      surface: '#073642',
      border: '#586e75',
      text: '#839496',
      muted: '#657b83',
      accent: '#268bd2',
      success: '#859900',
      warning: '#b58900',
      error: '#dc322f',
      sidebarBg: '#001e26',
      sidebarHover: '#002b36',
    },
  },
  'one-dark': {
    id: 'one-dark',
    name: 'One Dark',
    description: 'Atom editor inspired dark theme',
    isDark: true,
    colors: {
      bg: '#282c34',
      surface: '#21252b',
      border: '#3e4451',
      text: '#abb2bf',
      muted: '#5c6370',
      accent: '#61afef',
      success: '#98c379',
      warning: '#e5c07b',
      error: '#e06c75',
      sidebarBg: '#1e2127',
      sidebarHover: '#282c34',
    },
  },
};

export const defaultTheme: Theme = 'catppuccin-mocha';

export function applyTheme(theme: Theme): void {
  const config = themes[theme];
  if (!config) return;

  const root = document.documentElement;
  const { colors } = config;

  root.style.setProperty('--color-editor-bg', colors.bg);
  root.style.setProperty('--color-editor-surface', colors.surface);
  root.style.setProperty('--color-editor-border', colors.border);
  root.style.setProperty('--color-editor-text', colors.text);
  root.style.setProperty('--color-editor-muted', colors.muted);
  root.style.setProperty('--color-editor-accent', colors.accent);
  root.style.setProperty('--color-editor-success', colors.success);
  root.style.setProperty('--color-editor-warning', colors.warning);
  root.style.setProperty('--color-editor-error', colors.error);
  root.style.setProperty('--color-sidebar-bg', colors.sidebarBg);
  root.style.setProperty('--color-sidebar-hover', colors.sidebarHover);

  // Store in localStorage
  localStorage.setItem('theme', theme);
}

export function getStoredTheme(): Theme {
  const stored = localStorage.getItem('theme') as Theme | null;
  if (stored && themes[stored]) {
    return stored;
  }
  return defaultTheme;
}
