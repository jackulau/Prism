import { UsageChart } from '../components/usage/UsageChart';
import { PlanSelector } from '../components/usage/PlanSelector';
import { BillingInfo } from '../components/usage/BillingInfo';

export default function Usage() {
  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-8">
        {/* Header */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold text-editor-text">Usage & Billing</h1>
          <p className="text-editor-muted">
            Monitor your usage and manage your subscription
          </p>
        </div>

        {/* Billing Info */}
        <BillingInfo />

        {/* Usage Charts */}
        <UsageChart />

        {/* Plan Selector */}
        <PlanSelector />
      </div>
    </div>
  );
}
