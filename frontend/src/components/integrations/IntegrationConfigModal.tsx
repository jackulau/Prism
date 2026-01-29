import { useState } from 'react';
import { X, Loader2, CheckCircle, Trash2 } from 'lucide-react';
import { trpc } from '../../lib/trpc';
import { toast } from '../../store/toastStore';

type IntegrationType = 'discord' | 'slack' | 'posthog' | 'mcp';

interface IntegrationConfig {
  type: IntegrationType;
  name: string;
  description: string;
  icon: string;
  authType: 'webhook' | 'apiKey' | 'oauth';
}

interface IntegrationConfigModalProps {
  type: IntegrationType;
  config: IntegrationConfig;
  onClose: () => void;
  onSave: () => void;
}

export function IntegrationConfigModal({
  type,
  config,
  onClose,
  onSave,
}: IntegrationConfigModalProps) {
  const [webhookUrl, setWebhookUrl] = useState('');
  const [channelId, setChannelId] = useState('');
  const [enabled, setEnabled] = useState(true);

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const utils = (trpc as any).useUtils();

  // Discord mutation
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const discordMutation = (trpc as any).integrations.configureDiscord.useMutation({
    onSuccess: () => {
      toast.success('Discord configured successfully');
      utils.integrations.getStatus.invalidate();
      onSave();
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      toast.error(error.message);
    },
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const disconnectDiscord = (trpc as any).integrations.disconnectDiscord.useMutation({
    onSuccess: () => {
      toast.success('Discord disconnected');
      utils.integrations.getStatus.invalidate();
      onSave();
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      toast.error(error.message);
    },
  });

  // Slack mutation
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const slackMutation = (trpc as any).integrations.configureSlack.useMutation({
    onSuccess: () => {
      toast.success('Slack configured successfully');
      utils.integrations.getStatus.invalidate();
      onSave();
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      toast.error(error.message);
    },
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const disconnectSlack = (trpc as any).integrations.disconnectSlack.useMutation({
    onSuccess: () => {
      toast.success('Slack disconnected');
      utils.integrations.getStatus.invalidate();
      onSave();
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      toast.error(error.message);
    },
  });

  // PostHog mutation
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const posthogMutation = (trpc as any).integrations.configurePostHog.useMutation({
    onSuccess: () => {
      toast.success('PostHog configured successfully');
      utils.integrations.getStatus.invalidate();
      onSave();
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      toast.error(error.message);
    },
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const disconnectPostHog = (trpc as any).integrations.disconnectPostHog.useMutation({
    onSuccess: () => {
      toast.success('PostHog disconnected');
      utils.integrations.getStatus.invalidate();
      onSave();
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      toast.error(error.message);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    switch (type) {
      case 'discord':
        if (!webhookUrl) {
          toast.error('Please enter a webhook URL');
          return;
        }
        discordMutation.mutate({ webhookUrl, enabled });
        break;

      case 'slack':
        if (!webhookUrl) {
          toast.error('Please enter a webhook URL');
          return;
        }
        slackMutation.mutate({ webhookUrl, channelId: channelId || undefined, enabled });
        break;

      case 'posthog':
        posthogMutation.mutate({ enabled });
        break;

      default:
        break;
    }
  };

  const handleDisconnect = () => {
    switch (type) {
      case 'discord':
        disconnectDiscord.mutate();
        break;
      case 'slack':
        disconnectSlack.mutate();
        break;
      case 'posthog':
        disconnectPostHog.mutate();
        break;
      default:
        break;
    }
  };

  const isLoading =
    discordMutation.isPending ||
    slackMutation.isPending ||
    posthogMutation.isPending ||
    disconnectDiscord.isPending ||
    disconnectSlack.isPending ||
    disconnectPostHog.isPending;

  const renderForm = () => {
    switch (type) {
      case 'discord':
      case 'slack':
        return (
          <>
            <div className="space-y-2">
              <label className="block text-sm font-medium text-editor-text">
                Webhook URL
              </label>
              <input
                type="url"
                value={webhookUrl}
                onChange={(e) => setWebhookUrl(e.target.value)}
                placeholder={
                  type === 'discord'
                    ? 'https://discord.com/api/webhooks/...'
                    : 'https://hooks.slack.com/services/...'
                }
                className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
              />
            </div>

            {type === 'slack' && (
              <div className="space-y-2">
                <label className="block text-sm font-medium text-editor-text">
                  Channel ID (optional)
                </label>
                <input
                  type="text"
                  value={channelId}
                  onChange={(e) => setChannelId(e.target.value)}
                  placeholder="C01234567"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
                />
              </div>
            )}

            <div className="flex items-center gap-3">
              <input
                type="checkbox"
                id="enabled"
                checked={enabled}
                onChange={(e) => setEnabled(e.target.checked)}
                className="w-4 h-4 rounded border-editor-border bg-editor-surface text-editor-accent focus:ring-editor-accent"
              />
              <label htmlFor="enabled" className="text-sm text-editor-text">
                Enable notifications
              </label>
            </div>
          </>
        );

      case 'posthog':
        return (
          <div className="space-y-4">
            <p className="text-sm text-editor-muted">
              PostHog analytics will track usage events to help improve Prism.
            </p>
            <div className="flex items-center gap-3">
              <input
                type="checkbox"
                id="enabled"
                checked={enabled}
                onChange={(e) => setEnabled(e.target.checked)}
                className="w-4 h-4 rounded border-editor-border bg-editor-surface text-editor-accent focus:ring-editor-accent"
              />
              <label htmlFor="enabled" className="text-sm text-editor-text">
                Enable analytics
              </label>
            </div>
          </div>
        );

      case 'mcp':
        return (
          <div className="space-y-4">
            <p className="text-sm text-editor-muted">
              MCP servers can be configured in Settings {'>'} MCP Servers.
            </p>
            <button
              type="button"
              onClick={() => {
                onClose();
                window.location.href = '/settings';
              }}
              className="w-full px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
            >
              Go to Settings
            </button>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <>
      {/* Backdrop */}
      <div className="fixed inset-0 bg-black/50 z-40" onClick={onClose} />

      {/* Modal */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[480px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border">
          <div className="flex items-center gap-3">
            <span className="text-2xl">{config.icon}</span>
            <h2 className="text-lg font-semibold text-editor-text">
              Configure {config.name}
            </h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {renderForm()}
        </form>

        {/* Footer */}
        {type !== 'mcp' && (
          <div className="flex items-center justify-between px-6 py-4 border-t border-editor-border bg-editor-surface/50">
            <button
              type="button"
              onClick={handleDisconnect}
              disabled={isLoading}
              className="flex items-center gap-2 px-4 py-2 text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors disabled:opacity-50"
            >
              <Trash2 size={16} />
              Disconnect
            </button>
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSubmit}
                disabled={isLoading}
                className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 transition-colors"
              >
                {isLoading ? (
                  <>
                    <Loader2 size={16} className="animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    <CheckCircle size={16} />
                    Save
                  </>
                )}
              </button>
            </div>
          </div>
        )}
      </div>
    </>
  );
}
