import { trpc } from '../lib/trpc';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const t = trpc as any;

export const useRunTask = () => {
  const utils = t.useUtils();
  return t.workers.runTask.useMutation({
    onSuccess: () => {
      utils.workers.listExecutions.invalidate();
    },
  });
};

export const useRunParallel = () => {
  const utils = t.useUtils();
  return t.workers.runParallel.useMutation({
    onSuccess: () => {
      utils.workers.listExecutions.invalidate();
    },
  });
};

export const useRunSequential = () => {
  const utils = t.useUtils();
  return t.workers.runSequential.useMutation({
    onSuccess: () => {
      utils.workers.listExecutions.invalidate();
    },
  });
};

export const useRunSwarm = () => {
  const utils = t.useUtils();
  return t.workers.runSwarm.useMutation({
    onSuccess: () => {
      utils.workers.listSwarms.invalidate();
    },
  });
};

export const useExecution = (executionId?: string) => {
  return t.workers.getExecution.useQuery(
    { executionId: executionId! },
    { enabled: !!executionId, refetchInterval: 2000 }
  );
};

export const useExecutions = () => {
  return t.workers.listExecutions.useQuery();
};

export const useCancelExecution = () => {
  const utils = t.useUtils();
  return t.workers.cancelExecution.useMutation({
    onSuccess: () => {
      utils.workers.listExecutions.invalidate();
    },
  });
};

export const useSwarm = (swarmId?: string) => {
  return t.workers.getSwarm.useQuery(
    { swarmId: swarmId! },
    { enabled: !!swarmId, refetchInterval: 2000 }
  );
};

export const useSwarms = () => {
  return t.workers.listSwarms.useQuery();
};

export const useCancelSwarm = () => {
  const utils = t.useUtils();
  return t.workers.cancelSwarm.useMutation({
    onSuccess: () => {
      utils.workers.listSwarms.invalidate();
    },
  });
};

export const useWorkerStats = () => {
  return t.workers.getStats.useQuery();
};
