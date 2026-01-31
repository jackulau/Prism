import { create } from 'zustand';
import { ssoService } from '../services/sso';
import type {
  SSOProvider,
  SSOConfiguration,
  SSOStatus,
  SSOConnection,
  OrganizationSSOProviders,
  SSOTestResult,
} from '../types/sso';

interface SSOState {
  // Status
  status: SSOStatus | null;
  isLoadingStatus: boolean;
  statusError: string | null;

  // Connections for current user
  connections: SSOConnection[];
  isLoadingConnections: boolean;
  connectionsError: string | null;

  // Providers for login
  loginProviders: SSOProvider[];
  isLoadingLoginProviders: boolean;
  loginProvidersError: string | null;
  detectedOrganization: string | null;

  // Organization providers (admin view)
  organizationProviders: OrganizationSSOProviders | null;
  isLoadingOrgProviders: boolean;
  orgProvidersError: string | null;

  // Current configuration being edited
  currentConfiguration: SSOConfiguration | null;
  isLoadingConfiguration: boolean;
  configurationError: string | null;

  // Test results
  testResult: SSOTestResult | null;
  isTesting: boolean;
  testError: string | null;

  // SSO Login flow state
  ssoLoginInProgress: boolean;
  ssoLoginError: string | null;
  selectedProviderId: string | null;

  // Actions
  fetchStatus: () => Promise<void>;
  fetchConnections: () => Promise<void>;
  fetchLoginProviders: (email?: string, orgSlug?: string) => Promise<void>;
  fetchOrganizationProviders: (organizationId: string) => Promise<void>;
  fetchConfiguration: (organizationId: string, providerId: string) => Promise<void>;
  testConnection: (organizationId: string, providerId: string) => Promise<SSOTestResult | null>;
  initiateSSO: (params: { organization?: string; connectionId?: string; providerId?: string }) => Promise<void>;
  setSelectedProvider: (providerId: string | null) => void;
  setDetectedOrganization: (org: string | null) => void;
  clearLoginProviders: () => void;
  clearErrors: () => void;
  reset: () => void;
}

const initialState = {
  status: null,
  isLoadingStatus: false,
  statusError: null,
  connections: [],
  isLoadingConnections: false,
  connectionsError: null,
  loginProviders: [],
  isLoadingLoginProviders: false,
  loginProvidersError: null,
  detectedOrganization: null,
  organizationProviders: null,
  isLoadingOrgProviders: false,
  orgProvidersError: null,
  currentConfiguration: null,
  isLoadingConfiguration: false,
  configurationError: null,
  testResult: null,
  isTesting: false,
  testError: null,
  ssoLoginInProgress: false,
  ssoLoginError: null,
  selectedProviderId: null,
};

export const useSSOStore = create<SSOState>((set) => ({
  ...initialState,

  fetchStatus: async () => {
    set({ isLoadingStatus: true, statusError: null });
    try {
      const response = await ssoService.getStatus();
      if (response.error) {
        set({ statusError: response.error, isLoadingStatus: false });
      } else {
        set({ status: response.data || null, isLoadingStatus: false });
      }
    } catch (err) {
      set({
        statusError: err instanceof Error ? err.message : 'Failed to fetch SSO status',
        isLoadingStatus: false,
      });
    }
  },

  fetchConnections: async () => {
    set({ isLoadingConnections: true, connectionsError: null });
    try {
      const response = await ssoService.getConnections();
      if (response.error) {
        set({ connectionsError: response.error, isLoadingConnections: false });
      } else {
        set({
          connections: response.data?.connections || [],
          isLoadingConnections: false,
        });
      }
    } catch (err) {
      set({
        connectionsError: err instanceof Error ? err.message : 'Failed to fetch connections',
        isLoadingConnections: false,
      });
    }
  },

  fetchLoginProviders: async (email?: string, orgSlug?: string) => {
    set({ isLoadingLoginProviders: true, loginProvidersError: null });
    try {
      const response = await ssoService.getLoginProviders({ email, organizationSlug: orgSlug });
      if (response.error) {
        set({ loginProvidersError: response.error, isLoadingLoginProviders: false });
      } else {
        set({
          loginProviders: response.data?.providers || [],
          isLoadingLoginProviders: false,
        });
      }
    } catch (err) {
      set({
        loginProvidersError: err instanceof Error ? err.message : 'Failed to fetch login providers',
        isLoadingLoginProviders: false,
      });
    }
  },

  fetchOrganizationProviders: async (organizationId: string) => {
    set({ isLoadingOrgProviders: true, orgProvidersError: null });
    try {
      const response = await ssoService.getOrganizationProviders(organizationId);
      if (response.error) {
        set({ orgProvidersError: response.error, isLoadingOrgProviders: false });
      } else {
        set({
          organizationProviders: response.data || null,
          isLoadingOrgProviders: false,
        });
      }
    } catch (err) {
      set({
        orgProvidersError: err instanceof Error ? err.message : 'Failed to fetch organization providers',
        isLoadingOrgProviders: false,
      });
    }
  },

  fetchConfiguration: async (organizationId: string, providerId: string) => {
    set({ isLoadingConfiguration: true, configurationError: null });
    try {
      const response = await ssoService.getConfiguration(organizationId, providerId);
      if (response.error) {
        set({ configurationError: response.error, isLoadingConfiguration: false });
      } else {
        set({
          currentConfiguration: response.data || null,
          isLoadingConfiguration: false,
        });
      }
    } catch (err) {
      set({
        configurationError: err instanceof Error ? err.message : 'Failed to fetch configuration',
        isLoadingConfiguration: false,
      });
    }
  },

  testConnection: async (organizationId: string, providerId: string) => {
    set({ isTesting: true, testError: null, testResult: null });
    try {
      const response = await ssoService.testConnection(organizationId, providerId);
      if (response.error) {
        set({ testError: response.error, isTesting: false });
        return null;
      }
      set({ testResult: response.data || null, isTesting: false });
      return response.data || null;
    } catch (err) {
      set({
        testError: err instanceof Error ? err.message : 'Connection test failed',
        isTesting: false,
      });
      return null;
    }
  },

  initiateSSO: async (params) => {
    set({ ssoLoginInProgress: true, ssoLoginError: null });
    try {
      const response = await ssoService.authorize(params);
      if (response.error) {
        set({ ssoLoginError: response.error, ssoLoginInProgress: false });
        throw new Error(response.error);
      }
      if (response.data?.authorizationUrl) {
        window.location.href = response.data.authorizationUrl;
      } else {
        set({ ssoLoginError: 'No authorization URL received', ssoLoginInProgress: false });
        throw new Error('No authorization URL received');
      }
    } catch (err) {
      const error = err instanceof Error ? err.message : 'SSO initiation failed';
      set({ ssoLoginError: error, ssoLoginInProgress: false });
      throw err;
    }
  },

  setSelectedProvider: (providerId) => {
    set({ selectedProviderId: providerId });
  },

  setDetectedOrganization: (org) => {
    set({ detectedOrganization: org });
  },

  clearLoginProviders: () => {
    set({ loginProviders: [], detectedOrganization: null, loginProvidersError: null });
  },

  clearErrors: () => {
    set({
      statusError: null,
      connectionsError: null,
      loginProvidersError: null,
      orgProvidersError: null,
      configurationError: null,
      testError: null,
      ssoLoginError: null,
    });
  },

  reset: () => set(initialState),
}));
