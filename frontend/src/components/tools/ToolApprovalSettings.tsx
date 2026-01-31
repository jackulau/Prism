import { useState, useEffect, useMemo } from 'react';
import { useAuthStore } from '../../store/authStore';
import { Shield, Search, Check, X, RefreshCw, Info } from 'lucide-react';

interface ApprovalConfig {
  enabled: boolean;
  auto_approve_read_only: boolean;
  trusted_tools: string[];
  max_iterations: number;
  read_only_tools: string[];
}

interface Tool {
  id: string;
  display_name: string;
  slug_name: string;
  description: string;
}

type ToolCategory = 'file_ops' | 'code_exec' | 'search' | 'web' | 'integrations' | 'other';

const TOOL_CATEGORIES: Record<ToolCategory, { label: string; tools: string[] }> = {
  file_ops: {
    label: 'File Operations',
    tools: ['read_file', 'write_file', 'list_files', 'delete_file', 'create_directory'],
  },
  code_exec: {
    label: 'Code Execution',
    tools: ['execute_code', 'run_command', 'bash', 'terminal'],
  },
  search: {
    label: 'Search & Query',
    tools: ['search_code', 'grep', 'glob', 'database_query', 'get_info'],
  },
  web: {
    label: 'Web & External',
    tools: ['web_fetch', 'web_search', 'http_request'],
  },
  integrations: {
    label: 'Integrations',
    tools: ['posthog_query_run', 'posthog_generate_hogql', 'posthog_docs_search', 'github_api'],
  },
  other: {
    label: 'Other',
    tools: [],
  },
};

function getCategoryForTool(toolName: string): ToolCategory {
  for (const [category, config] of Object.entries(TOOL_CATEGORIES)) {
    if (config.tools.includes(toolName)) {
      return category as ToolCategory;
    }
  }
  return 'other';
}

