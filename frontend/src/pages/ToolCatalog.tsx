import { useState, useEffect } from 'react';
import { Plus, Wrench, Search, Trash2, Edit2, Loader2, Bot } from 'lucide-react';
import { apiService } from '../services/api';
import { useAuthStore } from '../store/authStore';
import { CustomToolForm } from '../components/tools/CustomToolForm';
import { ConfirmDialog } from '../components/ConfirmDialog';

interface Tool {
  id: string;
  display_name: string;
  slug_name: string;
  description?: string;
  is_model: boolean;
  is_builtin: boolean;
  provider_id?: string;
  parameters_schema?: string;
  created_at: string;
  updated_at: string;
}

export default function ToolCatalog() {
  const { accessToken } = useAuthStore();
  const [tools, setTools] = useState<Tool[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterType, setFilterType] = useState<'all' | 'tools' | 'models'>('all');

  const [showForm, setShowForm] = useState(false);
  const [editingTool, setEditingTool] = useState<Tool | null>(null);
  const [deletingTool, setDeletingTool] = useState<Tool | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);

  const fetchTools = async () => {
    if (!accessToken) return;

    setLoading(true);
    setError(null);

    try {
      const response = await apiService.listTools();
      if (response.error) {
        setError(response.error);
      } else if (response.data) {
        setTools(response.data.tools || []);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch tools');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTools();
  }, [accessToken]);

  const handleAddTool = () => {
    setEditingTool(null);
    setShowForm(true);
  };

  const handleEditTool = (tool: Tool) => {
    setEditingTool(tool);
    setShowForm(true);
  };

  const handleCloseForm = () => {
    setShowForm(false);
    setEditingTool(null);
  };

  const handleSaveForm = () => {
    setShowForm(false);
    setEditingTool(null);
    fetchTools();
  };

  const handleDeleteTool = async () => {
    if (!deletingTool) return;

    setDeleteLoading(true);
    try {
      const response = await apiService.deleteTool(deletingTool.id);
      if (response.error) {
        setError(response.error);
      } else {
        setDeletingTool(null);
        fetchTools();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete tool');
    } finally {
      setDeleteLoading(false);
    }
  };

  const filteredTools = tools.filter((tool) => {
    const matchesSearch =
      !searchQuery ||
      tool.display_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      tool.slug_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      tool.description?.toLowerCase().includes(searchQuery.toLowerCase());

    const matchesFilter =
      filterType === 'all' ||
      (filterType === 'models' && tool.is_model) ||
      (filterType === 'tools' && !tool.is_model);

    return matchesSearch && matchesFilter;
  });

  const customTools = filteredTools.filter((t) => !t.is_builtin);
  const builtinTools = filteredTools.filter((t) => t.is_builtin);

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-editor-text">Tool Catalog</h1>
            <p className="text-editor-muted">Manage custom tools and models for your agents</p>
          </div>
          <button
            onClick={handleAddTool}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            <Plus size={18} />
            Add Custom Tool
          </button>
        </div>

        {/* Filters */}
        <div className="flex items-center gap-4">
          <div className="relative flex-1 max-w-md">
            <Search
              size={18}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-editor-muted"
            />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search tools..."
              className="w-full pl-10 pr-4 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
            />
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setFilterType('all')}
              className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                filterType === 'all'
                  ? 'bg-editor-accent text-white'
                  : 'bg-editor-surface text-editor-muted hover:text-editor-text'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setFilterType('tools')}
              className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                filterType === 'tools'
                  ? 'bg-editor-accent text-white'
                  : 'bg-editor-surface text-editor-muted hover:text-editor-text'
              }`}
            >
              Tools
            </button>
            <button
              onClick={() => setFilterType('models')}
              className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                filterType === 'models'
                  ? 'bg-editor-accent text-white'
                  : 'bg-editor-surface text-editor-muted hover:text-editor-text'
              }`}
            >
              Models
            </button>
          </div>
        </div>

        {/* Error display */}
        {error && (
          <div className="p-3 bg-red-500/20 border border-red-500/50 rounded-lg">
            <p className="text-sm text-red-400">{error}</p>
          </div>
        )}

        {/* Loading state */}
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 size={32} className="animate-spin text-editor-accent" />
          </div>
        ) : (
          <div className="space-y-8">
            {/* Custom Tools Section */}
            <section>
              <h2 className="text-lg font-semibold text-editor-text mb-4">
                Custom Tools ({customTools.length})
              </h2>
              {customTools.length === 0 ? (
                <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
                  <Wrench size={48} className="mx-auto mb-4 text-editor-muted opacity-50" />
                  <p className="text-editor-muted">No custom tools yet</p>
                  <p className="text-sm text-editor-muted mt-1">
                    Click "Add Custom Tool" to create your first custom tool
                  </p>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {customTools.map((tool) => (
                    <ToolCard
                      key={tool.id}
                      tool={tool}
                      onEdit={() => handleEditTool(tool)}
                      onDelete={() => setDeletingTool(tool)}
                    />
                  ))}
                </div>
              )}
            </section>

            {/* Builtin Tools Section */}
            {builtinTools.length > 0 && (
              <section>
                <h2 className="text-lg font-semibold text-editor-text mb-4">
                  Built-in Tools ({builtinTools.length})
                </h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {builtinTools.map((tool) => (
                    <ToolCard key={tool.id} tool={tool} />
                  ))}
                </div>
              </section>
            )}
          </div>
        )}
      </div>

      {/* Custom Tool Form Modal */}
      {showForm && (
        <CustomToolForm tool={editingTool} onClose={handleCloseForm} onSave={handleSaveForm} />
      )}

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        isOpen={!!deletingTool}
        title="Delete Tool"
        message={`Are you sure you want to delete "${deletingTool?.display_name}"? This action cannot be undone.`}
        confirmText={deleteLoading ? 'Deleting...' : 'Delete'}
        onConfirm={handleDeleteTool}
        onCancel={() => setDeletingTool(null)}
        variant="danger"
      />
    </div>
  );
}

