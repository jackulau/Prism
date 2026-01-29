---
id: trpc-setup
name: tRPC API Layer Setup
wave: 1
priority: 1
dependencies: []
estimated_hours: 8
tags:
- backend
- frontend
- api
- trpc
---

## Objective

Add tRPC as an alternative API layer alongside existing REST endpoints for type-safe client-server communication.

## Context

The codebase currently uses:
- **Backend**: Go/Fiber with REST endpoints + WebSocket
- **Frontend**: React with custom ApiService class

Since the backend is Go (not Node.js), we have two options:
1. **Add a Node.js tRPC server** as a separate service that proxies to Go backend
2. **Use tRPC-like patterns** with generated TypeScript types from Go structs

**Recommended Approach**: Option 2 - Generate TypeScript types from Go structs and create a type-safe API client. This maintains the Go backend while providing tRPC-like developer experience.

## Implementation

### Backend Changes (Go)

1. **Add Type Generation** (`backend/cmd/typegen/main.go`)
   - Create a tool to export Go struct types to TypeScript
   - Parse repository DTOs and handler request/response types
   - Generate `frontend/src/types/api.generated.ts`

2. **Standardize API Response Types** (`backend/internal/api/types/responses.go`)
   - Create consistent response wrapper types
   - Define error response format
   - Export as TypeScript-compatible JSON schemas

3. **Add OpenAPI/Swagger Generation** (`backend/internal/api/docs/`)
   - Generate OpenAPI spec from handlers
   - Use spec to generate TypeScript types

### Frontend Changes

4. **Create Type-Safe API Client** (`frontend/src/services/trpc-client.ts`)
   - Create typed wrapper around existing ApiService
   - Use generated types for request/response typing
   - Implement procedure-style API calls

5. **Add Zod Schemas** (`frontend/src/schemas/api.ts`)
   - Define Zod schemas matching backend DTOs
   - Runtime validation for API responses
   - Type inference from schemas

6. **Create React Query Integration** (`frontend/src/hooks/useApi.ts`)
   - Create custom hooks using TanStack Query
   - Type-safe query and mutation hooks
   - Automatic cache invalidation

7. **Update Existing API Calls**
   - Migrate key API calls to use new typed client
   - Keep backward compatibility with existing ApiService

### Shared Types

8. **Generate Shared Type Definitions**
   - `frontend/src/types/api.generated.ts` - Auto-generated from Go
   - `frontend/src/types/api.ts` - Hand-written extensions

## Acceptance Criteria

- [ ] TypeScript types are generated from Go structs
- [ ] Frontend API client is fully typed
- [ ] Zod schemas validate API responses at runtime
- [ ] React Query hooks provide caching and loading states
- [ ] Existing API calls continue to work
- [ ] Type generation runs as part of build process
- [ ] Developer experience matches tRPC (autocomplete, type errors)

## Files to Create/Modify

**Create:**
- `backend/cmd/typegen/main.go` - Type generation tool
- `backend/internal/api/types/responses.go` - Standardized response types
- `frontend/src/services/trpc-client.ts` - Type-safe API client
- `frontend/src/schemas/api.ts` - Zod validation schemas
- `frontend/src/hooks/useApi.ts` - React Query hooks
- `frontend/src/types/api.generated.ts` - Generated types

**Modify:**
- `backend/Makefile` - Add type generation target
- `frontend/package.json` - Add zod, @tanstack/react-query dependencies
- `frontend/src/services/api.ts` - Update to use generated types

## Integration Points

- **Provides**: Type-safe API client, React Query hooks, runtime validation
- **Consumes**: Existing Go handlers and DTOs, ApiService
- **Conflicts**: None - additive changes only, existing API preserved
