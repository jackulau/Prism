import { useState } from 'react';
import {
  Plus,
  Trash2,
  GripVertical,
  ChevronDown,
  ChevronUp,
  Users,
  User,
  Clock,
  AlertTriangle,
  ArrowRight,
  Layers,
} from 'lucide-react';
import type { ApprovalStepConfig, EscalationPolicy } from '../../types/approval';

interface WorkflowStepEditorProps {
  steps: ApprovalStepConfig[];
  onChange: (steps: ApprovalStepConfig[]) => void;
  availableRoles?: string[];
}

const DEFAULT_STEP: Omit<ApprovalStepConfig, 'id' | 'order'> = {
  name: 'New Step',
  approverType: 'role',
  approverRoles: [],
  approverUserIds: [],
  requiredApprovals: 1,
  timeoutMinutes: 60,
  parallelWithPrevious: false,
};

const DEFAULT_ESCALATION: EscalationPolicy = {
  enabled: false,
  escalateAfterMinutes: 30,
  escalateTo: 'manager',
};

export function WorkflowStepEditor({ steps, onChange, availableRoles = [] }: WorkflowStepEditorProps) {
  const [expandedSteps, setExpandedSteps] = useState<Set<string>>(new Set());
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);

  const toggleExpanded = (stepId: string) => {
    setExpandedSteps(prev => {
      const next = new Set(prev);
      if (next.has(stepId)) {
        next.delete(stepId);
      } else {
        next.add(stepId);
      }
      return next;
    });
  };

  const addStep = () => {
    const newStep: ApprovalStepConfig = {
      ...DEFAULT_STEP,
      id: `step-${Date.now()}`,
      order: steps.length + 1,
    };
    onChange([...steps, newStep]);
    setExpandedSteps(prev => new Set([...prev, newStep.id]));
  };

  const removeStep = (index: number) => {
    const newSteps = steps.filter((_, i) => i !== index).map((step, i) => ({
      ...step,
      order: i + 1,
    }));
    onChange(newSteps);
  };

  const updateStep = (index: number, updates: Partial<ApprovalStepConfig>) => {
    const newSteps = steps.map((step, i) =>
      i === index ? { ...step, ...updates } : step
    );
    onChange(newSteps);
  };

  const handleDragStart = (index: number) => {
    setDraggedIndex(index);
  };

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    if (draggedIndex === null || draggedIndex === index) return;

    const newSteps = [...steps];
    const [removed] = newSteps.splice(draggedIndex, 1);
    newSteps.splice(index, 0, removed);

    // Update order numbers
    const reorderedSteps = newSteps.map((step, i) => ({
      ...step,
      order: i + 1,
    }));

    onChange(reorderedSteps);
    setDraggedIndex(index);
  };

  const handleDragEnd = () => {
    setDraggedIndex(null);
  };

  return (
    <div className="space-y-4">
      {/* Steps List */}
      <div className="space-y-3">
        {steps.map((step, index) => (
          <StepCard
            key={step.id}
            step={step}
            index={index}
            isExpanded={expandedSteps.has(step.id)}
            isFirst={index === 0}
            isLast={index === steps.length - 1}
            availableRoles={availableRoles}
            onToggleExpanded={() => toggleExpanded(step.id)}
            onUpdate={(updates) => updateStep(index, updates)}
            onRemove={() => removeStep(index)}
            onDragStart={() => handleDragStart(index)}
            onDragOver={(e) => handleDragOver(e, index)}
            onDragEnd={handleDragEnd}
            isDragging={draggedIndex === index}
          />
        ))}
      </div>

      {/* Add Step Button */}
      <button
        onClick={addStep}
        className="w-full flex items-center justify-center gap-2 p-3 border-2 border-dashed border-editor-border text-editor-muted hover:border-editor-accent/50 hover:text-editor-text rounded-lg transition-colors"
      >
        <Plus size={18} />
        Add Step
      </button>

      {/* Visual Flow Preview */}
      {steps.length > 0 && (
        <div className="mt-6">
          <h4 className="text-sm font-medium text-editor-text mb-3">Workflow Flow</h4>
          <div className="flex items-center gap-2 flex-wrap">
            {steps.map((step, index) => (
              <div key={step.id} className="flex items-center">
                {index > 0 && !step.parallelWithPrevious && (
                  <ArrowRight size={16} className="text-editor-muted mx-2" />
                )}
                {index > 0 && step.parallelWithPrevious && (
                  <div className="flex items-center gap-1 mx-2">
                    <Layers size={14} className="text-editor-accent" />
                    <span className="text-xs text-editor-accent">parallel</span>
                  </div>
                )}
                <div
                  className={`px-3 py-1.5 rounded-lg text-sm ${
                    step.approverType === 'any'
                      ? 'bg-yellow-500/10 text-yellow-400 border border-yellow-500/30'
                      : 'bg-editor-accent/10 text-editor-accent border border-editor-accent/30'
                  }`}
                >
                  {step.name}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

interface StepCardProps {
  step: ApprovalStepConfig;
  index: number;
  isExpanded: boolean;
  isFirst: boolean;
  isLast: boolean;
  availableRoles: string[];
  onToggleExpanded: () => void;
  onUpdate: (updates: Partial<ApprovalStepConfig>) => void;
  onRemove: () => void;
  onDragStart: () => void;
  onDragOver: (e: React.DragEvent) => void;
  onDragEnd: () => void;
  isDragging: boolean;
}

function StepCard({
  step,
  index,
  isExpanded,
  isFirst,
  availableRoles,
  onToggleExpanded,
  onUpdate,
  onRemove,
  onDragStart,
  onDragOver,
  onDragEnd,
  isDragging,
}: StepCardProps) {
  return (
    <div
      draggable
      onDragStart={onDragStart}
      onDragOver={onDragOver}
      onDragEnd={onDragEnd}
      className={`bg-editor-surface border border-editor-border rounded-lg transition-all ${
        isDragging ? 'opacity-50 scale-[0.98]' : ''
      }`}
    >
      {/* Header */}
      <div className="flex items-center gap-3 p-4">
        <div className="cursor-grab text-editor-muted hover:text-editor-text">
          <GripVertical size={18} />
        </div>

        <div className="w-8 h-8 flex items-center justify-center rounded-full bg-editor-accent/10 text-editor-accent text-sm font-medium">
          {index + 1}
        </div>

        <div className="flex-1 min-w-0">
          <input
            type="text"
            value={step.name}
            onChange={(e) => onUpdate({ name: e.target.value })}
            className="w-full bg-transparent text-editor-text font-medium focus:outline-none"
            placeholder="Step name"
          />
          <div className="flex items-center gap-2 mt-1 text-xs text-editor-muted">
            {step.approverType === 'role' && (
              <>
                <Users size={12} />
                <span>{step.approverRoles?.length || 0} roles</span>
              </>
            )}
            {step.approverType === 'user' && (
              <>
                <User size={12} />
                <span>{step.approverUserIds?.length || 0} users</span>
              </>
            )}
            {step.approverType === 'any' && (
              <>
                <Users size={12} />
                <span>Anyone</span>
              </>
            )}
            <span>·</span>
            <span>{step.requiredApprovals} approval(s) required</span>
          </div>
        </div>

        <button
          onClick={onToggleExpanded}
          className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-bg transition-colors"
        >
          {isExpanded ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
        </button>

        <button
          onClick={onRemove}
          className="p-1.5 rounded text-editor-muted hover:text-editor-error hover:bg-editor-error/10 transition-colors"
        >
          <Trash2 size={18} />
        </button>
      </div>

      {/* Expanded Content */}
      {isExpanded && (
        <div className="px-4 pb-4 space-y-4 border-t border-editor-border/50 pt-4">
          {/* Approver Type */}
          <div>
            <label className="block text-sm font-medium text-editor-text mb-2">
              Approver Type
            </label>
            <div className="flex gap-2">
              {(['role', 'user', 'any'] as const).map((type) => (
                <button
                  key={type}
                  onClick={() => onUpdate({ approverType: type })}
                  className={`flex-1 px-3 py-2 text-sm rounded-lg border transition-colors ${
                    step.approverType === type
                      ? 'bg-editor-accent/10 border-editor-accent text-editor-accent'
                      : 'bg-editor-bg border-editor-border text-editor-muted hover:text-editor-text'
                  }`}
                >
                  {type === 'role' && 'By Role'}
                  {type === 'user' && 'Specific Users'}
                  {type === 'any' && 'Anyone'}
                </button>
              ))}
            </div>
          </div>

          {/* Role Selection */}
          {step.approverType === 'role' && (
            <div>
              <label className="block text-sm font-medium text-editor-text mb-2">
                Roles
              </label>
              <div className="flex flex-wrap gap-2">
                {(availableRoles.length > 0 ? availableRoles : ['admin', 'manager', 'developer', 'reviewer']).map((role) => (
                  <button
                    key={role}
                    onClick={() => {
                      const roles = step.approverRoles || [];
                      if (roles.includes(role)) {
                        onUpdate({ approverRoles: roles.filter(r => r !== role) });
                      } else {
                        onUpdate({ approverRoles: [...roles, role] });
                      }
                    }}
                    className={`px-3 py-1.5 text-sm rounded-lg border transition-colors ${
                      step.approverRoles?.includes(role)
                        ? 'bg-editor-accent/10 border-editor-accent text-editor-accent'
                        : 'bg-editor-bg border-editor-border text-editor-muted hover:text-editor-text'
                    }`}
                  >
                    {role}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Required Approvals */}
          <div>
            <label className="block text-sm font-medium text-editor-text mb-2">
              Required Approvals
            </label>
            <input
              type="number"
              min={1}
              max={10}
              value={step.requiredApprovals}
              onChange={(e) => onUpdate({ requiredApprovals: parseInt(e.target.value) || 1 })}
              className="w-32 px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent"
            />
          </div>

          {/* Timeout */}
          <div>
            <label className="block text-sm font-medium text-editor-text mb-2">
              <div className="flex items-center gap-2">
                <Clock size={14} />
                Timeout (minutes)
              </div>
            </label>
            <input
              type="number"
              min={0}
              value={step.timeoutMinutes || 0}
              onChange={(e) => onUpdate({ timeoutMinutes: parseInt(e.target.value) || undefined })}
              className="w-32 px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent"
              placeholder="No timeout"
            />
            <p className="text-xs text-editor-muted mt-1">
              Leave empty or 0 for no timeout
            </p>
          </div>

          {/* Parallel Execution */}
          {!isFirst && (
            <div className="flex items-center justify-between">
              <div>
                <label className="text-sm font-medium text-editor-text">
                  Parallel with Previous
                </label>
                <p className="text-xs text-editor-muted">
                  Run this step at the same time as the previous step
                </p>
              </div>
              <button
                onClick={() => onUpdate({ parallelWithPrevious: !step.parallelWithPrevious })}
                className={`w-12 h-6 rounded-full transition-colors relative ${
                  step.parallelWithPrevious ? 'bg-editor-accent' : 'bg-editor-border'
                }`}
              >
                <div
                  className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-transform ${
                    step.parallelWithPrevious ? 'translate-x-7' : 'translate-x-1'
                  }`}
                />
              </button>
            </div>
          )}

          {/* Escalation Policy */}
          <div className="p-4 bg-editor-bg rounded-lg">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <AlertTriangle size={16} className="text-orange-400" />
                <span className="text-sm font-medium text-editor-text">Escalation</span>
              </div>
              <button
                onClick={() =>
                  onUpdate({
                    escalationPolicy: step.escalationPolicy?.enabled
                      ? { ...step.escalationPolicy, enabled: false }
                      : { ...DEFAULT_ESCALATION, enabled: true },
                  })
                }
                className={`w-12 h-6 rounded-full transition-colors relative ${
                  step.escalationPolicy?.enabled ? 'bg-orange-500' : 'bg-editor-border'
                }`}
              >
                <div
                  className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-transform ${
                    step.escalationPolicy?.enabled ? 'translate-x-7' : 'translate-x-1'
                  }`}
                />
              </button>
            </div>

            {step.escalationPolicy?.enabled && (
              <div className="space-y-3 mt-3">
                <div>
                  <label className="block text-xs text-editor-muted mb-1">
                    Escalate after (minutes)
                  </label>
                  <input
                    type="number"
                    min={1}
                    value={step.escalationPolicy.escalateAfterMinutes}
                    onChange={(e) =>
                      onUpdate({
                        escalationPolicy: {
                          ...step.escalationPolicy!,
                          escalateAfterMinutes: parseInt(e.target.value) || 30,
                        },
                      })
                    }
                    className="w-32 px-3 py-1.5 bg-editor-surface border border-editor-border rounded text-sm text-editor-text focus:outline-none focus:border-editor-accent"
                  />
                </div>

                <div>
                  <label className="block text-xs text-editor-muted mb-1">
                    Escalate to
                  </label>
                  <select
                    value={step.escalationPolicy.escalateTo}
                    onChange={(e) =>
                      onUpdate({
                        escalationPolicy: {
                          ...step.escalationPolicy!,
                          escalateTo: e.target.value as 'manager' | 'admin' | 'specific_users',
                        },
                      })
                    }
                    className="px-3 py-1.5 bg-editor-surface border border-editor-border rounded text-sm text-editor-text focus:outline-none focus:border-editor-accent"
                  >
                    <option value="manager">Manager</option>
                    <option value="admin">Admin</option>
                    <option value="specific_users">Specific Users</option>
                  </select>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
