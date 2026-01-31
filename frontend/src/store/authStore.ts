import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { wsService } from '../services/websocket';
import { apiService } from '../services/api';
import { Role, Permission, hasPermission as checkPermission } from '../types/rbac';

// SSO callback state interface
export interface SSOCallbackState {
  code: string;
  state: string;
}

export interface User {
  id: string;
  email: string;
  role: Role;
  createdAt: string;
  githubUsername?: string;
  githubConnectedAt?: string;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  // Actions
  setUser: (user: User | null) => void;
  setTokens: (accessToken: string, refreshToken: string) => void;
  clearAuth: () => void;
  setLoading: (loading: boolean) => void;
}

// DEV BYPASS - Remove in production
const DEV_BYPASS = true;
const MOCK_USER: User = {
  id: 'dev-user-123',
  email: 'dev@prism.local',
  role: 'admin', // Dev user has admin role for testing
  createdAt: new Date().toISOString(),
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: DEV_BYPASS ? MOCK_USER : null,
      accessToken: DEV_BYPASS ? 'dev-token' : null,
      refreshToken: DEV_BYPASS ? 'dev-refresh' : null,
      isAuthenticated: DEV_BYPASS ? true : false,
      isLoading: false,

      setUser: (user) => set({
        user,
        isAuthenticated: !!user,
        isLoading: false,
      }),

      setTokens: (accessToken, refreshToken) => set({
        accessToken,
        refreshToken,
      }),

      clearAuth: () => set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        isLoading: false,
      }),

      setLoading: (isLoading) => set({ isLoading }),
    }),
    {
      name: 'prism-auth',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        user: state.user,
      }),
    }
  )
);

// Auth API functions
const API_BASE = '/api/v1';

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterCredentials {
  email: string;
  password: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface MFARequiredResponse {
  mfa_required: true;
  session_token: string;
}

export const authApi = {
  async login(credentials: LoginCredentials): Promise<AuthResponse | MFARequiredResponse> {
    const response = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Login failed');
    }

    return response.json();
  },

  async guestLogin(): Promise<AuthResponse> {
    const response = await fetch(`${API_BASE}/auth/guest`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Guest login failed');
    }

    return response.json();
  },

  async isGuestModeEnabled(): Promise<boolean> {
    try {
      const response = await fetch(`${API_BASE}/guest-mode`);
      if (response.ok) {
        const data = await response.json();
        return data.enabled === true;
      }
    } catch {
      // Guest mode check failed
    }
    return false;
  },

  async register(credentials: RegisterCredentials): Promise<AuthResponse> {
    const response = await fetch(`${API_BASE}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Registration failed');
    }

    return response.json();
  },

  async refreshTokens(refreshToken: string): Promise<AuthResponse> {
    const response = await fetch(`${API_BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      throw new Error('Token refresh failed');
    }

    return response.json();
  },

  async logout(accessToken: string): Promise<void> {
    await fetch(`${API_BASE}/auth/logout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${accessToken}`,
      },
    });
  },

  async getMe(accessToken: string): Promise<User> {
    const response = await fetch(`${API_BASE}/auth/me`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to get user');
    }

    return response.json();
  },
};

// Helper to check if response is MFA required
export function isMFARequired(response: AuthResponse | MFARequiredResponse): response is MFARequiredResponse {
  return 'mfa_required' in response && response.mfa_required === true;
}

// Auth helper functions
export const loginUser = async (credentials: LoginCredentials): Promise<AuthResponse | MFARequiredResponse> => {
  const { setUser, setTokens } = useAuthStore.getState();

  const response = await authApi.login(credentials);

  // Check if MFA is required
  if (isMFARequired(response)) {
    return response;
  }

  setTokens(response.access_token, response.refresh_token);
  setUser(response.user);

  // Connect services with token
  apiService.setToken(response.access_token);
  wsService.connect(response.access_token);

  return response;
};

// Complete login after MFA verification
export const completeMFALogin = async (authData: {
  access_token: string;
  refresh_token: string;
  user: { id: string; email: string; created_at: string };
}) => {
  const { setUser, setTokens } = useAuthStore.getState();

  setTokens(authData.access_token, authData.refresh_token);
  setUser({
    id: authData.user.id,
    email: authData.user.email,
    createdAt: authData.user.created_at,
  });

  // Connect services with token
  apiService.setToken(authData.access_token);
  wsService.connect(authData.access_token);
};

export const registerUser = async (credentials: RegisterCredentials) => {
  const { setUser, setTokens } = useAuthStore.getState();

  const response = await authApi.register(credentials);
  setTokens(response.access_token, response.refresh_token);
  setUser(response.user);

  // Connect services with token
  apiService.setToken(response.access_token);
  wsService.connect(response.access_token);

  return response;
};

// Track ongoing guest login to prevent concurrent attempts
let guestLoginPromise: Promise<AuthResponse> | null = null;

export const loginAsGuest = async () => {
  // If guest login is already in progress, wait for it
  if (guestLoginPromise) {
    return guestLoginPromise;
  }

  const { setUser, setTokens } = useAuthStore.getState();

  guestLoginPromise = (async () => {
    try {
      const response = await authApi.guestLogin();
      setTokens(response.access_token, response.refresh_token);
      setUser(response.user);

      // Connect services with token
      apiService.setToken(response.access_token);
      wsService.connect(response.access_token);

      return response;
    } finally {
      guestLoginPromise = null;
    }
  })();

  return guestLoginPromise;
};

