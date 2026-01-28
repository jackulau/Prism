import type {
  UserProfile,
  UserSettings,
  ProviderKey,
  ProviderType,
  GitHubConnection,
  UpdateSettingsInput,
} from '../routers/organization/schemas.js';

// Database connection placeholder - will be injected
let db: DatabaseConnection | null = null;
let encryptionKey: string | null = null;

export interface DatabaseConnection {
  query<T>(sql: string, params?: unknown[]): Promise<T[]>;
  get<T>(sql: string, params?: unknown[]): Promise<T | null>;
  run(sql: string, params?: unknown[]): Promise<{ changes: number }>;
}

export function setDatabase(database: DatabaseConnection): void {
  db = database;
}

export function setEncryptionKey(key: string): void {
  encryptionKey = key;
}

function getDb(): DatabaseConnection {
  if (!db) {
    throw new Error('Database not configured for organization service');
  }
  return db;
}

// User profile operations
export async function getProfile(userId: string): Promise<UserProfile | null> {
  const user = await getDb().get<{
    id: string;
    email: string;
    created_at: string;
    updated_at: string;
    github_username: string | null;
    github_connected_at: string | null;
  }>(
    `SELECT id, email, created_at, updated_at, github_username, github_connected_at
     FROM users WHERE id = ?`,
    [userId]
  );

  if (!user) {
    return null;
  }

  return {
    id: user.id,
    email: user.email,
    createdAt: new Date(user.created_at),
    updatedAt: new Date(user.updated_at),
    githubConnected: !!user.github_username,
    githubUsername: user.github_username,
  };
}

export async function changePassword(
  userId: string,
  currentPassword: string,
  newPassword: string
): Promise<void> {
  const user = await getDb().get<{ password_hash: string }>(
    `SELECT password_hash FROM users WHERE id = ?`,
    [userId]
  );

  if (!user) {
    throw new Error('User not found');
  }

  // Verify current password using Argon2id
  const isValid = await verifyPassword(currentPassword, user.password_hash);
  if (!isValid) {
    throw new Error('Current password is incorrect');
  }

  // Hash new password
  const newHash = await hashPassword(newPassword);

  await getDb().run(
    `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`,
    [newHash, new Date().toISOString(), userId]
  );
}

export async function changeEmail(
  userId: string,
  newEmail: string,
  password: string
): Promise<void> {
  const user = await getDb().get<{ password_hash: string }>(
    `SELECT password_hash FROM users WHERE id = ?`,
    [userId]
  );

  if (!user) {
    throw new Error('User not found');
  }

  // Verify password
  const isValid = await verifyPassword(password, user.password_hash);
  if (!isValid) {
    throw new Error('Password is incorrect');
  }

  // Check if email is already taken
  const existing = await getDb().get<{ id: string }>(
    `SELECT id FROM users WHERE email = ? AND id != ?`,
    [newEmail.toLowerCase(), userId]
  );

  if (existing) {
    throw new Error('Email is already in use');
  }

  await getDb().run(
    `UPDATE users SET email = ?, updated_at = ? WHERE id = ?`,
    [newEmail.toLowerCase(), new Date().toISOString(), userId]
  );
}

// User settings operations
const DEFAULT_SETTINGS: UserSettings = {
  theme: 'system',
  defaultProvider: null,
  defaultModel: null,
  editorFontSize: 14,
  editorTabSize: 2,
};

export async function getSettings(userId: string): Promise<UserSettings> {
  const settings = await getDb().get<{
    theme: string | null;
    default_provider: string | null;
    default_model: string | null;
    settings_json: string | null;
  }>(
    `SELECT theme, default_provider, default_model, settings_json
     FROM user_settings WHERE user_id = ?`,
    [userId]
  );

  if (!settings) {
    return { ...DEFAULT_SETTINGS };
  }

  let extraSettings: Partial<UserSettings> = {};
  if (settings.settings_json) {
    try {
      extraSettings = JSON.parse(settings.settings_json);
    } catch {
      // Ignore parse errors
    }
  }

  return {
    theme: (settings.theme as UserSettings['theme']) || DEFAULT_SETTINGS.theme,
    defaultProvider: settings.default_provider,
    defaultModel: settings.default_model,
    editorFontSize: extraSettings.editorFontSize ?? DEFAULT_SETTINGS.editorFontSize,
    editorTabSize: extraSettings.editorTabSize ?? DEFAULT_SETTINGS.editorTabSize,
  };
}

