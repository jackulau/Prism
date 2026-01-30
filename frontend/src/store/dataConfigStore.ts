import { create } from 'zustand';
import { apiService } from '../services/api';

interface DataConfigState {
  // State
  configTypes: string[];
  selectedType: string | null;
  configKeys: string[];
  selectedKey: string | null;
  configData: Record<string, unknown> | null;
  configUpdatedAt: string | null;

  // Loading states
  typesLoading: boolean;
  keysLoading: boolean;
  dataLoading: boolean;
  saving: boolean;
  deleting: boolean;

  // Error state
  error: string | null;

  // Actions
  fetchConfigTypes: () => Promise<void>;
  fetchConfigKeys: (configType: string) => Promise<void>;
  fetchConfigData: (configType: string, configKey: string) => Promise<void>;
  setConfig: (
    configType: string,
    configKey: string,
    value: Record<string, unknown>
  ) => Promise<boolean>;
  deleteConfig: (configType: string, configKey: string) => Promise<boolean>;
  checkConfigExists: (configType: string, configKey: string) => Promise<boolean>;
  selectType: (configType: string | null) => void;
  selectKey: (configKey: string | null) => void;
  clearError: () => void;
  reset: () => void;
}

export const useDataConfigStore = create<DataConfigState>((set, get) => ({
  // Initial state
  configTypes: [],
  selectedType: null,
  configKeys: [],
  selectedKey: null,
  configData: null,
  configUpdatedAt: null,
  typesLoading: false,
  keysLoading: false,
  dataLoading: false,
  saving: false,
  deleting: false,
  error: null,

  fetchConfigTypes: async () => {
    set({ typesLoading: true, error: null });
    const response = await apiService.listDataConfigTypes();
    if (response.data) {
      set({
        configTypes: response.data.types || [],
        typesLoading: false,
      });
    } else {
      set({
        error: response.error || 'Failed to fetch config types',
        typesLoading: false,
      });
    }
  },

  fetchConfigKeys: async (configType: string) => {
    set({ keysLoading: true, error: null });
    const response = await apiService.listDataConfigs(configType);
    if (response.data) {
      set({
        configKeys: response.data.keys || [],
        keysLoading: false,
      });
    } else {
      set({
        error: response.error || 'Failed to fetch config keys',
        keysLoading: false,
      });
    }
  },

  fetchConfigData: async (configType: string, configKey: string) => {
    set({ dataLoading: true, error: null });
    const response = await apiService.getDataConfig(configType, configKey);
    if (response.data) {
      set({
        configData: response.data.data,
        configUpdatedAt: response.data.updated_at,
        dataLoading: false,
      });
    } else {
      set({
        error: response.error || 'Failed to fetch config data',
        dataLoading: false,
      });
    }
  },

  setConfig: async (configType, configKey, value) => {
    set({ saving: true, error: null });
    const response = await apiService.setDataConfig(configType, configKey, value);
    if (response.data?.success) {
      set({ saving: false });
      // Refresh the config types and keys lists
      const { selectedType } = get();
      await get().fetchConfigTypes();
      if (selectedType) {
        await get().fetchConfigKeys(selectedType);
      }
      return true;
    } else {
      set({
        error: response.error || 'Failed to save configuration',
        saving: false,
      });
      return false;
    }
  },

  deleteConfig: async (configType, configKey) => {
    set({ deleting: true, error: null });
    const response = await apiService.deleteDataConfig(configType, configKey);
    if (response.data?.success) {
      set({
        deleting: false,
        configData: null,
        configUpdatedAt: null,
        selectedKey: null,
      });
      // Refresh the lists
      const { selectedType } = get();
      await get().fetchConfigTypes();
      if (selectedType) {
        await get().fetchConfigKeys(selectedType);
      }
      return true;
    } else {
      set({
        error: response.error || 'Failed to delete configuration',
        deleting: false,
      });
      return false;
    }
  },

  checkConfigExists: async (configType, configKey) => {
    const response = await apiService.checkDataConfigExists(configType, configKey);
    return response.data?.exists ?? false;
  },

  selectType: (configType) => {
    set({
      selectedType: configType,
      selectedKey: null,
      configKeys: [],
      configData: null,
      configUpdatedAt: null,
    });
    if (configType) {
      get().fetchConfigKeys(configType);
    }
  },

  selectKey: (configKey) => {
    const { selectedType } = get();
    set({ selectedKey: configKey, configData: null, configUpdatedAt: null });
    if (selectedType && configKey) {
      get().fetchConfigData(selectedType, configKey);
    }
  },

  clearError: () => set({ error: null }),

  reset: () =>
    set({
      configTypes: [],
      selectedType: null,
      configKeys: [],
      selectedKey: null,
      configData: null,
      configUpdatedAt: null,
      typesLoading: false,
      keysLoading: false,
      dataLoading: false,
      saving: false,
      deleting: false,
      error: null,
    }),
}));
