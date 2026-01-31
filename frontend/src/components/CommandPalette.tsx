import { useEffect, useRef, useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, Clock } from 'lucide-react';
import { useCommandPaletteStore } from '../store/commandPaletteStore';
import { useAppStore } from '../store';
import { useWorkspaceStore } from '../store/workspaceStore';
import {
  createCommands,
  filterCommands,
  groupCommands,
  fuzzyMatch,
  type Command,
  type CommandGroup,
} from '../lib/commands';

export function CommandPalette() {
  const navigate = useNavigate();
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  // Command palette store
  const {
    isOpen,
    close,
    searchQuery,
    setSearchQuery,
    recentCommandIds,
    addRecentCommand,
    selectedIndex,
    setSelectedIndex,
  } = useCommandPaletteStore();

  // App store
  const {
    theme,
    setTheme,
    toggleSidebar,
    toggleSettingsPanel,
    extendedThinkingEnabled,
    setExtendedThinkingEnabled,
    clearMessages,
    createNewConversation,
  } = useAppStore();

  // Workspace store
  const { toggleFileTree, clearContextFiles, closeAllFiles } = useWorkspaceStore();

  // Create commands with current state
  const allCommands = useMemo(
    () =>
      createCommands({
        navigate,
        toggleTheme: () => {
          // Toggle between catppuccin-latte (light) and catppuccin-mocha (dark)
          setTheme(theme === 'catppuccin-latte' ? 'catppuccin-mocha' : 'catppuccin-latte');
        },
        currentTheme: theme,
        toggleSidebar,
        toggleSettingsPanel,
        toggleExtendedThinking: () => setExtendedThinkingEnabled(!extendedThinkingEnabled),
        extendedThinkingEnabled,
        clearMessages,
        createNewConversation,
        toggleFileTree,
        clearContextFiles,
        closeAllFiles,
      }),
    [
      navigate,
      theme,
      setTheme,
      toggleSidebar,
      toggleSettingsPanel,
      extendedThinkingEnabled,
      setExtendedThinkingEnabled,
      clearMessages,
      createNewConversation,
      toggleFileTree,
      clearContextFiles,
      closeAllFiles,
    ]
  );

  // Filter commands based on search
  const filteredCommands = useMemo(() => {
    return filterCommands(allCommands, searchQuery);
  }, [allCommands, searchQuery]);

  // Get recent commands
  const recentCommands = useMemo(() => {
    if (searchQuery) return [];
    return recentCommandIds
      .map((id) => allCommands.find((cmd) => cmd.id === id))
      .filter((cmd): cmd is Command => cmd !== undefined);
  }, [recentCommandIds, allCommands, searchQuery]);

  // Group filtered commands
  const groupedCommands = useMemo(() => {
    return groupCommands(filteredCommands);
  }, [filteredCommands]);

  // Flatten for keyboard navigation
  const flatCommands = useMemo(() => {
    const flat: Command[] = [];
    // Add recent commands first if showing
    if (recentCommands.length > 0) {
      flat.push(...recentCommands);
    }
    // Add grouped commands
    for (const group of groupedCommands) {
      flat.push(...group.commands);
    }
    return flat;
  }, [recentCommands, groupedCommands]);

  // Execute command
  const executeCommand = useCallback(
    (command: Command) => {
      addRecentCommand(command.id);
      close();
      // Execute after closing for smooth animation
      setTimeout(() => {
        command.action();
      }, 50);
    },
    [addRecentCommand, close]
  );

  // Focus input when opened
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen]);

  // Keyboard navigation
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex(Math.min(selectedIndex + 1, flatCommands.length - 1));
          break;
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex(Math.max(selectedIndex - 1, 0));
          break;
        case 'Enter':
          e.preventDefault();
          if (flatCommands[selectedIndex]) {
            executeCommand(flatCommands[selectedIndex]);
          }
          break;
        case 'Escape':
          e.preventDefault();
          close();
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, selectedIndex, flatCommands, setSelectedIndex, executeCommand, close]);

  // Scroll selected item into view
  useEffect(() => {
    if (!listRef.current || flatCommands.length === 0) return;

    const selectedElement = listRef.current.querySelector(`[data-index="${selectedIndex}"]`);
    if (selectedElement) {
      selectedElement.scrollIntoView({ block: 'nearest' });
    }
  }, [selectedIndex, flatCommands.length]);

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]"
      onClick={close}
      role="dialog"
      aria-modal="true"
      aria-label="Command palette"
    >
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />

      {/* Modal */}
      <div
        className="relative w-full max-w-lg mx-4 bg-editor-bg border border-editor-border rounded-lg shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Search Input */}
        <div className="flex items-center gap-3 px-4 py-3 border-b border-editor-border">
          <Search size={18} className="text-editor-muted flex-shrink-0" />
          <input
            ref={inputRef}
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Type a command or search..."
            className="flex-1 bg-transparent text-editor-text placeholder:text-editor-muted outline-none text-sm"
            aria-label="Search commands"
          />
          <kbd className="hidden sm:inline-flex px-2 py-0.5 text-xs text-editor-muted bg-editor-selection rounded">
            esc
          </kbd>
        </div>

        {/* Commands List */}
        <div
          ref={listRef}
          className="max-h-[50vh] overflow-y-auto overscroll-contain"
          role="listbox"
        >
          {flatCommands.length === 0 ? (
            <div className="px-4 py-8 text-center text-editor-muted text-sm">
              No commands found
            </div>
          ) : (
            <>
              {/* Recent Commands Section */}
              {recentCommands.length > 0 && (
                <div className="py-2">
                  <div className="px-4 py-1.5 text-xs font-semibold text-editor-muted uppercase tracking-wider flex items-center gap-2">
                    <Clock size={12} />
                    Recent
                  </div>
                  {recentCommands.map((cmd, idx) => (
                    <CommandItem
                      key={`recent-${cmd.id}`}
                      command={cmd}
                      isSelected={selectedIndex === idx}
                      index={idx}
                      onClick={() => executeCommand(cmd)}
                      searchQuery={searchQuery}
                    />
                  ))}
                </div>
              )}

              {/* Grouped Commands */}
              {groupedCommands.map((group) => (
                <CommandGroupSection
                  key={group.category}
                  group={group}
                  selectedIndex={selectedIndex}
                  startIndex={
                    recentCommands.length +
                    groupedCommands
                      .slice(0, groupedCommands.indexOf(group))
                      .reduce((acc, g) => acc + g.commands.length, 0)
                  }
                  onExecute={executeCommand}
                  searchQuery={searchQuery}
                />
              ))}
            </>
          )}
        </div>

        {/* Footer */}
        <div className="px-4 py-2 border-t border-editor-border bg-editor-selection/30 flex items-center gap-4 text-xs text-editor-muted">
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 bg-editor-selection rounded">↑↓</kbd>
            navigate
          </span>
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 bg-editor-selection rounded">↵</kbd>
            select
          </span>
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 bg-editor-selection rounded">esc</kbd>
            close
          </span>
        </div>
      </div>
    </div>
  );
}

