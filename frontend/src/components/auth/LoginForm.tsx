import { useState, useEffect, useCallback } from 'react';
import { loginUser, handleSSOCallback, getSSOCallbackParams } from '../../store/authStore';
import { SSOProviderList } from './SSOProviderList';
import { useSSOStore } from '../../store/ssoStore';
import { Loader2, AlertCircle } from 'lucide-react';

interface LoginFormProps {
  onSuccess?: () => void;
  onRegisterClick?: () => void;
}

export function LoginForm({ onSuccess, onRegisterClick }: LoginFormProps) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showSSO, setShowSSO] = useState(false);
  const [isHandlingSSOCallback, setIsHandlingSSOCallback] = useState(false);
  const [ssoCallbackError, setSSOCallbackError] = useState<string | null>(null);

  const { fetchLoginProviders, loginProviders, clearLoginProviders } = useSSOStore();

  // Handle SSO callback on mount
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);

    // Check for SSO error in URL
    const ssoError = params.get('sso');
    const errorMessage = params.get('message');
    if (ssoError === 'error' && errorMessage) {
      setSSOCallbackError(decodeURIComponent(errorMessage).replace(/_/g, ' '));
      window.history.replaceState({}, document.title, window.location.pathname);
      return;
    }

    // Check for SSO callback params (code + state)
    const callbackParams = getSSOCallbackParams();
    if (callbackParams) {
      handleSSOCallbackFlow(callbackParams.code, callbackParams.state);
    }
  }, []);

  const handleSSOCallbackFlow = async (code: string, state: string) => {
    setIsHandlingSSOCallback(true);
    setSSOCallbackError(null);

    try {
      await handleSSOCallback(code, state);
      onSuccess?.();
    } catch (err) {
      setSSOCallbackError(err instanceof Error ? err.message : 'SSO authentication failed');
    } finally {
      setIsHandlingSSOCallback(false);
    }
  };

  // Detect SSO providers when email domain changes
  const detectSSOProviders = useCallback(async (emailValue: string) => {
    if (!emailValue.includes('@')) {
      clearLoginProviders();
      return;
    }

    const domain = emailValue.split('@')[1];
    if (domain && domain.includes('.')) {
      await fetchLoginProviders(emailValue);
    }
  }, [fetchLoginProviders, clearLoginProviders]);

  // Debounce email change for SSO detection
  useEffect(() => {
    const timer = setTimeout(() => {
      if (email && !showSSO) {
        detectSSOProviders(email);
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [email, showSSO, detectSSOProviders]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await loginUser({ email, password });
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    setSSOCallbackError(null);
  };

  // Show loading state during SSO callback
  if (isHandlingSSOCallback) {
    return (
      <div className="w-full max-w-md mx-auto text-center py-8">
        <Loader2 className="w-8 h-8 animate-spin mx-auto text-primary mb-4" />
        <p className="text-editor-muted">Completing SSO authentication...</p>
      </div>
    );
  }

  if (showSSO) {
    return (
      <SSOProviderList
        onBack={() => setShowSSO(false)}
        onEmailLogin={() => setShowSSO(false)}
        initialEmail={email}
      />
    );
  }

  // Check if there are active SSO providers for the current email domain
  const hasActiveSSOProviders = loginProviders.some(p => p.enabled && p.state === 'active');

  return (
    <div className="w-full max-w-md mx-auto">
      <h2 className="text-2xl font-bold text-center mb-6">Sign In</h2>

      <form onSubmit={handleSubmit} className="space-y-4">
        {(error || ssoCallbackError) && (
          <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm flex items-start gap-2">
            <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
            <span>{error || ssoCallbackError}</span>
          </div>
        )}

        <div>
          <label htmlFor="email" className="block text-sm font-medium mb-1">
            Email
          </label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={handleEmailChange}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder="you@example.com"
            required
            disabled={loading}
          />
        </div>

        {/* Show SSO options if detected */}
        {hasActiveSSOProviders && (
          <div className="p-3 bg-primary/10 border border-primary/20 rounded-lg">
            <p className="text-sm text-primary mb-2">
              SSO is available for your organization
            </p>
            <button
              type="button"
              onClick={() => setShowSSO(true)}
              className="w-full py-2 px-4 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 transition-colors text-sm"
            >
              Continue with SSO
            </button>
          </div>
        )}

        <div>
          <label htmlFor="password" className="block text-sm font-medium mb-1">
            Password
          </label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder="••••••••"
            required
            disabled={loading}
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full py-2 px-4 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {loading ? 'Signing in...' : 'Sign In'}
        </button>

        <div className="relative my-4">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-editor-border" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-editor-surface px-2 text-editor-muted">Or</span>
          </div>
        </div>

        <button
          type="button"
          onClick={() => setShowSSO(true)}
          disabled={loading}
          className="w-full py-2 px-4 border border-editor-border text-editor-text rounded-lg font-medium hover:bg-editor-surface disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Sign in with SSO
        </button>
      </form>

      <p className="mt-4 text-center text-sm text-editor-muted">
        Don't have an account?{' '}
        <button
          type="button"
          onClick={onRegisterClick}
          className="text-primary hover:underline"
        >
          Sign up
        </button>
      </p>
    </div>
  );
}
