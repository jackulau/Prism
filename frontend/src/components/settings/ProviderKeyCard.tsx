import { CheckCircle, XCircle, ExternalLink } from 'lucide-react';
import { type ProviderKeyMetadata } from '../../store/apiKeysStore';

interface ProviderKeyCardProps {
  providerKey: ProviderKeyMetadata;
}

const PROVIDER_DISPLAY_NAMES: Record<string, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  google_ai: 'Google AI',
  google: 'Google AI',
  openrouter: 'OpenRouter',
  groq: 'Groq',
  deepseek: 'DeepSeek',
  together: 'Together AI',
  fireworks: 'Fireworks AI',
  mistral: 'Mistral AI',
  perplexity: 'Perplexity',
  huggingface: 'Hugging Face',
  ollama: 'Ollama',
};

const PROVIDER_SETTINGS_LINKS: Record<string, string> = {
  openai: 'https://platform.openai.com/api-keys',
  anthropic: 'https://console.anthropic.com/settings/keys',
  google_ai: 'https://aistudio.google.com/app/apikey',
  google: 'https://aistudio.google.com/app/apikey',
  openrouter: 'https://openrouter.ai/keys',
  groq: 'https://console.groq.com/keys',
  deepseek: 'https://platform.deepseek.com/api_keys',
};

function formatRelativeTime(dateString: string | null): string {
  if (!dateString) return 'Never';
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  if (diffDays < 30) return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
  return date.toLocaleDateString();
}

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
  return num.toLocaleString();
}

export function ProviderKeyCard({ providerKey }: ProviderKeyCardProps) {
  const displayName = PROVIDER_DISPLAY_NAMES[providerKey.provider.toLowerCase()] || providerKey.provider;
  const settingsLink = PROVIDER_SETTINGS_LINKS[providerKey.provider.toLowerCase()];

  return (
    <div className="p-4 bg-editor-bg rounded-lg border border-editor-border">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {providerKey.is_active ? (
            <CheckCircle className="w-5 h-5 text-green-500 flex-shrink-0" />
          ) : (
            <XCircle className="w-5 h-5 text-red-400 flex-shrink-0" />
          )}
          <div>
            <div className="flex items-center gap-2">
              <span className="font-medium">{displayName}</span>
              <span className={`px-2 py-0.5 text-xs rounded-full ${
                providerKey.is_active
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-red-500/20 text-red-400'
              }`}>
                {providerKey.is_active ? 'Active' : 'Inactive'}
              </span>
            </div>
            <div className="flex flex-wrap gap-x-4 gap-y-1 mt-1 text-xs text-editor-muted">
              <span>Added: {new Date(providerKey.created_at).toLocaleDateString()}</span>
              <span>Used: {formatNumber(providerKey.use_count)} times</span>
              <span>Last used: {formatRelativeTime(providerKey.last_used_at)}</span>
            </div>
          </div>
        </div>

        {settingsLink && (
          <a
            href={settingsLink}
            target="_blank"
            rel="noopener noreferrer"
            className="px-3 py-1.5 text-sm text-editor-accent border border-editor-accent/30 rounded-lg hover:bg-editor-accent/10 transition-colors flex items-center gap-1.5"
          >
            <ExternalLink className="w-3.5 h-3.5" />
            Manage
          </a>
        )}
      </div>
    </div>
  );
}
