import { useState, useMemo, useEffect, useRef } from 'react';
import { X, Search, Keyboard } from 'lucide-react';
import {
  useShortcutsStore,
  formatShortcut,
  SHORTCUT_CATEGORIES,
  ShortcutCategory,
  ShortcutDefinition,
} from '../store/shortcutsStore';

export function KeyboardShortcutsHelp() {
  const { isHelpModalOpen, closeHelpModal, shortcuts } = useShortcutsStore();
  const [searchQuery, setSearchQuery] = useState('');
  const searchInputRef = useRef<HTMLInputElement>(null);
  const modalRef = useRef<HTMLDivElement>(null);

  // Focus search input when modal opens
  useEffect(() => {
    if (isHelpModalOpen) {
      // Small delay to ensure modal is rendered
      setTimeout(() => {
        searchInputRef.current?.focus();
      }, 50);
    } else {
      setSearchQuery('');
    }
  }, [isHelpModalOpen]);

  // Handle Escape key to close modal
  useEffect(() => {
    if (!isHelpModalOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        e.stopPropagation();
        closeHelpModal();
      }
    };

    window.addEventListener('keydown', handleKeyDown, true);
    return () => window.removeEventListener('keydown', handleKeyDown, true);
  }, [isHelpModalOpen, closeHelpModal]);

  // Handle click outside to close
  useEffect(() => {
    if (!isHelpModalOpen) return;

    const handleClickOutside = (e: MouseEvent) => {
      if (modalRef.current && !modalRef.current.contains(e.target as Node)) {
        closeHelpModal();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isHelpModalOpen, closeHelpModal]);

  // Filter shortcuts by search query
  const filteredShortcuts = useMemo(() => {
    if (!searchQuery.trim()) return shortcuts;

    const query = searchQuery.toLowerCase();
    return shortcuts.filter(
      (shortcut) =>
        shortcut.description.toLowerCase().includes(query) ||
        shortcut.keys.some((key) => key.toLowerCase().includes(query)) ||
        SHORTCUT_CATEGORIES[shortcut.category].toLowerCase().includes(query)
    );
  }, [shortcuts, searchQuery]);

  // Group filtered shortcuts by category
  const groupedShortcuts = useMemo(() => {
    const groups: Record<ShortcutCategory, ShortcutDefinition[]> = {
      navigation: [],
      workspace: [],
      panels: [],
      editing: [],
      general: [],
    };

    filteredShortcuts.forEach((shortcut) => {
      groups[shortcut.category].push(shortcut);
    });

    return groups;
  }, [filteredShortcuts]);

  if (!isHelpModalOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div
        ref={modalRef}
        className="w-full max-w-2xl max-h-[80vh] bg-editor-bg border border-editor-border rounded-lg shadow-2xl flex flex-col overflow-hidden"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-editor-accent/10 rounded-lg">
              <Keyboard size={20} className="text-editor-accent" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-editor-text">
                Keyboard Shortcuts
              </h2>
              <p className="text-sm text-editor-muted">
                Quick actions to boost your productivity
              </p>
            </div>
          </div>
          <button
            onClick={closeHelpModal}
            className="p-2 rounded-md text-editor-muted hover:text-editor-text hover:bg-editor-hover transition-colors"
            title="Close (Esc)"
          >
            <X size={20} />
          </button>
        </div>

        {/* Search */}
        <div className="px-6 py-3 border-b border-editor-border">
          <div className="relative">
            <Search
              size={16}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-editor-muted"
            />
            <input
              ref={searchInputRef}
              type="text"
              placeholder="Search shortcuts..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-editor-input border border-editor-border rounded-md text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent focus:ring-1 focus:ring-editor-accent"
            />
          </div>
        </div>

        {/* Shortcuts List */}
        <div className="flex-1 overflow-y-auto px-6 py-4">
          {filteredShortcuts.length === 0 ? (
            <div className="text-center py-8 text-editor-muted">
              <p>No shortcuts found matching "{searchQuery}"</p>
            </div>
          ) : (
            <div className="space-y-6">
              {(Object.keys(SHORTCUT_CATEGORIES) as ShortcutCategory[]).map(
                (category) => {
                  const categoryShortcuts = groupedShortcuts[category];
                  if (categoryShortcuts.length === 0) return null;

                  return (
                    <div key={category}>
                      <h3 className="text-xs font-semibold text-editor-muted uppercase tracking-wider mb-3">
                        {SHORTCUT_CATEGORIES[category]}
                      </h3>
                      <div className="space-y-1">
                        {categoryShortcuts.map((shortcut) => (
                          <ShortcutRow
                            key={shortcut.id}
                            shortcut={shortcut}
                          />
                        ))}
                      </div>
                    </div>
                  );
                }
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-3 border-t border-editor-border bg-editor-bg/50">
          <p className="text-xs text-editor-muted text-center">
            Press <kbd className="px-1.5 py-0.5 bg-editor-hover rounded text-editor-text font-mono">Esc</kbd> to close
          </p>
        </div>
      </div>
    </div>
  );
}

interface ShortcutRowProps {
  shortcut: ShortcutDefinition;
}

function ShortcutRow({ shortcut }: ShortcutRowProps) {
  return (
    <div className="flex items-center justify-between py-2 px-3 rounded-md hover:bg-editor-hover/50 transition-colors">
      <span className={`text-sm ${shortcut.enabled ? 'text-editor-text' : 'text-editor-muted line-through'}`}>
        {shortcut.description}
      </span>
      <div className="flex items-center gap-1">
        {shortcut.keys.map((key, index) => (
          <span key={index}>
            <kbd className="px-2 py-1 bg-editor-hover border border-editor-border rounded text-xs font-mono text-editor-text min-w-[24px] text-center inline-block">
              {key === 'Mod' ? formatShortcut(['Mod']) : key}
            </kbd>
            {index < shortcut.keys.length - 1 && (
              <span className="text-editor-muted mx-0.5">+</span>
            )}
          </span>
        ))}
      </div>
    </div>
  );
}
