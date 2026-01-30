import { useState, useCallback } from 'react';
import { Plus, Trash2, ListTodo } from 'lucide-react';
import { BatchTaskItem } from './BatchTaskItem';
import { AddTaskModal } from './AddTaskModal';
import type { BatchTask, BatchTaskFormData } from '../../types/batch';

interface BatchTaskListProps {
  tasks: BatchTask[];
  onTasksChange: (tasks: BatchTask[]) => void;
  disabled?: boolean;
}

export function BatchTaskList({ tasks, onTasksChange, disabled = false }: BatchTaskListProps) {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<BatchTask | null>(null);

  const handleAddTask = useCallback((data: BatchTaskFormData) => {
    const newTask: BatchTask = {
      id: crypto.randomUUID(),
      prompt: data.prompt,
      context: data.context,
      priority: data.priority,
      status: 'pending',
      createdAt: new Date(),
    };
    onTasksChange([...tasks, newTask]);
  }, [tasks, onTasksChange]);

  const handleEditTask = useCallback((data: BatchTaskFormData) => {
    if (!editingTask) return;
    const updatedTasks = tasks.map((task) =>
      task.id === editingTask.id
        ? { ...task, prompt: data.prompt, context: data.context, priority: data.priority }
        : task
    );
    onTasksChange(updatedTasks);
    setEditingTask(null);
  }, [tasks, editingTask, onTasksChange]);

  const handleDeleteTask = useCallback((taskId: string): void => {
    onTasksChange(tasks.filter((task) => task.id !== taskId));
  }, [tasks, onTasksChange]);

  const handleClearAll = useCallback(() => {
    onTasksChange([]);
  }, [onTasksChange]);

  const openEditModal = useCallback((task: BatchTask): void => {
    setEditingTask(task);
    setIsModalOpen(true);
  }, []);

  const closeModal = useCallback(() => {
    setIsModalOpen(false);
    setEditingTask(null);
  }, []);

  const handleSave = useCallback((data: BatchTaskFormData) => {
    if (editingTask) {
      handleEditTask(data);
    } else {
      handleAddTask(data);
    }
  }, [editingTask, handleEditTask, handleAddTask]);

  const pendingCount = tasks.filter((t) => t.status === 'pending').length;
  const completedCount = tasks.filter((t) => t.status === 'completed').length;

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border">
        <div className="flex items-center gap-3">
          <ListTodo size={20} className="text-editor-accent" />
          <h2 className="text-lg font-semibold text-editor-text">Tasks</h2>
          <span className="px-2 py-0.5 text-xs bg-editor-surface rounded-full text-editor-muted">
            {tasks.length} {tasks.length === 1 ? 'task' : 'tasks'}
          </span>
          {completedCount > 0 && (
            <span className="px-2 py-0.5 text-xs bg-green-500/20 rounded-full text-green-400">
              {completedCount} done
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {tasks.length > 0 && (
            <button
              onClick={handleClearAll}
              disabled={disabled}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-muted hover:text-red-400 hover:bg-red-400/10 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              aria-label="Clear all tasks"
            >
              <Trash2 size={14} />
              Clear All
            </button>
          )}
          <button
            onClick={() => setIsModalOpen(true)}
            disabled={disabled}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            aria-label="Add task"
          >
            <Plus size={16} />
            Add Task
          </button>
        </div>
      </div>

      {/* Task List */}
      <div className="flex-1 overflow-y-auto p-4">
        {tasks.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center py-12">
            <div className="w-16 h-16 mb-4 rounded-full bg-editor-surface flex items-center justify-center">
              <ListTodo size={32} className="text-editor-muted" />
            </div>
            <h3 className="text-lg font-medium text-editor-text mb-2">No tasks yet</h3>
            <p className="text-sm text-editor-muted mb-4 max-w-xs">
              Add tasks to run them in parallel. Each task will be executed concurrently.
            </p>
            <button
              onClick={() => setIsModalOpen(true)}
              disabled={disabled}
              className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Plus size={16} />
              Add Your First Task
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {tasks.map((task) => (
              <BatchTaskItem
                key={task.id}
                task={task}
                onEdit={openEditModal}
                onDelete={handleDeleteTask}
                disabled={disabled}
              />
            ))}
          </div>
        )}
      </div>

      {/* Status Bar */}
      {tasks.length > 0 && (
        <div className="px-4 py-2 border-t border-editor-border bg-editor-surface/50 text-xs text-editor-muted">
          {pendingCount} pending • {completedCount} completed • {tasks.length - pendingCount - completedCount} in progress
        </div>
      )}

      {/* Modal */}
      <AddTaskModal
        isOpen={isModalOpen}
        editingTask={editingTask}
        onClose={closeModal}
        onSave={handleSave}
      />
    </div>
  );
}
