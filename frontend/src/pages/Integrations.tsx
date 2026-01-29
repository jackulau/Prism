import { useState } from 'react';
import { IntegrationCard } from '../components/integrations/IntegrationCard';
import { IntegrationConfigModal } from '../components/integrations/IntegrationConfigModal';
import { trpc } from '../lib/trpc';

type IntegrationType = 'discord' | 'slack' | 'posthog' | 'mcp';

interface IntegrationConfig {
  type: IntegrationType;
  name: string;
  description: string;
  icon: string;
  authType: 'webhook' | 'apiKey' | 'oauth';
}

const INTEGRATIONS: IntegrationConfig[] = [
  {
    type: 'discord',
    name: 'Discord',
    description: 'Receive notifications via Discord webhook',
    icon: '🎮',
    authType: 'webhook',
  },
  {
    type: 'slack',
    name: 'Slack',
    description: 'Receive notifications in Slack channels',
    icon: '💬',
    authType: 'webhook',
  },
  {
    type: 'posthog',
    name: 'PostHog',
    description: 'Track analytics and events',
    icon: '📊',
    authType: 'apiKey',
  },
  {
    type: 'mcp',
    name: 'MCP Servers',
    description: 'Connect Model Context Protocol servers',
    icon: '🔌',
    authType: 'apiKey',
  },
];

export default function Integrations() {
  const [configuring, setConfiguring] = useState<IntegrationType | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: status, isLoading, refetch } = (trpc as any).integrations.getStatus.useQuery();

  const handleConfigure = (type: IntegrationType) => {
    setConfiguring(type);
  };

  const handleClose = () => {
    setConfiguring(null);
  };

  const handleSave = () => {
    refetch();
    setConfiguring(null);
  };

  const getIntegrationStatus = (type: IntegrationType) => {
    if (!status) return { enabled: false, connected: false };

    switch (type) {
      case 'discord':
        return status.discord || { enabled: false, connected: false };
      case 'slack':
        return status.slack || { enabled: false, connected: false };
      case 'posthog':
        return status.posthog || { enabled: false, connected: false };
      case 'mcp':
        return { enabled: true, connected: (status.mcpServers || []).length > 0 };
      default:
        return { enabled: false, connected: false };
    }
  };

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold text-editor-text">Integrations</h1>
          <p className="text-editor-muted">
            Connect external services to extend Prism's capabilities
          </p>
        </div>

        {/* Integration Grid */}
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[1, 2, 3, 4].map((i) => (
              <div
                key={i}
                className="bg-editor-surface border border-editor-border rounded-lg p-6 animate-pulse"
              >
                <div className="w-12 h-12 bg-editor-border rounded-lg mb-4" />
                <div className="h-4 bg-editor-border rounded w-1/2 mb-2" />
                <div className="h-3 bg-editor-border rounded w-3/4" />
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {INTEGRATIONS.map((integration) => (
              <IntegrationCard
                key={integration.type}
                name={integration.name}
                description={integration.description}
                icon={integration.icon}
                status={getIntegrationStatus(integration.type)}
                onConfigure={() => handleConfigure(integration.type)}
              />
            ))}
          </div>
        )}

        {/* Config Modal */}
        {configuring && (
          <IntegrationConfigModal
            type={configuring}
            config={INTEGRATIONS.find((i) => i.type === configuring)!}
            onClose={handleClose}
            onSave={handleSave}
          />
        )}
      </div>
    </div>
  );
}