export function ToolApprovalSettings() {
  const { accessToken } = useAuthStore();
  const [config, setConfig] = useState<ApprovalConfig | null>(null);
  const [tools, setTools] = useState<Tool[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [showOnlyRequiringApproval, setShowOnlyRequiringApproval] = useState(false);

  useEffect(() => {
    fetchData();
  }, [accessToken]);

  const fetchData = async () => {
    if (!accessToken) return;

    setLoading(true);
    setError(null);

    try {
      const [configRes, toolsRes] = await Promise.all([
        fetch('/api/v1/tools/approval-config', {
          headers: { Authorization: `Bearer ${accessToken}` },
        }),
        fetch('/api/v1/tools', {
          headers: { Authorization: `Bearer ${accessToken}` },
        }),
      ]);

      if (configRes.ok) {
        const configData = await configRes.json();
        setConfig(configData);
      }

      if (toolsRes.ok) {
        const toolsData = await toolsRes.json();
        setTools(toolsData.tools || []);
      }
    } catch {
      setError('Failed to load approval settings');
    } finally {
      setLoading(false);
    }
  };

  const updateConfig = async (updates: Partial<ApprovalConfig>) => {
    if (!accessToken || !config) return;

    setSaving(true);
    setError(null);

    try {
      const response = await fetch('/api/v1/tools/approval-config', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify(updates),
      });

      if (response.ok) {
        const updatedConfig = await response.json();
        setConfig(updatedConfig);
      } else {
        const data = await response.json();
        setError(data.error || 'Failed to update settings');
      }
    } catch {
      setError('Failed to update settings');
    } finally {
      setSaving(false);
    }
  };

  const toggleTrustedTool = async (toolName: string) => {
    if (!config) return;

    const isTrusted = config.trusted_tools.includes(toolName);
    const newTrustedTools = isTrusted
      ? config.trusted_tools.filter((t) => t !== toolName)
      : [...config.trusted_tools, toolName];

    await updateConfig({ trusted_tools: newTrustedTools });
  };

  const isToolAutoApproved = (toolName: string): boolean => {
    if (!config) return false;
    if (!config.enabled) return false;
    if (config.trusted_tools.includes(toolName)) return true;
    if (config.auto_approve_read_only && config.read_only_tools.includes(toolName)) return true;
    return false;
  };

  const filteredTools = useMemo(() => {
    let filtered = tools;

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        (tool) =>
          tool.display_name.toLowerCase().includes(query) ||
          tool.slug_name.toLowerCase().includes(query) ||
          tool.description?.toLowerCase().includes(query)
      );
    }

    if (showOnlyRequiringApproval && config) {
      filtered = filtered.filter((tool) => !isToolAutoApproved(tool.slug_name));
    }

    return filtered;
  }, [tools, searchQuery, showOnlyRequiringApproval, config]);

  const toolsByCategory = useMemo(() => {
    const categories: Record<ToolCategory, Tool[]> = {
      file_ops: [],
      code_exec: [],
      search: [],
      web: [],
      integrations: [],
      other: [],
    };

    for (const tool of filteredTools) {
      const category = getCategoryForTool(tool.slug_name);
      categories[category].push(tool);
    }

    return categories;
  }, [filteredTools]);

  if (loading) {
    return (
      <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
        <div className="flex items-center justify-center py-4">
          <RefreshCw className="w-5 h-5 animate-spin text-editor-muted" />
        </div>
      </div>
    );
  }

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4 space-y-6">
      <p className="text-editor-muted text-sm">
        Configure which tools can run automatically without requiring your approval.
      </p>

      {error && (
        <div className="p-3 bg-red-500/20 border border-red-500/50 rounded-lg">
          <p className="text-sm text-red-400">{error}</p>
        </div>
      )}

      {/* Global Settings */}
      <div className="space-y-4">
        {/* Enable Auto-Approval */}
        <div className="flex items-center justify-between p-3 bg-editor-bg rounded-lg">
          <div className="flex items-center gap-3">
            <Shield className="w-5 h-5 text-editor-accent" />
            <div>
              <p className="font-medium">Enable Auto-Approval</p>
              <p className="text-sm text-editor-muted">
                Allow trusted tools to run without confirmation
              </p>
            </div>
          </div>
          <ToggleSwitch
            checked={config?.enabled ?? false}
            onChange={(enabled) => updateConfig({ enabled })}
            disabled={saving}
          />
        </div>

        {/* Auto-Approve Read-Only Tools */}
        <div className="flex items-center justify-between p-3 bg-editor-bg rounded-lg">
          <div className="flex items-center gap-3">
            <Info className="w-5 h-5 text-blue-400" />
            <div>
              <p className="font-medium">Auto-Approve Read-Only Tools</p>
              <p className="text-sm text-editor-muted">
                Automatically approve tools that only read data
              </p>
            </div>
          </div>
          <ToggleSwitch
            checked={config?.auto_approve_read_only ?? false}
            onChange={(auto_approve_read_only) => updateConfig({ auto_approve_read_only })}
            disabled={saving || !config?.enabled}
          />
        </div>

        {/* Max Iterations */}
        <div className="p-3 bg-editor-bg rounded-lg">
          <div className="flex items-center justify-between mb-2">
            <div>
              <p className="font-medium">Max Iterations</p>
              <p className="text-sm text-editor-muted">
                Pause for check-in after this many tool calls (0 = unlimited)
              </p>
            </div>
          </div>
          <input
            type="number"
            min="0"
            max="100"
            value={config?.max_iterations ?? 10}
            onChange={(e) => {
              const value = parseInt(e.target.value) || 0;
              updateConfig({ max_iterations: value });
            }}
            disabled={saving || !config?.enabled}
            className="w-24 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm disabled:opacity-50"
          />
        </div>
      </div>

      {/* Tool Trust List */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="font-medium">Tool Trust Settings</h3>
          <span className="text-xs text-editor-muted">
            {config?.trusted_tools.length ?? 0} trusted tools
          </span>
        </div>

        {/* Search and Filter */}
        <div className="flex gap-3">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-editor-muted" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search tools..."
              className="w-full pl-10 pr-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm"
            />
          </div>
          <label className="flex items-center gap-2 text-sm whitespace-nowrap">
            <input
              type="checkbox"
              checked={showOnlyRequiringApproval}
              onChange={(e) => setShowOnlyRequiringApproval(e.target.checked)}
              className="rounded border-editor-border"
            />
            Needs approval only
          </label>
        </div>

        {/* Tool List by Category */}
        <div className="space-y-4 max-h-96 overflow-y-auto">
          {(Object.entries(toolsByCategory) as [ToolCategory, Tool[]][]).map(
            ([category, categoryTools]) => {
              if (categoryTools.length === 0) return null;

              return (
                <div key={category} className="space-y-2">
                  <h4 className="text-xs font-medium text-editor-muted uppercase tracking-wider">
                    {TOOL_CATEGORIES[category].label}
                  </h4>
                  <div className="space-y-1">
                    {categoryTools.map((tool) => (
                      <ToolRow
                        key={tool.id}
                        tool={tool}
                        isTrusted={config?.trusted_tools.includes(tool.slug_name) ?? false}
                        isReadOnly={config?.read_only_tools.includes(tool.slug_name) ?? false}
                        isAutoApproved={isToolAutoApproved(tool.slug_name)}
                        autoApprovalEnabled={config?.enabled ?? false}
                        onToggleTrust={() => toggleTrustedTool(tool.slug_name)}
                        disabled={saving || !config?.enabled}
                      />
                    ))}
                  </div>
                </div>
              );
            }
          )}

          {filteredTools.length === 0 && (
            <p className="text-center text-editor-muted py-4">No tools found</p>
          )}
        </div>
      </div>
    </div>
  );
}

