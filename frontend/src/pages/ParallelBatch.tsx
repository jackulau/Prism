import { useState } from 'react';
import {
  Play,
  Square,
  Pause,
  PlayCircle,
  Layers,
  Settings2,
  ListTodo,
  Activity,
  FileSpreadsheet,
} from 'lucide-react';
import {
  BatchTaskList,
  BatchConfigForm,
  BatchProgressTracker,
  BatchResults,
} from '../components/batch';
import { useBatchStore } from '../store/batchStore';

type TabView = 'tasks' | 'config' | 'progress' | 'results';

export default function ParallelBatch() {
  const {
    tasks,
    isRunning,
    execution,
    startBatch,
    stopBatch,
    pauseBatch,
    resumeBatch,
    results,
  } = useBatchStore();

  const [activeTab, setActiveTab] = useState<TabView>('tasks');

  const canStart = tasks.length > 0 && !isRunning;
  const canStop = isRunning;
  const canPause = isRunning && execution?.status === 'running';
  const canResume = !isRunning && execution?.status === 'paused';

  const tabs: { id: TabView; label: string; icon: React.ReactNode; badge?: number }[] = [
    { id: 'tasks', label: 'Tasks', icon: <ListTodo size={16} />, badge: tasks.length },
    { id: 'config', label: 'Config', icon: <Settings2 size={16} /> },
    { id: 'progress', label: 'Progress', icon: <Activity size={16} /> },
    { id: 'results', label: 'Results', icon: <FileSpreadsheet size={16} />, badge: results.length },
  ];

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-7xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-editor-text flex items-center gap-3">
              <Layers className="text-editor-accent" />
              Parallel Batch Execution
            </h1>
            <p className="text-editor-muted">
              Run multiple AI tasks in parallel with shared configuration
            </p>
          </div>

          {/* Action Buttons */}
          <div className="flex items-center gap-2">
            {canResume && (
              <button
                onClick={resumeBatch}
                className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
              >
                <PlayCircle size={18} />
                Resume
              </button>
            )}
            {canPause && (
              <button
                onClick={pauseBatch}
                className="flex items-center gap-2 px-4 py-2 bg-editor-warning text-white rounded-lg hover:bg-editor-warning/90 transition-colors"
              >
                <Pause size={18} />
                Pause
              </button>
            )}
            {canStop && (
              <button
                onClick={stopBatch}
                className="flex items-center gap-2 px-4 py-2 bg-editor-error text-white rounded-lg hover:bg-editor-error/90 transition-colors"
              >
                <Square size={18} />
                Stop
              </button>
            )}
            {canStart && (
              <button
                onClick={startBatch}
                className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
              >
                <Play size={18} />
                Run Batch
              </button>
            )}
          </div>
        </div>

        {/* Status Banner */}
        {execution && (
          <div
            className={`p-4 rounded-lg border ${
              execution.status === 'running'
                ? 'bg-editor-accent/10 border-editor-accent/30'
                : execution.status === 'completed'
                ? 'bg-editor-success/10 border-editor-success/30'
                : execution.status === 'paused'
                ? 'bg-editor-warning/10 border-editor-warning/30'
                : 'bg-editor-error/10 border-editor-error/30'
            }`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span
                  className={`text-sm font-medium capitalize ${
                    execution.status === 'running'
                      ? 'text-editor-accent'
                      : execution.status === 'completed'
                      ? 'text-editor-success'
                      : execution.status === 'paused'
                      ? 'text-editor-warning'
                      : 'text-editor-error'
                  }`}
                >
                  Batch {execution.status}
                </span>
                <span className="text-xs text-editor-muted">
                  {execution.completedTasks + execution.failedTasks} / {execution.totalTasks} tasks processed
                </span>
              </div>
              {execution.status === 'running' && (
                <div className="h-2 w-32 bg-editor-surface rounded-full overflow-hidden">
                  <div
                    className="h-full bg-editor-accent transition-all duration-300"
                    style={{
                      width: `${((execution.completedTasks + execution.failedTasks) / execution.totalTasks) * 100}%`,
                    }}
                  />
                </div>
              )}
            </div>
          </div>
        )}

        {/* Mobile Tabs */}
        <div className="lg:hidden">
          <div className="flex border-b border-editor-border">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex-1 flex items-center justify-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === tab.id
                    ? 'border-editor-accent text-editor-accent'
                    : 'border-transparent text-editor-muted hover:text-editor-text'
                }`}
              >
                {tab.icon}
                <span>{tab.label}</span>
                {tab.badge !== undefined && tab.badge > 0 && (
                  <span className="px-1.5 py-0.5 text-xs bg-editor-surface rounded-full">
                    {tab.badge}
                  </span>
                )}
              </button>
            ))}
          </div>

          <div className="py-4">
            {activeTab === 'tasks' && <BatchTaskList />}
            {activeTab === 'config' && <BatchConfigForm />}
            {activeTab === 'progress' && <BatchProgressTracker />}
            {activeTab === 'results' && <BatchResults />}
          </div>
        </div>

        {/* Desktop Three-Column Layout */}
        <div className="hidden lg:grid lg:grid-cols-3 gap-6">
          {/* Left Column: Task List */}
          <div className="bg-editor-bg rounded-lg border border-editor-border p-4">
            <BatchTaskList />
          </div>

          {/* Center Column: Config & Progress */}
          <div className="space-y-6">
            <div className="bg-editor-bg rounded-lg border border-editor-border p-4">
              <BatchConfigForm />
            </div>
            <div className="bg-editor-bg rounded-lg border border-editor-border p-4">
              <BatchProgressTracker />
            </div>
          </div>

          {/* Right Column: Results */}
          <div className="bg-editor-bg rounded-lg border border-editor-border p-4">
            <BatchResults />
          </div>
        </div>
      </div>
    </div>
  );
}
