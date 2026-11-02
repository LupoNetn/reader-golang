'use client';

import { useState, useEffect, useCallback } from 'react';
import type { Message, ChatHistoryItem, DocumentMeta, SearchResult } from './types';

interface UseChatRetrievalProps {
  status: string;
  askDocument: (q: string) => Promise<SearchResult>;
  ingestDocument: (fn: string, c: string, ft: string) => Promise<{ document: DocumentMeta; chunkCount: number }>;
  deleteDocument: (id: string) => Promise<{ success: boolean; documentId: string }>;
  getDocuments: () => Promise<DocumentMeta[]>;
  getHistory: () => Promise<ChatHistoryItem[]>;
}

export function useChatRetrieval({
  status,
  askDocument,
  ingestDocument,
  deleteDocument,
  getDocuments,
  getHistory,
}: UseChatRetrievalProps) {
  const [query, setQuery] = useState<string>('');
  const [isStreaming, setIsStreaming] = useState<boolean>(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [history, setHistory] = useState<ChatHistoryItem[]>([]);
  const [documents, setDocuments] = useState<DocumentMeta[]>([]);
  const [isLoadingDocs, setIsLoadingDocs] = useState<boolean>(false);

  useEffect(() => {
    if (status === 'connected') {
      setIsLoadingDocs(true);
      getDocuments()
        .then((docs) => {
          if (docs && docs.length > 0) setDocuments(docs);
        })
        .catch(() => {})
        .finally(() => setIsLoadingDocs(false));

      getHistory()
        .then((hist) => {
          if (hist && hist.length > 0) setHistory(hist);
        })
        .catch(() => {});
    }
  }, [status, getDocuments, getHistory]);

  const handleSendQuery = useCallback(async (customQuery?: string) => {
    const text = (customQuery || query).trim();
    if (!text || isStreaming) return;

    setQuery('');
    const userMsg: Message = {
      id: `msg-u-${Date.now()}`,
      role: 'user',
      content: text,
      timestamp: Date.now(),
    };

    setMessages((prev) => [...prev, userMsg]);
    setIsStreaming(true);

    try {
      const searchResult = await askDocument(text);
      const primary = searchResult.primaryMatch || searchResult.primaryExcerpt;
      const passageText = primary?.content || primary?.text || 'No literal passage found.';

      const assistantMsg: Message = {
        id: `msg-a-${Date.now()}`,
        role: 'assistant',
        content: passageText,
        timestamp: Date.now(),
        result: searchResult,
        isStreaming: true,
      };

      setMessages((prev) => [...prev, assistantMsg]);

      const newHistoryItem: ChatHistoryItem = {
        id: `hist-${Date.now()}`,
        query: text,
        timestamp: Date.now(),
        result: searchResult,
      };

      setHistory((prev) => [newHistoryItem, ...prev.slice(0, 49)]);
    } catch (err: unknown) {
      const errorMsg: Message = {
        id: `msg-err-${Date.now()}`,
        role: 'assistant',
        content: `Error: ${err instanceof Error ? err.message : 'WebSocket error'}. Ensure backend is running on :3030.`,
        timestamp: Date.now(),
      };
      setMessages((prev) => [...prev, errorMsg]);
    } finally {
      setIsStreaming(false);
    }
  }, [query, isStreaming, askDocument]);

  const handleIngest = useCallback(async (filename: string, contentBase64: string, fileType: string) => {
    const res = await ingestDocument(filename, contentBase64, fileType);
    setDocuments((prev) => [res.document, ...prev.filter((d) => d.id !== res.document.id)]);
  }, [ingestDocument]);

  const handleDelete = useCallback(async (id: string) => {
    await deleteDocument(id);
    setDocuments((prev) => prev.filter((d) => d.id !== id));
  }, [deleteDocument]);

  const handleSelectHistory = useCallback((item: ChatHistoryItem) => {
    const userMsg: Message = {
      id: `msg-hist-u-${item.id}`,
      role: 'user',
      content: item.query,
      timestamp: item.timestamp,
    };
    const primary = item.result?.primaryMatch || item.result?.primaryExcerpt;
    const assistantMsg: Message = {
      id: `msg-hist-a-${item.id}`,
      role: 'assistant',
      content: primary?.content || primary?.text || '',
      timestamp: item.timestamp,
      result: item.result,
    };
    setMessages([userMsg, assistantMsg]);
  }, []);

  return {
    query,
    setQuery,
    isStreaming,
    messages,
    history,
    documents,
    isLoadingDocs,
    handleSendQuery,
    handleIngest,
    handleDelete,
    handleSelectHistory,
  };
}

