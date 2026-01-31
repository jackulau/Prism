import { useState, useEffect } from 'react';
import {
  ArrowLeft,
  Loader2,
  CheckCircle,
  XCircle,
  AlertCircle,
  Plus,
  Trash2,
  Play,
} from 'lucide-react';
import { ssoService } from '../../services/sso';
import { useSSOStore } from '../../store/ssoStore';
import type {
  SSOProvider,
  SSOProviderType,
  SSOConfiguration,
  SAMLConfiguration,
  OIDCConfiguration,
  OAuth2Configuration,
  AttributeMapping,
} from '../../types/sso';

interface SSOProviderFormProps {
  organizationId: string;
  provider?: SSOProvider;
  onSave: () => void;
  onCancel: () => void;
}

type FormStep = 'basic' | 'config' | 'mapping';

const PROVIDER_TYPES: { value: SSOProviderType; label: string; description: string }[] = [
  { value: 'saml', label: 'SAML 2.0', description: 'Enterprise identity providers like Okta, Azure AD, OneLogin' },
  { value: 'oidc', label: 'OpenID Connect', description: 'Modern identity providers with OAuth 2.0 + OIDC' },
  { value: 'oauth2', label: 'OAuth 2.0', description: 'Generic OAuth 2.0 providers' },
];

const DEFAULT_ATTRIBUTE_MAPPING: AttributeMapping = {
  email: 'email',
  firstName: 'given_name',
  lastName: 'family_name',
  displayName: 'name',
  groups: 'groups',
};

