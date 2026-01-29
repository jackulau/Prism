---
id: trpc-frontend-client
name: Frontend tRPC Client Integration
wave: 2
priority: 3
dependencies:
- trpc-core-setup
estimated_hours: 3
tags:
- frontend
- api
- trpc
- react
---

## Objective

Set up the tRPC client in the React frontend with proper type inference, React Query integration, and authentication handling.

## Context

The existing frontend uses plain fetch/axios for REST API calls via `/frontend/src/services/api.ts`. This task will:
- Add tRPC client with full type inference from the backend
- Integrate with React Query for caching and mutations
- Set up authentication token handling
- Create hooks for easy data fetching

## Implementation

### 1. Install Dependencies

```bash
cd frontend
npm install @trpc/client @trpc/react-query @tanstack/react-query superjson
```

### 2. Set Up tRPC Client

**File: `frontend/src/lib/trpc.ts`**
```typescript
import { createTRPCReact } from '@trpc/react-query';
import { httpBatchLink } from '@trpc/client';
import superjson from 'superjson';
import type { AppRouter } from '../../packages/trpc/src/router';

export const trpc = createTRPCReact<AppRouter>();

export const createTRPCClient = (getToken: () => string | null) => {
  return trpc.createClient({
    transformer: superjson,
    links: [
      httpBatchLink({
        url: `${import.meta.env.VITE_API_URL || 'http://localhost:3001'}/trpc`,
        headers: () => {
          const token = getToken();
          return token
            ? { Authorization: `Bearer ${token}` }
            : {};
        },
      }),
    ],
  });
};
```

### 3. Create Provider Wrapper

**File: `frontend/src/providers/TRPCProvider.tsx`**
```typescript
import React, { useState } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { trpc, createTRPCClient } from '../lib/trpc';
import { useAuthStore } from '../store/authStore';

interface TRPCProviderProps {
  children: React.ReactNode;
}

export const TRPCProvider: React.FC<TRPCProviderProps> = ({ children }) => {
  const { token } = useAuthStore();

  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 5 * 60 * 1000, // 5 minutes
            refetchOnWindowFocus: false,
          },
        },
      })
  );

  const [trpcClient] = useState(() =>
    createTRPCClient(() => token)
  );

  return (
    <trpc.Provider client={trpcClient} queryClient={queryClient}>
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    </trpc.Provider>
  );
};
```

### 4. Update App Entry Point

**File: `frontend/src/App.tsx`** (modification)
```typescript
import { TRPCProvider } from './providers/TRPCProvider';

function App() {
  return (
    <TRPCProvider>
      {/* existing app content */}
    </TRPCProvider>
  );
}
```

### 5. Create Typed Hooks

**File: `frontend/src/hooks/useWorkspace.ts`**
```typescript
import { trpc } from '../lib/trpc';

export const useCurrentWorkspace = () => {
  return trpc.workspace.getCurrent.useQuery();
};

export const useRecentWorkspaces = (limit = 10) => {
  return trpc.workspace.listRecent.useQuery({ limit });
};

export const useSetWorkspaceDirectory = () => {
  const utils = trpc.useUtils();
  return trpc.workspace.setDirectory.useMutation({
    onSuccess: () => {
      utils.workspace.getCurrent.invalidate();
      utils.workspace.listRecent.invalidate();
    },
  });
};

export const useBrowseDirectory = (path?: string, showHidden = false) => {
  return trpc.workspace.browse.useQuery(
    { path, showHidden },
    { enabled: path !== undefined }
  );
};

export const useRemoveWorkspace = () => {
  const utils = trpc.useUtils();
  return trpc.workspace.remove.useMutation({
    onSuccess: () => {
      utils.workspace.listRecent.invalidate();
    },
  });
};
```

**File: `frontend/src/hooks/useWorkers.ts`**
```typescript
import { trpc } from '../lib/trpc';

export const useRunTask = () => {
  const utils = trpc.useUtils();
  return trpc.workers.runTask.useMutation({
    onSuccess: () => {
      utils.workers.listExecutions.invalidate();
    },
  });
};

export const useExecutions = () => {
  return trpc.workers.listExecutions.useQuery();
};

export const useExecution = (executionId: string) => {
  return trpc.workers.getExecution.useQuery(
    { executionId },
    {
      enabled: !!executionId,
      refetchInterval: (data) =>
        data?.status === 'running' ? 1000 : false,
    }
  );
};

export const useCancelExecution = () => {
  const utils = trpc.useUtils();
  return trpc.workers.cancelExecution.useMutation({
    onSuccess: () => {
      utils.workers.listExecutions.invalidate();
    },
  });
};

export const useSwarms = () => {
  return trpc.workers.listSwarms.useQuery();
};
```

