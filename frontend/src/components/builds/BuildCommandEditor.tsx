import { useState, useRef } from 'react';
import {
  Plus,
  Trash2,
  GripVertical,
  Play,
  Pause,
  ChevronDown,
  ChevronRight,
  Folder,
} from 'lucide-react';
import { ConfirmDialog } from '../ConfirmDialog';
import type { BuildCommand } from '../../services/buildConfig';

interface BuildCommandEditorProps {
  commands: BuildCommand[];
  onAdd: (name: string, command: string, workingDirectory?: string) => Promise<void>;
  onUpdate: (cmdId: string, data: Partial<BuildCommand>) => Promise<void>;
  onDelete: (cmdId: string) => Promise<void>;
  onReorder: (order: string[]) => Promise<void>;
  disabled?: boolean;
}

// Common command templates
const COMMAND_TEMPLATES = [
  { label: 'npm install', command: 'npm install' },
  { label: 'npm run build', command: 'npm run build' },
  { label: 'npm test', command: 'npm test' },
  { label: 'go build', command: 'go build ./...' },
  { label: 'go test', command: 'go test ./...' },
  { label: 'cargo build', command: 'cargo build --release' },
  { label: 'make build', command: 'make build' },
  { label: 'docker build', command: 'docker build -t app .' },
];

