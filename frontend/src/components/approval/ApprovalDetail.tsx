import { useState } from 'react';
import { DiffEditor } from '@monaco-editor/react';
import {
  X,
  Check,
  Clock,
  AlertTriangle,
  User,
  MessageSquare,
  Wrench,
  Settings,
  Shield,
  Zap,
  ExternalLink,
} from 'lucide-react';
import type { ApprovalRequest, ApprovalDecision, ApprovalType } from '../../types/approval';

interface ApprovalDetailProps {
  approval: ApprovalRequest;
  isOpen: boolean;
  onClose: () => void;
  onApprove: (comment?: string) => void;
  onReject: (comment?: string) => void;
}

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
  custom: 'Custom Request',
};

const STATUS_STYLES = {
  pending: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30',
  approved: 'bg-green-500/10 text-green-400 border-green-500/30',
  rejected: 'bg-red-500/10 text-red-400 border-red-500/30',
  escalated: 'bg-orange-500/10 text-orange-400 border-orange-500/30',
  expired: 'bg-gray-500/10 text-gray-400 border-gray-500/30',
};

function formatDateTime(date: Date): string {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  }).format(date);
}

function formatTimeRemaining(expiresAt: Date): string {
  const now = new Date();
  const diffMs = expiresAt.getTime() - now.getTime();

  if (diffMs <= 0) return 'Expired';

  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const remainingMins = diffMins % 60;

  if (diffHours > 0) {
    return `${diffHours}h ${remainingMins}m remaining`;
  }
  return `${diffMins}m remaining`;
}

