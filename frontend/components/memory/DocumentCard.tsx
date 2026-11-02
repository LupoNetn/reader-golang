'use client';

import React, { useState } from 'react';
import { Icons } from '../ui/icons';
import type { DocumentMeta } from '../../lib/types';

interface DocumentCardProps {
  document: DocumentMeta;
  onDelete: (id: string) => Promise<void>;
}

export function DocumentCard({ document, onDelete }: DocumentCardProps) {
  const [isDeleting, setIsDeleting] = useState<boolean>(false);
  const [confirmDelete, setConfirmDelete] = useState<boolean>(false);

  const handleDelete = async () => {
    if (!confirmDelete) {
      setConfirmDelete(true);
      return;
    }

    try {
      setIsDeleting(true);
      await onDelete(document.id);
    } catch {
      setIsDeleting(false);
      setConfirmDelete(false);
    }
  };

  const formatBytes = (bytes?: number) => {
    if (!bytes) return null;
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatDate = (dateStr: string) => {
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
      });
    } catch {
      return dateStr;
    }
  };

  return (
    <div className="p-4 rounded border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-950 hover:border-neutral-300 dark:hover:border-neutral-700 transition-colors shadow-2xs flex flex-col justify-between gap-3">
      {/* Top row: Filename & Delete */}
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-start gap-2.5 min-w-0">
          <div className="p-2 rounded bg-neutral-100 dark:bg-neutral-900 text-neutral-800 dark:text-neutral-200 shrink-0 mt-0.5">
            <Icons.FileText className="w-4 h-4" />
          </div>

          <div className="min-w-0">
            <h4
              className="text-sm font-mono font-medium text-black dark:text-white truncate"
              title={document.filename}
            >
              {document.filename}
            </h4>
            <div className="flex items-center gap-2 mt-0.5 text-[11px] font-mono text-neutral-400">
              <span className="uppercase">{document.fileType || 'TXT'}</span>
              {document.sizeBytes && (
                <>
                  <span>•</span>
                  <span>{formatBytes(document.sizeBytes)}</span>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Delete action */}
        <div>
          {confirmDelete ? (
            <div className="flex items-center gap-1 font-mono text-[10px]">
              <button
                onClick={handleDelete}
                disabled={isDeleting}
                className="px-2 py-1 rounded bg-red-600 text-white hover:bg-red-700 transition-colors"
              >
                {isDeleting ? '...' : 'CONFIRM'}
              </button>
              <button
                onClick={() => setConfirmDelete(false)}
                className="px-1.5 py-1 text-neutral-400 hover:text-black dark:hover:text-white"
              >
                ✕
              </button>
            </div>
          ) : (
            <button
              onClick={handleDelete}
              aria-label={`Delete ${document.filename}`}
              className="p-1.5 rounded text-neutral-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950/30 transition-colors"
            >
              <Icons.Trash className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
      </div>

      {/* Preview if available */}
      {document.preview && (
        <p className="text-xs text-neutral-500 dark:text-neutral-400 line-clamp-2 font-mono bg-neutral-50 dark:bg-neutral-900/50 p-2 rounded border border-neutral-100 dark:border-neutral-900">
          {document.preview}
        </p>
      )}

      {/* Bottom stats row */}
      <div className="flex items-center justify-between pt-2 border-t border-neutral-100 dark:border-neutral-900 text-[11px] font-mono text-neutral-500 dark:text-neutral-400">
        <span className="flex items-center gap-1">
          <span className="font-semibold text-black dark:text-white">
            {document.chunkCount}
          </span>{' '}
          {document.chunkCount === 1 ? 'chunk' : 'chunks'} indexed
        </span>

        <span className="text-[10px] text-neutral-400">{formatDate(document.uploadedAt)}</span>
      </div>
    </div>
  );
}
