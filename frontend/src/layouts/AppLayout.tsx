import { useState, useEffect, useCallback } from 'react';
import { Outlet, useNavigate } from 'react-router-dom';
import { Sidebar } from '../components/sidebar/Sidebar';
import { SettingsPanel } from '../components/settings/SettingsPanel';
import { ToastContainer } from '../components/Toast';
import { KeyboardShortcutsHelp } from '../components/KeyboardShortcutsHelp';
import { FloatingEmergencyStop } from '../components/FloatingEmergencyStop';
import { AuthGuard } from '../components/auth/AuthGuard';
import { AuthPage } from '../components/auth/AuthPage';
import { TRPCProvider } from '../providers/TRPCProvider';
import { TeamSelector } from '../components/layout/TeamSelector';
import { CommandPalette } from '../components/CommandPalette';
import { useAppStore } from '../store';
import { useShortcutsStore } from '../store/shortcutsStore';
import { useCommandPaletteStore } from '../store/commandPaletteStore';
import { applyTheme } from '../config/themes';
import { useWorkspaceShortcuts } from '../hooks/useWorkspaceShortcuts';

export function AppLayout() {
  const navigate = useNavigate();
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const { theme, loadProviders, createNewConversation } = useAppStore();
  const { closeHelpModal, isHelpModalOpen } = useShortcutsStore();
  const { open: openCommandPalette } = useCommandPaletteStore();

  // Apply theme on mount and when it changes
  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  // Load providers on mount
  useEffect(() => {
    loadProviders();
  }, [loadProviders]);

  // Shortcut handlers
  const handleNewConversation = useCallback(async () => {
    const id = await createNewConversation();
    if (id) {
      navigate(`/workspace/${id}`);
    }
  }, [createNewConversation, navigate]);

  const handleNewWorker = useCallback(() => {
    navigate('/workers?action=new');
  }, [navigate]);

  const handleCloseModal = useCallback(() => {
    // Close help modal if open, otherwise close settings
    if (isHelpModalOpen) {
      closeHelpModal();
    } else {
      // Other modals can be closed via their own escape handlers
    }
  }, [isHelpModalOpen, closeHelpModal]);

  // Register global keyboard shortcuts
  useWorkspaceShortcuts({
    onNewConversation: handleNewConversation,
    onNewWorker: handleNewWorker,
    onClosePreview: handleCloseModal,
  });

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

          {/* Keyboard Shortcuts Help Modal */}
          <KeyboardShortcutsHelp />

          {/* Toast Notifications */}
          <ToastContainer />

          {/* Command Palette */}
          <CommandPalette />

          {/* Emergency Stop Button (floating) */}
          <FloatingEmergencyStop />
        </div>
      </AuthGuard>
    </TRPCProvider>
  );
}
