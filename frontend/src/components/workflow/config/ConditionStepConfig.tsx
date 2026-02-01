import { useState } from 'react';
import { Code, ArrowRight, Check, X } from 'lucide-react';
import { useWorkflowStore } from '../../../store/workflowStore';
import type { ConditionConfig } from '../../../types/workflow';
import { CONDITION_OPERATORS } from '../../../types/workflow';
import { StateVariablePicker } from './StateVariablePicker';

interface ConditionStepConfigProps {
  nodeId: string;
}

export function ConditionStepConfig({ nodeId }: ConditionStepConfigProps) {
  const { getSelectedNode, updateNodeConfig, nodes } = useWorkflowStore();
  const [showRawExpression, setShowRawExpression] = useState(false);

  const node = getSelectedNode();
  const nodeData = node?.data as { config?: { conditionConfig?: ConditionConfig } } | undefined;
  const config = nodeData?.config?.conditionConfig || node?.config?.conditionConfig;

  if (!node || !config) return null;

  const updateConfig = (updates: Partial<ConditionConfig>) => {
    updateNodeConfig(nodeId, {
      conditionConfig: { ...config, ...updates },
    });
  };

  // Parse expression into parts for the builder
  const parseExpression = (expr: string): { left: string; operator: string; right: string } => {
    // Try to parse expressions like "{{state.var}} == value"
    const match = expr.match(/^(.+?)\s*(==|!=|>|<|>=|<=|contains|exists)\s*(.*)$/);
    if (match) {
      return {
        left: match[1].trim(),
        operator: match[2],
        right: match[3].trim(),
      };
    }
    return { left: expr, operator: '==', right: '' };
  };

  const buildExpression = (left: string, operator: string, right: string): string => {
    if (operator === 'exists') {
      return `${left} exists`;
    }
    return `${left} ${operator} ${right}`;
  };

  const parsed = parseExpression(config.expression);
  const [leftOperand, setLeftOperand] = useState(parsed.left);
  const [operator, setOperator] = useState(parsed.operator);
  const [rightOperand, setRightOperand] = useState(parsed.right);

  const handleExpressionChange = (left: string, op: string, right: string) => {
    setLeftOperand(left);
    setOperator(op);
    setRightOperand(right);
    updateConfig({ expression: buildExpression(left, op, right) });
  };

  const handleInsertVariable = (variable: string) => {
    setLeftOperand(variable);
    handleExpressionChange(variable, operator, rightOperand);
  };

  // Get available steps for branch selection (excluding current node)
  const availableSteps = nodes
    .filter((n) => n.id !== nodeId)
    .map((n) => ({ id: n.id, name: n.name || (n.data as Record<string, unknown>)?.name as string || n.id }));

  return (
    <div className="space-y-4">
      {/* Expression Builder Toggle */}
      <div className="flex items-center justify-between">
        <label className="block text-sm font-medium text-editor-text">
          Condition Expression
        </label>
        <button
          type="button"
          onClick={() => setShowRawExpression(!showRawExpression)}
          className="flex items-center gap-1 px-2 py-1 text-xs text-editor-muted hover:text-editor-text transition-colors"
        >
          <Code size={12} />
          {showRawExpression ? 'Builder View' : 'Raw Expression'}
        </button>
      </div>

      {showRawExpression ? (
        /* Raw Expression Input */
        <div className="space-y-2">
          <textarea
            value={config.expression}
            onChange={(e) => updateConfig({ expression: e.target.value })}
            placeholder="{{state.result}} == 'success'"
            rows={3}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm resize-none"
          />
          <p className="text-xs text-editor-muted">
            Enter a boolean expression. Use <code className="bg-editor-surface px-1 rounded">{'{{state.var}}'}</code> for state variables.
          </p>
        </div>
      ) : (
        /* Expression Builder */
        <div className="space-y-3 p-3 bg-editor-surface/50 rounded-lg border border-editor-border">
          {/* Left Operand */}
          <div className="space-y-2">
            <label className="block text-xs text-editor-muted">Left Operand</label>
            <div className="flex gap-2">
              <input
                type="text"
                value={leftOperand}
                onChange={(e) => handleExpressionChange(e.target.value, operator, rightOperand)}
                placeholder="{{state.variable}}"
                className="flex-1 px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
              />
            </div>
            <StateVariablePicker nodeId={nodeId} onInsert={handleInsertVariable} />
          </div>

          {/* Operator */}
          <div className="space-y-2">
            <label className="block text-xs text-editor-muted">Operator</label>
            <select
              value={operator}
              onChange={(e) => handleExpressionChange(leftOperand, e.target.value, rightOperand)}
              className="w-full px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent text-sm"
            >
              {CONDITION_OPERATORS.map((op) => (
                <option key={op.value} value={op.value}>
                  {op.label} ({op.value})
                </option>
              ))}
            </select>
          </div>

          {/* Right Operand (hidden for 'exists' operator) */}
          {operator !== 'exists' && (
            <div className="space-y-2">
              <label className="block text-xs text-editor-muted">Right Operand</label>
              <input
                type="text"
                value={rightOperand}
                onChange={(e) => handleExpressionChange(leftOperand, operator, e.target.value)}
                placeholder="value or {{state.variable}}"
                className="w-full px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
              />
            </div>
          )}

          {/* Expression Preview */}
          <div className="pt-2 border-t border-editor-border">
            <div className="text-xs text-editor-muted mb-1">Preview:</div>
            <code className="block px-3 py-2 bg-editor-bg rounded-lg text-sm text-editor-accent font-mono">
              {config.expression || 'Enter condition...'}
            </code>
          </div>
        </div>
      )}

      {/* Branches */}
      <div className="space-y-4 pt-2">
        <h4 className="text-sm font-medium text-editor-text">Branches</h4>

        {/* True Branch */}
        <div className="space-y-2">
          <label className="flex items-center gap-2 text-sm font-medium text-editor-success">
            <Check size={14} />
            If True, go to:
          </label>
          <select
            value={config.trueBranch}
            onChange={(e) => updateConfig({ trueBranch: e.target.value })}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
          >
            <option value="">Select next step...</option>
            {availableSteps.map((step) => (
              <option key={step.id} value={step.id}>
                {step.name}
              </option>
            ))}
          </select>
        </div>

        {/* False Branch */}
        <div className="space-y-2">
          <label className="flex items-center gap-2 text-sm font-medium text-editor-error">
            <X size={14} />
            If False, go to:
          </label>
          <select
            value={config.falseBranch}
            onChange={(e) => updateConfig({ falseBranch: e.target.value })}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
          >
            <option value="">Select next step...</option>
            {availableSteps.map((step) => (
              <option key={step.id} value={step.id}>
                {step.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Visual Branch Preview */}
      <div className="p-3 bg-editor-surface/50 rounded-lg border border-editor-border">
        <div className="flex items-center justify-center gap-4 text-sm">
          <div className="flex items-center gap-2 text-editor-success">
            <Check size={14} />
            <span>{availableSteps.find((s) => s.id === config.trueBranch)?.name || 'Not set'}</span>
          </div>
          <div className="flex items-center gap-1 text-editor-muted">
            <ArrowRight size={14} />
          </div>
          <div className="px-2 py-1 bg-editor-accent/10 rounded text-editor-accent text-xs">
            Condition
          </div>
          <div className="flex items-center gap-1 text-editor-muted">
            <ArrowRight size={14} />
          </div>
          <div className="flex items-center gap-2 text-editor-error">
            <X size={14} />
            <span>{availableSteps.find((s) => s.id === config.falseBranch)?.name || 'Not set'}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
