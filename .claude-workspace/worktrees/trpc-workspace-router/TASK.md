---
id: trpc-workspace-router
name: Workspace Router Implementation
wave: 2
priority: 2
dependencies:
- trpc-core-setup
estimated_hours: 4
tags:
- backend
- api
- trpc
- workspace
---

## Objective

Implement the tRPC workspace router with full CRUD operations for workspace management, mirroring the existing Go REST endpoints.

## Context

The existing Go backend has workspace operations at `/api/v1/workspace/`. This router will provide type-safe tRPC procedures for:
- Getting/setting workspace directory
- Browsing directories
- Managing recent workspaces
- Folder picker integration

Existing Go endpoints to replicate:
- `GET /workspace/directory` - Get current workspace
- `POST /workspace/directory` - Set workspace directory
- `GET /workspace/browse` - Browse directories
- `POST /workspace/pick-folder` - Native folder picker
- `GET /workspace/recent` - List recent workspaces
- `POST /workspace/:id/current` - Set as current
- `DELETE /workspace/:id` - Remove from recent

## Implementation

### 1. Define Zod Schemas

**File: `packages/trpc/src/routers/workspace/schemas.ts`**
```typescript
import { z } from 'zod';

export const workspaceSchema = z.object({
  id: z.string(),
  userId: z.string(),
  path: z.string(),
  name: z.string(),
  isCurrent: z.boolean(),
  lastAccessedAt: z.date().nullable(),
  createdAt: z.date(),
});

export const setDirectoryInput = z.object({
  path: z.string().min(1, 'Path is required'),
});

export const browseDirectoryInput = z.object({
  path: z.string().optional(),
  showHidden: z.boolean().default(false),
});

export const browseDirectoryOutput = z.object({
  currentPath: z.string(),
  parentPath: z.string().nullable(),
  directories: z.array(z.object({
    name: z.string(),
    path: z.string(),
  })),
});

export const listRecentInput = z.object({
  limit: z.number().min(1).max(50).default(10),
});

export const workspaceIdInput = z.object({
  id: z.string().uuid(),
});

export type Workspace = z.infer<typeof workspaceSchema>;
export type SetDirectoryInput = z.infer<typeof setDirectoryInput>;
export type BrowseDirectoryInput = z.infer<typeof browseDirectoryInput>;
export type BrowseDirectoryOutput = z.infer<typeof browseDirectoryOutput>;
```

### 2. Implement Workspace Router

**File: `packages/trpc/src/routers/workspace/index.ts`**
```typescript
import { router, protectedProcedure } from '../../trpc';
import { TRPCError } from '@trpc/server';
import * as schemas from './schemas';

export const workspaceRouter = router({
  // Get current workspace
  getCurrent: protectedProcedure
    .output(schemas.workspaceSchema.nullable())
    .query(async ({ ctx }) => {
      // Call workspace service or database
      const workspace = await workspaceService.getCurrent(ctx.session.userId);
      return workspace;
    }),

  // Set workspace directory
  setDirectory: protectedProcedure
    .input(schemas.setDirectoryInput)
    .output(schemas.workspaceSchema)
    .mutation(async ({ ctx, input }) => {
      // Validate path exists
      // Create or update workspace
      const workspace = await workspaceService.setDirectory(
        ctx.session.userId,
        input.path
      );
      return workspace;
    }),

  // Browse directories
  browse: protectedProcedure
    .input(schemas.browseDirectoryInput)
    .output(schemas.browseDirectoryOutput)
    .query(async ({ ctx, input }) => {
      const result = await workspaceService.browse(
        input.path || process.env.HOME || '/',
        input.showHidden
      );
      return result;
    }),

  // Open native folder picker
  pickFolder: protectedProcedure
    .output(z.object({ path: z.string().nullable() }))
    .mutation(async ({ ctx }) => {
      // This may need to trigger electron dialog or use a web alternative
      const path = await workspaceService.openFolderPicker();
      return { path };
    }),

  // List recent workspaces
  listRecent: protectedProcedure
    .input(schemas.listRecentInput)
    .output(z.array(schemas.workspaceSchema))
    .query(async ({ ctx, input }) => {
      const workspaces = await workspaceService.listRecent(
        ctx.session.userId,
        input.limit
      );
      return workspaces;
    }),

  // Set workspace as current
  setCurrent: protectedProcedure
    .input(schemas.workspaceIdInput)
    .output(schemas.workspaceSchema)
    .mutation(async ({ ctx, input }) => {
      const workspace = await workspaceService.setCurrent(
        ctx.session.userId,
        input.id
      );
      if (!workspace) {
        throw new TRPCError({
          code: 'NOT_FOUND',
          message: 'Workspace not found',
        });
      }
      return workspace;
    }),

  // Remove from recent workspaces
  remove: protectedProcedure
    .input(schemas.workspaceIdInput)
    .mutation(async ({ ctx, input }) => {
      await workspaceService.delete(ctx.session.userId, input.id);
      return { success: true };
    }),

  // Get workspace by ID
  getById: protectedProcedure
    .input(schemas.workspaceIdInput)
    .output(schemas.workspaceSchema.nullable())
    .query(async ({ ctx, input }) => {
      const workspace = await workspaceService.getById(input.id);
      if (workspace && workspace.userId !== ctx.session.userId) {
        throw new TRPCError({
          code: 'FORBIDDEN',
          message: 'Access denied',
        });
      }
      return workspace;
    }),
});
```

