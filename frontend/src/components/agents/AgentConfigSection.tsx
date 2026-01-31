import { useState } from 'react';
import { ChevronDown, ChevronRight, Settings2, Wrench } from 'lucide-react';
import type { AgentConfigSectionProps, AvailableTool } from '../../types/agent';

/**
 * Collapsible configuration section for advanced agent settings.
 * Includes temperature slider, max tokens input, system prompt, and tool selection.
 */
export function AgentConfigSection({
  temperature,
  onTemperatureChange,
  maxTokens,
  onMaxTokensChange,
  systemPrompt,
  onSystemPromptChange,
  enabledTools,
  onToolsChange,
  availableTools = [],
  defaultCollapsed = true,
}: AgentConfigSectionProps) {
  const [isCollapsed, setIsCollapsed] = useState(defaultCollapsed);

  const handleToolToggle = (toolId: string) => {
    if (enabledTools.includes(toolId)) {
      onToolsChange(enabledTools.filter((id) => id !== toolId));
    } else {
      onToolsChange([...enabledTools, toolId]);
    }
  };

  // Group tools by category
  const toolsByCategory = availableTools.reduce<Record<string, AvailableTool[]>>(
    (acc, tool) => {
      const category = tool.category || 'General';
      if (!acc[category]) {
        acc[category] = [];
      }
      acc[category].push(tool);
      return acc;
    },
    {}
  );

  return (
    <div className="border border-editor-border rounded-lg overflow-hidden">
      {/* Header */}
      <button
        type="button"
        onClick={() => setIsCollapsed(!isCollapsed)}
        className="w-full flex items-center justify-between px-4 py-3 bg-editor-surface/50 hover:bg-editor-surface transition-colors"
      >
        <div className="flex items-center gap-2">
          <Settings2 size={16} className="text-editor-muted" />
          <span className="text-sm font-medium text-editor-text">
            Advanced Settings
          </span>
        </div>
        {isCollapsed ? (
          <ChevronRight size={16} className="text-editor-muted" />
        ) : (
          <ChevronDown size={16} className="text-editor-muted" />
        )}
      </button>

      {/* Content */}
      {!isCollapsed && (
        <div className="p-4 space-y-5 border-t border-editor-border">
          {/* Temperature and Max Tokens Row */}
          <div className="grid grid-cols-2 gap-4">
            {/* Temperature */}
            <div className="space-y-2">
              <label className="block text-xs font-medium text-editor-muted">
                Temperature ({temperature.toFixed(1)})
              </label>
              <input
                type="range"
                value={temperature}
                onChange={(e) => onTemperatureChange(Number(e.target.value))}
                min={0}
                max={2}
                step={0.1}
                className="w-full h-2 bg-editor-surface rounded-lg appearance-none cursor-pointer accent-editor-accent"
              />
              <div className="flex justify-between text-xs text-editor-muted">
                <span>Precise</span>
                <span>Creative</span>
              </div>
            </div>

            {/* Max Tokens */}
            <div className="space-y-2">
              <label className="block text-xs font-medium text-editor-muted">
                Max Tokens
              </label>
              <input
                type="number"
                value={maxTokens}
                onChange={(e) => onMaxTokensChange(Number(e.target.value))}
                min={1}
                max={200000}
                className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text text-sm focus:outline-none focus:border-editor-accent"
              />
            </div>
          </div>

          {/* System Prompt */}
          <div className="space-y-2">
            <label className="block text-xs font-medium text-editor-muted">
              System Prompt (Optional)
            </label>
            <textarea
              value={systemPrompt}
              onChange={(e) => onSystemPromptChange(e.target.value)}
              placeholder="Override the default system prompt..."
              rows={3}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text text-sm placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none"
            />
          </div>

          {/* Tool Selection */}
          {availableTools.length > 0 && (
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <Wrench size={14} className="text-editor-muted" />
                <label className="block text-xs font-medium text-editor-muted">
                  Tools ({enabledTools.length} selected)
                </label>
              </div>

              <div className="space-y-3">
                {Object.entries(toolsByCategory).map(([category, tools]) => (
                  <div key={category} className="space-y-2">
                    <span className="text-xs text-editor-muted uppercase tracking-wide">
                      {category}
                    </span>
                    <div className="grid grid-cols-2 gap-2">
                      {tools.map((tool) => (
                        <label
                          key={tool.id}
                          className="flex items-start gap-2 p-2 bg-editor-surface rounded-lg cursor-pointer hover:bg-editor-surface/80 transition-colors"
                        >
                          <input
                            type="checkbox"
                            checked={enabledTools.includes(tool.id)}
                            onChange={() => handleToolToggle(tool.id)}
                            className="mt-0.5 rounded border-editor-border text-editor-accent focus:ring-editor-accent focus:ring-offset-0"
                          />
                          <div className="flex-1 min-w-0">
                            <span className="text-sm text-editor-text block truncate">
                              {tool.name}
                            </span>
                            {tool.description && (
                              <span className="text-xs text-editor-muted block truncate">
                                {tool.description}
                              </span>
                            )}
                          </div>
                        </label>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
