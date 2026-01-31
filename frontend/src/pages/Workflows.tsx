import { Workflow } from 'lucide-react';

export default function Workflows() {
  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold text-editor-text">Workflows</h1>
          <p className="text-editor-muted">
            Create and manage automated workflows
          </p>
        </div>

        {/* Placeholder content */}
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Workflow size={48} className="text-editor-muted mb-4" />
          <h2 className="text-lg font-medium text-editor-text mb-2">No workflows created</h2>
          <p className="text-editor-muted max-w-md">
            Workflows enable you to orchestrate multi-step AI processes and automation.
          </p>
        </div>
      </div>
    </div>
  );
}
