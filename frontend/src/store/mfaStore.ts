import { create } from 'zustand';
import { apiService } from '../services/api';

interface MFASetupData {
  secret: string;
  qrCodeUrl: string;
}

interface MFAState {
  isEnabled: boolean;
  isLoading: boolean;
  setupData: MFASetupData | null;
  backupCodes: string[] | null;
  error: string | null;

  // Actions
  setToken: (token: string | null) => void;
  fetchStatus: () => Promise<void>;
  startSetup: () => Promise<void>;
  verifySetup: (code: string) => Promise<boolean>;
  disable: (password: string, code: string) => Promise<boolean>;
  regenerateBackupCodes: (code: string) => Promise<void>;
  validateLogin: (sessionToken: string, code: string) => Promise<{
    access_token: string;
    refresh_token: string;
    user: { id: string; email: string; created_at: string };
  } | null>;
  validateBackupCode: (sessionToken: string, code: string) => Promise<{
    access_token: string;
    refresh_token: string;
    user: { id: string; email: string; created_at: string };
  } | null>;
  clearSetupData: () => void;
  clearBackupCodes: () => void;
  clearError: () => void;
}

export const useMFAStore = create<MFAState>((set) => ({
  isEnabled: false,
  isLoading: false,
  setupData: null,
  backupCodes: null,
  error: null,

  setToken: (token) => {
    apiService.setToken(token);
  },

  fetchStatus: async () => {
    set({ isLoading: true, error: null });
    const response = await apiService.mfaGetStatus();
    if (response.data) {
      set({ isEnabled: response.data.enabled, isLoading: false });
    } else {
      set({ error: response.error || 'Failed to fetch MFA status', isLoading: false });
    }
  },

  startSetup: async () => {
    set({ isLoading: true, error: null, setupData: null });
    const response = await apiService.mfaStartSetup();
    if (response.data) {
      set({
        setupData: {
          secret: response.data.secret,
          qrCodeUrl: response.data.qr_url,
        },
        isLoading: false,
      });
    } else {
      set({ error: response.error || 'Failed to start MFA setup', isLoading: false });
    }
  },

  verifySetup: async (code: string) => {
    set({ isLoading: true, error: null });
    const response = await apiService.mfaVerifySetup(code);
    if (response.data) {
      set({
        isEnabled: true,
        setupData: null,
        backupCodes: response.data.backup_codes,
        isLoading: false,
      });
      return true;
    } else {
      set({ error: response.error || 'Invalid verification code', isLoading: false });
      return false;
    }
  },

  disable: async (password: string, code: string) => {
    set({ isLoading: true, error: null });
    const response = await apiService.mfaDisable(password, code);
    if (!response.error) {
      set({ isEnabled: false, isLoading: false });
      return true;
    } else {
      set({ error: response.error || 'Failed to disable MFA', isLoading: false });
      return false;
    }
  },

  regenerateBackupCodes: async (code: string) => {
    set({ isLoading: true, error: null, backupCodes: null });
    const response = await apiService.mfaRegenerateBackupCodes(code);
    if (response.data) {
      set({ backupCodes: response.data.backup_codes, isLoading: false });
    } else {
      set({ error: response.error || 'Failed to regenerate backup codes', isLoading: false });
    }
  },

  validateLogin: async (sessionToken: string, code: string) => {
    set({ isLoading: true, error: null });
    const response = await apiService.mfaValidate(sessionToken, code);
    if (response.data) {
      set({ isLoading: false });
      return response.data;
    } else {
      set({ error: response.error || 'Invalid verification code', isLoading: false });
      return null;
    }
  },

  validateBackupCode: async (sessionToken: string, code: string) => {
    set({ isLoading: true, error: null });
    const response = await apiService.mfaVerifyBackupCode(sessionToken, code);
    if (response.data) {
      set({ isLoading: false });
      return response.data;
    } else {
      set({ error: response.error || 'Invalid backup code', isLoading: false });
      return null;
    }
  },

  clearSetupData: () => set({ setupData: null }),
  clearBackupCodes: () => set({ backupCodes: null }),
  clearError: () => set({ error: null }),
}));
