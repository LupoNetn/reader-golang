export interface DocumentMeta {
  id: string;
  filename: string;
  fileType: string;
  chunkCount: number;
  sizeBytes?: number;
  uploadedAt: string;
  preview?: string;
}

export interface MatchedPassage {
  id?: string;
  chunkId?: string;
  documentId: string;
  filename: string;
  chunkIndex: number;
  content?: string;
  text?: string;
  score: number;
  matchPercentage?: number;
  matchedKeywords: string[];
  metadata?: Record<string, unknown>;
}

export interface SearchResult {
  query: string;
  status?: string;
  primaryMatch?: MatchedPassage;
  primaryExcerpt?: MatchedPassage;
  allMatches?: MatchedPassage[];
  alternativeExcerpts?: MatchedPassage[];
  totalChunksSearched?: number;
  executionTimeMs?: number;
  latencyMs?: number;
}

export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string; // for user: query text; for assistant: primary text or message
  timestamp: number;
  result?: SearchResult;
  isStreaming?: boolean;
}

export interface ChatHistoryItem {
  id: string;
  query: string;
  timestamp: number;
  result?: SearchResult;
}

export interface JSONRPCRequest {
  jsonrpc: '2.0';
  id: string;
  method: string;
  params?: Record<string, unknown>;
}

export interface JSONRPCError {
  code: number;
  message: string;
  data?: unknown;
}

export interface JSONRPCResponse<T = unknown> {
  jsonrpc: '2.0';
  id: string;
  result?: T;
  error?: JSONRPCError;
}

export interface ProgressNotification {
  jsonrpc: '2.0';
  method: 'notifications/progress' | 'progress' | '$/progress';
  params: {
    progressToken?: string | number;
    token?: string | number;
    message?: string;
    progress?: number;
    total?: number;
    percentage?: number;
  };
}

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected';
