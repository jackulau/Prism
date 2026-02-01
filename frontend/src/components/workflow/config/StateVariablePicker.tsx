import { useState, useRef, useEffect } from 'react';
import { Variable, ChevronDown, Copy, Check } from 'lucide-react';
import { useWorkflowStore } from '../../../store/workflowStore';

interface StateVariablePickerProps {
  nodeId: string;
  onInsert: (variable: string) => void;
  className?: string;
}

export function StateVariablePicker({ nodeId, onInsert, className = '' }: StateVariablePickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const { getAvailableStateVariables } = useWorkflowStore();

  // nodeId is available for future filtering by upstream nodes
  void nodeId;
  const variables = getAvailableStateVariables();

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

  const handleInsert = (variable: { name: string; type: string; sourceStepId?: string }) => {
    const template = `{{state.${variable.name}}}`;
    onInsert(template);
    setCopiedKey(variable.name);
    setTimeout(() => setCopiedKey(null), 1500);
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'string':
        return 'text-green-400';
      case 'number':
        return 'text-blue-400';
      case 'boolean':
        return 'text-yellow-400';
      case 'object':
        return 'text-purple-400';
      case 'array':
        return 'text-cyan-400';
      default:
        return 'text-editor-muted';
    }
  };

  if (variables.length === 0) {
    return (
      <div className={`flex items-center gap-2 px-3 py-2 text-sm text-editor-muted bg-editor-surface/50 rounded-lg border border-editor-border/50 ${className}`}>
        <Variable size={14} />
        <span>No state variables available</span>
      </div>
    );
  }

  return (
    <div ref={dropdownRef} className={`relative ${className}`}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 text-sm text-editor-text bg-editor-surface border border-editor-border rounded-lg hover:border-editor-accent transition-colors w-full"
      >
        <Variable size={14} className="text-editor-accent" />
        <span className="flex-1 text-left">Insert State Variable</span>
        <ChevronDown
          size={14}
          className={`text-editor-muted transition-transform ${isOpen ? 'rotate-180' : ''}`}
        />
      </button>

      {isOpen && (
        <div className="absolute top-full left-0 right-0 mt-1 bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 max-h-64 overflow-y-auto">
          <div className="p-2">
            <div className="text-xs text-editor-muted uppercase tracking-wide px-2 py-1 mb-1">
              Available Variables
            </div>
            {variables.map((variable) => (
              <button
                key={variable.name}
                onClick={() => handleInsert(variable)}
                className="w-full flex items-center gap-2 px-2 py-2 rounded-lg hover:bg-editor-surface text-left transition-colors group"
              >
                <code className="flex-1 text-sm font-mono text-editor-text">
                  <span className="text-editor-muted">{'{{state.'}</span>
                  <span className="text-editor-accent">{variable.name}</span>
                  <span className="text-editor-muted">{'}}'}</span>
                </code>
                <span className={`text-xs ${getTypeColor(variable.type)}`}>
                  {variable.type}
                </span>
                {copiedKey === variable.name ? (
                  <Check size={14} className="text-editor-success" />
                ) : (
                  <Copy size={14} className="text-editor-muted opacity-0 group-hover:opacity-100 transition-opacity" />
                )}
              </button>
            ))}
          </div>
          <div className="border-t border-editor-border px-3 py-2">
            <div className="text-xs text-editor-muted">
              From: {variables.map((v) => v.sourceStepId).filter((v, i, a) => v && a.indexOf(v) === i).join(', ') || 'workflow state'}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

interface StateVariableInputProps {
  value: string;
  onChange: (value: string) => void;
  nodeId: string;
  placeholder?: string;
  label?: string;
  rows?: number;
  className?: string;
}

export function StateVariableInput({
  value,
  onChange,
  nodeId,
  placeholder,
  label,
  rows,
  className = '',
}: StateVariableInputProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleInsert = (variable: string) => {
    const element = rows ? textareaRef.current : inputRef.current;
    if (!element) return;

    const start = element.selectionStart || 0;
    const end = element.selectionEnd || 0;
    const newValue = value.substring(0, start) + variable + value.substring(end);
    onChange(newValue);

    // Set cursor position after inserted variable
    setTimeout(() => {
      const newPos = start + variable.length;
      element.setSelectionRange(newPos, newPos);
      element.focus();
    }, 0);
  };

  return (
    <div className={`space-y-2 ${className}`}>
      {label && (
        <label className="block text-sm font-medium text-editor-text">
          {label}
        </label>
      )}
      <div className="space-y-2">
        {rows ? (
          <textarea
            ref={textareaRef}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            rows={rows}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm resize-none"
          />
        ) : (
          <input
            ref={inputRef}
            type="text"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
          />
        )}
        <StateVariablePicker nodeId={nodeId} onInsert={handleInsert} />
      </div>
      <p className="text-xs text-editor-muted">
        Use <code className="bg-editor-surface px-1 rounded">{'{{state.variableName}}'}</code> to reference previous step outputs
      </p>
    </div>
  );
}
