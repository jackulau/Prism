import { useState, useEffect } from 'react';
import {
  Plus,
  Trash2,
  Settings,
  ChevronRight,
  Loader2,
  Power,
  PowerOff,
  Edit,
  X,
  Check,
  Wrench,
  Zap,
  Shield,
  MessageSquare,
} from 'lucide-react';
import { useApprovalWorkflows } from '../../store/approvalStore';
import { WorkflowStepEditor } from './WorkflowStepEditor';
import { ConfirmDialog } from '../ConfirmDialog';
import { toast } from '../../store/toastStore';
import type { ApprovalWorkflow, ApprovalStepConfig, ApprovalType } from '../../types/approval';

const TYPE_ICONS: Record<ApprovalType, typeof Wrench> = {
  tool_execution: Wrench,
  config_change: Settings,
  deployment: Zap,
  access_request: Shield,
  custom: MessageSquare,
};

const TYPE_LABELS: Record<ApprovalType, string> = {
  tool_execution: 'Tool Execution',
  config_change: 'Configuration Change',
  deployment: 'Deployment',
  access_request: 'Access Request',
  custom: 'Custom',
};

export function ApprovalWorkflows() {
  const { workflows, isLoading, error, refresh, create, update, delete: deleteWorkflow, toggleActive } = useApprovalWorkflows();
  const [selectedWorkflow, setSelectedWorkflow] = useState<ApprovalWorkflow | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const handleCreate = () => {
    setIsCreating(true);
    setSelectedWorkflow(null);
  };

  const handleEdit = (workflow: ApprovalWorkflow) => {
    setSelectedWorkflow(workflow);
    setIsCreating(false);
  };

  const handleDelete = async () => {
    if (!deleteConfirmId) return;
    const success = await deleteWorkflow(deleteConfirmId);
    if (success) {
      toast.success('Workflow deleted');
      if (selectedWorkflow?.id === deleteConfirmId) {
        setSelectedWorkflow(null);
      }
    } else {
      toast.error('Failed to delete workflow');
    }
    setDeleteConfirmId(null);
  };

  const handleToggleActive = async (id: string, isActive: boolean) => {
    const success = await toggleActive(id, isActive);
    if (success) {
      toast.success(isActive ? 'Workflow activated' : 'Workflow deactivated');
    } else {
      toast.error('Failed to update workflow');
    }
  };

  const handleClose = () => {
    setSelectedWorkflow(null);
    setIsCreating(false);
  };

  if (isLoading && workflows.length === 0) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="w-6 h-6 animate-spin text-editor-muted" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400">
        <p>{error}</p>
        <button onClick={refresh} className="mt-2 text-sm underline hover:no-underline">
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-editor-text">Approval Workflows</h2>
          <p className="text-sm text-editor-muted">
            Configure automated approval workflows for different actions
          </p>
        </div>
        <button
          onClick={handleCreate}
          className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/80 transition-colors"
        >
          <Plus size={18} />
          New Workflow
        </button>
      </div>

      {/* Workflow List */}
      <div className="grid gap-4">
        {workflows.length === 0 ? (
          <div className="text-center py-12 bg-editor-surface rounded-lg border border-editor-border">
            <Settings size={40} className="mx-auto text-editor-muted mb-3" />
            <p className="text-editor-muted mb-4">No workflows configured yet</p>
            <button
              onClick={handleCreate}
              className="px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/80 transition-colors"
            >
              Create your first workflow
            </button>
          </div>
        ) : (
          workflows.map((workflow) => (
            <WorkflowCard
              key={workflow.id}
              workflow={workflow}
              onEdit={() => handleEdit(workflow)}
              onDelete={() => setDeleteConfirmId(workflow.id)}
              onToggleActive={(isActive) => handleToggleActive(workflow.id, isActive)}
            />
          ))
        )}
      </div>

      {/* Editor Modal */}
      {(isCreating || selectedWorkflow) && (
        <WorkflowEditor
          workflow={selectedWorkflow}
          isOpen={true}
          onClose={handleClose}
          onCreate={create}
          onUpdate={update}
        />
      )}

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={deleteConfirmId !== null}
        title="Delete Workflow"
        message="Are you sure you want to delete this workflow? This action cannot be undone."
        confirmText="Delete"
        variant="danger"
        onConfirm={handleDelete}
        onCancel={() => setDeleteConfirmId(null)}
      />
    </div>
  );
}

