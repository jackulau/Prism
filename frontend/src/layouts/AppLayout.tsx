import { useState, useEffect, useCallback } from 'react';
import { Outlet, useNavigate } from 'react-router-dom';
import { Sidebar } from '../components/sidebar/Sidebar';
import { SettingsPanel } from '../components/settings/SettingsPanel';
import { ToastContainer } from '../components/Toast';
import { AuthGuard } from '../components/auth/AuthGuard';
import { AuthPage } from '../components/auth/AuthPage';
import { TRPCProvider } from '../providers/TRPCProvider';
import { useAppStore } from '../store';
import { applyTheme } from '../config/themes';
import { useIdleDetection } from '../hooks/useIdleDetection';
import { useSessionStore } from '../store/sessionStore';
import { IdleWarningModal } from '../components/IdleWarningModal';
import { logoutUser, useAuthStore } from '../store/authStore';

// Session timeout configuration
// 25 minutes idle before warning (matching backend's 30 min timeout minus 5 min warning)
const IDLE_TIMEOUT_MS = 25 * 60 * 1000;
// 5 minute warning before actual logout
const WARNING_DURATION_MS = 5 * 60 * 1000;

function IdleDetectionWrapper({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const {
    idleWarningVisible,
    idleCountdown,
    showIdleWarning,
    hideIdleWarning,
    updateIdleCountdown,
    extendSession,
  } = useSessionStore();

  const handleIdle = useCallback(async () => {
    // User has been idle past the warning period, log them out
    await logoutUser();
    navigate('/login');
  }, [navigate]);

  const handleWarning = useCallback(() => {
    showIdleWarning(Math.ceil(WARNING_DURATION_MS / 1000));
  }, [showIdleWarning]);

  const handleActive = useCallback(() => {
    hideIdleWarning();
  }, [hideIdleWarning]);

  const { remainingTime, resetIdleTimer } = useIdleDetection({
    idleTimeout: IDLE_TIMEOUT_MS,
    warningDuration: WARNING_DURATION_MS,
    onIdle: handleIdle,
    onWarning: handleWarning,
    onActive: handleActive,
    enabled: isAuthenticated,
  });

  // Update countdown in store when it changes
  useEffect(() => {
    updateIdleCountdown(remainingTime);
  }, [remainingTime, updateIdleCountdown]);

  const handleStayLoggedIn = useCallback(async () => {
    resetIdleTimer();
    await extendSession();
  }, [resetIdleTimer, extendSession]);

  const handleLogoutNow = useCallback(async () => {
    await logoutUser();
    navigate('/login');
  }, [navigate]);

  return (
    <>
      {children}
      {idleWarningVisible && (
        <IdleWarningModal
          remainingTime={idleCountdown}
          onStayLoggedIn={handleStayLoggedIn}
          onLogout={handleLogoutNow}
        />
      )}
    </>
  );
}

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
        <IdleDetectionWrapper>
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
          </div>
        </IdleDetectionWrapper>
      </AuthGuard>
    </TRPCProvider>
  );
}
