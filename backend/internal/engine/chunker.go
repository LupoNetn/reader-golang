package engine

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	tokenRegex       = regexp.MustCompile(`[a-z0-9]+`)
	sentenceEndRegex = regexp.MustCompile(`[.?!](\s+|$)`)
	paragraphRegex   = regexp.MustCompile(`\r?\n+`)
)


// Tokenize converts raw text into lowercase alphanumeric tokens.
func Tokenize(text string) []string {
	lower := strings.ToLower(text)
	matches := tokenRegex.FindAllString(lower, -1)
	if matches == nil {
		return []string{}
	}
	return matches
}

// FilterStopwords removes stopwords from a token list while preserving order.
func FilterStopwords(tokens []string) []string {
	var filtered []string
	for _, t := range tokens {
		if _, isStop := EnglishStopwords[t]; !isStop {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// CountTerms creates a frequency map of tokens.
func CountTerms(tokens []string) map[string]int {
	counts := make(map[string]int, len(tokens))
	for _, t := range tokens {
		counts[t]++
	}
	return counts
}

// SplitIntoSentences splits a text block into individual sentences at punctuation boundaries [.?!].
func SplitIntoSentences(text string) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}

	indices := sentenceEndRegex.FindAllStringIndex(trimmed, -1)
	if len(indices) == 0 {
		return []string{trimmed}
	}

	var sentences []string
	currStart := 0

	for _, idx := range indices {
		// idx[0] is index of punctuation mark [.?!]
		// idx[1] is end of trailing whitespace (or end of string)
		sentenceEnd := idx[0] + 1
		sentence := strings.TrimSpace(trimmed[currStart:sentenceEnd])
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
		currStart = idx[1]
	}

	if currStart < len(trimmed) {
		remainder := strings.TrimSpace(trimmed[currStart:])
		if remainder != "" {
			sentences = append(sentences, remainder)
		}
	}

	return sentences
}

// ChunkText splits plain text into deterministic document chunks (200 - 600 characters).
func ChunkText(content, docID, filename string) []DocumentChunk {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil
	}

	rawParagraphs := paragraphRegex.Split(trimmed, -1)
	var units []string

	for _, p := range rawParagraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if len(p) > 600 {
			sentences := SplitIntoSentences(p)
			units = append(units, sentences...)
		} else {
			units = append(units, p)
		}
	}

	if len(units) == 0 {
		return nil
	}

	var chunkTexts []string
	var currentChunk strings.Builder

	for _, unit := range units {
		unit = strings.TrimSpace(unit)
		if unit == "" {
			continue
		}

		if currentChunk.Len() == 0 {
			currentChunk.WriteString(unit)
			continue
		}

		// If adding unit stays within upper bound (~600 chars)
		if currentChunk.Len()+1+len(unit) <= 600 {
			currentChunk.WriteString(" ")
			currentChunk.WriteString(unit)
		} else {
			// If current chunk has reached reasonable minimum size (>= 200 chars), flush it
			if currentChunk.Len() >= 200 {
				chunkTexts = append(chunkTexts, currentChunk.String())
				currentChunk.Reset()
				currentChunk.WriteString(unit)
			} else {
				// Current chunk is small (<200 chars), combine even if slightly over 600 to prevent fragmented fragments
				if currentChunk.Len()+1+len(unit) <= 800 {
					currentChunk.WriteString(" ")
					currentChunk.WriteString(unit)
				} else {
					chunkTexts = append(chunkTexts, currentChunk.String())
					currentChunk.Reset()
					currentChunk.WriteString(unit)
				}
			}
		}
	}

	if currentChunk.Len() > 0 {
		chunkTexts = append(chunkTexts, currentChunk.String())
	}

	chunks := make([]DocumentChunk, 0, len(chunkTexts))
	for idx, text := range chunkTexts {
		tokens := Tokenize(text)
		terms := CountTerms(tokens)
		wordCount := countWords(text)
		if wordCount == 0 {
			wordCount = len(tokens)
		}

		chunkID := generateChunkID(docID, idx)
		chunks = append(chunks, DocumentChunk{
			ID:         chunkID,
			DocumentID: docID,
			Filename:   filename,
			ChunkIndex: idx,
			Content:    text,
			WordCount:  wordCount,
			Terms:      terms,
		})
	}

	return chunks
}

func countWords(s string) int {
	fields := strings.FieldsFunc(s, unicode.IsSpace)
	return len(fields)
}

func generateChunkID(docID string, idx int) string {
	var b strings.Builder
	b.WriteString(docID)
	b.WriteString("-chk-")
	b.Grow(4)
	var digits [10]byte
	i := len(digits)
	v := idx
	if v == 0 {
		b.WriteByte('0')
		return b.String()
	}
	for v > 0 {
		i--
		digits[i] = byte('0' + (v % 10))
		v /= 10
	}
	b.Write(digits[i:])
	return b.String()
}
