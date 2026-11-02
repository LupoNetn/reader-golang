package parsers

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestParseCSV_BindsHeaders(t *testing.T) {
	csvData := "Name,Department,Salary\nAlice,Engineering,120000\nBob,Design,95000"

	parsed, err := ParseCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(parsed, "Name: Alice") || !strings.Contains(parsed, "Department: Engineering") {
		t.Errorf("expected CSV headers bound to values, got:\n%s", parsed)
	}

	if !strings.Contains(parsed, "Salary: 120000") {
		t.Errorf("expected salary field, got:\n%s", parsed)
	}
}

func TestParseJSON_ArrayOfObjects(t *testing.T) {
	jsonData := `[
		{"service": "auth-service", "port": 8080, "status": "active"},
		{"service": "payment-service", "port": 8081, "status": "active"}
	]`

	parsed, err := ParseJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(parsed, "service: auth-service") || !strings.Contains(parsed, "port: 8080") {
		t.Errorf("expected parsed JSON tokens, got:\n%s", parsed)
	}
}

func TestParseContent_Base64Encoded(t *testing.T) {
	raw := "This is a base64 encoded document test."
	b64 := base64.StdEncoding.EncodeToString([]byte(raw))

	parsed, err := ParseContent("notes.txt", "text/plain", b64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed != raw {
		t.Errorf("expected '%s', got '%s'", raw, parsed)
	}
}
