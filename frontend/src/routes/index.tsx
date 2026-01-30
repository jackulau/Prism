import { createBrowserRouter, Navigate } from 'react-router-dom';
import { AppLayout } from '../layouts/AppLayout';
import Dashboard from '../pages/Dashboard';
import Workers from '../pages/Workers';
import Integrations from '../pages/Integrations';
import Usage from '../pages/Usage';
import Organization from '../pages/Organization';
import { SettingsPage } from '../pages/Settings';
import { AuthPage } from '../components/auth/AuthPage';
import Workspace from '../pages/Workspace';
import AuditLogs from '../pages/settings/AuditLogs';
import DataExport from '../pages/settings/DataExport';
import ComplianceReports from '../pages/settings/ComplianceReports';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <AuthPage />,
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
        path: 'workers',
        element: <Workers />,
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
        path: 'settings',
        element: <SettingsPage />,
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
