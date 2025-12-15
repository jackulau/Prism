import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { wsService } from '../services/websocket';
import { apiService } from '../services/api';

export interface User {
  id: string;
  email: string;
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

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
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

export const authApi = {
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
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

// Auth helper functions
export const loginUser = async (credentials: LoginCredentials) => {
  const { setUser, setTokens } = useAuthStore.getState();

  const response = await authApi.login(credentials);
  setTokens(response.access_token, response.refresh_token);
  setUser(response.user);

  // Connect services with token
  apiService.setToken(response.access_token);
  wsService.connect(response.access_token);

  return response;
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
