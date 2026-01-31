import { useState, useEffect } from 'react';
import {
  Shield,
  Plus,
  Trash2,
  Settings,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2,
  GripVertical,
  ToggleLeft,
  ToggleRight,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { useOrganizationStore } from '../../store/organizationStore';
import { useSSOStore } from '../../store/ssoStore';
import { ssoService } from '../../services/sso';
import { useAuthStore } from '../../store/authStore';
import { SSOProviderForm } from './SSOProviderForm';
import type { SSOProvider, SSOProviderType } from '../../types/sso';

const PROVIDER_TYPE_LABELS: Record<SSOProviderType, string> = {
  saml: 'SAML 2.0',
  oidc: 'OpenID Connect',
  oauth2: 'OAuth 2.0',
};

const PROVIDER_STATE_COLORS: Record<string, string> = {
  active: 'text-green-500',
  inactive: 'text-editor-muted',
  pending: 'text-yellow-500',
  error: 'text-red-500',
};

export function SSOSettings() {
  const { currentOrg } = useOrganizationStore();
  const { accessToken } = useAuthStore();
  const {
    organizationProviders,
    isLoadingOrgProviders,
    orgProvidersError,
    fetchOrganizationProviders,
  } = useSSOStore();

  const [isAddingProvider, setIsAddingProvider] = useState(false);
  const [editingProvider, setEditingProvider] = useState<SSOProvider | null>(null);
  const [expandedProvider, setExpandedProvider] = useState<string | null>(null);
  const [isUpdating, setIsUpdating] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (accessToken) {
      ssoService.setToken(accessToken);
    }
  }, [accessToken]);

  useEffect(() => {
    if (currentOrg?.id) {
      fetchOrganizationProviders(currentOrg.id);
    }
  }, [currentOrg?.id, fetchOrganizationProviders]);

  const handleToggleProvider = async (provider: SSOProvider) => {
    if (!currentOrg?.id) return;

    setIsUpdating(provider.id);
    setError(null);

    try {
      const response = await ssoService.updateProvider(currentOrg.id, provider.id, {
        enabled: !provider.enabled,
      });

      if (response.error) {
        setError(response.error);
      } else {
        await fetchOrganizationProviders(currentOrg.id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update provider');
    } finally {
      setIsUpdating(null);
    }
  };

  const handleDeleteProvider = async (providerId: string) => {
    if (!currentOrg?.id) return;
    if (!confirm('Are you sure you want to delete this SSO provider? Users will no longer be able to sign in with it.')) {
      return;
    }

    setIsUpdating(providerId);
    setError(null);

    try {
      const response = await ssoService.deleteProvider(currentOrg.id, providerId);

      if (response.error) {
        setError(response.error);
      } else {
        await fetchOrganizationProviders(currentOrg.id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete provider');
    } finally {
      setIsUpdating(null);
    }
  };

  const handleEnforceSSO = async (enforce: boolean) => {
    if (!currentOrg?.id) return;

    setError(null);
    try {
      const response = await ssoService.updateOrganizationSSOSettings(currentOrg.id, {
        enforceSSO: enforce,
      });

      if (response.error) {
        setError(response.error);
      } else {
        await fetchOrganizationProviders(currentOrg.id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update settings');
    }
  };

  const handleAllowPassword = async (allow: boolean) => {
    if (!currentOrg?.id) return;

    setError(null);
    try {
      const response = await ssoService.updateOrganizationSSOSettings(currentOrg.id, {
        allowPasswordLogin: allow,
      });

      if (response.error) {
        setError(response.error);
      } else {
        await fetchOrganizationProviders(currentOrg.id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update settings');
    }
  };

  const handleProviderSaved = () => {
    setIsAddingProvider(false);
    setEditingProvider(null);
    if (currentOrg?.id) {
      fetchOrganizationProviders(currentOrg.id);
    }
  };

  if (!currentOrg) {
    return (
      <div className="text-center py-8">
        <Shield className="w-12 h-12 mx-auto text-editor-muted mb-3" />
        <p className="text-editor-muted">
          Select an organization to manage SSO settings
        </p>
      </div>
    );
  }

  if (isAddingProvider) {
    return (
      <SSOProviderForm
        organizationId={currentOrg.id}
        onSave={handleProviderSaved}
        onCancel={() => setIsAddingProvider(false)}
      />
    );
  }

  if (editingProvider) {
    return (
      <SSOProviderForm
        organizationId={currentOrg.id}
        provider={editingProvider}
        onSave={handleProviderSaved}
        onCancel={() => setEditingProvider(null)}
      />
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Shield className="w-5 h-5" />
          <h2 className="text-xl font-semibold">Single Sign-On (SSO)</h2>
        </div>
        <button
          onClick={() => setIsAddingProvider(true)}
          className="px-3 py-1.5 bg-primary text-white rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors flex items-center gap-1.5"
        >
          <Plus size={16} />
          Add Provider
        </button>
      </div>

      {/* Error */}
      {(error || orgProvidersError) && (
        <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm flex items-start gap-2">
          <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
          <span>{error || orgProvidersError}</span>
        </div>
      )}

      {/* Organization SSO Settings */}
      {organizationProviders && (
        <div className="bg-editor-bg border border-editor-border rounded-lg p-4 space-y-4">
          <h3 className="text-sm font-medium text-editor-muted">Organization Settings</h3>

          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Enforce SSO</p>
                <p className="text-sm text-editor-muted">
                  Require all users to sign in with SSO
                </p>
              </div>
              <button
                onClick={() => handleEnforceSSO(!organizationProviders.enforceSSO)}
                className="text-editor-text"
              >
                {organizationProviders.enforceSSO ? (
                  <ToggleRight className="w-8 h-8 text-primary" />
                ) : (
                  <ToggleLeft className="w-8 h-8 text-editor-muted" />
                )}
              </button>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Allow Password Login</p>
                <p className="text-sm text-editor-muted">
                  Allow users to sign in with email and password
                </p>
              </div>
              <button
                onClick={() => handleAllowPassword(!organizationProviders.allowPasswordLogin)}
                className="text-editor-text"
              >
                {organizationProviders.allowPasswordLogin ? (
                  <ToggleRight className="w-8 h-8 text-primary" />
                ) : (
                  <ToggleLeft className="w-8 h-8 text-editor-muted" />
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Loading */}
      {isLoadingOrgProviders && (
        <div className="flex items-center justify-center py-8">
          <Loader2 className="w-6 h-6 animate-spin text-editor-muted" />
        </div>
      )}

      {/* Provider List */}
      {!isLoadingOrgProviders && organizationProviders && (
        <div className="space-y-2">
          {organizationProviders.providers.length === 0 ? (
            <div className="text-center py-8 bg-editor-surface border border-editor-border rounded-lg">
              <Shield className="w-12 h-12 mx-auto text-editor-muted mb-3" />
              <p className="text-editor-muted mb-4">No SSO providers configured</p>
              <button
                onClick={() => setIsAddingProvider(true)}
                className="px-4 py-2 bg-primary text-white rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors"
              >
                Add your first provider
              </button>
            </div>
          ) : (
            organizationProviders.providers
              .sort((a, b) => a.priority - b.priority)
              .map((provider) => (
                <div
                  key={provider.id}
                  className="bg-editor-surface border border-editor-border rounded-lg overflow-hidden"
                >
                  <div className="p-4 flex items-center gap-3">
                    <GripVertical className="w-4 h-4 text-editor-muted cursor-grab" />

                    <button
                      onClick={() => setExpandedProvider(
                        expandedProvider === provider.id ? null : provider.id
                      )}
                      className="text-editor-muted hover:text-editor-text"
                    >
                      {expandedProvider === provider.id ? (
                        <ChevronDown size={16} />
                      ) : (
                        <ChevronRight size={16} />
                      )}
                    </button>

                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">
                          {provider.displayName || provider.name}
                        </span>
                        <span className="px-1.5 py-0.5 text-xs rounded bg-editor-bg border border-editor-border">
                          {PROVIDER_TYPE_LABELS[provider.type]}
                        </span>
                        <span className={`flex items-center gap-1 text-xs ${PROVIDER_STATE_COLORS[provider.state]}`}>
                          {provider.state === 'active' ? (
                            <CheckCircle size={12} />
                          ) : provider.state === 'error' ? (
                            <XCircle size={12} />
                          ) : (
                            <AlertCircle size={12} />
                          )}
                          {provider.state}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => handleToggleProvider(provider)}
                        disabled={isUpdating === provider.id}
                        className="text-editor-text"
                        title={provider.enabled ? 'Disable' : 'Enable'}
                      >
                        {isUpdating === provider.id ? (
                          <Loader2 className="w-5 h-5 animate-spin text-editor-muted" />
                        ) : provider.enabled ? (
                          <ToggleRight className="w-6 h-6 text-primary" />
                        ) : (
                          <ToggleLeft className="w-6 h-6 text-editor-muted" />
                        )}
                      </button>

                      <button
                        onClick={() => setEditingProvider(provider)}
                        className="p-1.5 rounded hover:bg-editor-bg text-editor-muted hover:text-editor-text transition-colors"
                        title="Configure"
                      >
                        <Settings size={16} />
                      </button>

                      <button
                        onClick={() => handleDeleteProvider(provider.id)}
                        disabled={isUpdating === provider.id}
                        className="p-1.5 rounded hover:bg-red-500/10 text-editor-muted hover:text-red-400 transition-colors"
                        title="Delete"
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </div>

                  {expandedProvider === provider.id && (
                    <div className="px-4 pb-4 pt-0 border-t border-editor-border mt-0">
                      <div className="pt-4 grid grid-cols-2 gap-4 text-sm">
                        <div>
                          <span className="text-editor-muted">Type:</span>{' '}
                          <span>{PROVIDER_TYPE_LABELS[provider.type]}</span>
                        </div>
                        <div>
                          <span className="text-editor-muted">Priority:</span>{' '}
                          <span>{provider.priority}</span>
                        </div>
                        <div>
                          <span className="text-editor-muted">Created:</span>{' '}
                          <span>{new Date(provider.createdAt).toLocaleDateString()}</span>
                        </div>
                        {provider.updatedAt && (
                          <div>
                            <span className="text-editor-muted">Updated:</span>{' '}
                            <span>{new Date(provider.updatedAt).toLocaleDateString()}</span>
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              ))
          )}
        </div>
      )}
    </div>
  );
}
