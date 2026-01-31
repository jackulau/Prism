import { useState, useEffect } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from '../components/sidebar/Sidebar';
import { SettingsPanel } from '../components/settings/SettingsPanel';
import { ToastContainer } from '../components/Toast';
import { AuthGuard } from '../components/auth/AuthGuard';
import { AuthPage } from '../components/auth/AuthPage';
import { TRPCProvider } from '../providers/TRPCProvider';
import { CommandPalette } from '../components/CommandPalette';
import { useAppStore } from '../store';
import { useCommandPaletteStore } from '../store/commandPaletteStore';
import { applyTheme } from '../config/themes';

export function AppLayout() {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const { theme, loadProviders } = useAppStore();
  const { open: openCommandPalette } = useCommandPaletteStore();

  // Apply theme on mount and when it changes
  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  // Load providers on mount
  useEffect(() => {
    loadProviders();
  }, [loadProviders]);

  // Global keyboard shortcut for command palette (Cmd/Ctrl + K)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        openCommandPalette();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [openCommandPalette]);

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
            <Outlet />
          </main>

          {/* Settings Panel (slide-out) */}
          <SettingsPanel />

          {/* Toast Notifications */}
          <ToastContainer />

          {/* Command Palette */}
          <CommandPalette />
        </div>
      </AuthGuard>
    </TRPCProvider>
  );
}