interface CommandGroupSectionProps {
  group: CommandGroup;
  selectedIndex: number;
  startIndex: number;
  onExecute: (command: Command) => void;
  searchQuery: string;
}

function CommandGroupSection({
  group,
  selectedIndex,
  startIndex,
  onExecute,
  searchQuery,
}: CommandGroupSectionProps) {
  return (
    <div className="py-2">
      <div className="px-4 py-1.5 text-xs font-semibold text-editor-muted uppercase tracking-wider">
        {group.label}
      </div>
      {group.commands.map((cmd, idx) => {
        const globalIndex = startIndex + idx;
        return (
          <CommandItem
            key={cmd.id}
            command={cmd}
            isSelected={selectedIndex === globalIndex}
            index={globalIndex}
            onClick={() => onExecute(cmd)}
            searchQuery={searchQuery}
          />
        );
      })}
    </div>
  );
}

interface CommandItemProps {
  command: Command;
  isSelected: boolean;
  index: number;
  onClick: () => void;
  searchQuery: string;
}

function CommandItem({ command, isSelected, index, onClick, searchQuery }: CommandItemProps) {
  const Icon = command.icon;

  // Highlight matched characters in title
  const highlightedTitle = useMemo(() => {
    if (!searchQuery) return command.title;

    const { indices } = fuzzyMatch(command.title, searchQuery);
    if (indices.length === 0) return command.title;

    const chars = command.title.split('');
    return chars.map((char, i) => {
      if (indices.includes(i)) {
        return (
          <span key={i} className="text-editor-accent font-medium">
            {char}
          </span>
        );
      }
      return char;
    });
  }, [command.title, searchQuery]);

  return (
    <button
      data-index={index}
      onClick={onClick}
      className={`w-full flex items-center gap-3 px-4 py-2 text-left transition-colors ${
        isSelected
          ? 'bg-editor-accent/10 text-editor-accent'
          : 'text-editor-text hover:bg-editor-selection/50'
      }`}
      role="option"
      aria-selected={isSelected}
    >
      {Icon && (
        <Icon
          size={16}
          className={isSelected ? 'text-editor-accent' : 'text-editor-muted'}
        />
      )}
      <div className="flex-1 min-w-0">
        <div className="text-sm truncate">{highlightedTitle}</div>
        {command.description && (
          <div className="text-xs text-editor-muted truncate">{command.description}</div>
        )}
      </div>
      {command.shortcut && (
        <kbd className="hidden sm:inline-flex px-1.5 py-0.5 text-xs text-editor-muted bg-editor-selection rounded">
          {command.shortcut}
        </kbd>
      )}
    </button>
  );
}
