package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luponetn/readr-go/internal/parsers"
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

func TestBM25Search_EmployeesStore(t *testing.T) {
	store, err := NewStore("../../data/store.json")
	if err != nil {
		t.Skip("skipping store test if data/store.json not found")
	}
	chunks := store.GetAllChunks()
	query := "Who is the Principal Systems Architect in Zurich and what is their salary?"
	result := Search(chunks, query)
	t.Logf("Result: status=%s, searched=%d, primary=%+v", result.Status, result.TotalChunksSearched, result.PrimaryMatch)
	if result.Status != "found" {
		t.Errorf("expected status 'found', got '%s'", result.Status)
	}
}

func TestBM25Search_AllFiveSampleFiles(t *testing.T) {
	sampleDir := "../../../text files"
	if _, err := os.Stat(sampleDir); err != nil {
		sampleDir = "../../text files"
	}
	if _, err := os.Stat(sampleDir); err != nil {
		sampleDir = "text files"
	}
	entries, err := os.ReadDir(sampleDir)
	if err != nil {
		t.Skip("sample directory not found")
	}

	var allChunks []DocumentChunk
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(sampleDir, entry.Name()))
		if err != nil {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		parsed, err := parsers.ParseContent(entry.Name(), ext, string(data))
		if err != nil {
			t.Fatalf("failed to parse %s: %v", entry.Name(), err)
		}
		chunks := ChunkText(parsed, "doc-"+entry.Name(), entry.Name())
		allChunks = append(allChunks, chunks...)
	}

	testCases := []struct {
		name          string
		query         string
		expectedInDoc string
	}{
		{
			name:          "Infrastructure JSON Query",
			query:         "Which cluster supports 250000 concurrent websockets in Northern Virginia",
			expectedInDoc: "250000",
		},
		{
			name:          "PDF Liquidated Damages",
			query:         "What are the liquidated damages for leaking the BM25 retrieval code?",
			expectedInDoc: "500,000",
		},
		{
			name:          "CSV Elena Salary",
			query:         "Who is the Principal Systems Architect in Zurich and what is their salary?",
			expectedInDoc: "195000",
		},
		{
			name:          "Offer Letter Severance",
			query:         "What is the severance continuation period if terminated without Cause?",
			expectedInDoc: "severance",
		},
		{
			name:          "Architecture MD Guards",
			query:         "What are the minimum confidence guards and keyword match ratios?",
			expectedInDoc: "0.30",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := Search(allChunks, tc.query)
			if res.Status != "found" {
				t.Fatalf("query '%s' returned status '%s', expected 'found'", tc.query, res.Status)
			}
			if res.PrimaryMatch == nil {
				t.Fatalf("expected non-nil PrimaryMatch for query: %s", tc.query)
			}
			if !strings.Contains(res.PrimaryMatch.Content, tc.expectedInDoc) {
				t.Errorf("expected '%s' in content, got: %s", tc.expectedInDoc, res.PrimaryMatch.Content)
			}
			t.Logf("PASS [%s]: score=%.2f matchPct=%.0f%% excerpt=%.80s...",
				tc.name, res.PrimaryMatch.Score, res.PrimaryMatch.MatchPercentage, res.PrimaryMatch.Content)
		})
	}
}