### 3. Create Workspace Service

**File: `packages/trpc/src/services/workspace.ts`**
```typescript
import { Workspace } from '../routers/workspace/schemas';

// Service that either:
// 1. Calls the Go backend REST API
// 2. Directly accesses the database
// 3. Uses a shared database client

export const workspaceService = {
  async getCurrent(userId: string): Promise<Workspace | null> {
    // Implementation
  },

  async setDirectory(userId: string, path: string): Promise<Workspace> {
    // Validate path exists on filesystem
    // Create or update workspace record
  },

  async browse(path: string, showHidden: boolean) {
    // Read directory contents
    // Filter hidden files if needed
    // Return directory listing
  },

  async openFolderPicker(): Promise<string | null> {
    // Electron dialog or alternative
  },

  async listRecent(userId: string, limit: number): Promise<Workspace[]> {
    // Get recent workspaces ordered by lastAccessedAt
  },

  async setCurrent(userId: string, workspaceId: string): Promise<Workspace | null> {
    // Set isCurrent = false for all user workspaces
    // Set isCurrent = true for specified workspace
    // Update lastAccessedAt
  },

  async delete(userId: string, workspaceId: string): Promise<void> {
    // Delete workspace from user's list
  },

  async getById(id: string): Promise<Workspace | null> {
    // Get workspace by ID
  },
};
```

### 4. Register Router

Update root router to include workspace router:

**File: `packages/trpc/src/router.ts`** (modification)
```typescript
import { workspaceRouter } from './routers/workspace';

export const appRouter = router({
  workspace: workspaceRouter,
});
```

## Acceptance Criteria

- [ ] All Zod schemas defined with proper validation
- [ ] All workspace procedures implemented and type-safe
- [ ] Protected procedures require authentication
- [ ] Proper error handling with TRPCError
- [ ] Path validation for filesystem operations
- [ ] User isolation (users can only access their own workspaces)
- [ ] Router registered in root appRouter

## Files to Create/Modify

- `packages/trpc/src/routers/workspace/schemas.ts` - Zod schemas
- `packages/trpc/src/routers/workspace/index.ts` - Router implementation
- `packages/trpc/src/services/workspace.ts` - Workspace service
- `packages/trpc/src/router.ts` - Add workspace router (modify)

## Integration Points

- **Provides**: Workspace CRUD operations via tRPC
- **Consumes**: trpc-core-setup (protectedProcedure, router, context)
- **Conflicts**: Avoid modifying Go workspace handlers

## Notes

- Existing Go workspace table: `user_workspaces`
- Fields: id, user_id, path, name, is_current, last_accessed_at, created_at
- Consider whether to call Go backend or access database directly
