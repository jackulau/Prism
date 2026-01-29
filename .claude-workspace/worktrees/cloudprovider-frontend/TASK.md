---
id: cloudprovider-frontend
name: CloudProvider TypeScript Types
wave: 1
priority: 1
dependencies: []
estimated_hours: 2
tags:
- frontend
- typescript
- types
---

## Objective

Define TypeScript types for the CloudProvider interface to be used in the frontend React application.

## Context

The frontend in `frontend/src/types/index.ts` already defines types for Provider, Message, ToolCall, etc. We need to add corresponding types for CloudProvider that mirror the backend interface.

## Implementation

1. **Add types to**: `frontend/src/types/index.ts`
   ```typescript
   // CloudProvider types
   export interface CloudProvider {
     name: string;
     hasCredentials: boolean;
   }
   
   export interface CloudAgent {
     id: string;
     providerId: string;
     providerName: string;
     name: string;
     status: CloudAgentStatus;
     createdAt: Date;
     updatedAt?: Date;
     model?: string;
     systemPrompt?: string;
   }
   
   export type CloudAgentStatus = 'active' | 'idle' | 'terminated' | 'error';
   
   export interface CreateCloudAgentParams {
     provider: string;
     name?: string;
     systemPrompt?: string;
     model?: string;
     tools?: string[];
     metadata?: Record<string, string>;
   }
   
   export interface CloudProviderMessage {
     id: string;
     role: 'user' | 'assistant' | 'system' | 'tool';
     content: string;
     timestamp: Date;
     toolCalls?: ToolCall[];
     images?: CloudImageData[];
   }
   
   export interface CloudImageData {
     url?: string;
     base64?: string;
     mimeType?: string;
   }
   
   export interface CloudMessageChunk {
     delta?: string;
     toolCalls?: ToolCall[];
     finishReason?: string;
     error?: string;
   }
   ```

2. **Add API methods to**: `frontend/src/services/api.ts`
   ```typescript
   // CloudProvider API methods
   listCloudProviders(): Promise<CloudProvider[]>
   
   createCloudAgent(params: CreateCloudAgentParams): Promise<CloudAgent>
   getCloudAgent(agentId: string): Promise<CloudAgent>
   deleteCloudAgent(agentId: string): Promise<void>
   
   getCloudAgentMessages(agentId: string): Promise<CloudProviderMessage[]>
   sendCloudAgentMessage(agentId: string, message: string, images?: CloudImageData[]): Promise<boolean>
   ```

3. **Add WebSocket message types**: Update `MessageType` union
   ```typescript
   | 'cloud_agent.created'
   | 'cloud_agent.message'
   | 'cloud_agent.chunk'
   | 'cloud_agent.complete'
   | 'cloud_agent.error'
   ```

## Acceptance Criteria

- [ ] All CloudProvider types are defined
- [ ] Types match the backend interface structure
- [ ] API service methods are typed correctly
- [ ] WebSocket message types are added
- [ ] Types follow existing naming conventions

## Files to Create/Modify

- `frontend/src/types/index.ts` - Add CloudProvider types
- `frontend/src/services/api.ts` - Add API methods (stubs)

## Integration Points

- **Provides**: TypeScript types for CloudProvider
- **Consumes**: None (independent frontend types)
- **Conflicts**: Coordinate with any other frontend type changes
