import { TRPCClientError } from '@trpc/client';
import type { AppRouter } from '@prism/trpc/router';

export function isTRPCError(
  error: unknown
): error is TRPCClientError<AppRouter> {
  return error instanceof TRPCClientError;
}

export function getTRPCErrorMessage(error: unknown): string {
  if (isTRPCError(error)) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'An unexpected error occurred';
}

export type TRPCErrorCode =
  | 'PARSE_ERROR'
  | 'BAD_REQUEST'
  | 'INTERNAL_SERVER_ERROR'
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'METHOD_NOT_SUPPORTED'
  | 'TIMEOUT'
  | 'CONFLICT'
  | 'PRECONDITION_FAILED'
  | 'PAYLOAD_TOO_LARGE'
  | 'UNPROCESSABLE_CONTENT'
  | 'TOO_MANY_REQUESTS'
  | 'CLIENT_CLOSED_REQUEST';

export function getTRPCErrorCode(error: unknown): TRPCErrorCode | null {
  if (isTRPCError(error)) {
    return error.data?.code as TRPCErrorCode ?? null;
  }
  return null;
}

export interface ErrorHandlerOptions {
  onUnauthorized?: () => void;
  onForbidden?: () => void;
  onNotFound?: () => void;
  onRateLimit?: () => void;
  onDefault?: (message: string) => void;
}

export function handleTRPCError(
  error: unknown,
  options: ErrorHandlerOptions = {}
) {
  const code = getTRPCErrorCode(error);
  const message = getTRPCErrorMessage(error);

  switch (code) {
    case 'UNAUTHORIZED':
      if (options.onUnauthorized) {
        options.onUnauthorized();
      } else {
        // Default: user should re-authenticate
        console.error('Unauthorized:', message);
      }
      break;
    case 'FORBIDDEN':
      if (options.onForbidden) {
        options.onForbidden();
      } else {
        console.error('Forbidden:', message);
      }
      break;
    case 'NOT_FOUND':
      if (options.onNotFound) {
        options.onNotFound();
      } else {
        console.error('Not found:', message);
      }
      break;
    case 'TOO_MANY_REQUESTS':
      if (options.onRateLimit) {
        options.onRateLimit();
      } else {
        console.error('Rate limited:', message);
      }
      break;
    default:
      if (options.onDefault) {
        options.onDefault(message);
      } else {
        console.error('tRPC Error:', message);
      }
  }
}
