import { useState, useEffect, useMemo } from 'react';
import { LayoutGrid, List, Package } from 'lucide-react';
import { ToolCard, ToolCardSkeleton } from '../components/tools/ToolCard';
import { ToolFilter } from '../components/tools/ToolFilter';
import { ToolDetailModal } from '../components/tools/ToolDetailModal';
import { useToolCatalogStore } from '../store/toolCatalogStore';
import { apiService } from '../services/api';
import type { Tool, ToolType } from '../types/tools';
import { getToolType } from '../types/tools';

export default function ToolCatalog() {
  const [tools, setTools] = useState<Tool[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedTool, setSelectedTool] = useState<Tool | null>(null);

  const {
    searchQuery,
    selectedProvider,
    selectedType,
    viewMode,
    setSearchQuery,
    setSelectedProvider,
    setSelectedType,
    setViewMode,
    resetFilters,
  } = useToolCatalogStore();

  // Load tools
  useEffect(() => {
    const loadTools = async () => {
      setIsLoading(true);
      setError(null);

      const result = await apiService.listTools();

      if (result.error) {
        setError(result.error);
      } else if (result.data) {
        setTools(result.data.tools);
      }

      setIsLoading(false);
    };

    loadTools();
  }, []);

  // Get unique providers for filter
  const providers = useMemo(() => {
    const providerSet = new Set<string>();
    tools.forEach((tool) => {
      if (tool.provider_id) {
        providerSet.add(tool.provider_id);
      }
    });
    return Array.from(providerSet).sort();
  }, [tools]);

  // Filter tools
  const filteredTools = useMemo(() => {
    return tools.filter((tool) => {
      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const matchesName = tool.display_name.toLowerCase().includes(query);
        const matchesSlug = tool.slug_name.toLowerCase().includes(query);
        const matchesDescription = tool.description?.toLowerCase().includes(query);
        if (!matchesName && !matchesSlug && !matchesDescription) {
          return false;
        }
      }

      // Provider filter
      if (selectedProvider && tool.provider_id !== selectedProvider) {
        return false;
      }

      // Type filter
      if (selectedType !== 'all') {
        const toolType = getToolType(tool);
        if (toolType !== selectedType) {
          return false;
        }
      }

      return true;
    });
  }, [tools, searchQuery, selectedProvider, selectedType]);

  const handleToolClick = (tool: Tool) => {
    setSelectedTool(tool);
  };

  const handleToolDeleted = () => {
    setTools((prev) => prev.filter((t) => t.id !== selectedTool?.id));
    setSelectedTool(null);
  };

  const handleTypeChange = (type: ToolType) => {
    setSelectedType(type);
  };

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-7xl mx-auto p-6">
        {/* Header */}
        <div className="flex items-start justify-between mb-6">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-editor-text">Tool Catalog</h1>
            <p className="text-editor-muted">
              Browse and manage available tools, models, and custom integrations
            </p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setViewMode('grid')}
              className={`p-2 rounded-lg transition-colors ${
                viewMode === 'grid'
                  ? 'bg-editor-accent text-white'
                  : 'text-editor-muted hover:text-editor-text hover:bg-editor-surface'
              }`}
              title="Grid view"
            >
              <LayoutGrid size={18} />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`p-2 rounded-lg transition-colors ${
                viewMode === 'list'
                  ? 'bg-editor-accent text-white'
                  : 'text-editor-muted hover:text-editor-text hover:bg-editor-surface'
              }`}
              title="List view"
            >
              <List size={18} />
            </button>
          </div>
        </div>

        <div className="flex gap-6">
          {/* Filter Sidebar */}
          <div className="w-64 flex-shrink-0">
            <ToolFilter
              searchQuery={searchQuery}
              selectedProvider={selectedProvider}
              selectedType={selectedType}
              providers={providers}
              onSearchChange={setSearchQuery}
              onProviderChange={setSelectedProvider}
              onTypeChange={handleTypeChange}
              onReset={resetFilters}
            />
          </div>

          {/* Main Content */}
          <div className="flex-1">
            {/* Error State */}
            {error && (
              <div className="p-4 mb-4 bg-editor-error/10 border border-editor-error/20 rounded-lg text-editor-error">
                {error}
              </div>
            )}

            {/* Loading State */}
            {isLoading ? (
              <div
                className={
                  viewMode === 'grid'
                    ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'
                    : 'space-y-3'
                }
              >
                {Array.from({ length: 6 }).map((_, i) => (
                  <ToolCardSkeleton key={i} viewMode={viewMode} />
                ))}
              </div>
            ) : filteredTools.length === 0 ? (
              /* Empty State */
              <div className="flex flex-col items-center justify-center py-16 text-center">
                <div className="w-16 h-16 flex items-center justify-center bg-editor-surface rounded-full mb-4">
                  <Package size={32} className="text-editor-muted" />
                </div>
                <h3 className="text-lg font-medium text-editor-text mb-1">No tools found</h3>
                <p className="text-editor-muted mb-4">
                  {searchQuery || selectedProvider || selectedType !== 'all'
                    ? 'Try adjusting your search or filters'
                    : 'No tools are available yet'}
                </p>
                {(searchQuery || selectedProvider || selectedType !== 'all') && (
                  <button
                    onClick={resetFilters}
                    className="px-4 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text hover:border-editor-accent/50 transition-colors"
                  >
                    Clear Filters
                  </button>
                )}
              </div>
            ) : (
              /* Tools Grid/List */
              <>
                <div className="text-sm text-editor-muted mb-4">
                  {filteredTools.length} tool{filteredTools.length !== 1 ? 's' : ''} found
                </div>
                <div
                  className={
                    viewMode === 'grid'
                      ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'
                      : 'space-y-3'
                  }
                >
                  {filteredTools.map((tool) => (
                    <ToolCard
                      key={tool.id}
                      tool={tool}
                      viewMode={viewMode}
                      onClick={() => handleToolClick(tool)}
                    />
                  ))}
                </div>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Detail Modal */}
      {selectedTool && (
        <ToolDetailModal
          tool={selectedTool}
          onClose={() => setSelectedTool(null)}
          onDelete={handleToolDeleted}
        />
      )}
    </div>
  );
}
