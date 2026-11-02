package parsers

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ParseContent decodes content (raw or base64) and parses it according to its file extension.
func ParseContent(filename, fileType, rawContent string) (string, error) {
	contentBytes := decodeContent(rawContent)
	ext := strings.ToLower(filepath.Ext(filename))
	ft := strings.ToLower(fileType)

	switch {
	case ext == ".pdf" || ft == "pdf" || ft == "application/pdf":
		return ParsePDF(contentBytes)
	case ext == ".csv" || ft == "csv" || ft == "text/csv":
		return ParseCSV(contentBytes)
	case ext == ".json" || ft == "json" || ft == "application/json":
		return ParseJSON(contentBytes)
	default:
		return ParseText(contentBytes)
	}
}

func decodeContent(raw string) []byte {
	trimmed := strings.TrimSpace(raw)
	if commaIdx := strings.Index(trimmed, ","); commaIdx != -1 && strings.HasPrefix(trimmed, "data:") {
		trimmed = trimmed[commaIdx+1:]
	}

	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err == nil {
		return decoded
	}
	return []byte(raw)
}

// ParsePDF extracts plain text content across all pages from PDF byte data.
func ParsePDF(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}

	var buf bytes.Buffer
	b, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("failed to extract PDF plain text: %w", err)
	}

	if _, err := buf.ReadFrom(b); err != nil {
		return "", fmt.Errorf("failed to read extracted PDF stream: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// ParseCSV extracts rows and pairs each column header with its corresponding cell value.
func ParseCSV(data []byte) (string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("csv parse error: %w", err)
	}
	if len(records) == 0 {
		return "", nil
	}
	if len(records) == 1 {
		return strings.Join(records[0], ", "), nil
	}

	headers := records[0]
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	var lines []string
	for _, row := range records[1:] {
		var pairs []string
		for colIdx, cell := range row {
			cellVal := strings.TrimSpace(cell)
			if cellVal == "" {
				continue
			}
			header := "Field"
			if colIdx < len(headers) && headers[colIdx] != "" {
				header = headers[colIdx]
			}
			pairs = append(pairs, fmt.Sprintf("%s: %s", header, cellVal))
		}
		if len(pairs) > 0 {
			lines = append(lines, strings.Join(pairs, ", "))
		}
	}
	return strings.Join(lines, "\n"), nil
}

// ParseText cleans line endings and UTF-8 characters.
func ParseText(data []byte) (string, error) {
	text := string(data)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.TrimSpace(text), nil
}

// ParseJSON parses array of objects or key-value structures into descriptive lines.
func ParseJSON(data []byte) (string, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return "", nil
	}

	if trimmed[0] == '[' {
		var list []map[string]interface{}
		if err := json.Unmarshal(trimmed, &list); err == nil {
			var lines []string
			for _, obj := range list {
				var pairs []string
				keys := make([]string, 0, len(obj))
				for k := range obj {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					pairs = append(pairs, fmt.Sprintf("%s: %v", k, obj[k]))
				}
				if len(pairs) > 0 {
					lines = append(lines, strings.Join(pairs, ", "))
				}
			}
			return strings.Join(lines, "\n"), nil
		}
	}

	var single map[string]interface{}
	if err := json.Unmarshal(trimmed, &single); err == nil {
		var pairs []string
		keys := make([]string, 0, len(single))
		for k := range single {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			pairs = append(pairs, fmt.Sprintf("%s: %v", k, single[k]))
		}
		return strings.Join(pairs, ", "), nil
	}

	return ParseText(data)
}
