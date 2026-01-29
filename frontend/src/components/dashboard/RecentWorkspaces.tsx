import { useNavigate } from 'react-router-dom';
import { Folder, Clock, ArrowRight, Plus } from 'lucide-react';
import { trpc } from '../../lib/trpc';


export function RecentWorkspaces() {
  const navigate = useNavigate();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: workspaces, isLoading, error } = (trpc as any).workspace.listRecent.useQuery({ limit: 6 });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleWorkspaceClick = (workspace: any) => {
    navigate(`/workspace/${workspace.id}`);
  };

  const handleNewWorkspace = () => {
    navigate('/workspace');
  };

  if (isLoading) {
    return (
      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-editor-text">Recent Workspaces</h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse"
            >
              <div className="h-4 bg-editor-border rounded w-3/4 mb-2" />
              <div className="h-3 bg-editor-border rounded w-1/2" />
            </div>
          ))}
        </div>
      </section>
    );
  }

  if (error) {
    return (
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-editor-text">Recent Workspaces</h2>
        <div className="bg-editor-surface border border-editor-error/20 rounded-lg p-6 text-center">
          <p className="text-editor-error">Failed to load workspaces</p>
          <p className="text-sm text-editor-muted mt-1">Please try again later</p>
        </div>
      </section>
    );
  }

  const workspaceList = workspaces || [];

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-editor-text">Recent Workspaces</h2>
        {workspaceList.length > 0 && (
          <button
            onClick={handleNewWorkspace}
            className="flex items-center gap-2 text-sm text-editor-accent hover:text-editor-accent/80 transition-colors"
          >
            <Plus size={16} />
            New Workspace
          </button>
        )}
      </div>

      {workspaceList.length === 0 ? (
        <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
          <Folder className="w-12 h-12 text-editor-muted mx-auto mb-4" />
          <h3 className="text-lg font-medium text-editor-text mb-2">No workspaces yet</h3>
          <p className="text-editor-muted mb-4">
            Create your first workspace to get started with Prism
          </p>
          <button
            onClick={handleNewWorkspace}
            className="inline-flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            <Plus size={18} />
            Create Workspace
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
          {workspaceList.map((workspace: any) => (
            <button
              key={workspace.id}
              onClick={() => handleWorkspaceClick(workspace)}
              className="group bg-editor-surface border border-editor-border rounded-lg p-4 text-left hover:border-editor-accent/50 hover:bg-editor-surface/80 transition-all"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-editor-accent/10 rounded-lg">
                    <Folder className="w-5 h-5 text-editor-accent" />
                  </div>
                  <div className="min-w-0">
                    <h3 className="font-medium text-editor-text truncate">
                      {workspace.name || workspace.path.split('/').pop()}
                    </h3>
                    <p className="text-sm text-editor-muted truncate">{workspace.path}</p>
                  </div>
                </div>
                <ArrowRight
                  size={18}
                  className="text-editor-muted opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
                />
              </div>
              <div className="flex items-center gap-1 mt-3 text-xs text-editor-muted">
                <Clock size={12} />
                <span>
                  {new Date(workspace.last_accessed).toLocaleDateString(undefined, {
                    month: 'short',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </span>
              </div>
            </button>
          ))}
        </div>
      )}
    </section>
  );
}
