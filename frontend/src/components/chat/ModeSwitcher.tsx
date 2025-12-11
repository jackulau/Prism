import { useState, useRef, useEffect } from 'react';
import { ChevronDown, Check, ListTodo, HelpCircle, Zap } from 'lucide-react';
import { useAppStore } from '../../store';
import type { ChatMode } from '../../types';

interface ModeOption {
  id: ChatMode;
  name: string;
  shortName: string;
  description: string;
  icon: React.ReactNode;
}

const MODES: ModeOption[] = [
  {
    id: 'plan',
    name: 'Plan Mode',
    shortName: 'Plan',
    description: 'AI plans thoroughly before executing tasks',
    icon: <ListTodo size={14} />,
  },
  {
    id: 'ask-before-edits',
    name: 'Ask Before Edits',
    shortName: 'Ask',
    description: 'AI proposes changes and waits for approval',
    icon: <HelpCircle size={14} />,
  },
  {
    id: 'edit-automatically',
    name: 'Edit Automatically',
    shortName: 'Auto',
    description: 'AI makes changes without asking',
    icon: <Zap size={14} />,
  },
];

export function ModeSwitcher() {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const { chatMode, setChatMode } = useAppStore();

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSelect = (mode: ChatMode) => {
    setChatMode(mode);
    setIsOpen(false);
  };

  const currentMode = MODES.find(m => m.id === chatMode) || MODES[1];

  return (
    <div ref={dropdownRef} className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-editor-surface border border-editor-border hover:border-editor-accent transition-colors text-xs"
      >
        <span className="text-editor-accent">{currentMode.icon}</span>
        <span className="text-editor-text">{currentMode.shortName}</span>
        <ChevronDown size={12} className={`text-editor-muted transition-transform ${isOpen ? 'rotate-180' : ''}`} />
      </button>

      {isOpen && (
        <div className="absolute bottom-full left-0 mb-1 bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 min-w-[220px]">
          <div className="px-2 py-1.5 text-xs font-semibold text-editor-muted uppercase border-b border-editor-border">
            Chat Mode
          </div>
          {MODES.map((mode) => (
            <button
              key={mode.id}
              onClick={() => handleSelect(mode.id)}
              className="w-full flex items-center gap-2 px-2 py-2 hover:bg-editor-surface text-left transition-colors"
            >
              <span className={`${chatMode === mode.id ? 'text-editor-accent' : 'text-editor-muted'}`}>
                {mode.icon}
              </span>
              <div className="flex-1 min-w-0">
                <div className="text-sm text-editor-text">{mode.name}</div>
                <div className="text-xs text-editor-muted">{mode.description}</div>
              </div>
              {chatMode === mode.id && (
                <Check size={14} className="text-editor-accent flex-shrink-0" />
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
