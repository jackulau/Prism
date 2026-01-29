import { ReactNode, useEffect, useState } from 'react';
import { useAuthStore, initAuth, getSSOCallbackParams, handleSSOCallback } from '../../store/authStore';

interface AuthGuardProps {
  children: ReactNode;
  fallback?: ReactNode;
}

export function AuthGuard({ children, fallback }: AuthGuardProps) {
  const { isAuthenticated, isLoading } = useAuthStore();
  const [ssoError, setSsoError] = useState<string | null>(null);
  const [ssoLoading, setSsoLoading] = useState(false);

  useEffect(() => {
    const ssoParams = getSSOCallbackParams();
    if (ssoParams) {
      setSsoLoading(true);
      handleSSOCallback(ssoParams.code, ssoParams.state)
        .catch((err) => {
          setSsoError(err instanceof Error ? err.message : 'SSO authentication failed');
        })
        .finally(() => {
          setSsoLoading(false);
        });
    } else {
      initAuth();
    }
  }, []);

  if (ssoLoading || isLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-4">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        {ssoLoading && <p className="text-editor-muted">Completing SSO sign in...</p>}
      </div>
    );
  }

  if (ssoError) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-4 p-4">
        <div className="max-w-md w-full bg-editor-surface border border-editor-border rounded-xl p-6 text-center">
          <h2 className="text-xl font-bold text-red-400 mb-2">SSO Login Failed</h2>
          <p className="text-editor-muted mb-4">{ssoError}</p>
          <button
            onClick={() => {
              setSsoError(null);
              window.history.replaceState({}, document.title, window.location.pathname);
            }}
            className="px-4 py-2 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 transition-colors"
          >
            Back to Login
          </button>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}
