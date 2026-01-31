import { useState, useEffect } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from '../components/sidebar/Sidebar';
import { SettingsPanel } from '../components/settings/SettingsPanel';
import { ToastContainer } from '../components/Toast';
import { AuthGuard } from '../components/auth/AuthGuard';
import { AuthPage } from '../components/auth/AuthPage';
import { TRPCProvider } from '../providers/TRPCProvider';
import { TeamSelector } from '../components/layout/TeamSelector';
import { useAppStore } from '../store';
import { applyTheme } from '../config/themes';

export function AppLayout() {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const { theme, loadProviders } = useAppStore();

  // Apply theme on mount and when it changes
  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  // Load providers on mount
  useEffect(() => {
    loadProviders();
  }, [loadProviders]);

  return (
    <TRPCProvider>
      <AuthGuard fallback={<AuthPage />}>
        <div className="h-screen w-screen flex bg-editor-bg text-editor-text overflow-hidden">
          {/* Sidebar */}
          <Sidebar
            isCollapsed={isSidebarCollapsed}
            onToggle={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
          />

          {/* Main Content Area */}
          <main className="flex-1 flex flex-col overflow-hidden">
            {/* Top Bar with Team Selector */}
            <div className="flex items-center justify-between px-4 py-2 border-b border-editor-border bg-editor-surface/50">
              <TeamSelector />
              <div className="flex items-center gap-2">
                {/* Additional header actions can go here */}
              </div>
            </div>
            <div className="flex-1 overflow-hidden">
              <Outlet />
            </div>
          </main>

          {/* Settings Panel (slide-out) */}
          <SettingsPanel />

          {/* Toast Notifications */}
          <ToastContainer />
        </div>
      </AuthGuard>
    </TRPCProvider>
  );
}
