import { trpc } from '../lib/trpc';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const t = trpc as any;

export const useCurrentWorkspace = () => {
  return t.workspace.getCurrent.useQuery();
};

export const useSetDirectory = () => {
  const utils = t.useUtils();
  return t.workspace.setDirectory.useMutation({
    onSuccess: () => {
      utils.workspace.getCurrent.invalidate();
      utils.workspace.listRecent.invalidate();
    },
  });
};

export const useBrowseDirectory = (path?: string, showHidden?: boolean) => {
  return t.workspace.browse.useQuery(
    { path, showHidden },
    { enabled: !!path }
  );
};

export const usePickFolder = () => {
  const utils = t.useUtils();
  return t.workspace.pickFolder.useMutation({
    onSuccess: () => {
      utils.workspace.getCurrent.invalidate();
      utils.workspace.listRecent.invalidate();
    },
  });
};

export const useRecentWorkspaces = (limit?: number) => {
  return t.workspace.listRecent.useQuery({ limit });
};

export const useSetCurrentWorkspace = () => {
  const utils = t.useUtils();
  return t.workspace.setCurrent.useMutation({
    onSuccess: () => {
      utils.workspace.getCurrent.invalidate();
    },
  });
};

export const useRemoveWorkspace = () => {
  const utils = t.useUtils();
  return t.workspace.remove.useMutation({
    onSuccess: () => {
      utils.workspace.listRecent.invalidate();
    },
  });
};

export const useWorkspaceById = (id?: string) => {
  return t.workspace.getById.useQuery(
    { id: id! },
    { enabled: !!id }
  );
};
