'use client';

import React, { useState } from 'react';
import { useMcpWebSocket } from '../lib/useMcpWebSocket';
import { useChatRetrieval } from '../lib/useChatRetrieval';
import { TopNav } from '../components/layout/TopNav';
import { Sidebar } from '../components/layout/Sidebar';
import { ChatContainer } from '../components/chat/ChatContainer';
import { ChatInput } from '../components/chat/ChatInput';
import { StreamProgress } from '../components/chat/StreamProgress';
import { MemoryView } from '../components/memory/MemoryView';
import { IngestModal } from '../components/modals/IngestModal';
import type { ChatHistoryItem } from '../lib/types';

export default function Home() {
  const ws = useMcpWebSocket('ws://localhost:3030/ws');

  const [activeView, setActiveView] = useState<'chat' | 'memory'>('chat');
  const [isSidebarOpen, setIsSidebarOpen] = useState<boolean>(false);
  const [isIngestModalOpen, setIsIngestModalOpen] = useState<boolean>(false);

  const retrieval = useChatRetrieval({
    status: ws.status,
    askDocument: ws.askDocument,
    ingestDocument: ws.ingestDocument,
    deleteDocument: ws.deleteDocument,
    getDocuments: ws.getDocuments,
    getHistory: ws.getHistory,
  });

  const onSelectHistoryItem = (item: ChatHistoryItem) => {
    setActiveView('chat');
    setIsSidebarOpen(false);
    retrieval.handleSelectHistory(item);
  };

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-white text-black dark:bg-black dark:text-white font-sans antialiased selection:bg-neutral-800 selection:text-white dark:selection:bg-neutral-200 dark:selection:text-black">
      <Sidebar
        isOpen={isSidebarOpen}
        onClose={() => setIsSidebarOpen(false)}
        activeView={activeView}
        onSelectView={(view) => {
          setActiveView(view);
          setIsSidebarOpen(false);
        }}
        history={retrieval.history}
        onSelectHistoryItem={onSelectHistoryItem}
        onClearHistory={() => {}}
        documentsCount={retrieval.documents.length}
      />

      <div className="flex flex-col flex-1 h-full min-w-0 overflow-hidden relative">
        <TopNav
          status={ws.status}
          isSidebarOpen={isSidebarOpen}
          onToggleSidebar={() => setIsSidebarOpen(!isSidebarOpen)}
          onOpenIngestModal={() => setIsIngestModalOpen(true)}
          activeView={activeView}
          onViewChange={setActiveView}
          onReconnect={ws.reconnect}
        />

        <main className="flex-1 overflow-hidden relative flex flex-col min-h-0">
          {activeView === 'chat' ? (
            <div className="flex flex-col h-full w-full relative">
              <ChatContainer
                messages={retrieval.messages}
                isStreaming={retrieval.isStreaming}
                onSelectSuggestion={(suggestion) => retrieval.handleSendQuery(suggestion)}
                onOpenIngest={() => setIsIngestModalOpen(true)}
              />

              <StreamProgress
                progress={ws.currentProgress}
                isStreaming={retrieval.isStreaming}
              />

              <ChatInput
                query={retrieval.query}
                onChange={retrieval.setQuery}
                onSubmit={() => retrieval.handleSendQuery()}
                isStreaming={retrieval.isStreaming}
                disabled={ws.status !== 'connected'}
              />
            </div>
          ) : (
            <MemoryView
              documents={retrieval.documents}
              isLoading={retrieval.isLoadingDocs}
              onDeleteDocument={retrieval.handleDelete}
              onOpenIngestModal={() => setIsIngestModalOpen(true)}
            />
          )}
        </main>
      </div>

      <IngestModal
        isOpen={isIngestModalOpen}
        onClose={() => setIsIngestModalOpen(false)}
        onIngest={retrieval.handleIngest}
      />
    </div>
  );
}
