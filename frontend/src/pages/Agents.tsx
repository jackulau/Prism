import { Bot } from 'lucide-react';

export default function Agents() {
  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold text-editor-text">Agents</h1>
          <p className="text-editor-muted">
            Configure and manage your AI agents
          </p>
        </div>

        {/* Placeholder content */}
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Bot size={48} className="text-editor-muted mb-4" />
          <h2 className="text-lg font-medium text-editor-text mb-2">No agents configured</h2>
          <p className="text-editor-muted max-w-md">
            Agents allow you to create specialized AI assistants with custom capabilities and tools.
          </p>
        </div>
      </div>
    </div>
  );
}
