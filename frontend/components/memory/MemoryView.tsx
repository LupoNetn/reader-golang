'use client';

import React, { useState } from 'react';
import { DocumentCard } from './DocumentCard';
import { Icons } from '../ui/icons';
import type { DocumentMeta } from '../../lib/types';

interface MemoryViewProps {
  documents: DocumentMeta[];
  onDeleteDocument: (id: string) => Promise<void>;
  onOpenIngestModal: () => void;
  isLoading?: boolean;
}

export function MemoryView({
  documents,
  onDeleteDocument,
  onOpenIngestModal,
  isLoading = false,
}: MemoryViewProps) {
  const [filterQuery, setFilterQuery] = useState('');

  const filteredDocs = documents.filter((doc) =>
    doc.filename.toLowerCase().includes(filterQuery.toLowerCase())
  );

  const totalChunks = documents.reduce((acc, d) => acc + (d.chunkCount || 0), 0);

  return (
    <div className="flex-1 overflow-y-auto px-4 py-8 max-w-5xl mx-auto w-full space-y-6">
      {/* Overview Stat Strip */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <div className="p-4 rounded border border-neutral-200 dark:border-neutral-800 bg-neutral-50/50 dark:bg-neutral-950/50 font-mono">
          <div className="text-[11px] text-neutral-400 dark:text-neutral-500 uppercase tracking-wider">
            Total Documents
          </div>
          <div className="text-2xl font-bold text-black dark:text-white mt-1">
            {documents.length}
          </div>
        </div>

        <div className="p-4 rounded border border-neutral-200 dark:border-neutral-800 bg-neutral-50/50 dark:bg-neutral-950/50 font-mono">
          <div className="text-[11px] text-neutral-400 dark:text-neutral-500 uppercase tracking-wider">
            Indexed Chunks
          </div>
          <div className="text-2xl font-bold text-black dark:text-white mt-1">
            {totalChunks}
          </div>
        </div>

        <div className="p-4 rounded border border-neutral-200 dark:border-neutral-800 bg-neutral-50/50 dark:bg-neutral-950/50 font-mono">
          <div className="text-[11px] text-neutral-400 dark:text-neutral-500 uppercase tracking-wider">
            Retrieval Mode
          </div>
          <div className="text-sm font-bold text-emerald-600 dark:text-emerald-400 mt-2 flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-emerald-500" />
            DETERMINISTIC_PASSAGE
          </div>
        </div>
      </div>

      {/* Action bar: Search filter + Add Document */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 pt-2">
        <div className="relative flex-1 max-w-md">
          <Icons.Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
          <input
            type="text"
            value={filterQuery}
            onChange={(e) => setFilterQuery(e.target.value)}
            placeholder="Filter documents by name..."
            className="w-full pl-9 pr-3 py-2 text-xs font-mono rounded border border-neutral-300 dark:border-neutral-800 bg-white dark:bg-neutral-950 text-black dark:text-white placeholder:text-neutral-400 focus:outline-none focus:border-black dark:focus:border-white transition-colors"
          />
        </div>

        <button
          onClick={onOpenIngestModal}
          className="flex items-center justify-center gap-1.5 px-3 py-2 text-xs font-mono font-medium rounded border border-neutral-900 bg-black text-white hover:bg-neutral-800 dark:border-neutral-200 dark:bg-white dark:text-black dark:hover:bg-neutral-100 transition-colors shadow-xs"
        >
          <Icons.Plus className="w-3.5 h-3.5" />
          <span>Ingest Document</span>
        </button>
      </div>

      {/* Document Grid */}
      {isLoading ? (
        <div className="py-16 text-center text-xs font-mono text-neutral-400">
          Loading document repository...
        </div>
      ) : filteredDocs.length === 0 ? (
        <div className="py-16 text-center rounded border border-dashed border-neutral-300 dark:border-neutral-800 p-8 space-y-3">
          <div className="inline-flex p-3 rounded-full bg-neutral-100 dark:bg-neutral-900 text-neutral-400">
            <Icons.FileText className="w-6 h-6" />
          </div>
          <p className="text-xs font-mono text-neutral-500 dark:text-neutral-400">
            {filterQuery
              ? `No documents matching "${filterQuery}"`
              : 'No documents ingested yet.'}
          </p>
          <button
            onClick={onOpenIngestModal}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono text-black dark:text-white underline hover:opacity-80"
          >
            + Upload your first source document
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {filteredDocs.map((doc) => (
            <DocumentCard
              key={doc.id}
              document={doc}
              onDelete={onDeleteDocument}
            />
          ))}
        </div>
      )}
    </div>
  );
}
