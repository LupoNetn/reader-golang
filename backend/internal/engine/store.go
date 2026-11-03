package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// StoreData represents the serializable representation of the store on disk.
type StoreData struct {
	Documents []DocumentMeta    `json:"documents"`
	Chunks    []DocumentChunk   `json:"chunks"`
	History   []ChatHistoryItem `json:"history"`
}

// Store provides a thread-safe in-memory cache of documents, chunks, and history with atomic disk persistence.
type Store struct {
	mu        sync.RWMutex
	dataPath  string
	documents map[string]DocumentMeta
	chunks    []DocumentChunk
	history   []ChatHistoryItem
}

func NewStore(dataPath string) (*Store, error) {
	s := &Store{
		dataPath:  dataPath,
		documents: make(map[string]DocumentMeta),
		chunks:    make([]DocumentChunk, 0),
		history:   make([]ChatHistoryItem, 0),
	}

	if err := s.load(); err != nil {
		return nil, fmt.Errorf("failed to load store from %s: %w", dataPath, err)
	}

	return s, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.dataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	fileBytes, err := os.ReadFile(s.dataPath)
	if os.IsNotExist(err) {
		// Initialize empty store file
		return s.flushLocked()
	} else if err != nil {
		return fmt.Errorf("failed to read store file: %w", err)
	}

	if len(fileBytes) == 0 {
		return s.flushLocked()
	}

	var data StoreData
	if err := json.Unmarshal(fileBytes, &data); err != nil {
		return fmt.Errorf("failed to unmarshal store data: %w", err)
	}

	for _, doc := range data.Documents {
		s.documents[doc.ID] = doc
	}
	s.chunks = data.Chunks
	s.history = data.History

	return nil
}

func (s *Store) flushLocked() error {
	dir := filepath.Dir(s.dataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data dir %s: %w", dir, err)
	}

	docsList := make([]DocumentMeta, 0, len(s.documents))
	for _, doc := range s.documents {
		docsList = append(docsList, doc)
	}

	// Sort documents by UploadedAt descending
	sort.Slice(docsList, func(i, j int) bool {
		return docsList[i].UploadedAt.After(docsList[j].UploadedAt)
	})

	data := StoreData{
		Documents: docsList,
		Chunks:    s.chunks,
		History:   s.history,
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal store data: %w", err)
	}

	tmpPath := s.dataPath + ".tmp"
	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return fmt.Errorf("failed to write tmp store file: %w", err)
	}

	_ = os.Remove(s.dataPath)
	if err := os.Rename(tmpPath, s.dataPath); err != nil {
		// Fallback for Windows locking anomalies
		if errDirect := os.WriteFile(s.dataPath, b, 0644); errDirect != nil {
			return fmt.Errorf("failed to replace store file: %w", err)
		}
		_ = os.Remove(tmpPath)
	}

	return nil
}

func (s *Store) Ingest(meta DocumentMeta, chunks []DocumentChunk) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.documents[meta.ID]; exists {
		filtered := make([]DocumentChunk, 0, len(s.chunks))
		for _, c := range s.chunks {
			if c.DocumentID != meta.ID {
				filtered = append(filtered, c)
			}
		}
		s.chunks = filtered
	}

	s.documents[meta.ID] = meta
	s.chunks = append(s.chunks, chunks...)

	return s.flushLocked()
}

func (s *Store) Delete(docID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.documents[docID]; !exists {
		return nil
	}

	delete(s.documents, docID)

	filtered := make([]DocumentChunk, 0, len(s.chunks))
	for _, c := range s.chunks {
		if c.DocumentID != docID {
			filtered = append(filtered, c)
		}
	}
	s.chunks = filtered

	return s.flushLocked()
}

func (s *Store) GetAllChunks() []DocumentChunk {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]DocumentChunk, len(s.chunks))
	copy(result, s.chunks)
	return result
}

func (s *Store) GetDocuments() []DocumentMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docs := make([]DocumentMeta, 0, len(s.documents))
	for _, doc := range s.documents {
		docs = append(docs, doc)
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].UploadedAt.After(docs[j].UploadedAt)
	})

	return docs
}

func (s *Store) GetDocument(id string) (DocumentMeta, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, exists := s.documents[id]
	return doc, exists
}

func (s *Store) AddHistory(item ChatHistoryItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = append(s.history, item)

	// Keep history bounded to 500 most recent queries
	if len(s.history) > 500 {
		s.history = s.history[len(s.history)-500:]
	}

	return s.flushLocked()
}

func (s *Store) GetHistory() []ChatHistoryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]ChatHistoryItem, len(s.history))
	copy(history, s.history)

	// Reverse to show newest first
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history
}

// Stats returns the total count of documents and chunks.
func (s *Store) Stats() (docCount int, chunkCount int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.documents), len(s.chunks)
}