interface ToolCardProps {
  tool: Tool;
  onEdit?: () => void;
  onDelete?: () => void;
}

function ToolCard({ tool, onEdit, onDelete }: ToolCardProps) {
  const isEditable = !tool.is_builtin && onEdit && onDelete;

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4 hover:border-editor-accent/50 transition-colors">
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          <div
            className={`p-2 rounded-lg ${
              tool.is_model ? 'bg-purple-500/20' : 'bg-editor-accent/10'
            }`}
          >
            {tool.is_model ? (
              <Bot size={20} className="text-purple-400" />
            ) : (
              <Wrench size={20} className="text-editor-accent" />
            )}
          </div>
          <div>
            <h3 className="font-medium text-editor-text">{tool.display_name}</h3>
            <p className="text-xs text-editor-muted font-mono">{tool.slug_name}</p>
          </div>
        </div>
        {isEditable && (
          <div className="flex items-center gap-1">
            <button
              onClick={onEdit}
              className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
              title="Edit tool"
            >
              <Edit2 size={16} />
            </button>
            <button
              onClick={onDelete}
              className="p-1.5 rounded text-editor-muted hover:text-red-400 hover:bg-red-500/10 transition-colors"
              title="Delete tool"
            >
              <Trash2 size={16} />
            </button>
          </div>
        )}
      </div>

      {tool.description && (
        <p className="text-sm text-editor-muted line-clamp-2 mb-3">{tool.description}</p>
      )}

      <div className="flex flex-wrap gap-2">
        {tool.is_model && (
          <span className="text-xs px-2 py-0.5 rounded bg-purple-500/20 text-purple-400">
            Model
          </span>
        )}
        {tool.is_builtin && (
          <span className="text-xs px-2 py-0.5 rounded bg-gray-500/20 text-gray-400">
            Built-in
          </span>
        )}
        {tool.provider_id && (
          <span className="text-xs px-2 py-0.5 rounded bg-blue-500/20 text-blue-400">
            {tool.provider_id}
          </span>
        )}
        {tool.parameters_schema && (
          <span className="text-xs px-2 py-0.5 rounded bg-green-500/20 text-green-400">
            Has Schema
          </span>
        )}
      </div>
    </div>
  );
}
