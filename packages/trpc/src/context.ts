import type { CreateExpressContextOptions } from '@trpc/server/adapters/express';
import * as jose from 'jose';

export interface Session {
  userId: string;
  email: string;
  token: string;
}

export interface Context {
  session: Session | null;
  token: string | null;
}

interface JWTPayload extends jose.JWTPayload {
  user_id: string;
  email: string;
  type: 'access' | 'refresh';
}

let jwtSecret: Uint8Array | null = null;

export function setJWTSecret(secret: string): void {
  jwtSecret = new TextEncoder().encode(secret);
}

async function validateToken(token: string): Promise<Session | null> {
  if (!jwtSecret) {
    console.warn('JWT secret not configured, skipping token validation');
    return null;
  }

  try {
    const { payload } = await jose.jwtVerify(token, jwtSecret, {
      issuer: 'prism',
    });

    const claims = payload as JWTPayload;

    if (claims.type !== 'access') {
      return null;
    }

    return {
      userId: claims.user_id,
      email: claims.email,
      token,
    };
  } catch {
    return null;
  }
}

export async function createContext(
  opts: CreateExpressContextOptions
): Promise<Context> {
  const authHeader = opts.req.headers.authorization;

  if (authHeader?.startsWith('Bearer ')) {
    const token = authHeader.slice(7);
    const session = await validateToken(token);
    return { session, token };
  }

  return { session: null, token: null };
}
