'use client';

import React, { useState, useRef, useCallback } from 'react';
import { Icons } from '../ui/icons';

interface IngestModalProps {
  isOpen: boolean;
  onClose: () => void;
  onIngest: (filename: string, contentBase64: string, fileType: string) => Promise<void>;
}

export function IngestModal({ isOpen, onClose, onIngest }: IngestModalProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [filePreview, setFilePreview] = useState<string>('');
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileProcess = useCallback((file: File) => {
    setError(null);
    const validExtensions = ['txt', 'md', 'csv', 'json'];
    const ext = file.name.split('.').pop()?.toLowerCase() || '';

    if (!validExtensions.includes(ext)) {
      setError(`Unsupported file type: .${ext}. Only .txt, .md, .csv, and .json are supported.`);
      return;
    }

    // Limit to 20MB
    if (file.size > 20 * 1024 * 1024) {
      setError('File is too large. Maximum size is 20MB.');
      return;
    }

    setSelectedFile(file);

    // Read first characters for preview
    if (file.name.toLowerCase().endsWith('.pdf')) {
      setFilePreview(`[PDF Document: ${file.name} - Text will be extracted deterministically upon ingestion]`);
    } else {
      const reader = new FileReader();
      reader.onload = (e) => {
        const text = (e.target?.result as string) || '';
        setFilePreview(text.slice(0, 500));
      };
      reader.readAsText(file);
    }
  }, []);

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFileProcess(e.dataTransfer.files[0]);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      handleFileProcess(e.target.files[0]);
    }
  };

  const handleSubmit = async () => {
    if (!selectedFile) return;

    try {
      setIsUploading(true);
      setError(null);

      const ext = selectedFile.name.split('.').pop()?.toLowerCase() || 'txt';

      // Convert file to Base64
      const reader = new FileReader();
      const base64Promise = new Promise<string>((resolve, reject) => {
        reader.onload = () => {
          const result = reader.result as string;
          // remove "data:*/*;base64," header
          const base64 = result.includes(',') ? result.split(',')[1] : result;
          resolve(base64);
        };
        reader.onerror = (err) => reject(err);
      });

      reader.readAsDataURL(selectedFile);
      const base64Data = await base64Promise;

      await onIngest(selectedFile.name, base64Data, ext);

      // Reset and close
      setSelectedFile(null);
      setFilePreview('');
      setIsUploading(false);
      onClose();
    } catch (err: unknown) {
      setIsUploading(false);
      setError(err instanceof Error ? err.message : 'Failed to ingest document');
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-150">
      <div className="w-full max-w-lg rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-black shadow-xl overflow-hidden font-mono flex flex-col">
        {/* Modal Header */}
        <div className="px-5 py-4 border-b border-neutral-200 dark:border-neutral-800 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Icons.UploadCloud className="w-4 h-4 text-black dark:text-white" />
            <h3 className="text-sm font-semibold text-black dark:text-white uppercase tracking-tight">
              Ingest Source Document
            </h3>
          </div>
          <button
            onClick={onClose}
            aria-label="Close modal"
            className="p-1 rounded text-neutral-400 hover:text-black dark:hover:text-white hover:bg-neutral-100 dark:hover:bg-neutral-900 transition-colors"
          >
            <Icons.X className="w-4 h-4" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="p-5 space-y-4">
          <input
            ref={fileInputRef}
            type="file"
            accept=".txt,.md,.csv,.json,.pdf"
            onChange={handleFileChange}
            className="hidden"
          />

          {/* Drag & drop zone */}
          <div
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            className={`border border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors flex flex-col items-center justify-center gap-2.5 ${
              isDragging
                ? 'border-black dark:border-white bg-neutral-100 dark:bg-neutral-900'
                : 'border-neutral-300 dark:border-neutral-800 hover:border-neutral-400 dark:hover:border-neutral-700 bg-neutral-50/50 dark:bg-neutral-950/50'
            }`}
          >
            <div className="p-3 rounded-full bg-neutral-100 dark:bg-neutral-900 text-neutral-600 dark:text-neutral-300">
              <Icons.UploadCloud className="w-6 h-6" />
            </div>

            <div>
              <p className="text-xs font-semibold text-black dark:text-white">
                Click to browse or drag file here
              </p>
              <p className="text-[11px] text-neutral-400 mt-1">
                Supports TXT, MD, CSV, JSON, PDF (max 20MB)
              </p>
            </div>
          </div>

          {/* File Selected feedback */}
          {selectedFile && (
            <div className="p-3 rounded border border-neutral-200 dark:border-neutral-800 bg-neutral-50 dark:bg-neutral-900/60 text-xs space-y-1">
              <div className="flex items-center justify-between text-black dark:text-white font-medium">
                <span className="truncate pr-2">{selectedFile.name}</span>
                <span className="text-[10px] text-neutral-400 shrink-0">
                  {(selectedFile.size / 1024).toFixed(1)} KB
                </span>
              </div>
              {filePreview && (
                <div className="text-[10px] text-neutral-500 dark:text-neutral-400 line-clamp-3 font-mono pt-1">
                  {filePreview}...
                </div>
              )}
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="p-2.5 rounded bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900 text-red-600 dark:text-red-400 text-xs">
              {error}
            </div>
          )}
        </div>

        {/* Modal Footer */}
        <div className="px-5 py-3 border-t border-neutral-200 dark:border-neutral-800 bg-neutral-50 dark:bg-neutral-950 flex items-center justify-end gap-2 text-xs">
          <button
            onClick={onClose}
            disabled={isUploading}
            className="px-3 py-1.5 rounded text-neutral-600 hover:text-black dark:text-neutral-400 dark:hover:text-white transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={!selectedFile || isUploading}
            className="px-4 py-1.5 rounded border border-neutral-900 bg-black text-white hover:bg-neutral-800 dark:border-neutral-200 dark:bg-white dark:text-black dark:hover:bg-neutral-100 disabled:opacity-40 transition-colors font-semibold flex items-center gap-1.5 shadow-xs"
          >
            {isUploading ? (
              <>
                <span className="animate-spin text-xs">⏳</span>
                <span>Ingesting...</span>
              </>
            ) : (
              <span>Ingest Document</span>
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