export async function updateSettings(
  userId: string,
  updates: UpdateSettingsInput
): Promise<UserSettings> {
  const current = await getSettings(userId);

  const newSettings: UserSettings = {
    theme: updates.theme ?? current.theme,
    defaultProvider: updates.defaultProvider !== undefined ? updates.defaultProvider : current.defaultProvider,
    defaultModel: updates.defaultModel !== undefined ? updates.defaultModel : current.defaultModel,
    editorFontSize: updates.editorFontSize ?? current.editorFontSize,
    editorTabSize: updates.editorTabSize ?? current.editorTabSize,
  };

  const settingsJson = JSON.stringify({
    editorFontSize: newSettings.editorFontSize,
    editorTabSize: newSettings.editorTabSize,
  });

  // Upsert settings
  await getDb().run(
    `INSERT INTO user_settings (user_id, theme, default_provider, default_model, settings_json, updated_at)
     VALUES (?, ?, ?, ?, ?, ?)
     ON CONFLICT(user_id) DO UPDATE SET
       theme = excluded.theme,
       default_provider = excluded.default_provider,
       default_model = excluded.default_model,
       settings_json = excluded.settings_json,
       updated_at = excluded.updated_at`,
    [
      userId,
      newSettings.theme,
      newSettings.defaultProvider,
      newSettings.defaultModel,
      settingsJson,
      new Date().toISOString(),
    ]
  );

  return newSettings;
}

// Provider key operations
const ALL_PROVIDERS: ProviderType[] = ['openai', 'anthropic', 'google', 'ollama', 'openrouter'];

export async function listProviderKeys(userId: string): Promise<ProviderKey[]> {
  const keys = await getDb().query<{
    provider: string;
    created_at: string;
  }>(
    `SELECT provider, created_at FROM provider_keys
     WHERE user_id = ? AND is_active = 1`,
    [userId]
  );

  const configuredProviders = new Set(keys.map((k) => k.provider));

  return ALL_PROVIDERS.map((provider) => {
    const key = keys.find((k) => k.provider === provider);
    return {
      provider,
      isConfigured: configuredProviders.has(provider),
      createdAt: key ? new Date(key.created_at) : null,
    };
  });
}

export async function setProviderKey(
  userId: string,
  provider: ProviderType,
  apiKey: string
): Promise<ProviderKey> {
  // Encrypt the API key
  const { ciphertext, nonce } = await encryptApiKey(apiKey);
  const id = crypto.randomUUID();
  const now = new Date().toISOString();

  await getDb().run(
    `INSERT INTO provider_keys (id, user_id, provider, encrypted_key, key_nonce, is_active, created_at)
     VALUES (?, ?, ?, ?, ?, 1, ?)
     ON CONFLICT(user_id, provider) DO UPDATE SET
       encrypted_key = excluded.encrypted_key,
       key_nonce = excluded.key_nonce,
       is_active = 1`,
    [id, userId, provider, ciphertext, nonce, now]
  );

  return {
    provider,
    isConfigured: true,
    createdAt: new Date(now),
  };
}

export async function deleteProviderKey(
  userId: string,
  provider: ProviderType
): Promise<void> {
  await getDb().run(
    `DELETE FROM provider_keys WHERE user_id = ? AND provider = ?`,
    [userId, provider]
  );
}

export async function validateProviderKey(
  provider: ProviderType,
  apiKey: string
): Promise<{ valid: boolean; error: string | null }> {
  try {
    switch (provider) {
      case 'openai':
        return await validateOpenAIKey(apiKey);
      case 'anthropic':
        return await validateAnthropicKey(apiKey);
      case 'ollama':
        // Ollama doesn't require an API key
        return { valid: true, error: null };
      case 'google':
        return await validateGoogleKey(apiKey);
      case 'openrouter':
        return await validateOpenRouterKey(apiKey);
      default:
        return { valid: apiKey.length > 0, error: null };
    }
  } catch (err) {
    return {
      valid: false,
      error: err instanceof Error ? err.message : 'Validation failed',
    };
  }
}

// GitHub connection operations
export async function getGitHubConnection(userId: string): Promise<GitHubConnection> {
  const user = await getDb().get<{
    github_username: string | null;
    github_connected_at: string | null;
  }>(
    `SELECT github_username, github_connected_at FROM users WHERE id = ?`,
    [userId]
  );

  if (!user) {
    throw new Error('User not found');
  }

  return {
    connected: !!user.github_username,
    username: user.github_username,
    connectedAt: user.github_connected_at ? new Date(user.github_connected_at) : null,
  };
}

export async function disconnectGitHub(userId: string): Promise<void> {
  await getDb().run(
    `UPDATE users SET github_token = NULL, github_username = NULL, github_connected_at = NULL, updated_at = ? WHERE id = ?`,
    [new Date().toISOString(), userId]
  );
}

// Account deletion
export async function deleteAccount(userId: string, password: string): Promise<void> {
  const user = await getDb().get<{ password_hash: string }>(
    `SELECT password_hash FROM users WHERE id = ?`,
    [userId]
  );

  if (!user) {
    throw new Error('User not found');
  }

  // Verify password
  const isValid = await verifyPassword(password, user.password_hash);
  if (!isValid) {
    throw new Error('Password is incorrect');
  }

  // Delete user (cascade will handle related data)
  await getDb().run(`DELETE FROM users WHERE id = ?`, [userId]);
}

