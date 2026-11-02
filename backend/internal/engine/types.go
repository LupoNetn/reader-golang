package engine

import "time"

// DocumentMeta holds top-level metadata for an ingested file.
type DocumentMeta struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	FileType   string    `json:"fileType"`
	UploadedAt time.Time `json:"uploadedAt"`
	ChunkCount int       `json:"chunkCount"`
	TotalChars int       `json:"totalChars"`
}

// DocumentChunk represents a distinct textual passage extracted from a document.
type DocumentChunk struct {
	ID         string         `json:"id"`
	DocumentID string         `json:"documentId"`
	Filename   string         `json:"filename"`
	ChunkIndex int            `json:"chunkIndex"`
	Content    string         `json:"content"`
	WordCount  int            `json:"wordCount"`
	Terms      map[string]int `json:"terms"` // Term frequencies for BM25 ranking
}

// MatchedPassage describes a chunk that satisfied search thresholds with its relevance metrics.
type MatchedPassage struct {
	ChunkID         string   `json:"chunkId"`
	DocumentID      string   `json:"documentId"`
	Filename        string   `json:"filename"`
	ChunkIndex      int      `json:"chunkIndex"`
	Content         string   `json:"content"`
	Score           float64  `json:"score"`
	MatchPercentage float64  `json:"matchPercentage"`
	MatchedKeywords []string `json:"matchedKeywords"`
}

// SearchResult represents the full output of a deterministic search execution.
type SearchResult struct {
	Query               string           `json:"query"`
	Status              string           `json:"status"` // "found" or "not_found"
	PrimaryMatch        *MatchedPassage  `json:"primaryMatch"`
	AllMatches          []MatchedPassage `json:"allMatches"`
	TotalChunksSearched int              `json:"totalChunksSearched"`
	ExecutionTimeMs     float64          `json:"executionTimeMs"`
}

// ChatHistoryItem records a single deterministic query and its result.
type ChatHistoryItem struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Query     string        `json:"query"`
	Result    *SearchResult `json:"result"`
}
