import { useState, useEffect } from 'react';
import { useAuthStore } from '../store/authStore';
import { useAppStore } from '../store';
import { apiService } from '../services/api';
import { Github, Key, Bell, Server, CheckCircle, XCircle, ExternalLink, X, Palette, Check, Cpu } from 'lucide-react';
import { themes, type ThemeConfig } from '../config/themes';

interface GitHubStatus {
  connected: boolean;
  username?: string;
  connected_at?: string;
}

interface IntegrationStatus {
  enabled: boolean;
  connected: boolean;
}

interface IntegrationsStatus {
  discord: IntegrationStatus;
  slack: IntegrationStatus;
  posthog: IntegrationStatus;
}

interface OllamaStatus {
  connected: boolean;
  models: string[];
}

export function SettingsPage() {
  const { accessToken } = useAuthStore();
  const { theme, setTheme } = useAppStore();
  const [githubStatus, setGithubStatus] = useState<GitHubStatus | null>(null);
  const [integrations, setIntegrations] = useState<IntegrationsStatus | null>(null);
  const [ollamaStatus, setOllamaStatus] = useState<OllamaStatus | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchStatuses();

    // Check for OAuth callback status
    const params = new URLSearchParams(window.location.search);
    const githubParam = params.get('github');
    if (githubParam) {
      // Clear the URL params
      window.history.replaceState({}, '', '/settings');
      fetchStatuses();
    }
  }, [accessToken]);

  const fetchStatuses = async () => {
    if (!accessToken) return;

    try {
      const [ghResponse, intResponse, providersResponse] = await Promise.allSettled([
        fetch('/api/v1/github/status', {
          headers: { Authorization: `Bearer ${accessToken}` },
        }),
        fetch('/api/v1/integrations/status', {
          headers: { Authorization: `Bearer ${accessToken}` },
        }),
        fetch('/api/v1/providers', {
          headers: { Authorization: `Bearer ${accessToken}` },
        }),
      ]);

      if (ghResponse.status === 'fulfilled' && ghResponse.value.ok) {
        setGithubStatus(await ghResponse.value.json());
      }

      if (intResponse.status === 'fulfilled' && intResponse.value.ok) {
        setIntegrations(await intResponse.value.json());
      }

      // Parse Ollama status from providers
      if (providersResponse.status === 'fulfilled' && providersResponse.value.ok) {
        const data = await providersResponse.value.json();
        const ollama = data.providers?.find((p: { name: string }) => p.name === 'ollama');
        if (ollama) {
          setOllamaStatus({
            connected: ollama.models.length > 0,
            models: ollama.models.map((m: { name: string }) => m.name),
          });
        } else {
          setOllamaStatus({ connected: false, models: [] });
        }
      }
    } catch {
      // Failed to fetch statuses - user will see default/empty state
    } finally {
      setLoading(false);
    }
  };

  const handleGitHubConnect = async () => {
    if (!accessToken) return;

    try {
      const response = await fetch('/api/v1/oauth/github/authorize', {
        headers: { Authorization: `Bearer ${accessToken}` },
      });

      if (response.ok) {
        const { url } = await response.json();
        window.location.href = url;
      }
    } catch {
      // OAuth initiation failed - page won't redirect
    }
  };

  const handleGitHubDisconnect = async () => {
    if (!accessToken) return;

    try {
      await fetch('/api/v1/github/disconnect', {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      setGithubStatus({ connected: false });
    } catch {
      // Disconnect failed - status won't update
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-8">
      <h1 className="text-2xl font-bold">Settings</h1>

      {/* Theme Selection */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Palette className="w-5 h-5" />
          <h2 className="text-xl font-semibold">Theme</h2>
        </div>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
          <p className="text-editor-muted text-sm mb-4">
            Choose your preferred color theme for the interface.
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {(Object.values(themes) as ThemeConfig[]).map((themeConfig) => (
              <ThemeCard
                key={themeConfig.id}
                config={themeConfig}
                isSelected={theme === themeConfig.id}
                onSelect={() => setTheme(themeConfig.id)}
              />
            ))}
          </div>
        </div>
      </section>

      {/* LLM Providers */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Key className="w-5 h-5" />
          <h2 className="text-xl font-semibold">LLM Providers</h2>
        </div>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
          <p className="text-editor-muted text-sm mb-4">
            Configure your API keys for different LLM providers.
          </p>
          <ProviderKeyInput provider="OpenAI" placeholder="sk-..." />
          <ProviderKeyInput provider="Anthropic" placeholder="sk-ant-..." />
          <ProviderKeyInput provider="Google AI" placeholder="AIza..." />
        </div>
      </section>

      {/* Ollama Local Models */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Cpu className="w-5 h-5" />
          <h2 className="text-xl font-semibold">Ollama (Local)</h2>
        </div>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
          <div className="flex items-center gap-3 mb-4">
            {ollamaStatus?.connected ? (
              <CheckCircle className="w-5 h-5 text-green-500 flex-shrink-0" />
            ) : (
              <XCircle className="w-5 h-5 text-red-400 flex-shrink-0" />
            )}
            <div>
              <p className="font-medium">
                {ollamaStatus?.connected ? 'Ollama Connected' : 'Ollama Not Available'}
              </p>
              <p className="text-sm text-editor-muted">
                {ollamaStatus?.connected
                  ? `${ollamaStatus.models.length} model(s) available locally`
                  : 'Make sure Ollama is running on localhost:11434'}
              </p>
            </div>
          </div>
          {ollamaStatus?.connected && ollamaStatus.models.length > 0 && (
            <div className="space-y-2">
              <p className="text-sm text-editor-muted">Available models:</p>
              <div className="flex flex-wrap gap-2">
                {ollamaStatus.models.map((model) => (
                  <span
                    key={model}
                    className="px-2 py-1 text-xs rounded-full bg-editor-bg border border-editor-border text-editor-text"
                  >
                    {model}
                  </span>
                ))}
              </div>
            </div>
          )}
          {!ollamaStatus?.connected && (
            <div className="mt-2 p-3 bg-editor-bg rounded-lg">
              <p className="text-sm text-editor-muted">
                To use Ollama, make sure it&apos;s installed and running:
              </p>
              <code className="block mt-2 text-xs bg-editor-surface p-2 rounded border border-editor-border text-editor-accent">
                ollama serve
              </code>
              <p className="text-xs text-editor-muted mt-2">
                Then pull a model: <code className="text-editor-accent">ollama pull llama2</code>
              </p>
            </div>
          )}
        </div>
      </section>

      {/* GitHub Integration */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Github className="w-5 h-5" />
          <h2 className="text-xl font-semibold">GitHub Integration</h2>
        </div>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
          {githubStatus?.connected ? (
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <CheckCircle className="w-5 h-5 text-green-500" />
                <div>
                  <p className="font-medium">Connected as @{githubStatus.username}</p>
                  <p className="text-sm text-editor-muted">
                    Connected on {new Date(githubStatus.connected_at!).toLocaleDateString()}
                  </p>
                </div>
              </div>
              <button
                onClick={handleGitHubDisconnect}
                className="px-4 py-2 text-red-400 border border-red-400/20 rounded-lg hover:bg-red-400/10 transition-colors"
              >
                Disconnect
              </button>
            </div>
          ) : (
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Connect your GitHub account</p>
                <p className="text-sm text-editor-muted">
                  Enable repository access and GitHub features
                </p>
              </div>
              <button
                onClick={handleGitHubConnect}
                className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors flex items-center gap-2"
              >
                <Github className="w-4 h-4" />
                Connect GitHub
              </button>
            </div>
          )}
        </div>
      </section>

      {/* MCP Servers */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Server className="w-5 h-5" />
          <h2 className="text-xl font-semibold">MCP Servers</h2>
        </div>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
          <p className="text-editor-muted text-sm mb-4">
            Connect to Model Context Protocol servers to extend agent capabilities.
          </p>
          <MCPServerList />
        </div>
      </section>

      {/* Notification Integrations */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Bell className="w-5 h-5" />
          <h2 className="text-xl font-semibold">Notifications</h2>
        </div>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-4 space-y-4">
          <IntegrationCard
            name="Discord"
            description="Receive notifications via Discord webhook"
            status={integrations?.discord}
            onSave={fetchStatuses}
          />
          <IntegrationCard
            name="Slack"
            description="Receive notifications in Slack"
            status={integrations?.slack}
            onSave={fetchStatuses}
          />
          <IntegrationCard
            name="PostHog"
            description="Track analytics with PostHog"
            status={integrations?.posthog}
            onSave={fetchStatuses}
          />
        </div>
      </section>
    </div>
  );
}

function ProviderKeyInput({ provider, placeholder }: { provider: string; placeholder: string }) {
  const [value, setValue] = useState('');
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { accessToken } = useAuthStore();

  // Set token on apiService when available
  useEffect(() => {
    if (accessToken) {
      apiService.setToken(accessToken);
    }
  }, [accessToken]);

  const handleSave = async () => {
    if (!value || !accessToken) return;

    setError(null);
    try {
      const providerKey = provider.toLowerCase().replace(' ', '_');
      const response = await apiService.setProviderKey(providerKey, value);
      if (response.error) {
        setError(response.error);
      } else {
        setSaved(true);
        setValue('');
        setTimeout(() => setSaved(false), 2000);
      }
    } catch {
      setError('Failed to save API key');
    }
  };

  return (
    <div className="space-y-1 mb-3 last:mb-0">
      <div className="flex items-center gap-3">
        <span className="w-24 text-sm font-medium">{provider}</span>
        <input
          type="password"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder={placeholder}
          className={`flex-1 px-3 py-2 bg-editor-bg border rounded-lg text-sm ${
            error ? 'border-red-500' : 'border-editor-border'
          }`}
        />
        <button
          onClick={handleSave}
          disabled={!value}
          className="px-3 py-2 bg-primary text-white text-sm rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {saved ? 'Saved!' : 'Save'}
        </button>
      </div>
      {error && (
        <div className="ml-24 pl-3 text-xs text-red-400">{error}</div>
      )}
    </div>
  );
}

function MCPServerList() {
  const [servers, setServers] = useState<Array<{ id: string; name: string; url: string; enabled: boolean }>>([]);
  const [newUrl, setNewUrl] = useState('');
  const [newName, setNewName] = useState('');
  const { accessToken } = useAuthStore();

  useEffect(() => {
    fetchServers();
  }, [accessToken]);

  const fetchServers = async () => {
    if (!accessToken) return;
    try {
      const response = await fetch('/api/v1/mcp/servers', {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      if (response.ok) {
        const data = await response.json();
        setServers(data.servers || []);
      }
    } catch {
      // Failed to fetch MCP servers - list will remain empty
    }
  };

  const handleAdd = async () => {
    if (!newUrl || !newName || !accessToken) return;

    try {
      await fetch('/api/v1/mcp/servers', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify({ name: newName, url: newUrl }),
      });
      setNewUrl('');
      setNewName('');
      fetchServers();
    } catch {
      // Failed to add MCP server - server won't appear in list
    }
  };

  const handleRemove = async (id: string) => {
    if (!accessToken) return;

    try {
      await fetch(`/api/v1/mcp/servers/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      fetchServers();
    } catch {
      // Failed to remove MCP server - server will remain in list
    }
  };

  return (
    <div className="space-y-3">
      {servers.map((server) => (
        <div key={server.id} className="flex items-center justify-between p-3 bg-editor-bg rounded-lg">
          <div>
            <p className="font-medium">{server.name}</p>
            <p className="text-sm text-editor-muted">{server.url}</p>
          </div>
          <button
            onClick={() => handleRemove(server.id)}
            className="text-red-400 hover:text-red-300 text-sm"
          >
            Remove
          </button>
        </div>
      ))}

      <div className="flex gap-2">
        <input
          type="text"
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          placeholder="Server name"
          className="flex-1 px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm"
        />
        <input
          type="url"
          value={newUrl}
          onChange={(e) => setNewUrl(e.target.value)}
          placeholder="https://mcp-server.example.com"
          className="flex-1 px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm"
        />
        <button
          onClick={handleAdd}
          disabled={!newUrl || !newName}
          className="px-4 py-2 bg-primary text-white text-sm rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Add
        </button>
      </div>
    </div>
  );
}

function IntegrationCard({
  name,
  description,
  status,
  onSave,
}: {
  name: string;
  description: string;
  status?: IntegrationStatus;
  onSave?: () => void;
}) {
  const [isConfiguring, setIsConfiguring] = useState(false);
  const [webhookUrl, setWebhookUrl] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { accessToken } = useAuthStore();

  const handleSave = async () => {
    if (!webhookUrl || !accessToken) return;

    setSaving(true);
    setError(null);
    try {
      const response = await fetch(`/api/v1/integrations/${name.toLowerCase()}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify({ webhook_url: webhookUrl }),
      });

      if (response.ok) {
        setIsConfiguring(false);
        setWebhookUrl('');
        onSave?.();
      } else {
        const data = await response.json();
        setError(data.error || 'Failed to save configuration');
      }
    } catch {
      setError('Failed to save configuration');
    } finally {
      setSaving(false);
    }
  };

  const handleDisconnect = async () => {
    if (!accessToken) return;

    try {
      await fetch(`/api/v1/integrations/${name.toLowerCase()}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      onSave?.();
    } catch {
      // Disconnect failed
    }
  };

  return (
    <>
      <div className="flex items-center justify-between p-3 bg-editor-bg rounded-lg">
        <div className="flex items-center gap-3">
          {status?.connected ? (
            <CheckCircle className="w-5 h-5 text-green-500" />
          ) : (
            <XCircle className="w-5 h-5 text-editor-muted" />
          )}
          <div>
            <p className="font-medium">{name}</p>
            <p className="text-sm text-editor-muted">{description}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {status?.connected && (
            <button
              onClick={handleDisconnect}
              className="text-sm text-red-400 hover:underline"
            >
              Disconnect
            </button>
          )}
          <button
            onClick={() => setIsConfiguring(true)}
            className="text-sm text-primary hover:underline flex items-center gap-1"
          >
            Configure
            <ExternalLink className="w-3 h-3" />
          </button>
        </div>
      </div>

      {/* Configuration Modal */}
      {isConfiguring && (
        <>
          <div
            className="fixed inset-0 bg-black/50 z-40"
            onClick={() => setIsConfiguring(false)}
          />
          <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Configure {name}</h3>
              <button
                onClick={() => setIsConfiguring(false)}
                className="p-1 hover:bg-editor-surface rounded"
              >
                <X className="w-5 h-5 text-editor-muted" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-2">
                  Webhook URL
                </label>
                <input
                  type="url"
                  value={webhookUrl}
                  onChange={(e) => setWebhookUrl(e.target.value)}
                  placeholder={`https://${name.toLowerCase()}.com/api/webhooks/...`}
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm"
                />
              </div>

              {error && (
                <p className="text-sm text-red-400">{error}</p>
              )}

              <div className="flex justify-end gap-2">
                <button
                  onClick={() => setIsConfiguring(false)}
                  className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSave}
                  disabled={!webhookUrl || saving}
                  className="px-4 py-2 bg-primary text-white text-sm rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {saving ? 'Saving...' : 'Save'}
                </button>
              </div>
            </div>
          </div>
        </>
      )}
    </>
  );
}

function ThemeCard({
  config,
  isSelected,
  onSelect,
}: {
  config: ThemeConfig;
  isSelected: boolean;
  onSelect: () => void;
}) {
  return (
    <button
      onClick={onSelect}
      className={`relative p-3 rounded-lg border-2 transition-all text-left ${
        isSelected
          ? 'border-editor-accent bg-editor-accent/10'
          : 'border-editor-border hover:border-editor-muted bg-editor-bg'
      }`}
    >
      {isSelected && (
        <div className="absolute top-2 right-2">
          <Check className="w-4 h-4 text-editor-accent" />
        </div>
      )}

      {/* Color preview */}
      <div className="flex gap-1 mb-2">
        <div
          className="w-6 h-6 rounded-full border border-white/20"
          style={{ backgroundColor: config.colors.bg }}
          title="Background"
        />
        <div
          className="w-6 h-6 rounded-full border border-white/20"
          style={{ backgroundColor: config.colors.accent }}
          title="Accent"
        />
        <div
          className="w-6 h-6 rounded-full border border-white/20"
          style={{ backgroundColor: config.colors.success }}
          title="Success"
        />
        <div
          className="w-6 h-6 rounded-full border border-white/20"
          style={{ backgroundColor: config.colors.error }}
          title="Error"
        />
      </div>

      <div className="font-medium text-sm">{config.name}</div>
      <div className="text-xs text-editor-muted">{config.description}</div>
      <div className="mt-1">
        <span className={`text-xs px-1.5 py-0.5 rounded ${
          config.isDark
            ? 'bg-editor-border text-editor-text'
            : 'bg-editor-warning/20 text-editor-warning'
        }`}>
          {config.isDark ? 'Dark' : 'Light'}
        </span>
      </div>
    </button>
  );
}

export default SettingsPage;
