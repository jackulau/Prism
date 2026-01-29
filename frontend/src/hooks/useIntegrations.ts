import { trpc } from '../lib/trpc';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const t = trpc as any;

export const useIntegrationStatus = () => {
  return t.integrations.getStatus.useQuery();
};

// Discord
export const useConfigureDiscord = () => {
  const utils = t.useUtils();
  return t.integrations.configureDiscord.useMutation({
    onSuccess: () => {
      utils.integrations.getStatus.invalidate();
    },
  });
};

export const useDisconnectDiscord = () => {
  const utils = t.useUtils();
  return t.integrations.disconnectDiscord.useMutation({
    onSuccess: () => {
      utils.integrations.getStatus.invalidate();
    },
  });
};

// Slack
export const useConfigureSlack = () => {
  const utils = t.useUtils();
  return t.integrations.configureSlack.useMutation({
    onSuccess: () => {
      utils.integrations.getStatus.invalidate();
    },
  });
};

export const useDisconnectSlack = () => {
  const utils = t.useUtils();
  return t.integrations.disconnectSlack.useMutation({
    onSuccess: () => {
      utils.integrations.getStatus.invalidate();
    },
  });
};

// PostHog
export const useConfigurePostHog = () => {
  const utils = t.useUtils();
  return t.integrations.configurePostHog.useMutation({
    onSuccess: () => {
      utils.integrations.getStatus.invalidate();
    },
  });
};

export const useDisconnectPostHog = () => {
  const utils = t.useUtils();
  return t.integrations.disconnectPostHog.useMutation({
    onSuccess: () => {
      utils.integrations.getStatus.invalidate();
    },
  });
};

// MCP Servers
export const useMcpServers = () => {
  return t.integrations.listMcpServers.useQuery();
};

export const useMcpServer = (id?: string) => {
  return t.integrations.getMcpServer.useQuery(
    { id: id! },
    { enabled: !!id }
  );
};

export const useAddMcpServer = () => {
  const utils = t.useUtils();
  return t.integrations.addMcpServer.useMutation({
    onSuccess: () => {
      utils.integrations.listMcpServers.invalidate();
      utils.integrations.getStatus.invalidate();
    },
  });
};

export const useUpdateMcpServer = () => {
  const utils = t.useUtils();
  return t.integrations.updateMcpServer.useMutation({
    onSuccess: () => {
      utils.integrations.listMcpServers.invalidate();
    },
  });
};

export const useRemoveMcpServer = () => {
  const utils = t.useUtils();
  return t.integrations.removeMcpServer.useMutation({
    onSuccess: () => {
      utils.integrations.listMcpServers.invalidate();
      utils.integrations.getStatus.invalidate();
    },
  });
};

export const useTestMcpServer = () => {
  return t.integrations.testMcpServer.useMutation();
};

export const useRefreshMcpServer = () => {
  const utils = t.useUtils();
  return t.integrations.refreshMcpServer.useMutation({
    onSuccess: () => {
      utils.integrations.listMcpServers.invalidate();
    },
  });
};

export const useEnableMcpServer = () => {
  const utils = t.useUtils();
  return t.integrations.enableMcpServer.useMutation({
    onSuccess: () => {
      utils.integrations.listMcpServers.invalidate();
    },
  });
};

export const useDisableMcpServer = () => {
  const utils = t.useUtils();
  return t.integrations.disableMcpServer.useMutation({
    onSuccess: () => {
      utils.integrations.listMcpServers.invalidate();
    },
  });
};

export const useMcpTools = () => {
  return t.integrations.listMcpTools.useQuery();
};

// MCP Stdio Servers
export const useStdioServers = () => {
  return t.integrations.listStdioServers.useQuery();
};

export const useStdioServer = (id?: string) => {
  return t.integrations.getStdioServer.useQuery(
    { id: id! },
    { enabled: !!id }
  );
};

export const useAddStdioServer = () => {
  const utils = t.useUtils();
  return t.integrations.addStdioServer.useMutation({
    onSuccess: () => {
      utils.integrations.listStdioServers.invalidate();
    },
  });
};

export const useUpdateStdioServer = () => {
  const utils = t.useUtils();
  return t.integrations.updateStdioServer.useMutation({
    onSuccess: () => {
      utils.integrations.listStdioServers.invalidate();
    },
  });
};

export const useRemoveStdioServer = () => {
  const utils = t.useUtils();
  return t.integrations.removeStdioServer.useMutation({
    onSuccess: () => {
      utils.integrations.listStdioServers.invalidate();
    },
  });
};

export const useStartStdioServer = () => {
  const utils = t.useUtils();
  return t.integrations.startStdioServer.useMutation({
    onSuccess: () => {
      utils.integrations.listStdioServers.invalidate();
    },
  });
};

export const useStopStdioServer = () => {
  const utils = t.useUtils();
  return t.integrations.stopStdioServer.useMutation({
    onSuccess: () => {
      utils.integrations.listStdioServers.invalidate();
    },
  });
};

export const useRestartStdioServer = () => {
  const utils = t.useUtils();
  return t.integrations.restartStdioServer.useMutation({
    onSuccess: () => {
      utils.integrations.listStdioServers.invalidate();
    },
  });
};
