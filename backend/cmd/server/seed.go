package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luponetn/readr-go/internal/engine"
	"github.com/luponetn/readr-go/internal/parsers"
)

func findTextFilesDir() string {
	candidates := []string{
		"text files",
		filepath.Join("..", "text files"),
		filepath.Join("..", "..", "text files"),
		filepath.Join("..", "..", "..", "text files"),
	}
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				return abs
			}
		}
	}
	return ""
}

// autoSeedTextFiles scans the sample text files directory and indexes any files not yet in the store.
func autoSeedTextFiles(store *engine.Store) {
	dir := findTextFilesDir()
	if dir == "" {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	existingDocs := store.GetDocuments()
	indexedMap := make(map[string]bool, len(existingDocs))
	for _, d := range existingDocs {
		indexedMap[strings.ToLower(d.Filename)] = true
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if indexedMap[strings.ToLower(filename)] {
			continue
		}

		filePath := filepath.Join(dir, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		ext := strings.ToLower(filepath.Ext(filename))
		parsedText, err := parsers.ParseContent(filename, ext, string(data))
		if err != nil || len(parsedText) == 0 {
			continue
		}

		docID := uuid.New().String()
		chunks := engine.ChunkText(parsedText, docID, filename)
		if len(chunks) == 0 {
			continue
		}

		meta := engine.DocumentMeta{
			ID:         docID,
			Filename:   filename,
			FileType:   strings.TrimPrefix(ext, "."),
			UploadedAt: time.Now(),
			ChunkCount: len(chunks),
			TotalChars: len(parsedText),
		}

		if err := store.Ingest(meta, chunks); err == nil {
			slog.Info("Auto-indexed sample document", "filename", filename, "chunks", len(chunks))
			indexedMap[strings.ToLower(filename)] = true
		}
	}
}
