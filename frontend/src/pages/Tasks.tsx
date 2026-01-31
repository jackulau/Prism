import { CheckSquare } from 'lucide-react';

export default function Tasks() {
  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold text-editor-text">Tasks</h1>
          <p className="text-editor-muted">
            View and manage your task queue
          </p>
        </div>

        {/* Placeholder content */}
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <CheckSquare size={48} className="text-editor-muted mb-4" />
          <h2 className="text-lg font-medium text-editor-text mb-2">No active tasks</h2>
          <p className="text-editor-muted max-w-md">
            Tasks represent work items that can be assigned to agents or workflows.
          </p>
        </div>
      </div>
    </div>
  );
}
