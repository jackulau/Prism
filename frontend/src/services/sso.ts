import type {
  SSOProvider,
  SSOConfiguration,
  SSOStatus,
  SSOConnection,
  OrganizationSSOProviders,
  SSOTestResult,
  SSOAuthorizationResponse,
  SSOProviderType,
  SAMLConfiguration,
  OIDCConfiguration,
  OAuth2Configuration,
  AttributeMapping,
} from '../types/sso';

const API_BASE_URL = '/api/v1';

interface ApiResponse<T> {
  data?: T;
  error?: string;
}

class SSOService {
  private token: string | null = null;

  setToken(token: string | null) {
    this.token = token;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers,
      });

      const contentType = response.headers.get('Content-Type');
      const hasJsonContent = contentType?.includes('application/json');

      let data: T | undefined;
      if (hasJsonContent) {
        const text = await response.text();
        if (text) {
          try {
            data = JSON.parse(text);
          } catch {
            if (!response.ok) {
              return { error: text || 'An error occurred' };
            }
          }
        }
      }

      if (!response.ok) {
        return { error: (data as { error?: string })?.error || 'An error occurred' };
      }

      return { data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }

  // Get SSO status for current user
  async getStatus(): Promise<ApiResponse<SSOStatus>> {
    return this.request<SSOStatus>('/auth/sso/status');
  }

  // Get SSO connections for user's organization
  async getConnections(): Promise<ApiResponse<{ connections: SSOConnection[] }>> {
    return this.request<{ connections: SSOConnection[] }>('/auth/sso/connections');
  }

  // Get SSO providers for an organization
  async getOrganizationProviders(
    organizationId: string
  ): Promise<ApiResponse<OrganizationSSOProviders>> {
    return this.request<OrganizationSSOProviders>(
      `/organizations/${organizationId}/sso/providers`
    );
  }

  // Get providers available for login (by email domain or org slug)
  async getLoginProviders(params: {
    email?: string;
    organizationSlug?: string;
  }): Promise<ApiResponse<{ providers: SSOProvider[] }>> {
    const query = new URLSearchParams();
    if (params.email) query.set('email', params.email);
    if (params.organizationSlug) query.set('org', params.organizationSlug);
    return this.request<{ providers: SSOProvider[] }>(
      `/auth/sso/providers?${query.toString()}`
    );
  }

  // Initiate SSO authorization
  async authorize(params: {
    organization?: string;
    connectionId?: string;
    providerId?: string;
  }): Promise<ApiResponse<SSOAuthorizationResponse>> {
    const query = new URLSearchParams();
    if (params.organization) query.set('organization', params.organization);
    if (params.connectionId) query.set('connection_id', params.connectionId);
    if (params.providerId) query.set('provider_id', params.providerId);
    return this.request<SSOAuthorizationResponse>(
      `/auth/sso/authorize?${query.toString()}`
    );
  }

  // Create SSO provider
  async createProvider(
    organizationId: string,
    provider: {
      name: string;
      type: SSOProviderType;
      displayName?: string;
      iconUrl?: string;
      priority?: number;
    }
  ): Promise<ApiResponse<SSOProvider>> {
    return this.request<SSOProvider>(
      `/organizations/${organizationId}/sso/providers`,
      {
        method: 'POST',
        body: JSON.stringify(provider),
      }
    );
  }

  // Update SSO provider
  async updateProvider(
    organizationId: string,
    providerId: string,
    updates: Partial<{
      name: string;
      displayName: string;
      iconUrl: string;
      priority: number;
      enabled: boolean;
    }>
  ): Promise<ApiResponse<SSOProvider>> {
    return this.request<SSOProvider>(
      `/organizations/${organizationId}/sso/providers/${providerId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(updates),
      }
    );
  }

  // Delete SSO provider
  async deleteProvider(
    organizationId: string,
    providerId: string
  ): Promise<ApiResponse<void>> {
    return this.request<void>(
      `/organizations/${organizationId}/sso/providers/${providerId}`,
      { method: 'DELETE' }
    );
  }

  // Get SSO configuration for a provider
  async getConfiguration(
    organizationId: string,
    providerId: string
  ): Promise<ApiResponse<SSOConfiguration>> {
    return this.request<SSOConfiguration>(
      `/organizations/${organizationId}/sso/providers/${providerId}/config`
    );
  }

  // Update SAML configuration
  async updateSAMLConfig(
    organizationId: string,
    providerId: string,
    config: SAMLConfiguration
  ): Promise<ApiResponse<SSOConfiguration>> {
    return this.request<SSOConfiguration>(
      `/organizations/${organizationId}/sso/providers/${providerId}/config/saml`,
      {
        method: 'PUT',
        body: JSON.stringify(config),
      }
    );
  }

  // Update OIDC configuration
  async updateOIDCConfig(
    organizationId: string,
    providerId: string,
    config: OIDCConfiguration
  ): Promise<ApiResponse<SSOConfiguration>> {
    return this.request<SSOConfiguration>(
      `/organizations/${organizationId}/sso/providers/${providerId}/config/oidc`,
      {
        method: 'PUT',
        body: JSON.stringify(config),
      }
    );
  }

  // Update OAuth2 configuration
  async updateOAuth2Config(
    organizationId: string,
    providerId: string,
    config: OAuth2Configuration
  ): Promise<ApiResponse<SSOConfiguration>> {
    return this.request<SSOConfiguration>(
      `/organizations/${organizationId}/sso/providers/${providerId}/config/oauth2`,
      {
        method: 'PUT',
        body: JSON.stringify(config),
      }
    );
  }

  // Update attribute mapping
  async updateAttributeMapping(
    organizationId: string,
    providerId: string,
    mapping: AttributeMapping
  ): Promise<ApiResponse<SSOConfiguration>> {
    return this.request<SSOConfiguration>(
      `/organizations/${organizationId}/sso/providers/${providerId}/config/mapping`,
      {
        method: 'PUT',
        body: JSON.stringify(mapping),
      }
    );
  }

  // Test SSO connection
  async testConnection(
    organizationId: string,
    providerId: string
  ): Promise<ApiResponse<SSOTestResult>> {
    return this.request<SSOTestResult>(
      `/organizations/${organizationId}/sso/providers/${providerId}/test`,
      { method: 'POST' }
    );
  }

  // Update organization SSO settings
  async updateOrganizationSSOSettings(
    organizationId: string,
    settings: {
      enforceSSO?: boolean;
      allowPasswordLogin?: boolean;
      defaultProviderId?: string;
    }
  ): Promise<ApiResponse<OrganizationSSOProviders>> {
    return this.request<OrganizationSSOProviders>(
      `/organizations/${organizationId}/sso/settings`,
      {
        method: 'PATCH',
        body: JSON.stringify(settings),
      }
    );
  }

  // Get metadata/well-known for SAML SP
  async getSAMLMetadata(organizationId: string): Promise<ApiResponse<string>> {
    return this.request<string>(
      `/organizations/${organizationId}/sso/saml/metadata`
    );
  }
}

export const ssoService = new SSOService();
