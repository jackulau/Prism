import { trpc } from '../lib/trpc';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const t = trpc as any;

// User Profile
export const useProfile = () => {
  return t.organization.getProfile.useQuery();
};

export const useChangePassword = () => {
  return t.organization.changePassword.useMutation();
};

export const useChangeEmail = () => {
  return t.organization.changeEmail.useMutation();
};

// User Settings
export const useSettings = () => {
  return t.organization.getSettings.useQuery();
};

export const useUpdateSettings = () => {
  const utils = t.useUtils();
  return t.organization.updateSettings.useMutation({
    onSuccess: () => {
      utils.organization.getSettings.invalidate();
    },
  });
};

// Provider Keys
export const useProviderKeys = () => {
  return t.organization.listProviderKeys.useQuery();
};

export const useSetProviderKey = () => {
  const utils = t.useUtils();
  return t.organization.setProviderKey.useMutation({
    onSuccess: () => {
      utils.organization.listProviderKeys.invalidate();
    },
  });
};

export const useDeleteProviderKey = () => {
  const utils = t.useUtils();
  return t.organization.deleteProviderKey.useMutation({
    onSuccess: () => {
      utils.organization.listProviderKeys.invalidate();
    },
  });
};

export const useValidateProviderKey = () => {
  return t.organization.validateProviderKey.useMutation();
};

// GitHub Connection
export const useGitHubConnection = () => {
  return t.organization.getGitHubConnection.useQuery();
};

export const useDisconnectGitHub = () => {
  const utils = t.useUtils();
  return t.organization.disconnectGitHub.useMutation({
    onSuccess: () => {
      utils.organization.getGitHubConnection.invalidate();
    },
  });
};

// Delete Account
export const useDeleteAccount = () => {
  return t.organization.deleteAccount.useMutation();
};