export function BuildCommandEditor({
  commands,
  onAdd,
  onUpdate,
  onDelete,
  onReorder,
  disabled = false,
}: BuildCommandEditorProps) {
  const [isAddingNew, setIsAddingNew] = useState(false);
  const [newName, setNewName] = useState('');
  const [newCommand, setNewCommand] = useState('');
  const [newWorkDir, setNewWorkDir] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [showTemplates, setShowTemplates] = useState(false);

  // Drag and drop state
  const [draggedId, setDraggedId] = useState<string | null>(null);
  const dragOverRef = useRef<string | null>(null);

  const sortedCommands = [...commands].sort((a, b) => a.runOrder - b.runOrder);

  const handleAdd = async () => {
    if (!newName || !newCommand) return;

    setIsSubmitting(true);
    try {
      await onAdd(newName, newCommand, newWorkDir || undefined);
      setNewName('');
      setNewCommand('');
      setNewWorkDir('');
      setIsAddingNew(false);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleTemplateSelect = (template: { label: string; command: string }) => {
    setNewName(template.label);
    setNewCommand(template.command);
    setShowTemplates(false);
  };

  // Drag handlers
  const handleDragStart = (e: React.DragEvent, id: string) => {
    setDraggedId(id);
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', id);
  };

  const handleDragOver = (e: React.DragEvent, id: string) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    dragOverRef.current = id;
  };

  const handleDragEnd = async () => {
    if (draggedId && dragOverRef.current && draggedId !== dragOverRef.current) {
      const newOrder = [...sortedCommands.map((c) => c.id)];
      const fromIndex = newOrder.indexOf(draggedId);
      const toIndex = newOrder.indexOf(dragOverRef.current);

      if (fromIndex !== -1 && toIndex !== -1) {
        newOrder.splice(fromIndex, 1);
        newOrder.splice(toIndex, 0, draggedId);
        await onReorder(newOrder);
      }
    }
    setDraggedId(null);
    dragOverRef.current = null;
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-editor-text">Build Commands</h3>
        <span className="text-xs text-editor-muted">
          {commands.length} command{commands.length !== 1 ? 's' : ''}
        </span>
      </div>

      {/* Commands list */}
      <div className="space-y-2">
        {sortedCommands.length > 0 ? (
          sortedCommands.map((cmd) => (
            <CommandItem
              key={cmd.id}
              command={cmd}
              isExpanded={expandedId === cmd.id}
              isDragging={draggedId === cmd.id}
              disabled={disabled}
              onToggleExpand={() => setExpandedId(expandedId === cmd.id ? null : cmd.id)}
              onUpdate={(data) => onUpdate(cmd.id, data)}
              onDelete={() => setDeleteId(cmd.id)}
              onDragStart={(e) => handleDragStart(e, cmd.id)}
              onDragOver={(e) => handleDragOver(e, cmd.id)}
              onDragEnd={handleDragEnd}
            />
          ))
        ) : (
          <div className="text-center py-8 text-editor-muted text-sm border border-editor-border border-dashed rounded-lg">
            No build commands defined.
          </div>
        )}
      </div>

      {/* Add new command form */}
      {isAddingNew ? (
        <div className="p-4 bg-editor-surface rounded-lg space-y-3 border border-editor-accent/30">
          <div className="flex items-center justify-between">
            <h4 className="text-sm font-medium text-editor-text">Add Command</h4>
            <div className="relative">
              <button
                type="button"
                onClick={() => setShowTemplates(!showTemplates)}
                className="text-xs text-editor-accent hover:underline"
              >
                Use template
              </button>
              {showTemplates && (
                <div className="absolute right-0 top-full mt-1 w-48 bg-editor-bg border border-editor-border rounded-lg shadow-lg z-10 py-1">
                  {COMMAND_TEMPLATES.map((tmpl) => (
                    <button
                      key={tmpl.label}
                      type="button"
                      onClick={() => handleTemplateSelect(tmpl)}
                      className="w-full px-3 py-2 text-left text-sm hover:bg-editor-surface transition-colors"
                    >
                      {tmpl.label}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <input
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Command name (e.g., Build)"
              disabled={isSubmitting}
              className="px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
            />
            <div className="flex items-center gap-2">
              <Folder size={14} className="text-editor-muted flex-shrink-0" />
              <input
                type="text"
                value={newWorkDir}
                onChange={(e) => setNewWorkDir(e.target.value)}
                placeholder="Working directory (optional)"
                disabled={isSubmitting}
                className="flex-1 px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
              />
            </div>
          </div>

          <textarea
            value={newCommand}
            onChange={(e) => setNewCommand(e.target.value)}
            placeholder="Command (e.g., npm run build)"
            rows={2}
            disabled={isSubmitting}
            className="w-full px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm font-mono text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none"
          />

          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => {
                setIsAddingNew(false);
                setNewName('');
                setNewCommand('');
                setNewWorkDir('');
              }}
              disabled={isSubmitting}
              className="px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleAdd}
              disabled={disabled || isSubmitting || !newName || !newCommand}
              className="flex items-center gap-1 px-4 py-1.5 bg-editor-accent text-white text-sm rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <Plus size={14} />
              Add Command
            </button>
          </div>
        </div>
      ) : (
        <button
          onClick={() => setIsAddingNew(true)}
          disabled={disabled}
          className="w-full flex items-center justify-center gap-2 px-4 py-3 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface border border-editor-border border-dashed rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Plus size={16} />
          Add Command
        </button>
      )}

      {/* Delete confirmation */}
      <ConfirmDialog
        isOpen={!!deleteId}
        title="Delete Command"
        message="Are you sure you want to delete this command? This action cannot be undone."
        confirmText="Delete"
        variant="danger"
        onConfirm={async () => {
          if (deleteId) {
            await onDelete(deleteId);
            setDeleteId(null);
          }
        }}
        onCancel={() => setDeleteId(null)}
      />
    </div>
  );
}

interface CommandItemProps {
  command: BuildCommand;
  isExpanded: boolean;
  isDragging: boolean;
  disabled: boolean;
  onToggleExpand: () => void;
  onUpdate: (data: Partial<BuildCommand>) => Promise<void>;
  onDelete: () => void;
  onDragStart: (e: React.DragEvent) => void;
  onDragOver: (e: React.DragEvent) => void;
  onDragEnd: () => void;
}

function CommandItem({
  command,
  isExpanded,
  isDragging,
  disabled,
  onToggleExpand,
  onUpdate,
  onDelete,
  onDragStart,
  onDragOver,
  onDragEnd,
}: CommandItemProps) {
  const [editingName, setEditingName] = useState(false);
  const [editingCommand, setEditingCommand] = useState(false);
  const [nameValue, setNameValue] = useState(command.name);
  const [commandValue, setCommandValue] = useState(command.command);
  const [workDirValue, setWorkDirValue] = useState(command.workingDirectory || '');

  const handleSaveName = async () => {
    if (nameValue && nameValue !== command.name) {
      await onUpdate({ name: nameValue });
    }
    setEditingName(false);
  };

  const handleSaveCommand = async () => {
    if (commandValue && commandValue !== command.command) {
      await onUpdate({ command: commandValue });
    }
    setEditingCommand(false);
  };

  const handleSaveWorkDir = async () => {
    const newVal = workDirValue || undefined;
    if (newVal !== command.workingDirectory) {
      await onUpdate({ workingDirectory: newVal });
    }
  };

  return (
    <div
      draggable={!disabled}
      onDragStart={onDragStart}
      onDragOver={onDragOver}
      onDragEnd={onDragEnd}
      className={`border rounded-lg transition-all ${
        isDragging
          ? 'border-editor-accent bg-editor-accent/10 opacity-50'
          : 'border-editor-border bg-editor-surface/50 hover:bg-editor-surface'
      } ${!command.isEnabled ? 'opacity-60' : ''}`}
    >
      {/* Header row */}
      <div className="flex items-center gap-2 px-3 py-2">
        <div className="cursor-grab text-editor-muted hover:text-editor-text">
          <GripVertical size={16} />
        </div>

        <button
          onClick={onToggleExpand}
          className="p-0.5 text-editor-muted hover:text-editor-text"
        >
          {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
        </button>

        <button
          onClick={() => onUpdate({ isEnabled: !command.isEnabled })}
          disabled={disabled}
          className={`p-1 rounded transition-colors ${
            command.isEnabled
              ? 'text-green-500 hover:bg-green-500/10'
              : 'text-editor-muted hover:bg-editor-surface'
          }`}
          title={command.isEnabled ? 'Enabled' : 'Disabled'}
        >
          {command.isEnabled ? <Play size={14} /> : <Pause size={14} />}
        </button>

        {editingName ? (
          <input
            type="text"
            value={nameValue}
            onChange={(e) => setNameValue(e.target.value)}
            onBlur={handleSaveName}
            onKeyDown={(e) => e.key === 'Enter' && handleSaveName()}
            autoFocus
            className="flex-1 px-2 py-0.5 bg-editor-bg border border-editor-accent rounded text-sm text-editor-text focus:outline-none"
          />
        ) : (
          <button
            onClick={() => setEditingName(true)}
            disabled={disabled}
            className="flex-1 text-left text-sm font-medium text-editor-text hover:text-editor-accent truncate"
          >
            {command.name}
          </button>
        )}

        <code className="flex-1 text-xs font-mono text-editor-muted truncate">
          {command.command}
        </code>

        <button
          onClick={onDelete}
          disabled={disabled}
          className="p-1 text-editor-muted hover:text-red-400 transition-colors disabled:opacity-50"
          title="Delete command"
        >
          <Trash2 size={14} />
        </button>
      </div>

      {/* Expanded details */}
      {isExpanded && (
        <div className="px-3 pb-3 pt-1 border-t border-editor-border space-y-3">
          <div>
            <label className="block text-xs text-editor-muted mb-1">Command</label>
            {editingCommand ? (
              <textarea
                value={commandValue}
                onChange={(e) => setCommandValue(e.target.value)}
                onBlur={handleSaveCommand}
                rows={2}
                autoFocus
                className="w-full px-3 py-2 bg-editor-bg border border-editor-accent rounded-lg text-sm font-mono text-editor-text focus:outline-none resize-none"
              />
            ) : (
              <button
                onClick={() => setEditingCommand(true)}
                disabled={disabled}
                className="w-full text-left px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm font-mono text-editor-text hover:border-editor-accent transition-colors"
              >
                {command.command}
              </button>
            )}
          </div>

          <div>
            <label className="block text-xs text-editor-muted mb-1">Working Directory</label>
            <div className="flex items-center gap-2">
              <Folder size={14} className="text-editor-muted flex-shrink-0" />
              <input
                type="text"
                value={workDirValue}
                onChange={(e) => setWorkDirValue(e.target.value)}
                onBlur={handleSaveWorkDir}
                placeholder="(project root)"
                disabled={disabled}
                className="flex-1 px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
              />
            </div>
          </div>

          <div className="flex items-center gap-4 text-xs text-editor-muted">
            <span>Run order: {command.runOrder + 1}</span>
            {command.workingDirectory && (
              <span className="flex items-center gap-1">
                <Folder size={12} />
                {command.workingDirectory}
              </span>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
