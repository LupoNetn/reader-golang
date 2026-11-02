package engine

import (
	"strings"
	"testing"
)

func createSampleChunks() []DocumentChunk {
	doc1Text := `The Raft consensus algorithm is designed to be easy to understand. It is equivalent to Paxos in fault-tolerance and performance. The key elements are leader election, log replication, and safety. Raft decomposes consensus into relatively independent subproblems.`
	doc2Text := `Database indexing is a data structure technique used to quickly locate and access data in a database. B-Trees and LSM-Trees are standard storage engine architectures. LSM-Trees optimize for high write throughput using memory tables and SSTables.`
	doc3Text := `Sourdough bread is made by the fermentation of dough using wild lactobacillaceae and yeast. Sourdough baking requires flour, water, salt, and a thriving active sourdough starter culture.`

	chunks1 := ChunkText(doc1Text, "doc-raft", "raft.txt")
	chunks2 := ChunkText(doc2Text, "doc-db", "db.txt")
	chunks3 := ChunkText(doc3Text, "doc-bread", "bread.txt")

	all := append([]DocumentChunk{}, chunks1...)
	all = append(all, chunks2...)
	all = append(all, chunks3...)
	return all
}

func TestBM25Search_Found(t *testing.T) {
	chunks := createSampleChunks()
	query := "Raft consensus algorithm leader election"

	result := Search(chunks, query)

	if result.Status != "found" {
		t.Fatalf("expected status 'found', got '%s'", result.Status)
	}

	if result.PrimaryMatch == nil {
		t.Fatal("expected non-nil PrimaryMatch")
	}

	if result.PrimaryMatch.DocumentID != "doc-raft" {
		t.Errorf("expected doc-raft to be primary match, got %s", result.PrimaryMatch.DocumentID)
	}

	if !strings.Contains(result.PrimaryMatch.Content, "Raft consensus algorithm") {
		t.Errorf("primary match excerpt does not contain verbatim text: %s", result.PrimaryMatch.Content)
	}

	if result.PrimaryMatch.Score < MinimumConfidenceThreshold {
		t.Errorf("score %f below minimum threshold %f", result.PrimaryMatch.Score, MinimumConfidenceThreshold)
	}

	if result.PrimaryMatch.MatchPercentage < 40.0 {
		t.Errorf("match percentage %f below 40.0", result.PrimaryMatch.MatchPercentage)
	}

	if len(result.AllMatches) == 0 {
		t.Error("expected AllMatches to contain at least 1 match")
	}
}

func TestBM25Search_DatabaseIndexing(t *testing.T) {
	chunks := createSampleChunks()
	query := "LSM-Trees storage engine write throughput"

	result := Search(chunks, query)

	if result.Status != "found" {
		t.Fatalf("expected status 'found', got '%s'", result.Status)
	}

	if result.PrimaryMatch == nil {
		t.Fatal("expected non-nil PrimaryMatch")
	}

	if result.PrimaryMatch.DocumentID != "doc-db" {
		t.Errorf("expected doc-db to be primary match, got %s", result.PrimaryMatch.DocumentID)
	}
}

func TestBM25Search_NotFound_Irrelevant(t *testing.T) {
	chunks := createSampleChunks()
	// None of these terms exist in the sample corpus
	query := "quantum teleportation black hole singularity astrophysics"

	result := Search(chunks, query)

	if result.Status != "not_found" {
		t.Errorf("expected status 'not_found', got '%s'", result.Status)
	}

	if result.PrimaryMatch != nil {
		t.Errorf("expected nil PrimaryMatch, got %+v", result.PrimaryMatch)
	}

	if len(result.AllMatches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(result.AllMatches))
	}
}

func TestBM25Search_KeywordMatchRatioGuard(t *testing.T) {
	chunks := createSampleChunks()
	// Only 1 of 5 keywords matches ("bread" matches doc-bread, but 1/5 = 20% < 40% threshold)
	query := "bread quantum galaxy nebula gravitational"

	result := Search(chunks, query)

	// Since only 20% of keywords match, it must fail the 40% guard
	if result.Status != "not_found" {
		t.Errorf("expected guard to reject query with <40%% keyword coverage, got status '%s'", result.Status)
	}
}

func TestBM25Search_EmptyCorpus(t *testing.T) {
	result := Search([]DocumentChunk{}, "anything")
	if result.Status != "not_found" {
		t.Errorf("expected 'not_found' for empty corpus, got '%s'", result.Status)
	}
}

func TestBM25Search_EmptyQuery(t *testing.T) {
	chunks := createSampleChunks()
	result := Search(chunks, "   ")
	if result.Status != "not_found" {
		t.Errorf("expected 'not_found' for empty query, got '%s'", result.Status)
	}
}

func TestChunker_SentenceSplitting(t *testing.T) {
	longParagraph := "Distributed systems require fault tolerance. Network partitions will inevitably happen. Therefore, consensus algorithms such as Raft and Paxos provide safety guarantees even under adversity. Every node maintains a state machine and a replicated log."

	chunks := ChunkText(longParagraph, "doc-test", "test.txt")
	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk generated")
	}

	for _, chunk := range chunks {
		if chunk.WordCount <= 0 {
			t.Errorf("chunk %s has invalid word count: %d", chunk.ID, chunk.WordCount)
		}
		if len(chunk.Terms) == 0 {
			t.Errorf("chunk %s has empty terms map", chunk.ID)
		}
		if chunk.Content == "" {
			t.Errorf("chunk %s has empty content", chunk.ID)
		}
	}
}
