// SSO Provider Types - matches backend WorkOS/SAML/OIDC configurations

export type SSOProviderType = 'saml' | 'oidc' | 'oauth2';
export type SSOProviderState = 'active' | 'inactive' | 'pending' | 'error';

export interface SSOProvider {
  id: string;
  name: string;
  type: SSOProviderType;
  state: SSOProviderState;
  connectionId?: string;
  organizationId: string;
  displayName?: string;
  iconUrl?: string;
  priority: number;
  enabled: boolean;
  createdAt: string;
  updatedAt?: string;
}

export interface SAMLConfiguration {
  metadataUrl?: string;
  metadataXml?: string;
  entityId: string;
  acsUrl: string;
  sloUrl?: string;
  certificate: string;
  signRequests: boolean;
  signatureAlgorithm: 'SHA256' | 'SHA384' | 'SHA512';
}

export interface OIDCConfiguration {
  issuer: string;
  clientId: string;
  clientSecret?: string;
  authorizationEndpoint?: string;
  tokenEndpoint?: string;
  userinfoEndpoint?: string;
  jwksUri?: string;
  scopes: string[];
  responseType: 'code' | 'id_token' | 'code id_token';
}

export interface OAuth2Configuration {
  authorizationEndpoint: string;
  tokenEndpoint: string;
  userinfoEndpoint?: string;
  clientId: string;
  clientSecret?: string;
  scopes: string[];
  pkceEnabled: boolean;
}

export interface AttributeMapping {
  email: string;
  firstName?: string;
  lastName?: string;
  displayName?: string;
  groups?: string;
  role?: string;
  customAttributes?: Record<string, string>;
}

export interface SSOConfiguration {
  id: string;
  providerId: string;
  type: SSOProviderType;
  saml?: SAMLConfiguration;
  oidc?: OIDCConfiguration;
  oauth2?: OAuth2Configuration;
  attributeMapping: AttributeMapping;
  jitProvisioning: boolean;
  defaultRole?: string;
  allowedDomains?: string[];
  createdAt: string;
  updatedAt?: string;
}

export interface SSOStatus {
  enabled: boolean;
  connected: boolean;
  organizationId?: string;
  connectionType?: string;
}

export interface SSOConnection {
  id: string;
  name: string;
  connectionType: string;
  state: string;
}

export interface OrganizationSSOProviders {
  organizationId: string;
  providers: SSOProvider[];
  defaultProviderId?: string;
  enforceSSO: boolean;
  allowPasswordLogin: boolean;
}

export interface SSOTestResult {
  success: boolean;
  message: string;
  details?: {
    responseTime?: number;
    userAttributes?: Record<string, string>;
    errors?: string[];
  };
}

export interface SSOLoginRequest {
  email?: string;
  organizationSlug?: string;
  providerId?: string;
  connectionId?: string;
}

export interface SSOAuthorizationResponse {
  authorizationUrl: string;
  state: string;
}
