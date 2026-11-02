'use client';

import React from 'react';
import { Icons } from '../ui/icons';
import type { MatchedPassage } from '../../lib/types';

interface CitationBadgeProps {
  citation: MatchedPassage;
  formatScore: (score: number) => string;
}

export function CitationBadge({ citation, formatScore }: CitationBadgeProps) {
  return (
    <div className="mt-1 flex flex-wrap items-center gap-2 pt-2 border-t border-neutral-100 dark:border-neutral-900 text-xs font-mono">
      {/* File Source */}
      <div className="flex items-center gap-1.5 px-2 py-0.5 rounded bg-neutral-100 dark:bg-neutral-900 text-neutral-800 dark:text-neutral-200 border border-neutral-200 dark:border-neutral-800">
        <Icons.FileText className="w-3 h-3 text-neutral-400" />
        <span className="font-medium truncate max-w-[200px]" title={citation.filename}>
          {citation.filename}
        </span>
      </div>

      {/* Chunk Index */}
      <div className="px-1.5 py-0.5 rounded bg-neutral-100 dark:bg-neutral-900 text-neutral-600 dark:text-neutral-400 border border-neutral-200 dark:border-neutral-800 text-[11px]">
        chunk #{citation.chunkIndex}
      </div>

      {/* Match Score */}
      <div className="px-1.5 py-0.5 rounded bg-neutral-100 dark:bg-neutral-900 text-black dark:text-white font-semibold border border-neutral-200 dark:border-neutral-800 text-[11px]">
        {formatScore(citation.score)} match
      </div>

      {/* Matched Keywords */}
      {citation.matchedKeywords && citation.matchedKeywords.length > 0 && (
        <div className="flex items-center gap-1 flex-wrap">
          {citation.matchedKeywords.map((kw, idx) => (
            <span
              key={idx}
              className="text-[10px] px-1.5 py-0.5 rounded border border-neutral-200 dark:border-neutral-800 text-neutral-500 dark:text-neutral-400"
            >
              #{kw}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

