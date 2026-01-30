import { useState, useEffect } from 'react';
import {
  Shield,
  Clock,
  AlertTriangle,
  Lock,
  Unlock,
  Save,
  RotateCcw,
  Info,
} from 'lucide-react';
import { useAuthStore } from '../../store/authStore';
import { useAuditStore } from '../../store/auditStore';
import { auditService } from '../../services/audit';
import type { RetentionPolicy as RetentionPolicyType, DataType } from '../../types/audit';

const dataTypeLabels: Record<DataType, string> = {
  conversations: 'Conversations',
  messages: 'Messages',
  audit_logs: 'Audit Logs',
  exports: 'Export Files',
  user_data: 'User Data',
  analytics: 'Analytics Data',
};

const dataTypeDescriptions: Record<DataType, string> = {
  conversations: 'Chat conversations and their metadata',
  messages: 'Individual messages within conversations',
  audit_logs: 'Activity logs and compliance audit trail',
  exports: 'Generated export files and downloads',
  user_data: 'User profile and preferences',
  analytics: 'Usage metrics and analytics data',
};

const retentionOptions = [
  { days: 30, label: '30 days' },
  { days: 90, label: '90 days' },
  { days: 180, label: '6 months' },
  { days: 365, label: '1 year' },
  { days: 730, label: '2 years' },
  { days: 1095, label: '3 years' },
  { days: 2555, label: '7 years' },
  { days: -1, label: 'Forever' },
];

interface PolicyRowProps {
  policy: RetentionPolicyType;
  onUpdate: (dataType: DataType, days: number, legalHold: boolean) => Promise<void>;
  isUpdating: boolean;
}

