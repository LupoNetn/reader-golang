'use client';

import React, { useRef, useEffect } from 'react';
import { Icons } from '../ui/icons';

interface ChatInputProps {
  query: string;
  onChange: (val: string) => void;
  onSubmit: () => void;
  disabled?: boolean;
  isStreaming?: boolean;
}

export function ChatInput({
  query,
  onChange,
  onSubmit,
  disabled = false,
  isStreaming = false,
}: ChatInputProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Auto-resize textarea height up to 160px
  useEffect(() => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    textarea.style.height = 'auto';
    const nextHeight = Math.min(textarea.scrollHeight, 160);
    textarea.style.height = `${Math.max(nextHeight, 44)}px`;
  }, [query]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (query.trim() && !disabled && !isStreaming) {
        onSubmit();
      }
    }
  };

  const isButtonDisabled = disabled || isStreaming || !query.trim();

  return (
    <div className="w-full max-w-3xl mx-auto px-4 pb-4">
      <div className="relative flex items-end rounded-lg border border-neutral-300 dark:border-neutral-800 bg-white dark:bg-neutral-950 focus-within:border-black dark:focus-within:border-white focus-within:ring-1 focus-within:ring-black dark:focus-within:ring-white transition-all shadow-xs">
        <textarea
          ref={textareaRef}
          value={query}
          onChange={(e) => onChange(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Ask a question to extract verbatim passages... (Enter to submit)"
          disabled={disabled || isStreaming}
          rows={1}
          className="w-full resize-none bg-transparent py-3 pl-3.5 pr-12 text-sm font-mono placeholder:text-neutral-400 dark:placeholder:text-neutral-600 focus:outline-none disabled:opacity-50 text-black dark:text-white leading-relaxed"
        />

        <div className="absolute right-2 bottom-2 flex items-center">
          <button
            onClick={onSubmit}
            disabled={isButtonDisabled}
            aria-label="Send Query"
            className="p-1.5 rounded bg-black text-white hover:bg-neutral-800 dark:bg-white dark:text-black dark:hover:bg-neutral-200 disabled:opacity-20 disabled:hover:bg-black dark:disabled:hover:bg-white transition-all"
          >
            <Icons.ArrowUp className="w-4 h-4" />
          </button>
        </div>
      </div>

      <div className="mt-1.5 flex items-center justify-between px-1 text-[10px] font-mono text-neutral-400 dark:text-neutral-600 select-none">
        <span>Deterministic verbatim matching • Zero hallucination</span>
        <span className="hidden sm:inline">Press ↵ Enter to search</span>
      </div>
    </div>
  );
}
