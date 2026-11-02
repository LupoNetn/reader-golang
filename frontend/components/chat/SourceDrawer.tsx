'use client';

import React, { useState } from 'react';
import { Icons } from '../ui/icons';
import type { MatchedPassage } from '../../lib/types';

interface SourceDrawerProps {
  excerpts: MatchedPassage[];
  formatScore: (score: number) => string;
}

export function SourceDrawer({ excerpts, formatScore }: SourceDrawerProps) {
  const [isOpen, setIsOpen] = useState<boolean>(false);

  if (excerpts.length === 0) return null;

  return (
    <div className="mt-2">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-1.5 text-xs font-mono text-neutral-500 hover:text-black dark:hover:text-white transition-colors py-1 group"
      >
        <span>
          [{excerpts.length} more {excerpts.length === 1 ? 'source' : 'sources'}]
        </span>
        <Icons.ChevronDown
          className={`w-3.5 h-3.5 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`}
        />
      </button>

      {isOpen && (
        <div className="mt-2 space-y-3 pl-3 border-l-2 border-neutral-200 dark:border-neutral-800">
          {excerpts.map((alt: MatchedPassage, index: number) => (
            <div
              key={index}
              className="p-3 rounded border border-neutral-200 dark:border-neutral-800 bg-neutral-50/50 dark:bg-neutral-950/50 space-y-2 text-xs font-mono"
            >
              <div className="flex items-center justify-between text-neutral-400 text-[11px]">
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-neutral-800 dark:text-neutral-200">
                    {alt.filename}
                  </span>
                  <span>•</span>
                  <span>chunk #{alt.chunkIndex}</span>
                </div>
                <span className="text-black dark:text-white font-medium">
                  {formatScore(alt.score)} match
                </span>
              </div>

              <p className="text-neutral-800 dark:text-neutral-200 whitespace-pre-wrap leading-relaxed font-sans text-xs">
                {alt.text}
              </p>

              {alt.matchedKeywords && alt.matchedKeywords.length > 0 && (
                <div className="flex items-center gap-1 flex-wrap pt-1">
                  {alt.matchedKeywords.map((kw, kwIdx) => (
                    <span
                      key={kwIdx}
                      className="text-[9px] px-1 py-0.5 rounded border border-neutral-200 dark:border-neutral-800 text-neutral-500"
                    >
                      #{kw}
                    </span>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

