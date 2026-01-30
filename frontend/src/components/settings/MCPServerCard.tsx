import { useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  Settings,
  Trash2,
  Power,
  PowerOff,
  ExternalLink,
  Key,
  Clock,
  Wrench,
} from 'lucide-react';
import { useMCPServerStore, type MCPServer } from '../../store/mcpServerStore';
import { MCPServerStatus } from './MCPServerStatus';
import { MCPToolList } from './MCPToolList';
import { MCPServerStats } from './MCPServerStats';

interface MCPServerCardProps {
  server: MCPServer;
  onEdit?: (server: MCPServer) => void;
}

export function MCPServerCard({ server, onEdit }: MCPServerCardProps) {
  const [expanded, setExpanded] = useState(false);
  const [activeTab, setActiveTab] = useState<'tools' | 'stats'>('tools');
  const [deleting, setDeleting] = useState(false);
  const { removeServer, enableServer, disableServer, serverStatuses } = useMCPServerStore();

  const status = serverStatuses[server.id];

  const handleDelete = async () => {
    if (deleting) return;
    setDeleting(true);
    await removeServer(server.id);
    setDeleting(false);
  };

  const handleToggleEnabled = async () => {
    if (server.enabled) {
      await disableServer(server.id);
    } else {
      await enableServer(server.id);
    }
  };

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return 'Never';
    const date = new Date(dateStr);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="border border-editor-border rounded-lg overflow-hidden bg-editor-surface">
      {/* Card header */}
      <div className="p-4">
        <div className="flex items-start justify-between gap-4">
          {/* Left side - server info */}
          <div className="flex items-start gap-3 min-w-0 flex-1">
            <button
              onClick={() => setExpanded(!expanded)}
              className="p-1 text-editor-muted hover:text-editor-text transition-colors mt-0.5"
            >
              {expanded ? (
                <ChevronDown className="w-5 h-5" />
              ) : (
                <ChevronRight className="w-5 h-5" />
              )}
            </button>

            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 mb-1">
                <h3 className="font-semibold text-editor-text truncate">{server.name}</h3>
                <MCPServerStatus
                  serverId={server.id}
                  serverName={server.name}
                  enabled={server.enabled}
                  lastError={server.last_error}
                  compact
                />
              </div>

              <div className="flex items-center gap-2 text-sm text-editor-muted">
                <a
                  href={server.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-editor-accent hover:underline flex items-center gap-1 truncate"
                >
                  {server.url}
                  <ExternalLink className="w-3 h-3 flex-shrink-0" />
                </a>
              </div>

              {/* Metadata row */}
              <div className="flex flex-wrap items-center gap-3 mt-2 text-xs text-editor-muted">
                {server.has_api_key && (
                  <span className="flex items-center gap-1">
                    <Key className="w-3 h-3" />
                    API Key set
                  </span>
                )}
                {server.manifest && (
                  <span className="flex items-center gap-1">
                    <Wrench className="w-3 h-3" />
                    {server.manifest.tool_count} tools
                  </span>
                )}
                {server.last_sync && (
                  <span className="flex items-center gap-1">
                    <Clock className="w-3 h-3" />
                    Synced: {formatDate(server.last_sync)}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Right side - actions */}
          <div className="flex items-center gap-1 flex-shrink-0">
            <button
              onClick={handleToggleEnabled}
              className={`p-2 rounded-lg transition-colors ${
                server.enabled
                  ? 'text-green-500 hover:bg-green-500/10'
                  : 'text-editor-muted hover:bg-editor-bg'
              }`}
              title={server.enabled ? 'Disable server' : 'Enable server'}
            >
              {server.enabled ? (
                <Power className="w-4 h-4" />
              ) : (
                <PowerOff className="w-4 h-4" />
              )}
            </button>

            {onEdit && (
              <button
                onClick={() => onEdit(server)}
                className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded-lg transition-colors"
                title="Edit server"
              >
                <Settings className="w-4 h-4" />
              </button>
            )}

            <button
              onClick={handleDelete}
              disabled={deleting}
              className="p-2 text-red-400 hover:text-red-300 hover:bg-red-400/10 rounded-lg transition-colors disabled:opacity-50"
              title="Remove server"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Manifest info */}
        {server.manifest && (
          <div className="mt-3 p-2 bg-editor-bg rounded-lg text-sm">
            <div className="font-medium text-editor-text">
              {server.manifest.name} v{server.manifest.version}
            </div>
            {server.manifest.description && (
              <p className="text-editor-muted text-xs mt-1 line-clamp-2">
                {server.manifest.description}
              </p>
            )}
          </div>
        )}
      </div>

      {/* Expanded content */}
      {expanded && (
        <div className="border-t border-editor-border">
          {/* Status details */}
          <div className="p-4 border-b border-editor-border bg-editor-bg/50">
            <MCPServerStatus
              serverId={server.id}
              serverName={server.name}
              enabled={server.enabled}
              lastError={server.last_error}
            />
          </div>

          {/* Tab navigation */}
          <div className="flex items-center gap-1 px-4 pt-3 border-b border-editor-border">
            <button
              onClick={() => setActiveTab('tools')}
              className={`px-3 py-2 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'tools'
                  ? 'border-editor-accent text-editor-accent'
                  : 'border-transparent text-editor-muted hover:text-editor-text'
              }`}
            >
              Tools
            </button>
            <button
              onClick={() => setActiveTab('stats')}
              className={`px-3 py-2 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'stats'
                  ? 'border-editor-accent text-editor-accent'
                  : 'border-transparent text-editor-muted hover:text-editor-text'
              }`}
            >
              Statistics
            </button>
          </div>

          {/* Tab content */}
          <div className="p-4">
            {activeTab === 'tools' && (
              <MCPToolList serverId={server.id} serverName={server.name} />
            )}
            {activeTab === 'stats' && <MCPServerStats serverId={server.id} />}
          </div>
        </div>
      )}
    </div>
  );
}

export default MCPServerCard;
