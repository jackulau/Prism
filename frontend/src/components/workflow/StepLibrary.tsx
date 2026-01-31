import { useState, useMemo, useCallback } from 'react';
import { Search, ChevronDown, ChevronRight, PanelLeftClose, PanelLeft, FileCode } from 'lucide-react';
import { StepCard } from './StepCard';
import {
  STEP_TYPE_DEFINITIONS,
  STEP_CATEGORIES,
  type StepCategory,
  type StepTypeDefinition,
} from '../../config/workflowStepTypes';
import type { WorkflowTemplate } from '../../hooks/useWorkflowTemplates';

interface StepLibraryProps {
  isCollapsed?: boolean;
  onToggle?: () => void;
  onDragStart?: (stepType: StepTypeDefinition) => void;
  onDragEnd?: () => void;
  templates?: WorkflowTemplate[];
  onTemplateClick?: (template: WorkflowTemplate) => void;
  onTemplatePreview?: (template: WorkflowTemplate) => void;
}

const CATEGORY_ORDER: StepCategory[] = ['ai', 'logic', 'control', 'data'];

export function StepLibrary({
  isCollapsed = false,
  onToggle,
  onDragStart,
  onDragEnd,
  templates = [],
  onTemplateClick,
  onTemplatePreview,
}: StepLibraryProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedCategories, setExpandedCategories] = useState<Set<StepCategory>>(
    new Set(CATEGORY_ORDER)
  );
  const [showTemplates, setShowTemplates] = useState(true);

  const filteredSteps = useMemo(() => {
    if (!searchQuery.trim()) {
      return STEP_TYPE_DEFINITIONS;
    }
    const query = searchQuery.toLowerCase();
    return STEP_TYPE_DEFINITIONS.filter(
      (step) =>
        step.name.toLowerCase().includes(query) ||
        step.description.toLowerCase().includes(query) ||
        step.detailedDescription.toLowerCase().includes(query)
    );
  }, [searchQuery]);

  const filteredTemplates = useMemo(() => {
    if (!searchQuery.trim()) {
      return templates;
    }
    const query = searchQuery.toLowerCase();
    return templates.filter(
      (template) =>
        template.name.toLowerCase().includes(query) ||
        template.description.toLowerCase().includes(query)
    );
  }, [searchQuery, templates]);

  const toggleCategory = useCallback((category: StepCategory) => {
    setExpandedCategories((prev) => {
      const next = new Set(prev);
      if (next.has(category)) {
        next.delete(category);
      } else {
        next.add(category);
      }
      return next;
    });
  }, []);

  if (isCollapsed) {
    return (
      <div className="h-full w-12 bg-sidebar-bg border-r border-editor-border flex flex-col">
        <button
          onClick={onToggle}
          className="p-3 text-editor-muted hover:text-editor-text hover:bg-sidebar-hover transition-colors"
          title="Expand step library"
        >
          <PanelLeft size={18} />
        </button>
        <div className="flex-1 flex flex-col items-center py-2 space-y-2">
          {STEP_TYPE_DEFINITIONS.slice(0, 6).map((step) => {
            const Icon = step.icon;
            return (
              <div
                key={step.type}
                className={`p-2 rounded-lg ${step.bgColor} cursor-grab`}
                draggable
                onDragStart={(e) => {
                  e.dataTransfer.effectAllowed = 'copy';
                  e.dataTransfer.setData('application/workflow-step', JSON.stringify({
                    type: step.type,
                    name: step.name,
                    defaultConfig: step.defaultConfig,
                  }));
                  onDragStart?.(step);
                }}
                onDragEnd={onDragEnd}
                title={step.name}
              >
                <Icon size={16} className={step.color} />
              </div>
            );
          })}
        </div>
      </div>
    );
  }

  return (
    <div className="h-full w-72 bg-sidebar-bg border-r border-editor-border flex flex-col">
      <div className="flex items-center justify-between p-3 border-b border-editor-border">
        <h2 className="text-sm font-semibold text-editor-text">Step Library</h2>
        <button
          onClick={onToggle}
          className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-sidebar-hover rounded transition-colors"
          title="Collapse step library"
        >
          <PanelLeftClose size={16} />
        </button>
      </div>

      <div className="p-3 border-b border-editor-border">
        <div className="relative">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-editor-muted" />
          <input
            type="text"
            placeholder="Search steps..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
          />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {CATEGORY_ORDER.map((category) => {
          const categorySteps = filteredSteps.filter((s) => s.category === category);
          if (categorySteps.length === 0) return null;

          const isExpanded = expandedCategories.has(category);
          const categoryInfo = STEP_CATEGORIES[category];

          return (
            <div key={category} className="border-b border-editor-border/50">
              <button
                onClick={() => toggleCategory(category)}
                className="w-full flex items-center justify-between px-3 py-2.5 text-left hover:bg-sidebar-hover transition-colors"
              >
                <div className="flex items-center gap-2">
                  {isExpanded ? (
                    <ChevronDown size={14} className="text-editor-muted" />
                  ) : (
                    <ChevronRight size={14} className="text-editor-muted" />
                  )}
                  <span className="text-xs font-semibold text-editor-muted uppercase tracking-wider">
                    {categoryInfo.label}
                  </span>
                  <span className="text-xs text-editor-muted/60">
                    ({categorySteps.length})
                  </span>
                </div>
              </button>

              {isExpanded && (
                <div className="px-2 pb-2 space-y-1.5">
                  {categorySteps.map((step) => (
                    <StepCard
                      key={step.type}
                      stepType={step}
                      onDragStart={onDragStart}
                      onDragEnd={onDragEnd}
                    />
                  ))}
                </div>
              )}
            </div>
          );
        })}

        {templates.length > 0 && (
          <div className="border-b border-editor-border/50">
            <button
              onClick={() => setShowTemplates(!showTemplates)}
              className="w-full flex items-center justify-between px-3 py-2.5 text-left hover:bg-sidebar-hover transition-colors"
            >
              <div className="flex items-center gap-2">
                {showTemplates ? (
                  <ChevronDown size={14} className="text-editor-muted" />
                ) : (
                  <ChevronRight size={14} className="text-editor-muted" />
                )}
                <span className="text-xs font-semibold text-editor-muted uppercase tracking-wider">
                  Templates
                </span>
                <span className="text-xs text-editor-muted/60">
                  ({filteredTemplates.length})
                </span>
              </div>
            </button>

            {showTemplates && (
              <div className="px-2 pb-2 space-y-1.5">
                {filteredTemplates.map((template) => (
                  <TemplateCard
                    key={template.id}
                    template={template}
                    onClick={() => onTemplateClick?.(template)}
                    onPreview={() => onTemplatePreview?.(template)}
                  />
                ))}
              </div>
            )}
          </div>
        )}

        {filteredSteps.length === 0 && filteredTemplates.length === 0 && (
          <div className="p-6 text-center">
            <Search size={24} className="mx-auto mb-2 text-editor-muted/50" />
            <p className="text-sm text-editor-muted">
              No steps or templates match "{searchQuery}"
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

interface TemplateCardProps {
  template: WorkflowTemplate;
  onClick?: () => void;
  onPreview?: () => void;
}

function TemplateCard({ template, onClick, onPreview }: TemplateCardProps) {
  return (
    <div
      className="group flex items-center gap-3 p-3 bg-editor-surface/50 hover:bg-editor-surface border border-editor-border/50 hover:border-editor-border rounded-lg cursor-pointer transition-all"
      onClick={onClick}
    >
      <div className="flex-shrink-0 p-2 rounded-lg bg-editor-accent/20">
        <FileCode size={18} className="text-editor-accent" />
      </div>

      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium text-editor-text truncate">
          {template.name}
        </div>
        <div className="text-xs text-editor-muted truncate">
          {template.description}
        </div>
      </div>

      <button
        onClick={(e) => {
          e.stopPropagation();
          onPreview?.();
        }}
        className="flex-shrink-0 opacity-0 group-hover:opacity-100 p-1.5 text-editor-muted hover:text-editor-text hover:bg-sidebar-hover rounded transition-all"
        title="Preview template"
      >
        <Search size={14} />
      </button>
    </div>
  );
}
