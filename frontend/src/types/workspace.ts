export interface Workspace {
  id: string;
  name: string;
  path: string;
  files: WorkspaceFileNode[];
}

export interface WorkspaceFileNode {
  id: string;
  name: string;
  path: string;
  type: 'file' | 'directory';
  children?: WorkspaceFileNode[];
  size?: number;
  modified?: Date;
}

export interface FileContext {
  path: string;
  content: string;
  language: string;
  selection?: Range;
}

export interface EditSuggestion {
  id: string;
  path: string;
  original: string;
  modified: string;
  startLine: number;
  endLine: number;
  description: string;
  status: 'pending' | 'accepted' | 'rejected';
}

export interface Range {
  startLine: number;
  startColumn: number;
  endLine: number;
  endColumn: number;
}

export interface FileReference {
  path: string;
  line?: number;
  column?: number;
}

export interface MessagePart {
  type: 'text' | 'file-ref' | 'code-edit';
  content?: string;
  path?: string;
  line?: number;
  original?: string;
  modified?: string;
  language?: string;
}