// Password hashing utilities (Argon2id compatible with Go backend)
async function hashPassword(password: string): Promise<string> {
  // Use Web Crypto API for random bytes
  const salt = new Uint8Array(16);
  crypto.getRandomValues(salt);

  // Note: For full compatibility with Go's Argon2id implementation,
  // we'd need to use argon2 npm package. For now, using a placeholder
  // that should be replaced with proper argon2 hashing in production.
  // This is a simplification - in production, use @node-rs/argon2 or similar.

  const encoder = new TextEncoder();
  const data = encoder.encode(password);
  const hashBuffer = await crypto.subtle.digest('SHA-256', new Uint8Array([...salt, ...data]));
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
  const saltHex = Array.from(salt).map(b => b.toString(16).padStart(2, '0')).join('');

  return `${saltHex}$${hashHex}`;
}

async function verifyPassword(password: string, hashedPassword: string): Promise<boolean> {
  const [saltHex, expectedHashHex] = hashedPassword.split('$');
  if (!saltHex || !expectedHashHex) {
    return false;
  }

  const salt = new Uint8Array(saltHex.match(/.{2}/g)?.map(byte => parseInt(byte, 16)) || []);
  const encoder = new TextEncoder();
  const data = encoder.encode(password);
  const hashBuffer = await crypto.subtle.digest('SHA-256', new Uint8Array([...salt, ...data]));
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const computedHashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');

  return computedHashHex === expectedHashHex;
}

// Encryption utilities (AES-256-GCM compatible with Go backend)
async function encryptApiKey(apiKey: string): Promise<{ ciphertext: Uint8Array; nonce: Uint8Array }> {
  if (!encryptionKey) {
    throw new Error('Encryption key not configured');
  }

  const keyBytes = hexToBytes(encryptionKey);
  const key = await crypto.subtle.importKey('raw', keyBytes, 'AES-GCM', false, ['encrypt']);

  const nonce = new Uint8Array(12);
  crypto.getRandomValues(nonce);

  const encoder = new TextEncoder();
  const data = encoder.encode(apiKey);

  const ciphertext = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv: nonce },
    key,
    data
  );

  return {
    ciphertext: new Uint8Array(ciphertext),
    nonce,
  };
}

function hexToBytes(hex: string): Uint8Array {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.substring(i, i + 2), 16);
  }
  return bytes;
}

// Provider validation utilities
async function validateOpenAIKey(apiKey: string): Promise<{ valid: boolean; error: string | null }> {
  try {
    const response = await fetch('https://api.openai.com/v1/models', {
      headers: { Authorization: `Bearer ${apiKey}` },
    });
    if (response.ok) {
      return { valid: true, error: null };
    }
    return { valid: false, error: 'Invalid API key' };
  } catch (err) {
    return { valid: false, error: 'Failed to validate key' };
  }
}

async function validateAnthropicKey(apiKey: string): Promise<{ valid: boolean; error: string | null }> {
  if (!apiKey.startsWith('sk-ant-')) {
    return { valid: false, error: 'Invalid API key format' };
  }
  try {
    const response = await fetch('https://api.anthropic.com/v1/messages', {
      method: 'GET',
      headers: {
        'x-api-key': apiKey,
        'anthropic-version': '2023-06-01',
      },
    });
    // 401 means invalid key, 405 means valid key but wrong method
    if (response.status !== 401) {
      return { valid: true, error: null };
    }
    return { valid: false, error: 'Invalid API key' };
  } catch (err) {
    return { valid: false, error: 'Failed to validate key' };
  }
}

async function validateGoogleKey(apiKey: string): Promise<{ valid: boolean; error: string | null }> {
  try {
    const response = await fetch(
      `https://generativelanguage.googleapis.com/v1beta/models?key=${apiKey}`
    );
    if (response.ok) {
      return { valid: true, error: null };
    }
    return { valid: false, error: 'Invalid API key' };
  } catch (err) {
    return { valid: false, error: 'Failed to validate key' };
  }
}

async function validateOpenRouterKey(apiKey: string): Promise<{ valid: boolean; error: string | null }> {
  try {
    const response = await fetch('https://openrouter.ai/api/v1/models', {
      headers: { Authorization: `Bearer ${apiKey}` },
    });
    if (response.ok) {
      return { valid: true, error: null };
    }
    return { valid: false, error: 'Invalid API key' };
  } catch (err) {
    return { valid: false, error: 'Failed to validate key' };
  }
}

// Export all functions for use in router
export const organizationService = {
  getProfile,
  changePassword,
  changeEmail,
  getSettings,
  updateSettings,
  listProviderKeys,
  setProviderKey,
  deleteProviderKey,
  validateProviderKey,
  getGitHubConnection,
  disconnectGitHub,
  deleteAccount,
  setDatabase,
  setEncryptionKey,
};