**File: `frontend/src/hooks/useIntegrations.ts`**
```typescript
import { trpc } from '../lib/trpc';

export const useIntegrationStatus = () => {
  return trpc.integrations.getStatus.useQuery();
};

export const useConfigureDiscord = () => {
  const utils = trpc.useUtils();
  return trpc.integrations.configureDiscord.useMutation({
    onSuccess: () => {
      utils.integrations.getStatus.invalidate();
    },
  });
};

export const useDisconnectDiscord = () => {
  const utils = trpc.useUtils();
  return trpc.integrations.disconnectDiscord.useMutation({
    onSuccess: () => {
      utils.integrations.getStatus.invalidate();
    },
  });
};

// Similar hooks for Slack, PostHog, MCP...
```

**File: `frontend/src/hooks/useOrganization.ts`**
```typescript
import { trpc } from '../lib/trpc';

export const useProfile = () => {
  return trpc.organization.getProfile.useQuery();
};

export const useSettings = () => {
  return trpc.organization.getSettings.useQuery();
};

export const useUpdateSettings = () => {
  const utils = trpc.useUtils();
  return trpc.organization.updateSettings.useMutation({
    onSuccess: () => {
      utils.organization.getSettings.invalidate();
    },
  });
};

export const useProviderKeys = () => {
  return trpc.organization.listProviderKeys.useQuery();
};

export const useSetProviderKey = () => {
  const utils = trpc.useUtils();
  return trpc.organization.setProviderKey.useMutation({
    onSuccess: () => {
      utils.organization.listProviderKeys.invalidate();
    },
  });
};
```

**File: `frontend/src/hooks/usePayment.ts`**
```typescript
import { trpc } from '../lib/trpc';

export const usePlans = () => {
  return trpc.payment.listPlans.useQuery();
};

export const useSubscription = () => {
  return trpc.payment.getSubscription.useQuery();
};

export const useUsage = () => {
  return trpc.payment.getUsage.useQuery();
};

export const useCreateCheckout = () => {
  return trpc.payment.createCheckoutSession.useMutation({
    onSuccess: (data) => {
      // Redirect to Stripe Checkout
      window.location.href = data.url;
    },
  });
};

export const useCreatePortal = () => {
  return trpc.payment.createPortalSession.useMutation({
    onSuccess: (data) => {
      window.location.href = data.url;
    },
  });
};

export const useInvoices = (limit = 10) => {
  return trpc.payment.listInvoices.useQuery({ limit });
};
```

### 6. Type Export Configuration

**File: `packages/trpc/package.json`** (ensure exports)
```json
{
  "name": "@prism/trpc",
  "exports": {
    ".": "./src/index.ts",
    "./router": "./src/router.ts"
  },
  "types": "./src/index.ts"
}
```

**File: `frontend/tsconfig.json`** (add path alias)
```json
{
  "compilerOptions": {
    "paths": {
      "@prism/trpc/*": ["../packages/trpc/src/*"]
    }
  }
}
```

### 7. Error Handling

**File: `frontend/src/lib/trpcErrorHandler.ts`**
```typescript
import { TRPCClientError } from '@trpc/client';
import type { AppRouter } from '@prism/trpc/router';

export const isTRPCError = (error: unknown): error is TRPCClientError<AppRouter> => {
  return error instanceof TRPCClientError;
};

export const getTRPCErrorMessage = (error: unknown): string => {
  if (isTRPCError(error)) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'An unexpected error occurred';
};

export const handleTRPCError = (error: unknown) => {
  if (isTRPCError(error)) {
    switch (error.data?.code) {
      case 'UNAUTHORIZED':
        // Redirect to login or refresh token
        break;
      case 'FORBIDDEN':
        // Show access denied message
        break;
      case 'NOT_FOUND':
        // Show not found message
        break;
      default:
        // Show generic error
    }
  }
};
```

## Acceptance Criteria

- [ ] tRPC client configured with proper types
- [ ] React Query integration working
- [ ] Authentication token included in requests
- [ ] Query invalidation on mutations
- [ ] Typed hooks for all routers (workspace, workers, integrations, organization, payment)
- [ ] Error handling utilities
- [ ] TypeScript path aliases configured
- [ ] Provider wrapped around app

## Files to Create/Modify

- `frontend/src/lib/trpc.ts` - tRPC client setup
- `frontend/src/providers/TRPCProvider.tsx` - React provider
- `frontend/src/App.tsx` - Add provider (modify)
- `frontend/src/hooks/useWorkspace.ts` - Workspace hooks
- `frontend/src/hooks/useWorkers.ts` - Worker hooks
- `frontend/src/hooks/useIntegrations.ts` - Integration hooks
- `frontend/src/hooks/useOrganization.ts` - Organization hooks
- `frontend/src/hooks/usePayment.ts` - Payment hooks
- `frontend/src/lib/trpcErrorHandler.ts` - Error handling
- `frontend/tsconfig.json` - Path aliases (modify)
- `frontend/package.json` - Dependencies (modify)

## Integration Points

- **Provides**: Type-safe React hooks for all tRPC routers
- **Consumes**: trpc-core-setup (AppRouter type)
- **Conflicts**: May need to gradually migrate from existing api.ts service

## Notes

- Existing auth store: `/frontend/src/store/authStore.ts` with token persistence
- Current API service: `/frontend/src/services/api.ts`
- Consider gradual migration - keep existing REST calls working during transition
- React Query devtools can be added for debugging
