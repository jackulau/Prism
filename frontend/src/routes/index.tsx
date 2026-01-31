import { createBrowserRouter, Navigate } from 'react-router-dom';
import { AppLayout } from '../layouts/AppLayout';
import Dashboard from '../pages/Dashboard';
import Workers from '../pages/Workers';
import Integrations from '../pages/Integrations';
import Usage from '../pages/Usage';
import Organization from '../pages/Organization';
import TeamsPage from '../pages/settings/Teams';
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
