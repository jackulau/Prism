import { createBrowserRouter, Navigate } from 'react-router-dom';
import { AppLayout } from '../layouts/AppLayout';
import Dashboard from '../pages/Dashboard';
import Agents from '../pages/Agents';
import Workers from '../pages/Workers';
import Workflows from '../pages/Workflows';
import Tasks from '../pages/Tasks';
import Tools from '../pages/Tools';
import Integrations from '../pages/Integrations';
import Usage from '../pages/Usage';
import Organization from '../pages/Organization';
import TeamsPage from '../pages/settings/Teams';
import AuditDashboard from '../pages/AuditDashboard';
import { SettingsPage } from '../pages/Settings';
import { AuthPage } from '../components/auth/AuthPage';
import Workspace from '../pages/Workspace';
import Onboarding from '../pages/Onboarding';
import Activity from '../pages/Activity';
import Approvals from '../pages/Approvals';
import AuditLogs from '../pages/settings/AuditLogs';
import DataExport from '../pages/settings/DataExport';
import ComplianceReports from '../pages/settings/ComplianceReports';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <AuthPage />,
  },
  {
    path: '/onboarding',
    element: <Onboarding />,
  },
  {
    path: '/',
    element: <AppLayout />,
    children: [
      {
        index: true,
        element: <Dashboard />,
      },
      {
        path: 'workspace/:id?',
        element: <Workspace />,
      },
      {
        path: 'agents',
        element: <Agents />,
      },
      {
        path: 'approvals',
        element: <Approvals />,
      },
      {
        path: 'workers',
        element: <Workers />,
      },
      {
        path: 'agents/:id',
        element: <Agents />,
      },
      {
        path: 'workflows',
        element: <Workflows />,
      },
      {
        path: 'tasks',
        element: <Tasks />,
      },
      {
        path: 'tools',
        element: <Tools />,
      },
      {
        path: 'integrations',
        element: <Integrations />,
      },
      {
        path: 'usage',
        element: <Usage />,
      },
      {
        path: 'organization',
        element: <Organization />,
      },
      {
        path: 'teams',
        element: <TeamsPage />,
      },
      {
        path: 'audit',
        element: <AuditDashboard />,
      },
      {
        path: 'settings',
        element: <SettingsPage />,
      },
      {
        path: 'activity',
        element: <Activity />,
      },
      {
        path: 'settings/audit-logs',
        element: <AuditLogs />,
      },
      {
        path: 'settings/data-export',
        element: <DataExport />,
      },
      {
        path: 'settings/compliance',
        element: <ComplianceReports />,
      },
    ],
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);
