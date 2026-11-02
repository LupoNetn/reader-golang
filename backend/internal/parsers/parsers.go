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
	"unicode/utf8"
)

// ParseContent decodes content (raw or base64) and parses it into plain text according to its file type.
func ParseContent(filename, fileType, rawContent string) (string, error) {
	contentBytes := decodeContent(rawContent)

	ext := strings.ToLower(filepath.Ext(filename))
	ft := strings.ToLower(fileType)

	switch {
	case ext == ".csv" || ft == "csv" || ft == "text/csv":
		return ParseCSV(contentBytes)
	case ext == ".json" || ft == "json" || ft == "application/json":
		return ParseJSON(contentBytes)
	case ext == ".md" || ft == "md" || ft == "text/markdown":
		return ParseText(contentBytes)
	default:
		// Default to plain text / utf8 clean
		return ParseText(contentBytes)
	}
}

// decodeContent attempts to decode base64, falling back to raw bytes if decoding fails.
func decodeContent(raw string) []byte {
	trimmed := strings.TrimSpace(raw)
	// If it has a data URL prefix like "data:text/csv;base64,...", strip it
	if commaIdx := strings.Index(trimmed, ","); commaIdx != -1 && strings.HasPrefix(trimmed, "data:") {
		trimmed = trimmed[commaIdx+1:]
	}

	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err == nil && utf8.Valid(decoded) {
		return decoded
	}

	return []byte(raw)
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

	var sb strings.Builder
	for rowIdx := 1; rowIdx < len(records); rowIdx++ {
		row := records[rowIdx]
		var pairs []string

		for colIdx, cell := range row {
			cell = strings.TrimSpace(cell)
			if cell == "" {
				continue
			}

			if colIdx < len(headers) && headers[colIdx] != "" {
				pairs = append(pairs, fmt.Sprintf("%s: %s", headers[colIdx], cell))
			} else {
				pairs = append(pairs, cell)
			}
		}

		if len(pairs) > 0 {
			sb.WriteString(strings.Join(pairs, ", "))
			sb.WriteString("\n")
		}
	}

	return strings.TrimSpace(sb.String()), nil
}

// ParseJSON parses JSON arrays or objects into descriptive, token-rich text representations.
func ParseJSON(data []byte) (string, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return "", nil
	}

	// Try array of objects first
	var arr []interface{}
	if err := json.Unmarshal(trimmed, &arr); err == nil {
		var lines []string
		for idx, item := range arr {
			switch v := item.(type) {
			case map[string]interface{}:
				lines = append(lines, formatMap(v))
			default:
				lines = append(lines, fmt.Sprintf("Item %d: %v", idx+1, v))
			}
		}
		return strings.Join(lines, "\n\n"), nil
	}

	// Try single object
	var obj map[string]interface{}
	if err := json.Unmarshal(trimmed, &obj); err == nil {
		return formatMap(obj), nil
	}

	// Fallback to pretty-printed JSON
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, trimmed, "", "  "); err == nil {
		return pretty.String(), nil
	}

	return string(data), nil
}

func formatMap(m map[string]interface{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		val := m[k]
		switch v := val.(type) {
		case map[string]interface{}:
			pairs = append(pairs, fmt.Sprintf("%s: (%s)", k, formatMap(v)))
		case []interface{}:
			var items []string
			for _, elem := range v {
				items = append(items, fmt.Sprintf("%v", elem))
			}
			pairs = append(pairs, fmt.Sprintf("%s: [%s]", k, strings.Join(items, ", ")))
		default:
			pairs = append(pairs, fmt.Sprintf("%s: %v", k, v))
		}
	}
	return strings.Join(pairs, ", ")
}

// ParseText normalizes plain text and markdown by standardizing line endings and sanitizing UTF-8.
func ParseText(data []byte) (string, error) {
	str := string(data)
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, "\r", "\n")
	str = strings.ReplaceAll(str, "\x00", "")
	str = strings.ToValidUTF8(str, "")
	return strings.TrimSpace(str), nil
}