export function ApprovalDetail({
  approval,
  isOpen,
  onClose,
  onApprove,
  onReject,
}: ApprovalDetailProps) {
  const [comment, setComment] = useState('');
  const [activeTab, setActiveTab] = useState<'details' | 'history' | 'diff'>('details');

  if (!isOpen) return null;

  const TypeIcon = TYPE_ICONS[approval.type];
  const isPending = approval.status === 'pending' || approval.status === 'escalated';

  const handleApprove = () => {
    onApprove(comment || undefined);
    setComment('');
  };

  const handleReject = () => {
    onReject(comment || undefined);
    setComment('');
  };

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/60 z-50 animate-fade-in"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="fixed inset-4 md:inset-auto md:top-1/2 md:left-1/2 md:-translate-x-1/2 md:-translate-y-1/2 md:w-[800px] md:max-h-[85vh] bg-editor-bg border border-editor-border rounded-xl shadow-2xl z-50 flex flex-col overflow-hidden animate-scale-in">
        {/* Header */}
        <div className="flex items-start justify-between p-6 border-b border-editor-border">
          <div className="flex items-start gap-4">
            <div className="p-3 rounded-xl bg-editor-accent/10 text-editor-accent">
              <TypeIcon size={24} />
            </div>
            <div>
              <h2 className="text-xl font-semibold text-editor-text">{approval.title}</h2>
              <div className="flex items-center gap-3 mt-2">
                <span className="text-sm text-editor-muted">{TYPE_LABELS[approval.type]}</span>
                <span
                  className={`px-2 py-0.5 text-xs rounded-full border ${STATUS_STYLES[approval.status]}`}
                >
                  {approval.status.charAt(0).toUpperCase() + approval.status.slice(1)}
                </span>
                {approval.currentStep && approval.totalSteps && (
                  <span className="text-sm text-editor-muted">
                    Step {approval.currentStep} of {approval.totalSteps}
                  </span>
                )}
              </div>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-surface transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-editor-border">
          <button
            onClick={() => setActiveTab('details')}
            className={`px-6 py-3 text-sm font-medium transition-colors relative ${
              activeTab === 'details'
                ? 'text-editor-accent'
                : 'text-editor-muted hover:text-editor-text'
            }`}
          >
            Details
            {activeTab === 'details' && (
              <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-editor-accent" />
            )}
          </button>
          <button
            onClick={() => setActiveTab('history')}
            className={`px-6 py-3 text-sm font-medium transition-colors relative ${
              activeTab === 'history'
                ? 'text-editor-accent'
                : 'text-editor-muted hover:text-editor-text'
            }`}
          >
            History ({approval.decisions.length})
            {activeTab === 'history' && (
              <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-editor-accent" />
            )}
          </button>
          {approval.configDiff && (
            <button
              onClick={() => setActiveTab('diff')}
              className={`px-6 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'diff'
                  ? 'text-editor-accent'
                  : 'text-editor-muted hover:text-editor-text'
              }`}
            >
              Changes
              {activeTab === 'diff' && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-editor-accent" />
              )}
            </button>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {activeTab === 'details' && (
            <div className="space-y-6">
              {/* Description */}
              {approval.description && (
                <div>
                  <h3 className="text-sm font-medium text-editor-text mb-2">Description</h3>
                  <p className="text-sm text-editor-muted">{approval.description}</p>
                </div>
              )}

              {/* Tool info */}
              {approval.toolName && (
                <div>
                  <h3 className="text-sm font-medium text-editor-text mb-2">Tool</h3>
                  <div className="p-3 bg-editor-surface rounded-lg">
                    <div className="flex items-center gap-2 mb-2">
                      <Wrench size={14} className="text-editor-muted" />
                      <code className="text-sm text-editor-accent">{approval.toolName}</code>
                    </div>
                    {approval.toolParameters && (
                      <pre className="text-xs text-editor-muted bg-editor-bg p-2 rounded overflow-x-auto">
                        {JSON.stringify(approval.toolParameters, null, 2)}
                      </pre>
                    )}
                  </div>
                </div>
              )}

              {/* Requestor */}
              <div>
                <h3 className="text-sm font-medium text-editor-text mb-2">Requested By</h3>
                <div className="flex items-center gap-3 p-3 bg-editor-surface rounded-lg">
                  <div className="w-10 h-10 rounded-full bg-editor-accent/10 flex items-center justify-center">
                    <User size={18} className="text-editor-accent" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-editor-text">
                      {approval.requestedBy.name || approval.requestedBy.email}
                    </p>
                    {approval.requestedBy.role && (
                      <p className="text-xs text-editor-muted">{approval.requestedBy.role}</p>
                    )}
                  </div>
                </div>
              </div>

              {/* Timing */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h3 className="text-sm font-medium text-editor-text mb-2">Requested At</h3>
                  <div className="flex items-center gap-2 text-sm text-editor-muted">
                    <Clock size={14} />
                    {formatDateTime(approval.requestedAt)}
                  </div>
                </div>
                {approval.expiresAt && (
                  <div>
                    <h3 className="text-sm font-medium text-editor-text mb-2">Expires</h3>
                    <div className="flex items-center gap-2 text-sm text-editor-warning">
                      <AlertTriangle size={14} />
                      {formatTimeRemaining(approval.expiresAt)}
                    </div>
                  </div>
                )}
              </div>

              {/* Context links */}
              {(approval.conversationId || approval.workspaceId) && (
                <div>
                  <h3 className="text-sm font-medium text-editor-text mb-2">Context</h3>
                  <div className="flex flex-wrap gap-2">
                    {approval.conversationId && (
                      <a
                        href={`/workspace/${approval.conversationId}`}
                        className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-accent bg-editor-accent/10 rounded-lg hover:bg-editor-accent/20 transition-colors"
                      >
                        <MessageSquare size={14} />
                        View Conversation
                        <ExternalLink size={12} />
                      </a>
                    )}
                  </div>
                </div>
              )}

              {/* Escalation info */}
              {approval.status === 'escalated' && approval.escalatedTo && (
                <div className="p-4 bg-orange-500/5 border border-orange-500/20 rounded-lg">
                  <div className="flex items-center gap-2 mb-2">
                    <AlertTriangle size={16} className="text-orange-400" />
                    <span className="font-medium text-orange-400">Escalated</span>
                  </div>
                  <p className="text-sm text-editor-muted mb-2">
                    This request has been escalated to:
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {approval.escalatedTo.map((approver) => (
                      <span
                        key={approver.id}
                        className="px-2 py-1 text-xs bg-editor-surface rounded"
                      >
                        {approver.name || approver.email}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {activeTab === 'history' && (
            <div className="space-y-4">
              {approval.decisions.length === 0 ? (
                <p className="text-sm text-editor-muted text-center py-8">
                  No decisions have been made yet.
                </p>
              ) : (
                <div className="relative">
                  {/* Timeline line */}
                  <div className="absolute left-4 top-0 bottom-0 w-0.5 bg-editor-border" />

                  {approval.decisions.map((decision, index) => (
                    <DecisionTimelineItem
                      key={decision.id}
                      decision={decision}
                      isLast={index === approval.decisions.length - 1}
                    />
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'diff' && approval.configDiff && (
            <div className="h-[400px] rounded-lg overflow-hidden border border-editor-border">
              <DiffEditor
                height="100%"
                language="json"
                original={approval.configDiff.before}
                modified={approval.configDiff.after}
                theme="vs-dark"
                options={{
                  fontSize: 13,
                  fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
                  readOnly: true,
                  renderSideBySide: true,
                  automaticLayout: true,
                  minimap: { enabled: false },
                  scrollBeyondLastLine: false,
                }}
              />
            </div>
          )}
        </div>

        {/* Footer with actions */}
        {isPending && (
          <div className="p-6 border-t border-editor-border bg-editor-surface/50">
            <div className="mb-4">
              <textarea
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                placeholder="Add a comment (optional)..."
                className="w-full px-4 py-3 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text placeholder:text-editor-muted resize-none focus:outline-none focus:border-editor-accent"
                rows={2}
              />
            </div>
            <div className="flex items-center justify-end gap-3">
              <button
                onClick={onClose}
                className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleReject}
                className="flex items-center gap-2 px-4 py-2 text-sm text-white bg-editor-error hover:bg-editor-error/80 rounded-lg transition-colors"
              >
                <X size={16} />
                Reject
              </button>
              <button
                onClick={handleApprove}
                className="flex items-center gap-2 px-4 py-2 text-sm text-white bg-editor-success hover:bg-editor-success/80 rounded-lg transition-colors"
              >
                <Check size={16} />
                Approve
              </button>
            </div>
          </div>
        )}
      </div>
    </>
  );
}

interface DecisionTimelineItemProps {
  decision: ApprovalDecision;
  isLast: boolean;
}

function DecisionTimelineItem({ decision, isLast }: DecisionTimelineItemProps) {
  const isApproved = decision.decision === 'approved';

  return (
    <div className={`relative pl-10 ${isLast ? '' : 'pb-6'}`}>
      {/* Timeline dot */}
      <div
        className={`absolute left-2.5 w-3 h-3 rounded-full border-2 ${
          isApproved
            ? 'bg-green-500 border-green-500'
            : 'bg-red-500 border-red-500'
        }`}
      />

      <div className="p-4 bg-editor-surface rounded-lg">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            {isApproved ? (
              <Check size={16} className="text-green-400" />
            ) : (
              <X size={16} className="text-red-400" />
            )}
            <span className={`font-medium ${isApproved ? 'text-green-400' : 'text-red-400'}`}>
              {isApproved ? 'Approved' : 'Rejected'}
            </span>
          </div>
          <span className="text-xs text-editor-muted">
            {formatDateTime(decision.decidedAt)}
          </span>
        </div>

        <div className="flex items-center gap-2 mb-2">
          <User size={14} className="text-editor-muted" />
          <span className="text-sm text-editor-text">
            {decision.decidedBy.name || decision.decidedBy.email}
          </span>
        </div>

        {decision.comment && (
          <div className="mt-2 p-3 bg-editor-bg rounded text-sm text-editor-muted">
            <MessageSquare size={12} className="inline mr-2" />
            {decision.comment}
          </div>
        )}
      </div>
    </div>
  );
}