export const logoutUser = async () => {
  const { accessToken, clearAuth } = useAuthStore.getState();

  // Disconnect WebSocket and clear API token
  wsService.disconnect();
  apiService.setToken(null);

  if (accessToken) {
    try {
      await authApi.logout(accessToken);
    } catch {
      // Logout API call failed - still clearing local auth
    }
  }

  clearAuth();
};

// Track ongoing refresh to prevent concurrent refresh attempts
let refreshPromise: Promise<boolean> | null = null;

export const refreshAuth = async (): Promise<boolean> => {
  // If refresh is already in progress, wait for it
  if (refreshPromise) {
    return refreshPromise;
  }

  const { refreshToken, setUser, setTokens, clearAuth } = useAuthStore.getState();

  if (!refreshToken) {
    clearAuth();
    return false;
  }

  // Create the refresh promise
  refreshPromise = (async () => {
    try {
      const response = await authApi.refreshTokens(refreshToken);
      setTokens(response.access_token, response.refresh_token);
      setUser(response.user);

      // Reconnect services with new token
      apiService.setToken(response.access_token);
      wsService.disconnect();
      wsService.connect(response.access_token);

      return true;
    } catch {
      // Token refresh failed - clearing auth
      wsService.disconnect();
      apiService.setToken(null);
      clearAuth();
      return false;
    } finally {
      // Clear the promise when done
      refreshPromise = null;
    }
  })();

  return refreshPromise;
};

// SSO Authentication
export const initiateSSO = async (organization: string): Promise<void> => {
  const response = await apiService.ssoAuthorize(organization);
  if (response.error) {
    throw new Error(response.error);
  }
  if (response.data?.authorization_url) {
    window.location.href = response.data.authorization_url;
  } else {
    throw new Error('No authorization URL received');
  }
};

export const handleSSOCallback = async (code: string, state: string): Promise<void> => {
  const { setUser, setTokens } = useAuthStore.getState();

  const response = await apiService.ssoCallback(code, state);
  if (response.error) {
    throw new Error(response.error);
  }
  if (!response.data) {
    throw new Error('Invalid response from SSO callback');
  }

  setTokens(response.data.access_token, response.data.refresh_token);
  setUser({
    id: response.data.user.id,
    email: response.data.user.email,
    createdAt: response.data.user.created_at,
  });

  // Connect services with token
  apiService.setToken(response.data.access_token);
  wsService.connect(response.data.access_token);

  // Clear URL params after successful SSO
  window.history.replaceState({}, document.title, window.location.pathname);
};

// Check for SSO callback params in URL
export const getSSOCallbackParams = (): SSOCallbackState | null => {
  const params = new URLSearchParams(window.location.search);
  const code = params.get('code');
  const state = params.get('state');

  if (code && state) {
    return { code, state };
  }
  return null;
};

// Track ongoing init to prevent concurrent attempts
let initAuthPromise: Promise<void> | null = null;

// Initialize auth on app load
export const initAuth = async () => {
  // If init is already in progress, wait for it
  if (initAuthPromise) {
    return initAuthPromise;
  }

  initAuthPromise = (async () => {
    try {
      const { accessToken, user, setUser, setLoading, clearAuth } = useAuthStore.getState();

      // If no token, check if guest mode is enabled and auto-login
      if (!accessToken) {
        const guestModeEnabled = await authApi.isGuestModeEnabled();
        if (guestModeEnabled) {
          try {
            await loginAsGuest();
            return;
          } catch {
            // Guest login failed, user will see login screen
          }
        }
        setLoading(false);
        return;
      }

      // If we already have user data from localStorage, skip loading state
      // and validate token in background
      if (user) {
        // Connect services with existing token
        apiService.setToken(accessToken);
        wsService.connect(accessToken);

        try {
          const freshUser = await authApi.getMe(accessToken);
          setUser(freshUser);
        } catch (e) {
          const refreshed = await refreshAuth();
          if (!refreshed) {
            wsService.disconnect();
            apiService.setToken(null);
            clearAuth();
          }
        }
        return;
      }

      // No cached user but have token - show loading while fetching
      setLoading(true);
      try {
        const fetchedUser = await authApi.getMe(accessToken);
        setUser(fetchedUser);

        // Connect services with validated token
        apiService.setToken(accessToken);
        wsService.connect(accessToken);
      } catch (e) {
        const refreshed = await refreshAuth();
        if (!refreshed) {
          wsService.disconnect();
          apiService.setToken(null);
          clearAuth();
        }
      }
      setLoading(false);
    } finally {
      initAuthPromise = null;
    }
  })();

  return initAuthPromise;
};

// Role and Permission helpers

/**
 * Check if the current user is an admin
 */
export const isAdmin = (): boolean => {
  const { user } = useAuthStore.getState();
  return user?.role === 'admin';
};

/**
 * Check if the current user has a specific permission
 */
export const hasPermission = (permission: Permission): boolean => {
  const { user } = useAuthStore.getState();
  return checkPermission(user?.role, permission);
};

/**
 * Get the current user's role
 */
export const getUserRole = (): Role | undefined => {
  const { user } = useAuthStore.getState();
  return user?.role;
};

/**
 * Hook to check if user is admin (for use in components)
 */
export const useIsAdmin = (): boolean => {
  const user = useAuthStore((state) => state.user);
  return user?.role === 'admin';
};

/**
 * Hook to check if user has permission (for use in components)
 */
export const useHasPermission = (permission: Permission): boolean => {
  const user = useAuthStore((state) => state.user);
  return checkPermission(user?.role, permission);
};

/**
 * Hook to get user role (for use in components)
 */
export const useUserRole = (): Role | undefined => {
  const user = useAuthStore((state) => state.user);
  return user?.role;
};
