'use client';

import React, { useEffect, useState, useRef } from 'react';
import { Icons } from '../ui/icons';
import { CitationBadge } from './CitationBadge';
import { SourceDrawer } from './SourceDrawer';
import type { Message } from '../../lib/types';

interface MessageItemProps {
  message: Message;
  isLatest: boolean;
}

export function MessageItem({ message, isLatest }: MessageItemProps) {
  const isUser = message.role === 'user';
  const fullText = message.content;

  const [displayedWordCount, setDisplayedWordCount] = useState<number>(() => {
    if (!message.isStreaming && !isLatest) {
      return Infinity;
    }
    return 0;
  });
  const [isTyping, setIsTyping] = useState<boolean>(Boolean(message.isStreaming || isLatest));
  const [copied, setCopied] = useState<boolean>(false);

  const words = useRef<string[]>([]);
  words.current = fullText.split(/(\s+)/);

  useEffect(() => {
    if (!isUser && isTyping) {
      const totalTokens = words.current.length;
      if (displayedWordCount >= totalTokens) {
        setIsTyping(false);
        return;
      }

      const timer = setTimeout(() => {
        setDisplayedWordCount((prev) => {
          const next = prev + 1;
          if (next >= totalTokens) {
            setIsTyping(false);
          }
          return next;
        });
      }, 22);

      return () => clearTimeout(timer);
    }
  }, [displayedWordCount, isTyping, isUser]);

  useEffect(() => {
    if (message.isStreaming) {
      setIsTyping(true);
    }
  }, [message.isStreaming, message.content]);

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // ignore clipboard error
    }
  };

  const formatScore = (score: number) => {
    if (score <= 1) {
      return `${Math.round(score * 100)}%`;
    }
    return `${Math.round(score)}%`;
  };

  if (isUser) {
    return (
      <div className="w-full max-w-3xl mx-auto px-4 py-4">
        <div className="flex flex-col gap-1 items-start">
          <div className="font-mono text-[11px] uppercase tracking-wider text-neutral-400 dark:text-neutral-500">
            Query
          </div>
          <p className="font-mono text-base font-semibold text-black dark:text-white leading-relaxed">
            {message.content}
          </p>
        </div>
      </div>
    );
  }

  const searchResult = message.result;
  const primaryCitation = searchResult?.primaryExcerpt;
  const alternativeExcerpts = searchResult?.alternativeExcerpts || [];

  const textToDisplay =
    displayedWordCount === Infinity || !isTyping
      ? fullText
      : words.current.slice(0, displayedWordCount).join('');

  return (
    <div className="w-full max-w-3xl mx-auto px-4 py-5 border-t border-neutral-100 dark:border-neutral-900">
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between font-mono text-[11px] text-neutral-400 dark:text-neutral-500">
          <div className="flex items-center gap-2">
            <span className="uppercase tracking-wider font-semibold text-neutral-900 dark:text-neutral-100">
              VERBATIM PASSAGE
            </span>
            {searchResult?.latencyMs !== undefined && (
              <span className="text-neutral-400 dark:text-neutral-600">
                ({searchResult.latencyMs}ms)
              </span>
            )}
          </div>

          <button
            onClick={() => copyToClipboard(fullText)}
            aria-label="Copy extracted passage"
            className="flex items-center gap-1 hover:text-black dark:hover:text-white transition-colors"
          >
            {copied ? (
              <>
                <Icons.Check className="w-3.5 h-3.5 text-emerald-500" />
                <span className="text-emerald-500 text-[10px]">COPIED</span>
              </>
            ) : (
              <>
                <Icons.Copy className="w-3.5 h-3.5" />
                <span className="text-[10px]">COPY</span>
              </>
            )}
          </button>
        </div>

        <div className="relative text-base leading-relaxed text-neutral-900 dark:text-neutral-100 whitespace-pre-wrap selection:bg-neutral-800 selection:text-white dark:selection:bg-neutral-200 dark:selection:text-black font-serif sm:font-sans">
          <span>{textToDisplay}</span>
          {isTyping && <span className="typewriter-cursor" />}
        </div>

        {primaryCitation && (
          <CitationBadge citation={primaryCitation} formatScore={formatScore} />
        )}

        {alternativeExcerpts.length > 0 && (
          <SourceDrawer excerpts={alternativeExcerpts} formatScore={formatScore} />
        )}
      </div>
    </div>
  );
}
