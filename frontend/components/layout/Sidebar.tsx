'use client';

import React from 'react';
import { Icons } from '../ui/icons';
import type { ChatHistoryItem } from '../../lib/types';

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
  activeView: 'chat' | 'memory';
  onSelectView: (view: 'chat' | 'memory') => void;
  history: ChatHistoryItem[];
  onSelectHistoryItem: (item: ChatHistoryItem) => void;
  onClearHistory: () => void;
  documentsCount: number;
}

export function Sidebar({
  isOpen,
  onClose,
  activeView,
  onSelectView,
  history,
  onSelectHistoryItem,
  onClearHistory,
  documentsCount,
}: SidebarProps) {
  return (
    <>
      {/* Mobile backdrop */}
      {isOpen && (
        <div
          onClick={onClose}
          className="fixed inset-0 z-30 bg-black/40 backdrop-blur-xs md:hidden"
        />
      )}

      {/* Sidebar container */}
      <aside
        className={`fixed md:sticky top-14 left-0 z-30 h-[calc(100vh-3.5rem)] w-72 bg-white dark:bg-black border-r border-neutral-200 dark:border-neutral-800 flex flex-col transition-transform duration-200 ease-in-out ${
          isOpen ? 'translate-x-0' : '-translate-x-full md:-translate-x-full'
        } ${isOpen ? 'block' : 'hidden md:hidden'}`}
      >
        {/* Navigation list */}
        <div className="p-3 border-b border-neutral-200 dark:border-neutral-800 space-y-1">
          <button
            onClick={() => {
              onSelectView('chat');
            }}
            className={`w-full flex items-center justify-between px-3 py-2 text-xs font-mono rounded transition-colors ${
              activeView === 'chat'
                ? 'bg-neutral-100 dark:bg-neutral-900 text-black dark:text-white font-medium'
                : 'text-neutral-500 hover:text-black dark:hover:text-white hover:bg-neutral-50 dark:hover:bg-neutral-950'
            }`}
          >
            <div className="flex items-center gap-2">
              <Icons.Terminal className="w-3.5 h-3.5" />
              <span>PASSAGE_QUERY</span>
            </div>
            <span className="text-[10px] text-neutral-400">/ask</span>
          </button>

          <button
            onClick={() => {
              onSelectView('memory');
            }}
            className={`w-full flex items-center justify-between px-3 py-2 text-xs font-mono rounded transition-colors ${
              activeView === 'memory'
                ? 'bg-neutral-100 dark:bg-neutral-900 text-black dark:text-white font-medium'
                : 'text-neutral-500 hover:text-black dark:hover:text-white hover:bg-neutral-50 dark:hover:bg-neutral-950'
            }`}
          >
            <div className="flex items-center gap-2">
              <Icons.Database className="w-3.5 h-3.5" />
              <span>DOCUMENT_MEMORY</span>
            </div>
            <span className="font-mono text-[10px] px-1.5 py-0.5 rounded border border-neutral-200 dark:border-neutral-800 text-neutral-500">
              {documentsCount}
            </span>
          </button>
        </div>

        {/* History Header */}
        <div className="px-3 pt-4 pb-2 flex items-center justify-between">
          <div className="flex items-center gap-1.5 text-neutral-400 dark:text-neutral-500 font-mono text-[11px] tracking-wider uppercase">
            <Icons.Clock className="w-3 h-3" />
            <span>Search History</span>
          </div>
          {history.length > 0 && (
            <button
              onClick={onClearHistory}
              title="Clear search history"
              className="text-[10px] font-mono text-neutral-400 hover:text-red-500 dark:text-neutral-600 dark:hover:text-red-400 transition-colors"
            >
              CLEAR
            </button>
          )}
        </div>

        {/* Past Search History List */}
        <div className="flex-1 overflow-y-auto px-2 space-y-0.5">
          {history.length === 0 ? (
            <div className="px-3 py-6 text-center text-xs font-mono text-neutral-400 dark:text-neutral-600">
              No previous searches recorded.
            </div>
          ) : (
            history.map((item) => (
              <button
                key={item.id}
                onClick={() => {
                  onSelectHistoryItem(item);
                }}
                className="w-full text-left px-2.5 py-2 rounded text-xs font-mono truncate text-neutral-600 dark:text-neutral-400 hover:text-black dark:hover:text-white hover:bg-neutral-100 dark:hover:bg-neutral-900 transition-colors group flex items-center justify-between"
              >
                <span className="truncate pr-2">{item.query}</span>
                <Icons.ChevronRight className="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity text-neutral-400 shrink-0" />
              </button>
            ))
          )}
        </div>

        {/* Footer info */}
        <div className="p-3 border-t border-neutral-200 dark:border-neutral-800 text-[10px] font-mono text-neutral-400 dark:text-neutral-600 flex items-center justify-between">
          <span>READR_CORE::v1.0</span>
          <span>DETERMINISTIC</span>
        </div>
      </aside>
    </>
  );
}
