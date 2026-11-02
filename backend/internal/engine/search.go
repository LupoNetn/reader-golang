package engine

import (
	"math"
	"sort"
	"time"
)

const (
	BM25K1                     = 1.5
	BM25B                      = 0.75
	MinimumConfidenceThreshold = 0.30
	MinimumKeywordMatchRatio   = 0.40
	MaxSearchResults           = 5
)

// Search performs a mathematical BM25 ranking over all chunks in memory with deterministic thresholds.
func Search(chunks []DocumentChunk, query string) *SearchResult {
	startTime := time.Now()

	result := &SearchResult{
		Query:               query,
		Status:              "not_found",
		PrimaryMatch:        nil,
		AllMatches:          []MatchedPassage{},
		TotalChunksSearched: len(chunks),
		ExecutionTimeMs:     0,
	}

	defer func() {
		ms := float64(time.Since(startTime).Microseconds()) / 1000.0
		result.ExecutionTimeMs = ms
		result.LatencyMs = ms
	}()

	totalChunks := len(chunks)
	if totalChunks == 0 {
		return result
	}

	rawTokens := Tokenize(query)
	if len(rawTokens) == 0 {
		return result
	}

	// Extract unique tokens
	uniqueTokenMap := make(map[string]struct{}, len(rawTokens))
	var uniqueTokens []string
	for _, tok := range rawTokens {
		if _, exists := uniqueTokenMap[tok]; !exists {
			uniqueTokenMap[tok] = struct{}{}
			uniqueTokens = append(uniqueTokens, tok)
		}
	}

	// Filter stopwords for query keywords
	queryKeywords := FilterStopwords(uniqueTokens)
	if len(queryKeywords) == 0 {
		// If query contained only stopwords, fall back to unique raw tokens
		queryKeywords = uniqueTokens
	}

	// Compute average document length (avgdl)
	var totalDocLength float64
	for i := range chunks {
		length := float64(chunks[i].WordCount)
		if length <= 0 {
			length = 1.0
		}
		totalDocLength += length
	}
	avgdl := totalDocLength / float64(totalChunks)
	if avgdl <= 0 {
		avgdl = 1.0
	}

	// Calculate Document Frequency n(q) for each query keyword
	// and calculate IDF(q) = ln( (N - n(q) + 0.5) / (n(q) + 0.5) + 1 )
	idf := make(map[string]float64, len(queryKeywords))
	for _, q := range queryKeywords {
		var nQ int
		for i := range chunks {
			if count, ok := chunks[i].Terms[q]; ok && count > 0 {
				nQ++
			}
		}

		numerator := float64(totalChunks-nQ) + 0.5
		denominator := float64(nQ) + 0.5
		val := (numerator / denominator) + 1.0
		idf[q] = math.Log(val)
	}

	var candidateMatches []MatchedPassage

	// Evaluate BM25 for each chunk
	for i := range chunks {
		chunk := &chunks[i]

		// 1. Keyword match check
		var matchedKeywords []string
		for _, q := range queryKeywords {
			if count, ok := chunk.Terms[q]; ok && count > 0 {
				matchedKeywords = append(matchedKeywords, q)
			}
		}

		keywordMatchRatio := float64(len(matchedKeywords)) / float64(len(queryKeywords))
		if keywordMatchRatio < MinimumKeywordMatchRatio {
			continue
		}

		// 2. BM25 score calculation
		docLen := float64(chunk.WordCount)
		if docLen <= 0 {
			docLen = 1.0
		}

		var score float64
		for _, q := range queryKeywords {
			freq := float64(chunk.Terms[q])
			if freq <= 0 {
				continue
			}

			qIDF := idf[q]
			numerator := freq * (BM25K1 + 1.0)
			denominator := freq + BM25K1*(1.0-BM25B+BM25B*(docLen/avgdl))
			score += qIDF * (numerator / denominator)
		}

		if score < MinimumConfidenceThreshold {
			continue
		}

		// Match percentage represents normalized confidence based on keyword coverage and relative relevance
		matchPercentage := math.Round(keywordMatchRatio*1000.0) / 10.0

		candidateMatches = append(candidateMatches, MatchedPassage{
			ChunkID:         chunk.ID,
			DocumentID:      chunk.DocumentID,
			Filename:        chunk.Filename,
			ChunkIndex:      chunk.ChunkIndex,
			Content:         chunk.Content,
			Text:            chunk.Content,
			Score:           math.Round(score*10000.0) / 10000.0,
			MatchPercentage: matchPercentage,
			MatchedKeywords: matchedKeywords,
		})
	}

	if len(candidateMatches) == 0 {
		return result
	}

	// Sort descending by Score, then MatchPercentage descending, then ChunkIndex ascending
	sort.Slice(candidateMatches, func(i, j int) bool {
		if candidateMatches[i].Score != candidateMatches[j].Score {
			return candidateMatches[i].Score > candidateMatches[j].Score
		}
		if candidateMatches[i].MatchPercentage != candidateMatches[j].MatchPercentage {
			return candidateMatches[i].MatchPercentage > candidateMatches[j].MatchPercentage
		}
		return candidateMatches[i].ChunkIndex < candidateMatches[j].ChunkIndex
	})

	// Keep Top-K
	limit := MaxSearchResults
	if len(candidateMatches) < limit {
		limit = len(candidateMatches)
	}

	topMatches := candidateMatches[:limit]

	result.Status = "found"
	result.PrimaryMatch = &topMatches[0]
	result.PrimaryExcerpt = &topMatches[0]
	result.AllMatches = topMatches
	if len(topMatches) > 1 {
		result.AlternativeExcerpts = topMatches[1:]
	} else {
		result.AlternativeExcerpts = []MatchedPassage{}
	}

	return result
}
