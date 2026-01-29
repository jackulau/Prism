import { useState } from 'react';
import { Plus } from 'lucide-react';
import { WorkerList } from '../components/workers/WorkerList';
import { WorkerForm } from '../components/workers/WorkerForm';

export default function Workers() {
  const [showForm, setShowForm] = useState(false);
  const [editingWorker, setEditingWorker] = useState<string | null>(null);

  const handleCreateNew = () => {
    setEditingWorker(null);
    setShowForm(true);
  };

  const handleClose = () => {
    setShowForm(false);
    setEditingWorker(null);
  };

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-editor-text">Workers</h1>
            <p className="text-editor-muted">
              Manage your AI agent workers and their configurations
            </p>
          </div>
          <button
            onClick={handleCreateNew}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            <Plus size={18} />
            New Worker
          </button>
        </div>

        {/* Worker List */}
        <WorkerList />

        {/* Worker Form Modal */}
        {showForm && (
          <WorkerForm
            workerId={editingWorker}
            onClose={handleClose}
          />
        )}
      </div>
    </div>
  );
}
