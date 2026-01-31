import { useState, useEffect } from 'react';
import { Wrench, ChevronDown, ChevronRight, Copy, Check, Code } from 'lucide-react';
import { useMCPServerStore, type MCPTool } from '../../store/mcpServerStore';

interface MCPToolListProps {
  serverId: string;
  serverName: string;
}

export function MCPToolList({ serverId, serverName: _serverName }: MCPToolListProps) {
  const { serverTools, toolsLoading, fetchServerTools } = useMCPServerStore();
  const tools = serverTools[serverId] || [];
  const loading = toolsLoading[serverId];

  useEffect(() => {
    fetchServerTools(serverId);
  }, [serverId, fetchServerTools]);

  if (loading) {
    return (
      <div className="flex items-center gap-2 p-4 text-editor-muted">
        <div className="animate-spin rounded-full h-4 w-4 border-2 border-editor-muted border-t-transparent" />
        <span className="text-sm">Loading tools...</span>
      </div>
    );
  }

  if (tools.length === 0) {
    return (
      <div className="p-4 text-center text-editor-muted">
        <Wrench className="w-8 h-8 mx-auto mb-2 opacity-50" />
        <p className="text-sm">No tools available from this server</p>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 text-sm text-editor-muted mb-3">
        <Wrench className="w-4 h-4" />
        <span>{tools.length} tool{tools.length !== 1 ? 's' : ''} available</span>
      </div>
      <div className="space-y-1">
        {tools.map((tool) => (
          <ToolItem key={tool.name} tool={tool} />
        ))}
      </div>
    </div>
  );
}

interface ToolItemProps {
  tool: MCPTool;
}

function ToolItem({ tool }: ToolItemProps) {
  const [expanded, setExpanded] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    const toolDefinition = {
      name: tool.name,
      description: tool.description,
      parameters: tool.parameters,
    };
    await navigator.clipboard.writeText(JSON.stringify(toolDefinition, null, 2));
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="border border-editor-border rounded-lg overflow-hidden">
      {/* Tool header */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-2 p-3 bg-editor-bg hover:bg-editor-surface transition-colors text-left"
      >
        <span className="text-editor-muted">
          {expanded ? (
            <ChevronDown className="w-4 h-4" />
          ) : (
            <ChevronRight className="w-4 h-4" />
          )}
        </span>
        <Code className="w-4 h-4 text-editor-accent" />
        <span className="font-mono text-sm text-editor-text">{tool.name}</span>
      </button>

      {/* Expanded content */}
      {expanded && (
        <div className="p-3 border-t border-editor-border bg-editor-surface/50 space-y-3">
          {/* Description */}
          <div>
            <p className="text-sm text-editor-text">{tool.description || 'No description provided'}</p>
          </div>

          {/* Parameters schema */}
          {tool.parameters && Object.keys(tool.parameters).length > 0 && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <h4 className="text-xs font-medium text-editor-muted uppercase tracking-wide">
                  Parameters
                </h4>
                <button
                  onClick={handleCopy}
                  className="flex items-center gap-1 px-2 py-1 text-xs text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded transition-colors"
                  title="Copy tool definition"
                >
                  {copied ? (
                    <>
                      <Check className="w-3 h-3 text-green-500" />
                      <span className="text-green-500">Copied!</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3 h-3" />
                      <span>Copy</span>
                    </>
                  )}
                </button>
              </div>
              <div className="bg-editor-bg rounded-lg p-3 overflow-auto max-h-60">
                <pre className="text-xs font-mono text-editor-text whitespace-pre-wrap">
                  {JSON.stringify(tool.parameters, null, 2)}
                </pre>
              </div>
            </div>
          )}

          {/* Server info */}
          <div className="text-xs text-editor-muted">
            From: {tool.server_name}
          </div>
        </div>
      )}
    </div>
  );
}

export default MCPToolList;
