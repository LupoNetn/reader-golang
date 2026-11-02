'use client';

import React, { useEffect, useRef } from 'react';
import { MessageItem } from './MessageItem';
import { Icons } from '../ui/icons';
import type { Message } from '../../lib/types';

interface ChatContainerProps {
  messages: Message[];
  isStreaming: boolean;
  onSelectSuggestion?: (query: string) => void;
  onOpenIngest?: () => void;
}

export function ChatContainer({
  messages,
  isStreaming,
  onSelectSuggestion,
  onOpenIngest,
}: ChatContainerProps) {
  const bottomRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom whenever messages change or during streaming
  useEffect(() => {
    if (bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, isStreaming]);

  const suggestions = [
    'What was the total revenue reported in Q4?',
    'What are the termination conditions in Section 12?',
    'Find exact specifications for maximum memory payload.',
  ];

  if (messages.length === 0) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center px-4 py-12 text-center">
        <div className="w-full max-w-md space-y-6">
          <div className="inline-flex p-3 rounded-full bg-neutral-100 dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 text-neutral-800 dark:text-neutral-200">
            <Icons.Terminal className="w-6 h-6" />
          </div>

          <div className="space-y-2">
            <h2 className="text-xl font-mono font-semibold tracking-tight text-black dark:text-white">
              Deterministic Passage Retrieval
            </h2>
            <p className="text-xs text-neutral-500 dark:text-neutral-400 font-sans leading-relaxed">
              Literal source passage extraction from ingested documents. Strict verbatim matches
              with citations, zero creative synthesis, zero hallucinations.
            </p>
          </div>

          {/* Quick suggestions */}
          <div className="space-y-2 pt-2 text-left">
            <div className="font-mono text-[10px] uppercase tracking-wider text-neutral-400 dark:text-neutral-600 px-1">
              Suggested queries
            </div>
            {suggestions.map((sug, i) => (
              <button
                key={i}
                onClick={() => onSelectSuggestion?.(sug)}
                className="w-full text-left px-3 py-2 rounded border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-950 hover:bg-neutral-50 dark:hover:bg-neutral-900 text-xs font-mono text-neutral-700 dark:text-neutral-300 transition-colors flex items-center justify-between group shadow-2xs"
              >
                <span className="truncate">{sug}</span>
                <Icons.ChevronRight className="w-3 h-3 text-neutral-400 opacity-0 group-hover:opacity-100 transition-opacity shrink-0" />
              </button>
            ))}
          </div>

          {/* Prompt to add doc if memory is empty */}
          {onOpenIngest && (
            <div className="pt-2">
              <button
                onClick={onOpenIngest}
                className="inline-flex items-center gap-1.5 text-xs font-mono text-neutral-500 hover:text-black dark:hover:text-white transition-colors"
              >
                <Icons.UploadCloud className="w-3.5 h-3.5" />
                <span>Upload a document to search (.txt, .md, .csv, .json)</span>
              </button>
            </div>
          )}
        </div>
      </div>
    );
  }

  return (
    <div ref={containerRef} className="flex-1 overflow-y-auto py-4">
      {messages.map((message, index) => (
        <MessageItem
          key={message.id}
          message={message}
          isLatest={index === messages.length - 1}
        />
      ))}
      <div ref={bottomRef} className="h-4" />
    </div>
  );
}