function PolicyRow({ policy, onUpdate, isUpdating }: PolicyRowProps) {
  const [days, setDays] = useState(policy.retentionDays);
  const [legalHold, setLegalHold] = useState(policy.legalHold);
  const [hasChanges, setHasChanges] = useState(false);

  useEffect(() => {
    setDays(policy.retentionDays);
    setLegalHold(policy.legalHold);
    setHasChanges(false);
  }, [policy]);

  const handleDaysChange = (newDays: number) => {
    setDays(newDays);
    setHasChanges(newDays !== policy.retentionDays || legalHold !== policy.legalHold);
  };

  const handleLegalHoldChange = (newLegalHold: boolean) => {
    setLegalHold(newLegalHold);
    setHasChanges(days !== policy.retentionDays || newLegalHold !== policy.legalHold);
  };

  const handleSave = async () => {
    await onUpdate(policy.dataType, days, legalHold);
    setHasChanges(false);
  };

  const handleReset = () => {
    setDays(policy.retentionDays);
    setLegalHold(policy.legalHold);
    setHasChanges(false);
  };

  return (
    <div className={`p-4 rounded-lg border ${legalHold ? 'bg-editor-warning/5 border-editor-warning/30' : 'bg-editor-bg border-editor-border'}`}>
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h4 className="font-medium text-editor-text">
              {dataTypeLabels[policy.dataType]}
            </h4>
            {legalHold && (
              <span className="flex items-center gap-1 px-2 py-0.5 text-xs font-medium bg-editor-warning/20 text-editor-warning rounded-full">
                <Lock size={10} />
                Legal Hold
              </span>
            )}
          </div>
          <p className="text-sm text-editor-muted mt-1">
            {dataTypeDescriptions[policy.dataType]}
          </p>
        </div>

        <div className="flex items-center gap-3">
          {/* Retention Period Selector */}
          <div className="flex flex-col items-end gap-1">
            <label className="text-xs text-editor-muted">Retention</label>
            <select
              value={days}
              onChange={(e) => handleDaysChange(Number(e.target.value))}
              disabled={!policy.isConfigurable || legalHold}
              className="px-3 py-1.5 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {retentionOptions.map((opt) => (
                <option key={opt.days} value={opt.days}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>

          {/* Legal Hold Toggle */}
          <div className="flex flex-col items-center gap-1">
            <label className="text-xs text-editor-muted">Legal Hold</label>
            <button
              onClick={() => handleLegalHoldChange(!legalHold)}
              disabled={!policy.isConfigurable}
              className={`p-2 rounded-lg transition-colors ${
                legalHold
                  ? 'bg-editor-warning text-white'
                  : 'bg-editor-surface text-editor-muted hover:text-editor-text'
              } disabled:opacity-50 disabled:cursor-not-allowed`}
              title={legalHold ? 'Release legal hold' : 'Enable legal hold'}
            >
              {legalHold ? <Lock size={16} /> : <Unlock size={16} />}
            </button>
          </div>

          {/* Save/Reset Buttons */}
          {hasChanges && (
            <div className="flex items-center gap-1 ml-2">
              <button
                onClick={handleSave}
                disabled={isUpdating}
                className="p-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50"
                title="Save changes"
              >
                <Save size={16} />
              </button>
              <button
                onClick={handleReset}
                disabled={isUpdating}
                className="p-2 bg-editor-surface text-editor-muted rounded-lg hover:text-editor-text"
                title="Reset changes"
              >
                <RotateCcw size={16} />
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Legal Hold Warning */}
      {legalHold && (
        <div className="flex items-center gap-2 mt-3 p-2 bg-editor-warning/10 rounded-lg">
          <AlertTriangle size={14} className="text-editor-warning flex-shrink-0" />
          <span className="text-xs text-editor-warning">
            Data under legal hold will not be automatically deleted regardless of retention settings.
          </span>
        </div>
      )}

      {/* Last Updated */}
      {policy.lastUpdated && (
        <div className="flex items-center gap-1 mt-3 text-xs text-editor-muted">
          <Clock size={12} />
          <span>
            Last updated {new Date(policy.lastUpdated).toLocaleDateString()}
            {policy.updatedBy && ` by ${policy.updatedBy.email}`}
          </span>
        </div>
      )}
    </div>
  );
}

export function RetentionPolicy() {
  const { accessToken } = useAuthStore();
  const {
    retentionPolicies,
    retentionLoading,
    retentionError,
    setRetentionPolicies,
    updateRetentionPolicy,
    setRetentionLoading,
    setRetentionError,
  } = useAuditStore();
  const [updatingPolicy, setUpdatingPolicy] = useState<DataType | null>(null);

  useEffect(() => {
    if (accessToken) {
      auditService.setToken(accessToken);
      fetchPolicies();
    }
  }, [accessToken]);

  const fetchPolicies = async () => {
    setRetentionLoading(true);
    setRetentionError(null);

    const response = await auditService.getRetentionPolicies();
    if (response.error) {
      setRetentionError(response.error);
    } else if (response.data) {
      setRetentionPolicies(response.data.policies);
    }
    setRetentionLoading(false);
  };

  const handleUpdate = async (dataType: DataType, days: number, legalHold: boolean) => {
    setUpdatingPolicy(dataType);
    setRetentionError(null);

    const response = await auditService.updateRetentionPolicy({
      dataType,
      retentionDays: days,
      legalHold,
    });

    if (response.error) {
      setRetentionError(response.error);
    } else if (response.data) {
      updateRetentionPolicy(dataType, response.data);
    }
    setUpdatingPolicy(null);
  };

  if (retentionLoading && retentionPolicies.length === 0) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Shield className="w-5 h-5" />
          <h3 className="text-lg font-semibold">Data Retention</h3>
        </div>
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="h-24 bg-editor-surface border border-editor-border rounded-lg animate-pulse"
            />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Shield className="w-5 h-5" />
          <h3 className="text-lg font-semibold">Data Retention</h3>
        </div>
      </div>

      {/* Info Banner */}
      <div className="flex items-start gap-3 p-4 bg-editor-accent/10 border border-editor-accent/20 rounded-lg">
        <Info size={18} className="text-editor-accent flex-shrink-0 mt-0.5" />
        <div className="text-sm text-editor-text">
          <p>
            Configure how long different types of data are retained. Data exceeding the retention
            period will be automatically deleted unless under legal hold.
          </p>
          <p className="mt-2 text-editor-muted">
            Note: Some retention periods may be subject to minimum requirements based on your
            subscription plan or regulatory obligations.
          </p>
        </div>
      </div>

      {/* Error Message */}
      {retentionError && (
        <div className="flex items-center gap-2 p-3 bg-editor-error/10 border border-editor-error/20 rounded-lg">
          <AlertTriangle size={16} className="text-editor-error" />
          <span className="text-sm text-editor-error">{retentionError}</span>
        </div>
      )}

      {/* Policy List */}
      <div className="space-y-3">
        {retentionPolicies.map((policy) => (
          <PolicyRow
            key={policy.dataType}
            policy={policy}
            onUpdate={handleUpdate}
            isUpdating={updatingPolicy === policy.dataType}
          />
        ))}
      </div>

      {/* Empty State */}
      {retentionPolicies.length === 0 && !retentionLoading && (
        <div className="text-center py-8 text-editor-muted">
          <Shield size={48} className="mx-auto mb-4 opacity-50" />
          <p>No retention policies configured.</p>
          <p className="text-sm mt-1">
            Contact your administrator to set up data retention policies.
          </p>
        </div>
      )}
    </div>
  );
}
