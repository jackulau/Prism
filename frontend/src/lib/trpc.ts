import { createTRPCReact } from '@trpc/react-query';
import { httpBatchLink } from '@trpc/client';
import superjson from 'superjson';

// Placeholder router type until @prism/trpc package is ready
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AppRouter = any;

export const trpc = createTRPCReact<AppRouter>();

function getTRPCUrl(): string {
  // In production, use the configured API URL
  if (import.meta.env.VITE_API_URL) {
    return `${import.meta.env.VITE_API_URL}/trpc`;
  }
  // In development, use the Vite proxy
  if (import.meta.env.DEV) {
    return '/trpc';
  }
  // Fallback
  return 'http://localhost:3001/trpc';
}

export function createTRPCClient(getToken: () => string | null) {
  // @ts-expect-error - tRPC requires a proper router type, using placeholder until backend is ready
  return trpc.createClient({
    links: [
      httpBatchLink({
        url: getTRPCUrl(),
        headers: () => {
          const token = getToken();
          return token ? { Authorization: `Bearer ${token}` } : {};
        },
        transformer: superjson,
      }),
    ],
  });
}
