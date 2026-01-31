import { createBrowserRouter, Navigate } from 'react-router-dom';
import { AppLayout } from '../layouts/AppLayout';
import Dashboard from '../pages/Dashboard';
import Workers from '../pages/Workers';
import Workflows from '../pages/Workflows';
import WorkflowDetail from '../pages/WorkflowDetail';
import Integrations from '../pages/Integrations';
import Usage from '../pages/Usage';
import Organization from '../pages/Organization';
import { SettingsPage } from '../pages/Settings';
import { AuthPage } from '../components/auth/AuthPage';
import Workspace from '../pages/Workspace';

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
        path: 'workflows',
        element: <Workflows />,
      },
      {
        path: 'workflows/:id',
        element: <WorkflowDetail />,
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
    ],
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);