interface WorkflowCardProps {
  workflow: ApprovalWorkflow;
  onEdit: () => void;
  onDelete: () => void;
  onToggleActive: (isActive: boolean) => void;
}

function WorkflowCard({ workflow, onEdit, onDelete, onToggleActive }: WorkflowCardProps) {
  const TypeIcon = TYPE_ICONS[workflow.triggerType];

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg hover:border-editor-accent/30 transition-colors">
      <div className="p-4">
        <div className="flex items-start justify-between">
          <div className="flex items-start gap-3">
            <div className="p-2 rounded-lg bg-editor-accent/10 text-editor-accent">
              <TypeIcon size={20} />
            </div>
            <div>
              <h3 className="font-medium text-editor-text">{workflow.name}</h3>
              <p className="text-sm text-editor-muted mt-0.5">
                {TYPE_LABELS[workflow.triggerType]}
              </p>
              {workflow.description && (
                <p className="text-sm text-editor-muted mt-1">{workflow.description}</p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {/* Active Toggle */}
            <button
              onClick={() => onToggleActive(!workflow.isActive)}
              className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-sm transition-colors ${
                workflow.isActive
                  ? 'bg-green-500/10 text-green-400 hover:bg-green-500/20'
                  : 'bg-editor-muted/10 text-editor-muted hover:bg-editor-muted/20'
              }`}
              title={workflow.isActive ? 'Deactivate' : 'Activate'}
            >
              {workflow.isActive ? <Power size={14} /> : <PowerOff size={14} />}
              {workflow.isActive ? 'Active' : 'Inactive'}
            </button>

            <button
              onClick={onEdit}
              className="p-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-bg transition-colors"
              title="Edit"
            >
              <Edit size={16} />
            </button>

            <button
              onClick={onDelete}
              className="p-2 rounded-lg text-editor-muted hover:text-editor-error hover:bg-editor-error/10 transition-colors"
              title="Delete"
            >
              <Trash2 size={16} />
            </button>
          </div>
        </div>

        {/* Steps Preview */}
        {workflow.steps.length > 0 && (
          <div className="mt-4 flex items-center gap-2 text-sm text-editor-muted">
            {workflow.steps.map((step, index) => (
              <div key={step.id} className="flex items-center">
                {index > 0 && <ChevronRight size={14} className="mx-1" />}
                <span className="px-2 py-0.5 bg-editor-bg rounded">{step.name}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

interface WorkflowEditorProps {
  workflow: ApprovalWorkflow | null;
  isOpen: boolean;
  onClose: () => void;
  onCreate: (workflow: Omit<ApprovalWorkflow, 'id' | 'createdAt' | 'updatedAt' | 'createdBy'>) => Promise<ApprovalWorkflow | null>;
  onUpdate: (id: string, updates: Partial<ApprovalWorkflow>) => Promise<boolean>;
}

function WorkflowEditor({ workflow, isOpen, onClose, onCreate, onUpdate }: WorkflowEditorProps) {
  const [name, setName] = useState(workflow?.name || '');
  const [description, setDescription] = useState(workflow?.description || '');
  const [triggerType, setTriggerType] = useState<ApprovalType>(workflow?.triggerType || 'tool_execution');
  const [steps, setSteps] = useState<ApprovalStepConfig[]>(workflow?.steps || []);
  const [isActive, setIsActive] = useState(workflow?.isActive ?? true);
  const [isSaving, setIsSaving] = useState(false);

  const isEditing = !!workflow;

  useEffect(() => {
    if (workflow) {
      setName(workflow.name);
      setDescription(workflow.description || '');
      setTriggerType(workflow.triggerType);
      setSteps(workflow.steps);
      setIsActive(workflow.isActive);
    } else {
      setName('');
      setDescription('');
      setTriggerType('tool_execution');
      setSteps([]);
      setIsActive(true);
    }
  }, [workflow]);

  const handleSave = async () => {
    if (!name.trim()) {
      toast.error('Please enter a workflow name');
      return;
    }

    if (steps.length === 0) {
      toast.error('Please add at least one step');
      return;
    }

    setIsSaving(true);

    try {
      if (isEditing && workflow) {
        const success = await onUpdate(workflow.id, {
          name,
          description,
          triggerType,
          steps,
          isActive,
        });
        if (success) {
          toast.success('Workflow updated');
          onClose();
        } else {
          toast.error('Failed to update workflow');
        }
      } else {
        const result = await onCreate({
          name,
          description,
          triggerType,
          steps,
          isActive,
        });
        if (result) {
          toast.success('Workflow created');
          onClose();
        } else {
          toast.error('Failed to create workflow');
        }
      }
    } finally {
      setIsSaving(false);
    }
  };

  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/60 z-50"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="fixed inset-4 md:inset-auto md:top-1/2 md:left-1/2 md:-translate-x-1/2 md:-translate-y-1/2 md:w-[800px] md:max-h-[85vh] bg-editor-bg border border-editor-border rounded-xl shadow-2xl z-50 flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-editor-border">
          <h2 className="text-xl font-semibold text-editor-text">
            {isEditing ? 'Edit Workflow' : 'Create Workflow'}
          </h2>
          <button
            onClick={onClose}
            className="p-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-surface transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Basic Info */}
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-editor-text mb-2">
                Workflow Name
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Production Deployment Approval"
                className="w-full px-4 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder:text-editor-muted focus:outline-none focus:border-editor-accent"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-editor-text mb-2">
                Description
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional description..."
                rows={2}
                className="w-full px-4 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder:text-editor-muted focus:outline-none focus:border-editor-accent resize-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-editor-text mb-2">
                Trigger Type
              </label>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
                {(Object.keys(TYPE_LABELS) as ApprovalType[]).map((type) => {
                  const Icon = TYPE_ICONS[type];
                  return (
                    <button
                      key={type}
                      onClick={() => setTriggerType(type)}
                      className={`flex items-center gap-2 p-3 rounded-lg border transition-colors ${
                        triggerType === type
                          ? 'bg-editor-accent/10 border-editor-accent text-editor-accent'
                          : 'bg-editor-surface border-editor-border text-editor-muted hover:text-editor-text'
                      }`}
                    >
                      <Icon size={16} />
                      <span className="text-sm">{TYPE_LABELS[type].split(' ')[0]}</span>
                    </button>
                  );
                })}
              </div>
            </div>
          </div>

          {/* Steps Editor */}
          <div>
            <h3 className="text-sm font-medium text-editor-text mb-4">Approval Steps</h3>
            <WorkflowStepEditor steps={steps} onChange={setSteps} />
          </div>

          {/* Active Toggle */}
          <div className="flex items-center justify-between p-4 bg-editor-surface rounded-lg">
            <div>
              <p className="font-medium text-editor-text">Workflow Status</p>
              <p className="text-sm text-editor-muted">
                {isActive ? 'This workflow will process requests' : 'This workflow is disabled'}
              </p>
            </div>
            <button
              onClick={() => setIsActive(!isActive)}
              className={`w-14 h-7 rounded-full transition-colors relative ${
                isActive ? 'bg-editor-success' : 'bg-editor-border'
              }`}
            >
              <div
                className={`absolute top-1 w-5 h-5 rounded-full bg-white transition-transform ${
                  isActive ? 'translate-x-8' : 'translate-x-1'
                }`}
              />
            </button>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 p-6 border-t border-editor-border bg-editor-surface/50">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={isSaving}
            className="flex items-center gap-2 px-4 py-2 text-sm text-white bg-editor-accent hover:bg-editor-accent/80 rounded-lg transition-colors disabled:opacity-50"
          >
            {isSaving ? (
              <Loader2 size={16} className="animate-spin" />
            ) : (
              <Check size={16} />
            )}
            {isEditing ? 'Save Changes' : 'Create Workflow'}
          </button>
        </div>
      </div>
    </>
  );
}
