import { OrgNameEditor } from '../components/organization/OrgNameEditor';
import { MemberList } from '../components/organization/MemberList';
import { SubscriptionInfo } from '../components/organization/SubscriptionInfo';
import { OrgWorkspaceManager } from '../components/organization/OrgWorkspaceManager';

export default function Organization() {
  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-4xl mx-auto p-6 space-y-8">
        {/* Header */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold text-editor-text">Organization</h1>
          <p className="text-editor-muted">
            Manage your organization settings and team members
          </p>
        </div>

        {/* Organization Name */}
        <OrgNameEditor />

        {/* Subscription Info */}
        <SubscriptionInfo />

        {/* Workspaces */}
        <OrgWorkspaceManager />

        {/* Member List */}
        <MemberList />
      </div>
    </div>
  );
}
