import { TRPCError } from '@trpc/server';

export function notFound(resource: string): TRPCError {
  return new TRPCError({
    code: 'NOT_FOUND',
    message: `${resource} not found`,
  });
}

export function unauthorized(message = 'Unauthorized'): TRPCError {
  return new TRPCError({
    code: 'UNAUTHORIZED',
    message,
  });
}

export function forbidden(message = 'Forbidden'): TRPCError {
  return new TRPCError({
    code: 'FORBIDDEN',
    message,
  });
}

export function badRequest(message: string): TRPCError {
  return new TRPCError({
    code: 'BAD_REQUEST',
    message,
  });
}

export function internalError(message = 'Internal server error'): TRPCError {
  return new TRPCError({
    code: 'INTERNAL_SERVER_ERROR',
    message,
  });
}

export function conflict(message: string): TRPCError {
  return new TRPCError({
    code: 'CONFLICT',
    message,
  });
}
