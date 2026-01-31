import { AuditStatsCard } from '../components/admin/AuditStatsCard';
import { AuditLogTable } from '../components/admin/AuditLogTable';
import { Shield } from 'lucide-react';

export default function AuditDashboard() {
  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-8">
        {/* Header */}
        <div className="space-y-2">
          <div className="flex items-center gap-3">
            <Shield className="w-7 h-7 text-editor-accent" />
            <h1 className="text-2xl font-bold text-editor-text">Audit Dashboard</h1>
          </div>
          <p className="text-editor-muted">
            Monitor security events and access detailed audit logs across your organization.
          </p>
        </div>

        {/* Stats Overview */}
        <AuditStatsCard />

        {/* Audit Log Table */}
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Audit Logs</h2>
          </div>
          <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
            <AuditLogTable />
          </div>
        </section>
      </div>
    </div>
  );
}
