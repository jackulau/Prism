import { z } from 'zod';

export const workspaceSchema = z.object({
  id: z.string().uuid(),
  userId: z.string(),
  path: z.string(),
  name: z.string(),
  isCurrent: z.boolean(),
  lastAccessedAt: z.string().datetime().nullable(),
  createdAt: z.string().datetime(),
});

export const setDirectoryInput = z.object({
  path: z.string().min(1, 'Path is required'),
});

export const browseDirectoryInput = z.object({
  path: z.string().optional(),
  showHidden: z.boolean().default(false),
});

export const directoryEntrySchema = z.object({
  name: z.string(),
  path: z.string(),
});

export const browseDirectoryOutput = z.object({
  currentPath: z.string(),
  parentPath: z.string().nullable(),
  directories: z.array(directoryEntrySchema),
});

export const listRecentInput = z.object({
  limit: z.number().min(1).max(50).default(10),
});

export const workspaceIdInput = z.object({
  id: z.string().uuid(),
});

export const pickFolderOutput = z.object({
  path: z.string().nullable(),
  cancelled: z.boolean().optional(),
});

export const successOutput = z.object({
  success: z.boolean(),
});

export const setDirectoryOutput = z.object({
  success: z.boolean(),
  path: z.string(),
});

export type Workspace = z.infer<typeof workspaceSchema>;
export type SetDirectoryInput = z.infer<typeof setDirectoryInput>;
export type BrowseDirectoryInput = z.infer<typeof browseDirectoryInput>;
export type BrowseDirectoryOutput = z.infer<typeof browseDirectoryOutput>;
export type ListRecentInput = z.infer<typeof listRecentInput>;
export type WorkspaceIdInput = z.infer<typeof workspaceIdInput>;
export type PickFolderOutput = z.infer<typeof pickFolderOutput>;
