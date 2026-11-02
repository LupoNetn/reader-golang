'use client';

import React from 'react';
import type { ProgressState } from '../../lib/useMcpWebSocket';

interface StreamProgressProps {
  progress: ProgressState | null;
  isStreaming: boolean;
}

export function StreamProgress({ progress, isStreaming }: StreamProgressProps) {
  if (!isStreaming && !progress) return null;

  const message = progress?.message || 'Scanning ingested chunks & computing deterministic match...';
  const pct = Math.min(Math.max(progress?.percentage ?? 0, 0), 100);

  return (
    <div className="w-full max-w-3xl mx-auto px-4 py-2">
      <div className="rounded border border-neutral-300 dark:border-neutral-800 bg-neutral-50 dark:bg-neutral-950 p-2.5 font-mono text-xs flex flex-col gap-2 shadow-xs transition-all">
        <div className="flex items-center justify-between text-neutral-600 dark:text-neutral-400">
          <div className="flex items-center gap-2">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-neutral-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-neutral-900 dark:bg-neutral-100" />
            </span>
            <span className="truncate">{message}</span>
          </div>
          <span className="text-[11px] font-semibold text-black dark:text-white shrink-0">
            {pct > 0 ? `${pct}%` : 'SCANNING'}
          </span>
        </div>

        {/* Progress bar track */}
        <div className="w-full h-1 bg-neutral-200 dark:bg-neutral-800 rounded-full overflow-hidden">
          <div
            className={`h-full bg-black dark:bg-white rounded-full transition-all duration-200 ease-out ${
              pct === 0 ? 'w-1/3 animate-[pulse_1.2s_ease-in-out_infinite]' : ''
            }`}
            style={{ width: pct > 0 ? `${pct}%` : undefined }}
          />
        </div>
      </div>
    </div>
  );
}
