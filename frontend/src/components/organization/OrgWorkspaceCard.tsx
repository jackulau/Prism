import { GitBranch, Github, Hash, Pencil, Trash2, ExternalLink } from 'lucide-react';
import type { OrgWorkspace } from '../../types/organization';

interface OrgWorkspaceCardProps {
  workspace: OrgWorkspace;
  onEdit: (workspace: OrgWorkspace) => void;
  onDelete: (workspace: OrgWorkspace) => void;
  onOpen?: (workspace: OrgWorkspace) => void;
}

export function OrgWorkspaceCard({
  workspace,
  onEdit,
  onDelete,
  onOpen,
}: OrgWorkspaceCardProps) {
  const hasGitHub = !!workspace.githubRepositoryName;
  const hasWorker = !!workspace.workerId;

  return (
    <div className="flex items-center justify-between p-4 hover:bg-editor-bg/50 transition-colors">
      <div className="flex items-center gap-4 min-w-0 flex-1">
        <div className="w-10 h-10 bg-editor-accent/10 rounded-lg flex items-center justify-center flex-shrink-0">
          <span className="text-editor-accent font-medium">
            {workspace.name.charAt(0).toUpperCase()}
          </span>
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="font-medium text-editor-text truncate">
              {workspace.name}
            </span>
            {hasWorker && (
              <span className="flex items-center gap-1 px-2 py-0.5 text-xs rounded-full bg-editor-success/10 text-editor-success">
                <span className="w-1.5 h-1.5 rounded-full bg-editor-success animate-pulse" />
                Active
              </span>
            )}
          </div>

          <div className="flex items-center gap-3 mt-1 text-sm text-editor-muted">
            {hasGitHub && (
              <a
                href={`https://github.com/${workspace.githubRepositoryName}`}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1 hover:text-editor-text transition-colors"
                onClick={(e) => e.stopPropagation()}
              >
                <Github size={14} />
                <span className="truncate max-w-[200px]">
                  {workspace.githubRepositoryName}
                </span>
              </a>
            )}

            {workspace.currentBranch && (
              <span className="flex items-center gap-1">
                <GitBranch size={14} />
                <span className="truncate max-w-[100px]">
                  {workspace.currentBranch}
                </span>
              </span>
            )}

            {workspace.slackChannelId && (
              <span className="flex items-center gap-1">
                <Hash size={14} />
                <span className="truncate max-w-[100px]">
                  {workspace.slackChannelId}
                </span>
              </span>
            )}
          </div>
        </div>
      </div>

      <div className="flex items-center gap-1 flex-shrink-0 ml-4">
        {onOpen && (
          <button
            onClick={() => onOpen(workspace)}
            className="p-2 text-editor-muted hover:text-editor-accent hover:bg-editor-accent/10 rounded-lg transition-colors"
            title="Open workspace"
          >
            <ExternalLink size={16} />
          </button>
        )}
        <button
          onClick={() => onEdit(workspace)}
          className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          title="Edit workspace"
        >
          <Pencil size={16} />
        </button>
        <button
          onClick={() => onDelete(workspace)}
          className="p-2 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
          title="Delete workspace"
        >
          <Trash2 size={16} />
        </button>
      </div>
    </div>
  );
}
