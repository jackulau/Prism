import { create } from 'zustand';
import { apiService } from '../services/api';

export interface APIKey {
  id: string;
  name: string;
  prefix: string;
  scopes: string[];
  created_at: string;
  expires_at: string | null;
  last_used_at: string | null;
}

export interface ProviderKeyMetadata {
  provider: string;
  is_active: boolean;
  created_at: string;
  last_used_at: string | null;
  use_count: number;
}

interface APIKeysState {
  // Data
  keys: APIKey[];
  providerKeys: ProviderKeyMetadata[];
  newKeyValue: string | null;

  // Loading states
  isLoading: boolean;
  isCreating: boolean;
  isDeleting: string | null;
  isRotating: string | null;
  isRenaming: string | null;
  error: string | null;

  // Actions
  setToken: (token: string | null) => void;
  fetchKeys: () => Promise<void>;
  fetchProviderKeys: () => Promise<void>;
  createKey: (name: string, expiresInDays?: number, scopes?: string[]) => Promise<boolean>;
  deleteKey: (id: string) => Promise<boolean>;
  rotateKey: (id: string) => Promise<boolean>;
  renameKey: (id: string, name: string) => Promise<boolean>;
  clearNewKey: () => void;
  clearError: () => void;
}

export const useAPIKeysStore = create<APIKeysState>((set) => ({
  // Initial state
  keys: [],
  providerKeys: [],
  newKeyValue: null,
  isLoading: false,
  isCreating: false,
  isDeleting: null,
  isRotating: null,
  isRenaming: null,
  error: null,

  setToken: (token) => {
    apiService.setToken(token);
  },

  fetchKeys: async () => {
    set({ isLoading: true, error: null });
    const response = await apiService.listAPIKeys();
    if (response.data) {
      set({
        keys: response.data.api_keys.map((k) => ({
          id: k.id,
          name: k.name,
          prefix: k.prefix,
          scopes: k.scopes || [],
          created_at: k.created_at,
          expires_at: k.expires_at,
          last_used_at: k.last_used_at,
        })),
        isLoading: false,
      });
    } else {
      set({ error: response.error || 'Failed to fetch API keys', isLoading: false });
    }
  },

  fetchProviderKeys: async () => {
    const response = await apiService.listProviderKeyMetadata();
    if (response.data) {
      set({
        providerKeys: response.data.provider_keys || [],
      });
    }
  },

  createKey: async (name, expiresInDays, scopes) => {
    set({ isCreating: true, error: null });
    const response = await apiService.createAPIKey(name, expiresInDays, scopes);
    if (response.data) {
      const newKey: APIKey = {
        id: response.data.id,
        name: response.data.name,
        prefix: response.data.prefix,
        scopes: response.data.scopes || [],
        created_at: response.data.created_at,
        expires_at: response.data.expires_at,
        last_used_at: response.data.last_used_at,
      };
      set((state) => ({
        keys: [newKey, ...state.keys],
        newKeyValue: response.data!.key,
        isCreating: false,
      }));
      return true;
    } else {
      set({ error: response.error || 'Failed to create API key', isCreating: false });
      return false;
    }
  },

  deleteKey: async (id) => {
    set({ isDeleting: id, error: null });
    const response = await apiService.deleteAPIKey(id);
    if (response.data?.success) {
      set((state) => ({
        keys: state.keys.filter((k) => k.id !== id),
        isDeleting: null,
      }));
      return true;
    } else {
      set({ error: response.error || 'Failed to delete API key', isDeleting: null });
      return false;
    }
  },

  rotateKey: async (id) => {
    set({ isRotating: id, error: null });
    const response = await apiService.rotateAPIKey(id);
    if (response.data) {
      const rotatedKey: APIKey = {
        id: response.data.id,
        name: response.data.name,
        prefix: response.data.prefix,
        scopes: response.data.scopes || [],
        created_at: response.data.created_at,
        expires_at: response.data.expires_at,
        last_used_at: response.data.last_used_at,
      };
      set((state) => ({
        keys: state.keys.map((k) => (k.id === id ? rotatedKey : k)).filter((k) => k.id !== id).concat([rotatedKey]),
        newKeyValue: response.data!.key,
        isRotating: null,
      }));
      // Re-sort keys to put rotated at top
      set((state) => ({
        keys: [rotatedKey, ...state.keys.filter((k) => k.id !== rotatedKey.id)],
      }));
      return true;
    } else {
      set({ error: response.error || 'Failed to rotate API key', isRotating: null });
      return false;
    }
  },

  renameKey: async (id, name) => {
    set({ isRenaming: id, error: null });
    const response = await apiService.updateAPIKeyName(id, name);
    if (response.data?.success) {
      set((state) => ({
        keys: state.keys.map((k) => (k.id === id ? { ...k, name } : k)),
        isRenaming: null,
      }));
      return true;
    } else {
      set({ error: response.error || 'Failed to rename API key', isRenaming: null });
      return false;
    }
  },

  clearNewKey: () => {
    set({ newKeyValue: null });
  },

  clearError: () => {
    set({ error: null });
  },
}));
