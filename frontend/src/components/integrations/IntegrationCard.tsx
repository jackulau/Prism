import { CheckCircle, XCircle, Settings } from 'lucide-react';

interface IntegrationStatus {
  enabled: boolean;
  connected: boolean;
}

interface IntegrationCardProps {
  name: string;
  description: string;
  icon: string;
  status: IntegrationStatus;
  onConfigure: () => void;
}

export function IntegrationCard({
  name,
  description,
  icon,
  status,
  onConfigure,
}: IntegrationCardProps) {
  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-6 hover:border-editor-accent/30 transition-colors">
      <div className="flex items-start justify-between mb-4">
        <div className="text-4xl">{icon}</div>
        {status.connected ? (
          <div className="flex items-center gap-1 px-2 py-1 bg-editor-success/10 text-editor-success rounded-full text-xs">
            <CheckCircle size={12} />
            Connected
          </div>
        ) : (
          <div className="flex items-center gap-1 px-2 py-1 bg-editor-muted/10 text-editor-muted rounded-full text-xs">
            <XCircle size={12} />
            Not Connected
          </div>
        )}
      </div>

      <h3 className="font-medium text-editor-text mb-1">{name}</h3>
      <p className="text-sm text-editor-muted mb-4">{description}</p>

      <button
        onClick={onConfigure}
        className="flex items-center gap-2 w-full justify-center px-4 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text hover:border-editor-accent/50 hover:bg-editor-surface/80 transition-colors"
      >
        <Settings size={16} />
        {status.connected ? 'Configure' : 'Connect'}
      </button>
    </div>
  );
}