export function SSOProviderForm({
  organizationId,
  provider,
  onSave,
  onCancel,
}: SSOProviderFormProps) {
  const [step, setStep] = useState<FormStep>('basic');
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Basic info
  const [name, setName] = useState(provider?.name || '');
  const [displayName, setDisplayName] = useState(provider?.displayName || '');
  const [providerType, setProviderType] = useState<SSOProviderType>(provider?.type || 'saml');
  const [priority, setPriority] = useState(provider?.priority || 0);

  // Configuration (used internally by loadConfiguration)
  const [, setConfiguration] = useState<SSOConfiguration | null>(null);

  // SAML Config
  const [samlMetadataUrl, setSamlMetadataUrl] = useState('');
  const [samlEntityId, setSamlEntityId] = useState('');
  const [samlAcsUrl, setSamlAcsUrl] = useState('');
  const [samlCertificate, setSamlCertificate] = useState('');
  const [samlSignRequests, setSamlSignRequests] = useState(false);

  // OIDC Config
  const [oidcIssuer, setOidcIssuer] = useState('');
  const [oidcClientId, setOidcClientId] = useState('');
  const [oidcClientSecret, setOidcClientSecret] = useState('');
  const [oidcScopes, setOidcScopes] = useState('openid profile email');

  // OAuth2 Config
  const [oauth2AuthEndpoint, setOauth2AuthEndpoint] = useState('');
  const [oauth2TokenEndpoint, setOauth2TokenEndpoint] = useState('');
  const [oauth2UserinfoEndpoint, setOauth2UserinfoEndpoint] = useState('');
  const [oauth2ClientId, setOauth2ClientId] = useState('');
  const [oauth2ClientSecret, setOauth2ClientSecret] = useState('');
  const [oauth2Scopes, setOauth2Scopes] = useState('');
  const [oauth2PkceEnabled, setOauth2PkceEnabled] = useState(true);

  // Attribute Mapping
  const [attributeMapping, setAttributeMapping] = useState<AttributeMapping>(DEFAULT_ATTRIBUTE_MAPPING);
  const [customAttributes, setCustomAttributes] = useState<Array<{ key: string; value: string }>>([]);
  const [jitProvisioning, setJitProvisioning] = useState(true);
  const [defaultRole, setDefaultRole] = useState('member');

  // Test results
  const { testResult, isTesting, testError, testConnection } = useSSOStore();

  useEffect(() => {
    if (provider) {
      loadConfiguration();
    }
  }, [provider]);

  const loadConfiguration = async () => {
    if (!provider) return;

    setIsLoading(true);
    try {
      const response = await ssoService.getConfiguration(organizationId, provider.id);
      if (response.data) {
        setConfiguration(response.data);
        populateConfigurationFields(response.data);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load configuration');
    } finally {
      setIsLoading(false);
    }
  };

  const populateConfigurationFields = (config: SSOConfiguration) => {
    setAttributeMapping(config.attributeMapping);
    setJitProvisioning(config.jitProvisioning);
    setDefaultRole(config.defaultRole || 'member');

    if (config.saml) {
      setSamlMetadataUrl(config.saml.metadataUrl || '');
      setSamlEntityId(config.saml.entityId);
      setSamlAcsUrl(config.saml.acsUrl);
      setSamlCertificate(config.saml.certificate);
      setSamlSignRequests(config.saml.signRequests);
    }

    if (config.oidc) {
      setOidcIssuer(config.oidc.issuer);
      setOidcClientId(config.oidc.clientId);
      setOidcClientSecret(config.oidc.clientSecret || '');
      setOidcScopes(config.oidc.scopes.join(' '));
    }

    if (config.oauth2) {
      setOauth2AuthEndpoint(config.oauth2.authorizationEndpoint);
      setOauth2TokenEndpoint(config.oauth2.tokenEndpoint);
      setOauth2UserinfoEndpoint(config.oauth2.userinfoEndpoint || '');
      setOauth2ClientId(config.oauth2.clientId);
      setOauth2ClientSecret(config.oauth2.clientSecret || '');
      setOauth2Scopes(config.oauth2.scopes.join(' '));
      setOauth2PkceEnabled(config.oauth2.pkceEnabled);
    }

    if (config.attributeMapping.customAttributes) {
      setCustomAttributes(
        Object.entries(config.attributeMapping.customAttributes).map(([key, value]) => ({
          key,
          value,
        }))
      );
    }
  };

  const handleSaveBasic = async () => {
    if (!name.trim()) {
      setError('Provider name is required');
      return;
    }

    setIsSaving(true);
    setError(null);

    try {
      if (provider) {
        const response = await ssoService.updateProvider(organizationId, provider.id, {
          name,
          displayName: displayName || undefined,
          priority,
        });
        if (response.error) {
          setError(response.error);
          return;
        }
      } else {
        const response = await ssoService.createProvider(organizationId, {
          name,
          type: providerType,
          displayName: displayName || undefined,
          priority,
        });
        if (response.error) {
          setError(response.error);
          return;
        }
      }

      if (provider) {
        setStep('config');
      } else {
        onSave();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save provider');
    } finally {
      setIsSaving(false);
    }
  };

  const handleSaveConfig = async () => {
    if (!provider) return;

    setIsSaving(true);
    setError(null);

    try {
      let response;

      switch (providerType) {
        case 'saml': {
          const samlConfig: SAMLConfiguration = {
            metadataUrl: samlMetadataUrl || undefined,
            entityId: samlEntityId,
            acsUrl: samlAcsUrl,
            certificate: samlCertificate,
            signRequests: samlSignRequests,
            signatureAlgorithm: 'SHA256',
          };
          response = await ssoService.updateSAMLConfig(organizationId, provider.id, samlConfig);
          break;
        }
        case 'oidc': {
          const oidcConfig: OIDCConfiguration = {
            issuer: oidcIssuer,
            clientId: oidcClientId,
            clientSecret: oidcClientSecret || undefined,
            scopes: oidcScopes.split(/\s+/).filter(Boolean),
            responseType: 'code',
          };
          response = await ssoService.updateOIDCConfig(organizationId, provider.id, oidcConfig);
          break;
        }
        case 'oauth2': {
          const oauth2Config: OAuth2Configuration = {
            authorizationEndpoint: oauth2AuthEndpoint,
            tokenEndpoint: oauth2TokenEndpoint,
            userinfoEndpoint: oauth2UserinfoEndpoint || undefined,
            clientId: oauth2ClientId,
            clientSecret: oauth2ClientSecret || undefined,
            scopes: oauth2Scopes.split(/\s+/).filter(Boolean),
            pkceEnabled: oauth2PkceEnabled,
          };
          response = await ssoService.updateOAuth2Config(organizationId, provider.id, oauth2Config);
          break;
        }
      }

      if (response?.error) {
        setError(response.error);
        return;
      }

      setStep('mapping');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save configuration');
    } finally {
      setIsSaving(false);
    }
  };

  const handleSaveMapping = async () => {
    if (!provider) return;

    setIsSaving(true);
    setError(null);

    try {
      const mapping: AttributeMapping = {
        ...attributeMapping,
        customAttributes:
          customAttributes.length > 0
            ? Object.fromEntries(customAttributes.map((a) => [a.key, a.value]))
            : undefined,
      };

      const response = await ssoService.updateAttributeMapping(organizationId, provider.id, mapping);

      if (response?.error) {
        setError(response.error);
        return;
      }

      onSave();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save attribute mapping');
    } finally {
      setIsSaving(false);
    }
  };

  const handleTestConnection = async () => {
    if (!provider) return;
    await testConnection(organizationId, provider.id);
  };

  const addCustomAttribute = () => {
    setCustomAttributes([...customAttributes, { key: '', value: '' }]);
  };

  const removeCustomAttribute = (index: number) => {
    setCustomAttributes(customAttributes.filter((_, i) => i !== index));
  };

  const updateCustomAttribute = (index: number, field: 'key' | 'value', value: string) => {
    const updated = [...customAttributes];
    updated[index][field] = value;
    setCustomAttributes(updated);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="w-8 h-8 animate-spin text-editor-muted" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <button
          onClick={onCancel}
          className="p-2 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
        >
          <ArrowLeft size={20} />
        </button>
        <div>
          <h2 className="text-xl font-semibold">
            {provider ? 'Configure SSO Provider' : 'Add SSO Provider'}
          </h2>
          <p className="text-sm text-editor-muted">
            {step === 'basic' && 'Basic information'}
            {step === 'config' && 'Protocol configuration'}
            {step === 'mapping' && 'Attribute mapping'}
          </p>
        </div>
      </div>

      {/* Steps indicator */}
      {provider && (
        <div className="flex items-center gap-2">
          {(['basic', 'config', 'mapping'] as FormStep[]).map((s, i) => (
            <div key={s} className="flex items-center">
              <button
                onClick={() => setStep(s)}
                className={`px-3 py-1 rounded-full text-sm ${
                  step === s
                    ? 'bg-primary text-white'
                    : 'bg-editor-surface text-editor-muted hover:text-editor-text'
                }`}
              >
                {i + 1}. {s.charAt(0).toUpperCase() + s.slice(1)}
              </button>
              {i < 2 && <div className="w-8 h-px bg-editor-border mx-2" />}
            </div>
          ))}
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm flex items-start gap-2">
          <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Basic Step */}
      {step === 'basic' && (
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Provider Name *</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Company Okta"
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Display Name</label>
            <input
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="e.g., Sign in with Okta"
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            />
            <p className="mt-1 text-xs text-editor-muted">
              Shown on the login button
            </p>
          </div>

          {!provider && (
            <div>
              <label className="block text-sm font-medium mb-2">Provider Type *</label>
              <div className="space-y-2">
                {PROVIDER_TYPES.map((type) => (
                  <label
                    key={type.value}
                    className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${
                      providerType === type.value
                        ? 'bg-primary/10 border-primary'
                        : 'bg-editor-surface border-editor-border hover:border-editor-muted'
                    }`}
                  >
                    <input
                      type="radio"
                      name="providerType"
                      value={type.value}
                      checked={providerType === type.value}
                      onChange={() => setProviderType(type.value)}
                      className="mt-1"
                    />
                    <div>
                      <p className="font-medium">{type.label}</p>
                      <p className="text-sm text-editor-muted">{type.description}</p>
                    </div>
                  </label>
                ))}
              </div>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium mb-1">Priority</label>
            <input
              type="number"
              value={priority}
              onChange={(e) => setPriority(parseInt(e.target.value) || 0)}
              min={0}
              className="w-24 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            />
            <p className="mt-1 text-xs text-editor-muted">
              Lower numbers appear first on the login page
            </p>
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <button
              onClick={onCancel}
              className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleSaveBasic}
              disabled={isSaving || !name.trim()}
              className="px-4 py-2 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              {isSaving && <Loader2 className="w-4 h-4 animate-spin" />}
              {provider ? 'Next' : 'Create Provider'}
            </button>
          </div>
        </div>
      )}

      {/* Config Step */}
      {step === 'config' && provider && (
        <div className="space-y-4">
          {providerType === 'saml' && (
            <>
              <div>
                <label className="block text-sm font-medium mb-1">Metadata URL</label>
                <input
                  type="url"
                  value={samlMetadataUrl}
                  onChange={(e) => setSamlMetadataUrl(e.target.value)}
                  placeholder="https://idp.example.com/metadata.xml"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
                <p className="mt-1 text-xs text-editor-muted">
                  URL to fetch IdP metadata automatically
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Entity ID *</label>
                <input
                  type="text"
                  value={samlEntityId}
                  onChange={(e) => setSamlEntityId(e.target.value)}
                  placeholder="https://idp.example.com"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">ACS URL *</label>
                <input
                  type="url"
                  value={samlAcsUrl}
                  onChange={(e) => setSamlAcsUrl(e.target.value)}
                  placeholder="https://your-app.com/api/v1/auth/saml/acs"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
                <p className="mt-1 text-xs text-editor-muted">
                  Assertion Consumer Service URL
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">X.509 Certificate *</label>
                <textarea
                  value={samlCertificate}
                  onChange={(e) => setSamlCertificate(e.target.value)}
                  placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                  rows={4}
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary font-mono text-sm"
                />
              </div>

              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={samlSignRequests}
                  onChange={(e) => setSamlSignRequests(e.target.checked)}
                  className="rounded"
                />
                <span className="text-sm">Sign authentication requests</span>
              </label>
            </>
          )}

          {providerType === 'oidc' && (
            <>
              <div>
                <label className="block text-sm font-medium mb-1">Issuer URL *</label>
                <input
                  type="url"
                  value={oidcIssuer}
                  onChange={(e) => setOidcIssuer(e.target.value)}
                  placeholder="https://accounts.google.com"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
                <p className="mt-1 text-xs text-editor-muted">
                  The OpenID Connect issuer URL
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Client ID *</label>
                <input
                  type="text"
                  value={oidcClientId}
                  onChange={(e) => setOidcClientId(e.target.value)}
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Client Secret</label>
                <input
                  type="password"
                  value={oidcClientSecret}
                  onChange={(e) => setOidcClientSecret(e.target.value)}
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Scopes</label>
                <input
                  type="text"
                  value={oidcScopes}
                  onChange={(e) => setOidcScopes(e.target.value)}
                  placeholder="openid profile email"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
                <p className="mt-1 text-xs text-editor-muted">Space-separated list of scopes</p>
              </div>
            </>
          )}

          {providerType === 'oauth2' && (
            <>
              <div>
                <label className="block text-sm font-medium mb-1">Authorization Endpoint *</label>
                <input
                  type="url"
                  value={oauth2AuthEndpoint}
                  onChange={(e) => setOauth2AuthEndpoint(e.target.value)}
                  placeholder="https://provider.com/oauth/authorize"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Token Endpoint *</label>
                <input
                  type="url"
                  value={oauth2TokenEndpoint}
                  onChange={(e) => setOauth2TokenEndpoint(e.target.value)}
                  placeholder="https://provider.com/oauth/token"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Userinfo Endpoint</label>
                <input
                  type="url"
                  value={oauth2UserinfoEndpoint}
                  onChange={(e) => setOauth2UserinfoEndpoint(e.target.value)}
                  placeholder="https://provider.com/oauth/userinfo"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Client ID *</label>
                <input
                  type="text"
                  value={oauth2ClientId}
                  onChange={(e) => setOauth2ClientId(e.target.value)}
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Client Secret</label>
                <input
                  type="password"
                  value={oauth2ClientSecret}
                  onChange={(e) => setOauth2ClientSecret(e.target.value)}
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Scopes</label>
                <input
                  type="text"
                  value={oauth2Scopes}
                  onChange={(e) => setOauth2Scopes(e.target.value)}
                  placeholder="user:email"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={oauth2PkceEnabled}
                  onChange={(e) => setOauth2PkceEnabled(e.target.checked)}
                  className="rounded"
                />
                <span className="text-sm">Enable PKCE (Recommended)</span>
              </label>
            </>
          )}

          {/* Test Connection */}
          <div className="pt-4 border-t border-editor-border">
            <button
              onClick={handleTestConnection}
              disabled={isTesting}
              className="px-4 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm font-medium hover:bg-editor-bg transition-colors flex items-center gap-2"
            >
              {isTesting ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Play className="w-4 h-4" />
              )}
              Test Connection
            </button>

            {testResult && (
              <div
                className={`mt-3 p-3 rounded-lg flex items-start gap-2 ${
                  testResult.success
                    ? 'bg-green-500/10 border border-green-500/20 text-green-400'
                    : 'bg-red-500/10 border border-red-500/20 text-red-400'
                }`}
              >
                {testResult.success ? (
                  <CheckCircle className="w-4 h-4 mt-0.5" />
                ) : (
                  <XCircle className="w-4 h-4 mt-0.5" />
                )}
                <span className="text-sm">{testResult.message}</span>
              </div>
            )}

            {testError && (
              <div className="mt-3 p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">
                {testError}
              </div>
            )}
          </div>

          <div className="flex justify-between gap-3 pt-4">
            <button
              onClick={() => setStep('basic')}
              className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
            >
              Back
            </button>
            <button
              onClick={handleSaveConfig}
              disabled={isSaving}
              className="px-4 py-2 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              {isSaving && <Loader2 className="w-4 h-4 animate-spin" />}
              Next
            </button>
          </div>
        </div>
      )}

      {/* Mapping Step */}
      {step === 'mapping' && provider && (
        <div className="space-y-4">
          <div className="bg-editor-bg border border-editor-border rounded-lg p-4 space-y-3">
            <h3 className="text-sm font-medium text-editor-muted">Standard Attributes</h3>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs text-editor-muted mb-1">Email</label>
                <input
                  type="text"
                  value={attributeMapping.email}
                  onChange={(e) =>
                    setAttributeMapping({ ...attributeMapping, email: e.target.value })
                  }
                  placeholder="email"
                  className="w-full px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-xs text-editor-muted mb-1">Display Name</label>
                <input
                  type="text"
                  value={attributeMapping.displayName || ''}
                  onChange={(e) =>
                    setAttributeMapping({ ...attributeMapping, displayName: e.target.value })
                  }
                  placeholder="name"
                  className="w-full px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-xs text-editor-muted mb-1">First Name</label>
                <input
                  type="text"
                  value={attributeMapping.firstName || ''}
                  onChange={(e) =>
                    setAttributeMapping({ ...attributeMapping, firstName: e.target.value })
                  }
                  placeholder="given_name"
                  className="w-full px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-xs text-editor-muted mb-1">Last Name</label>
                <input
                  type="text"
                  value={attributeMapping.lastName || ''}
                  onChange={(e) =>
                    setAttributeMapping({ ...attributeMapping, lastName: e.target.value })
                  }
                  placeholder="family_name"
                  className="w-full px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-xs text-editor-muted mb-1">Groups</label>
                <input
                  type="text"
                  value={attributeMapping.groups || ''}
                  onChange={(e) =>
                    setAttributeMapping({ ...attributeMapping, groups: e.target.value })
                  }
                  placeholder="groups"
                  className="w-full px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-xs text-editor-muted mb-1">Role</label>
                <input
                  type="text"
                  value={attributeMapping.role || ''}
                  onChange={(e) =>
                    setAttributeMapping({ ...attributeMapping, role: e.target.value })
                  }
                  placeholder="role"
                  className="w-full px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                />
              </div>
            </div>
          </div>

          {/* Custom Attributes */}
          <div className="bg-editor-bg border border-editor-border rounded-lg p-4 space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-medium text-editor-muted">Custom Attributes</h3>
              <button
                onClick={addCustomAttribute}
                className="text-sm text-primary hover:underline flex items-center gap-1"
              >
                <Plus size={14} />
                Add
              </button>
            </div>

            {customAttributes.map((attr, index) => (
              <div key={index} className="flex items-center gap-2">
                <input
                  type="text"
                  value={attr.key}
                  onChange={(e) => updateCustomAttribute(index, 'key', e.target.value)}
                  placeholder="Attribute key"
                  className="flex-1 px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                />
                <span className="text-editor-muted">=</span>
                <input
                  type="text"
                  value={attr.value}
                  onChange={(e) => updateCustomAttribute(index, 'value', e.target.value)}
                  placeholder="IdP claim name"
                  className="flex-1 px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                />
                <button
                  onClick={() => removeCustomAttribute(index)}
                  className="p-1 text-editor-muted hover:text-red-400"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            ))}
          </div>

          {/* Provisioning Settings */}
          <div className="bg-editor-bg border border-editor-border rounded-lg p-4 space-y-3">
            <h3 className="text-sm font-medium text-editor-muted">Provisioning</h3>

            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={jitProvisioning}
                onChange={(e) => setJitProvisioning(e.target.checked)}
                className="rounded"
              />
              <span className="text-sm">Just-in-time provisioning</span>
            </label>
            <p className="text-xs text-editor-muted ml-5">
              Automatically create user accounts on first SSO login
            </p>

            <div className="pt-2">
              <label className="block text-sm mb-1">Default Role</label>
              <select
                value={defaultRole}
                onChange={(e) => setDefaultRole(e.target.value)}
                className="px-3 py-1.5 bg-editor-surface border border-editor-border rounded text-sm focus:outline-none focus:ring-1 focus:ring-primary"
              >
                <option value="member">Member</option>
                <option value="admin">Admin</option>
                <option value="viewer">Viewer</option>
              </select>
            </div>
          </div>

          <div className="flex justify-between gap-3 pt-4">
            <button
              onClick={() => setStep('config')}
              className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
            >
              Back
            </button>
            <button
              onClick={handleSaveMapping}
              disabled={isSaving}
              className="px-4 py-2 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              {isSaving && <Loader2 className="w-4 h-4 animate-spin" />}
              Save Configuration
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
