import { RecentWorkspaces } from '../components/dashboard/RecentWorkspaces';
import { QuickActions } from '../components/dashboard/QuickActions';
import { UsageSummary } from '../components/dashboard/UsageSummary';
import { SystemStatusDashboard } from '../components/monitoring';

export default function Dashboard() {
  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-8">
        {/* Header */}
        <div className="space-y-2">
          <h1 className="text-2xl font-bold text-editor-text">Dashboard</h1>
          <p className="text-editor-muted">
            Welcome to Prism. Start a new workspace or continue where you left off.
          </p>
        </div>

        {/* System Status */}
        <SystemStatusDashboard />

        {/* Quick Actions */}
        <QuickActions />

        {/* Usage Summary */}
        <UsageSummary />

        {/* Recent Workspaces */}
        <RecentWorkspaces />
      </div>
    </div>
  );
}
