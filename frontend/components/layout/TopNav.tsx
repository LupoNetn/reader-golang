'use client';

import React, { useEffect, useState } from 'react';
import { Icons } from '../ui/icons';
import type { ConnectionStatus } from '../../lib/types';

interface TopNavProps {
  onToggleSidebar: () => void;
  isSidebarOpen: boolean;
  status: ConnectionStatus;
  activeView: 'chat' | 'memory';
  onViewChange: (view: 'chat' | 'memory') => void;
  onOpenIngestModal: () => void;
  onReconnect: () => void;
}

export function TopNav({
  onToggleSidebar,
  isSidebarOpen,
  status,
  activeView,
  onViewChange,
  onOpenIngestModal,
  onReconnect,
}: TopNavProps) {
  const [isDark, setIsDark] = useState<boolean>(true);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      setIsDark(document.documentElement.classList.contains('dark'));
    }
  }, []);

  const toggleTheme = () => {
    const nextDark = !isDark;
    setIsDark(nextDark);
    if (nextDark) {
      document.documentElement.classList.add('dark');
      localStorage.setItem('readr-theme', 'dark');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('readr-theme', 'light');
    }
  };

  return (
    <header className="sticky top-0 z-40 w-full h-14 border-b border-neutral-200 dark:border-neutral-800 bg-white/80 dark:bg-black/80 backdrop-blur-md px-4 flex items-center justify-between select-none">
      {/* Left: Sidebar toggle + Logo */}
      <div className="flex items-center gap-3">
        <button
          onClick={onToggleSidebar}
          aria-label="Toggle Navigation Sidebar"
          aria-expanded={isSidebarOpen}
          className="p-1.5 rounded text-neutral-600 hover:text-black dark:text-neutral-400 dark:hover:text-white hover:bg-neutral-100 dark:hover:bg-neutral-900 transition-colors focus:outline-none focus-visible:ring-1 focus-visible:ring-neutral-400"
        >
          <Icons.Sidebar className="w-4 h-4" />
        </button>

        <div className="flex items-center gap-2">
          <span className="font-mono font-bold tracking-tight text-sm uppercase text-black dark:text-white">
            READR
          </span>
          <span className="hidden sm:inline-block font-mono text-[10px] uppercase tracking-wider text-neutral-400 dark:text-neutral-500 border border-neutral-200 dark:border-neutral-800 px-1.5 py-0.5 rounded">
            PASSAGE_RETRIEVAL
          </span>
        </div>
      </div>

      {/* Center: View Switcher */}
      <div className="flex items-center bg-neutral-100 dark:bg-neutral-900 p-0.5 rounded border border-neutral-200 dark:border-neutral-800">
        <button
          onClick={() => onViewChange('chat')}
          className={`flex items-center gap-1.5 px-3 py-1 text-xs font-mono rounded transition-all ${
            activeView === 'chat'
              ? 'bg-white dark:bg-black text-black dark:text-white shadow-sm font-semibold'
              : 'text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-200'
          }`}
        >
          <Icons.Terminal className="w-3 h-3" />
          <span>QUERY</span>
        </button>
        <button
          onClick={() => onViewChange('memory')}
          className={`flex items-center gap-1.5 px-3 py-1 text-xs font-mono rounded transition-all ${
            activeView === 'memory'
              ? 'bg-white dark:bg-black text-black dark:text-white shadow-sm font-semibold'
              : 'text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-200'
          }`}
        >
          <Icons.Database className="w-3 h-3" />
          <span>DOCUMENTS</span>
        </button>
      </div>

      {/* Right: WS Status + Add Doc Button + Theme Switcher */}
      <div className="flex items-center gap-3">
        {/* WS Status Badge */}
        <button
          onClick={status === 'disconnected' ? onReconnect : undefined}
          title={status === 'disconnected' ? 'Disconnected. Click to reconnect' : `WebSocket: ${status}`}
          className={`flex items-center gap-1.5 px-2 py-1 rounded border text-[11px] font-mono transition-colors ${
            status === 'connected'
              ? 'border-neutral-200 dark:border-neutral-800 text-neutral-700 dark:text-neutral-300'
              : status === 'connecting'
              ? 'border-amber-500/30 text-amber-600 dark:text-amber-400 bg-amber-500/5'
              : 'border-red-500/30 text-red-600 dark:text-red-400 bg-red-500/5 cursor-pointer hover:bg-red-500/10'
          }`}
        >
          <span className="relative flex h-2 w-2">
            {status === 'connected' && (
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
            )}
            <span
              className={`relative inline-flex rounded-full h-2 w-2 ${
                status === 'connected'
                  ? 'bg-emerald-500'
                  : status === 'connecting'
                  ? 'bg-amber-400 animate-pulse'
                  : 'bg-red-500'
              }`}
            />
          </span>
          <span className="hidden md:inline uppercase text-[10px] tracking-wider">
            {status === 'connected' ? 'WS:CONNECTED' : status === 'connecting' ? 'WS:SYNCING' : 'RECONNECT'}
          </span>
        </button>

        {/* Add Document button */}
        <button
          onClick={onOpenIngestModal}
          className="flex items-center gap-1.5 px-2.5 py-1 text-xs font-mono font-medium rounded border border-neutral-900 bg-black text-white hover:bg-neutral-800 dark:border-neutral-200 dark:bg-white dark:text-black dark:hover:bg-neutral-100 transition-colors shadow-sm"
        >
          <Icons.Plus className="w-3.5 h-3.5" />
          <span className="hidden sm:inline">Add Document</span>
        </button>

        {/* Theme toggle */}
        <button
          onClick={toggleTheme}
          aria-label="Toggle color theme"
          className="p-1.5 rounded text-neutral-600 hover:text-black dark:text-neutral-400 dark:hover:text-white hover:bg-neutral-100 dark:hover:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 transition-colors"
        >
          {isDark ? <Icons.Sun className="w-4 h-4" /> : <Icons.Moon className="w-4 h-4" />}
        </button>
      </div>
    </header>
  );
}
