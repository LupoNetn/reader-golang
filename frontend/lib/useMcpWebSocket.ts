'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { McpWebSocketClient, type ProgressState } from './wsClient';
import type {
  ConnectionStatus,
  DocumentMeta,
  SearchResult,
  ChatHistoryItem,
} from './types';

export type { ProgressState };
export function useMcpWebSocket(url: string = 'ws://localhost:3034/ws') {
  const [status, setStatus] = useState<ConnectionStatus>('connecting');
  const [currentProgress, setCurrentProgress] = useState<ProgressState | null>(null);

  const clientRef = useRef<McpWebSocketClient | null>(null);

  if (!clientRef.current) {
    clientRef.current = new McpWebSocketClient(url);
  }

  useEffect(() => {
    const client = clientRef.current;
    if (!client) return;

    client.setListeners(
      (newStatus) => setStatus(newStatus),
      (newProgress) => setCurrentProgress(newProgress)
    );

    client.connect();

    return () => {
      client.disconnect();
    };
  }, []);

  const reconnect = useCallback(() => {
    clientRef.current?.connect();
  }, []);

  const askDocument = useCallback(
    async (query: string, onProgress?: (msg: string, pct: number) => void): Promise<SearchResult> => {
      const client = clientRef.current;
      if (!client) throw new Error('Client not initialized');

      const res = await client.sendRequest<SearchResult>(
        'tools/call',
        {
          name: 'ask_document',
          arguments: { query },
        },
        onProgress
      );
      return res;
    },
    []
  );

  const ingestDocument = useCallback(
    async (filename: string, content: string, fileType: string = 'text/plain'): Promise<{ document: DocumentMeta; chunkCount: number }> => {
      const client = clientRef.current;
      if (!client) throw new Error('Client not initialized');

      return await client.sendRequest<{ document: DocumentMeta; chunkCount: number }>(
        'tools/call',
        {
          name: 'ingest_document',
          arguments: { filename, content, fileType },
        }
      );
    },
    []
  );

  const deleteDocument = useCallback(
    async (documentId: string): Promise<{ success: boolean; documentId: string }> => {
      const client = clientRef.current;
      if (!client) throw new Error('Client not initialized');

      return await client.sendRequest<{ success: boolean; documentId: string }>(
        'tools/call',
        {
          name: 'delete_document',
          arguments: { documentId },
        }
      );
    },
    []
  );

  const getDocuments = useCallback(async (): Promise<DocumentMeta[]> => {
    const client = clientRef.current;
    if (!client) return [];

    try {
      const res = await client.sendRequest<{ documents?: DocumentMeta[] }>(
        'resources/read',
        { uri: 'readr://documents' }
      );
      return res?.documents || [];
    } catch {
      return [];
    }
  }, []);

  const getHistory = useCallback(async (): Promise<ChatHistoryItem[]> => {
    const client = clientRef.current;
    if (!client) return [];

    try {
      const res = await client.sendRequest<{ history?: ChatHistoryItem[] }>(
        'resources/read',
        { uri: 'readr://history' }
      );
      return res?.history || [];
    } catch {
      return [];
    }
  }, []);

  return {
    status,
    currentProgress,
    askDocument,
    ingestDocument,
    deleteDocument,
    getDocuments,
    getHistory,
    reconnect,
  };
}
