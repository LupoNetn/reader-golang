package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestStore_IngestAndReload(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "store.json")

	store, err := NewStore(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	docMeta := DocumentMeta{
		ID:         "doc-1",
		Filename:   "sample.txt",
		FileType:   "text/plain",
		UploadedAt: time.Now(),
		ChunkCount: 2,
		TotalChars: 150,
	}

	chunks := []DocumentChunk{
		{
			ID:         "doc-1-chk-0",
			DocumentID: "doc-1",
			Filename:   "sample.txt",
			ChunkIndex: 0,
			Content:    "First paragraph of sample text.",
			WordCount:  5,
			Terms:      map[string]int{"first": 1, "paragraph": 1, "sample": 1, "text": 1},
		},
		{
			ID:         "doc-1-chk-1",
			DocumentID: "doc-1",
			Filename:   "sample.txt",
			ChunkIndex: 1,
			Content:    "Second paragraph of sample text.",
			WordCount:  5,
			Terms:      map[string]int{"second": 1, "paragraph": 1, "sample": 1, "text": 1},
		},
	}

	if err := store.Ingest(docMeta, chunks); err != nil {
		t.Fatalf("failed to ingest doc: %v", err)
	}

	docs, totalChunks := store.Stats()
	if docs != 1 || totalChunks != 2 {
		t.Fatalf("expected 1 doc and 2 chunks, got %d docs and %d chunks", docs, totalChunks)
	}

	// Verify persistence by loading into a new store instance
	reloadedStore, err := NewStore(storePath)
	if err != nil {
		t.Fatalf("failed to reload store: %v", err)
	}

	rDocs, rChunks := reloadedStore.Stats()
	if rDocs != 1 || rChunks != 2 {
		t.Fatalf("reloaded store expected 1 doc and 2 chunks, got %d and %d", rDocs, rChunks)
	}

	loadedChunks := reloadedStore.GetAllChunks()
	if len(loadedChunks) != 2 {
		t.Fatalf("expected 2 loaded chunks, got %d", len(loadedChunks))
	}
}

func TestStore_CascadeDelete(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "store.json")

	store, err := NewStore(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	doc1 := DocumentMeta{ID: "doc-1", Filename: "a.txt", UploadedAt: time.Now()}
	chunks1 := []DocumentChunk{{ID: "c1", DocumentID: "doc-1", Content: "Hello"}}

	doc2 := DocumentMeta{ID: "doc-2", Filename: "b.txt", UploadedAt: time.Now()}
	chunks2 := []DocumentChunk{{ID: "c2", DocumentID: "doc-2", Content: "World"}}

	_ = store.Ingest(doc1, chunks1)
	_ = store.Ingest(doc2, chunks2)

	docs, chunks := store.Stats()
	if docs != 2 || chunks != 2 {
		t.Fatalf("expected 2 docs and 2 chunks, got %d and %d", docs, chunks)
	}

	// Delete doc-1
	if err := store.Delete("doc-1"); err != nil {
		t.Fatalf("failed to delete doc-1: %v", err)
	}

	docs, chunks = store.Stats()
	if docs != 1 || chunks != 1 {
		t.Fatalf("expected 1 doc and 1 chunk after delete, got %d and %d", docs, chunks)
	}

	remaining := store.GetAllChunks()
	if remaining[0].DocumentID != "doc-2" {
		t.Errorf("expected doc-2 to remain, got %s", remaining[0].DocumentID)
	}
}

func TestStore_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "concurrent_store.json")

	store, err := NewStore(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	var wg sync.WaitGroup
	workers := 10
	iterations := 20

	// Concurrent writers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				docID := fmt.Sprintf("doc-%d-%d", workerID, j)
				meta := DocumentMeta{
					ID:         docID,
					Filename:   fmt.Sprintf("file_%d_%d.txt", workerID, j),
					UploadedAt: time.Now(),
				}
				chunks := []DocumentChunk{
					{
						ID:         fmt.Sprintf("%s_chk", docID),
						DocumentID: docID,
						Content:    "Sample concurrent chunk text content",
						WordCount:  5,
						Terms:      map[string]int{"sample": 1, "concurrent": 1},
					},
				}
				_ = store.Ingest(meta, chunks)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = store.GetAllChunks()
				_ = store.GetDocuments()
				_, _ = store.Stats()
			}
		}()
	}

	wg.Wait()

	docs, chunks := store.Stats()
	if docs != workers*iterations || chunks != workers*iterations {
		t.Fatalf("expected %d docs and chunks, got %d docs and %d chunks", workers*iterations, docs, chunks)
	}

	// Verify file on disk exists and is valid
	if _, err := os.Stat(storePath); err != nil {
		t.Fatalf("persisted store file does not exist: %v", err)
	}
}
