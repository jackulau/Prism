import { useCallback } from 'react';
import { GripVertical } from 'lucide-react';
import type { StepTypeDefinition } from '../../config/workflowStepTypes';

interface StepCardProps {
  stepType: StepTypeDefinition;
  onDragStart?: (stepType: StepTypeDefinition) => void;
  onDragEnd?: () => void;
}

export function StepCard({ stepType, onDragStart, onDragEnd }: StepCardProps) {
  const Icon = stepType.icon;

  const handleDragStart = useCallback(
    (e: React.DragEvent<HTMLDivElement>) => {
      e.dataTransfer.effectAllowed = 'copy';
      e.dataTransfer.setData('application/workflow-step', JSON.stringify({
        type: stepType.type,
        name: stepType.name,
        defaultConfig: stepType.defaultConfig,
      }));

      const dragPreview = document.createElement('div');
      dragPreview.className = 'bg-editor-surface border border-editor-border rounded-lg p-2 shadow-lg flex items-center gap-2';
      dragPreview.innerHTML = `
        <div class="${stepType.bgColor} ${stepType.color} p-1.5 rounded">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
          </svg>
        </div>
        <span class="text-editor-text text-sm font-medium">${stepType.name}</span>
      `;
      dragPreview.style.position = 'absolute';
      dragPreview.style.top = '-1000px';
      document.body.appendChild(dragPreview);
      e.dataTransfer.setDragImage(dragPreview, 0, 0);
      setTimeout(() => document.body.removeChild(dragPreview), 0);

      onDragStart?.(stepType);
    },
    [stepType, onDragStart]
  );

  const handleDragEnd = useCallback(() => {
    onDragEnd?.();
  }, [onDragEnd]);

  return (
    <div
      draggable
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      className="group flex items-center gap-3 p-3 bg-editor-surface/50 hover:bg-editor-surface border border-editor-border/50 hover:border-editor-border rounded-lg cursor-grab active:cursor-grabbing transition-all"
      title={stepType.detailedDescription}
    >
      <div className="flex-shrink-0 text-editor-muted/50 group-hover:text-editor-muted transition-colors">
        <GripVertical size={14} />
      </div>

      <div className={`flex-shrink-0 p-2 rounded-lg ${stepType.bgColor}`}>
        <Icon size={18} className={stepType.color} />
      </div>

      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium text-editor-text truncate">
          {stepType.name}
        </div>
        <div className="text-xs text-editor-muted truncate">
          {stepType.description}
        </div>
      </div>
    </div>
  );
}
