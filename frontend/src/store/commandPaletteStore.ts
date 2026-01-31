import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface CommandPaletteState {
  // Open/Close state
  isOpen: boolean;
  open: () => void;
  close: () => void;
  toggle: () => void;

  // Search state
  searchQuery: string;
  setSearchQuery: (query: string) => void;

  // Recent commands (persisted)
  recentCommandIds: string[];
  addRecentCommand: (commandId: string) => void;
  clearRecentCommands: () => void;

  // Selected index for keyboard navigation
  selectedIndex: number;
  setSelectedIndex: (index: number) => void;
  resetSelectedIndex: () => void;
}

const MAX_RECENT_COMMANDS = 5;

export const useCommandPaletteStore = create<CommandPaletteState>()(
  persist(
    (set, get) => ({
      // Open/Close state
      isOpen: false,
      open: () => set({ isOpen: true, searchQuery: '', selectedIndex: 0 }),
      close: () => set({ isOpen: false, searchQuery: '', selectedIndex: 0 }),
      toggle: () => {
        const { isOpen } = get();
        if (isOpen) {
          set({ isOpen: false, searchQuery: '', selectedIndex: 0 });
        } else {
          set({ isOpen: true, searchQuery: '', selectedIndex: 0 });
        }
      },

      // Search state
      searchQuery: '',
      setSearchQuery: (query) => set({ searchQuery: query, selectedIndex: 0 }),

      // Recent commands
      recentCommandIds: [],
      addRecentCommand: (commandId) => {
        const { recentCommandIds } = get();
        // Remove if already exists, then add to front
        const filtered = recentCommandIds.filter((id) => id !== commandId);
        const updated = [commandId, ...filtered].slice(0, MAX_RECENT_COMMANDS);
        set({ recentCommandIds: updated });
      },
      clearRecentCommands: () => set({ recentCommandIds: [] }),

      // Selected index
      selectedIndex: 0,
      setSelectedIndex: (index) => set({ selectedIndex: index }),
      resetSelectedIndex: () => set({ selectedIndex: 0 }),
    }),
    {
      name: 'command-palette-storage',
      partialize: (state) => ({
        recentCommandIds: state.recentCommandIds,
      }),
    }
  )
);