function ToggleSwitch({
  checked,
  onChange,
  disabled,
}: {
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
        disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'
      } ${checked ? 'bg-primary' : 'bg-editor-border'}`}
    >
      <span
        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
          checked ? 'translate-x-6' : 'translate-x-1'
        }`}
      />
    </button>
  );
}

function ToolRow({
  tool,
  isTrusted,
  isReadOnly,
  isAutoApproved,
  autoApprovalEnabled,
  onToggleTrust,
  disabled,
}: {
  tool: Tool;
  isTrusted: boolean;
  isReadOnly: boolean;
  isAutoApproved: boolean;
  autoApprovalEnabled: boolean;
  onToggleTrust: () => void;
  disabled?: boolean;
}) {
  return (
    <div
      className={`flex items-center justify-between p-2 rounded-lg transition-colors ${
        isAutoApproved ? 'bg-green-500/10' : 'bg-editor-bg'
      }`}
    >
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-mono text-sm truncate">{tool.slug_name}</span>
          {isReadOnly && (
            <span className="text-xs px-1.5 py-0.5 rounded bg-blue-500/20 text-blue-400">
              Read-only
            </span>
          )}
          {isTrusted && (
            <span className="text-xs px-1.5 py-0.5 rounded bg-green-500/20 text-green-400">
              Trusted
            </span>
          )}
        </div>
        {tool.description && (
          <p className="text-xs text-editor-muted truncate">{tool.description}</p>
        )}
      </div>
      <div className="flex items-center gap-2 ml-2">
        {isAutoApproved && autoApprovalEnabled ? (
          <Check className="w-4 h-4 text-green-400" />
        ) : (
          <X className="w-4 h-4 text-editor-muted" />
        )}
        <button
          onClick={onToggleTrust}
          disabled={disabled}
          className={`px-2 py-1 text-xs rounded transition-colors ${
            disabled
              ? 'opacity-50 cursor-not-allowed'
              : isTrusted
              ? 'bg-red-500/20 text-red-400 hover:bg-red-500/30'
              : 'bg-green-500/20 text-green-400 hover:bg-green-500/30'
          }`}
        >
          {isTrusted ? 'Remove' : 'Trust'}
        </button>
      </div>
    </div>
  );
}

export default ToolApprovalSettings;
