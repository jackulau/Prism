export interface OrgWorkspace {
  id: string;
  name: string;
  organizationId: string;
  githubRepositoryName?: string;
  workerId?: string;
  currentBranch?: string;
  slackChannelId?: string;
  createdAt: string;
  updatedAt?: string;
}

export interface CreateOrgWorkspaceInput {
  name: string;
  githubRepositoryName?: string;
  slackChannelId?: string;
}

export interface UpdateOrgWorkspaceInput {
  name?: string;
  githubRepositoryName?: string;
  slackChannelId?: string;
}

export interface OrgWorkspaceListParams {
  limit?: number;
  offset?: number;
}

export interface OrgWorkspaceListResponse {
  workspaces: OrgWorkspace[];
  total: number;
  hasMore: boolean;
}
