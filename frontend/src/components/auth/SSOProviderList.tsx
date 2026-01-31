import { useState, useEffect } from 'react';
import { ArrowLeft, Building2, Key, Mail, Loader2, Shield, AlertCircle } from 'lucide-react';
import { useSSOStore } from '../../store/ssoStore';
import { ssoService } from '../../services/sso';
import { useAuthStore } from '../../store/authStore';
import type { SSOProvider } from '../../types/sso';

interface SSOProviderListProps {
  onBack?: () => void;
  onEmailLogin?: () => void;
  initialEmail?: string;
}

const PROVIDER_ICONS: Record<string, React.ReactNode> = {
  saml: <Shield size={20} />,
  oidc: <Key size={20} />,
  oauth2: <Building2 size={20} />,
};

const PROVIDER_COLORS: Record<string, string> = {
  saml: 'bg-blue-500/10 border-blue-500/30 hover:bg-blue-500/20',
  oidc: 'bg-purple-500/10 border-purple-500/30 hover:bg-purple-500/20',
  oauth2: 'bg-green-500/10 border-green-500/30 hover:bg-green-500/20',
};

export function SSOProviderList({ onBack, onEmailLogin, initialEmail }: SSOProviderListProps) {
  const [email, setEmail] = useState(initialEmail || '');
  const [isDetecting, setIsDetecting] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState<SSOProvider | null>(null);
  const { accessToken } = useAuthStore();

  const {
    loginProviders,
    isLoadingLoginProviders,
    loginProvidersError,
    detectedOrganization,
    ssoLoginInProgress,
    ssoLoginError,
    fetchLoginProviders,
    initiateSSO,
    setDetectedOrganization,
    clearLoginProviders,
    clearErrors,
  } = useSSOStore();

  // Set token on SSO service when available
  useEffect(() => {
    if (accessToken) {
      ssoService.setToken(accessToken);
    }
  }, [accessToken]);

  // Detect organization by email domain
  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newEmail = e.target.value;
    setEmail(newEmail);

    // Clear previous results when email changes
    if (loginProviders.length > 0) {
      clearLoginProviders();
    }
  };

  const handleDetectOrganization = async () => {
    if (!email.includes('@')) return;

    const domain = email.split('@')[1];
    if (!domain) return;

    setIsDetecting(true);
    clearErrors();

    try {
      await fetchLoginProviders(email);
      setDetectedOrganization(domain);
    } finally {
      setIsDetecting(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && email.includes('@')) {
      handleDetectOrganization();
    }
  };

  const handleProviderClick = async (provider: SSOProvider) => {
    setSelectedProvider(provider);
    clearErrors();

    try {
      await initiateSSO({
        providerId: provider.id,
        connectionId: provider.connectionId,
      });
    } catch {
      setSelectedProvider(null);
    }
  };

  const handleBack = () => {
    clearLoginProviders();
    clearErrors();
    onBack?.();
  };

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="flex items-center gap-2 mb-6">
        {onBack && (
          <button
            type="button"
            onClick={handleBack}
            className="p-2 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
            aria-label="Back to login"
          >
            <ArrowLeft size={20} />
          </button>
        )}
        <h2 className="text-2xl font-bold">Sign in with SSO</h2>
      </div>

      <div className="space-y-4">
        {/* Email input for organization detection */}
        {loginProviders.length === 0 && (
          <div className="space-y-3">
            <div>
              <label htmlFor="sso-email" className="block text-sm font-medium mb-1">
                Work Email
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-editor-muted" />
                <input
                  id="sso-email"
                  type="email"
                  value={email}
                  onChange={handleEmailChange}
                  onKeyDown={handleKeyDown}
                  className="w-full pl-10 pr-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="you@company.com"
                  disabled={isDetecting || isLoadingLoginProviders}
                />
              </div>
              <p className="mt-1 text-xs text-editor-muted">
                Enter your work email to find your organization&apos;s SSO options
              </p>
            </div>

            <button
              type="button"
              onClick={handleDetectOrganization}
              disabled={!email.includes('@') || isDetecting || isLoadingLoginProviders}
              className="w-full py-2 px-4 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
            >
              {isDetecting || isLoadingLoginProviders ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Finding your organization...
                </>
              ) : (
                'Continue'
              )}
            </button>
          </div>
        )}

        {/* Error message */}
        {(loginProvidersError || ssoLoginError) && (
          <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm flex items-start gap-2">
            <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
            <span>{loginProvidersError || ssoLoginError}</span>
          </div>
        )}

        {/* Provider list */}
        {loginProviders.length > 0 && (
          <div className="space-y-3">
            {detectedOrganization && (
              <p className="text-sm text-editor-muted">
                SSO options for <span className="font-medium text-editor-text">{detectedOrganization}</span>
              </p>
            )}

            <div className="space-y-2">
              {loginProviders
                .filter(p => p.enabled && p.state === 'active')
                .sort((a, b) => a.priority - b.priority)
                .map((provider) => (
                  <button
                    key={provider.id}
                    type="button"
                    onClick={() => handleProviderClick(provider)}
                    disabled={ssoLoginInProgress}
                    className={`w-full p-3 border rounded-lg flex items-center gap-3 transition-colors ${
                      PROVIDER_COLORS[provider.type] || 'bg-editor-surface border-editor-border hover:bg-editor-surface/80'
                    } ${ssoLoginInProgress && selectedProvider?.id === provider.id ? 'opacity-70' : ''}`}
                  >
                    <div className="w-10 h-10 rounded-lg bg-editor-bg flex items-center justify-center text-editor-muted">
                      {provider.iconUrl ? (
                        <img
                          src={provider.iconUrl}
                          alt={provider.displayName || provider.name}
                          className="w-6 h-6"
                        />
                      ) : (
                        PROVIDER_ICONS[provider.type] || <Building2 size={20} />
                      )}
                    </div>
                    <div className="flex-1 text-left">
                      <p className="font-medium">
                        {provider.displayName || provider.name}
                      </p>
                      <p className="text-xs text-editor-muted capitalize">
                        {provider.type.toUpperCase()}
                      </p>
                    </div>
                    {ssoLoginInProgress && selectedProvider?.id === provider.id && (
                      <Loader2 className="w-4 h-4 animate-spin text-editor-muted" />
                    )}
                  </button>
                ))}
            </div>

            {/* Divider */}
            {onEmailLogin && (
              <>
                <div className="relative my-4">
                  <div className="absolute inset-0 flex items-center">
                    <div className="w-full border-t border-editor-border" />
                  </div>
                  <div className="relative flex justify-center text-xs uppercase">
                    <span className="bg-editor-bg px-2 text-editor-muted">Or</span>
                  </div>
                </div>

                <button
                  type="button"
                  onClick={onEmailLogin}
                  disabled={ssoLoginInProgress}
                  className="w-full py-2 px-4 border border-editor-border text-editor-text rounded-lg font-medium hover:bg-editor-surface disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
                >
                  <Mail className="w-4 h-4" />
                  Sign in with email instead
                </button>
              </>
            )}

            {/* Back to email detection */}
            <button
              type="button"
              onClick={() => clearLoginProviders()}
              className="w-full py-2 px-4 text-editor-muted hover:text-editor-text text-sm transition-colors"
            >
              Use a different email
            </button>
          </div>
        )}

        {/* No providers found */}
        {detectedOrganization && loginProviders.length === 0 && !isLoadingLoginProviders && !loginProvidersError && (
          <div className="text-center py-4">
            <Building2 className="w-12 h-12 mx-auto text-editor-muted mb-3" />
            <p className="text-editor-muted">
              No SSO providers configured for this organization
            </p>
            {onEmailLogin && (
              <button
                type="button"
                onClick={onEmailLogin}
                className="mt-3 text-primary hover:underline text-sm"
              >
                Sign in with email instead
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
