package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/luponetn/readr-go/internal/engine"
)

func TestWebSocketJSONRPC_EndToEnd(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "ws_store.json")

	store, err := engine.NewStore(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	hub := NewHub(store)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWS(hub, store, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer ws.Close()

	// 1. Ingest a document via tools/call
	ingestReq := RPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`"req-1"`),
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "ingest_document",
			"arguments": {
				"filename": "consensus.txt",
				"content": "Raft is a consensus algorithm that manages a replicated log. Paxos is another consensus protocol.",
				"fileType": "text/plain"
			}
		}`),
	}

	if err := ws.WriteJSON(ingestReq); err != nil {
		t.Fatalf("failed to write ingest request: %v", err)
	}

	var ingestResp RPCResponse
	if err := ws.ReadJSON(&ingestResp); err != nil {
		t.Fatalf("failed to read ingest response: %v", err)
	}

	if ingestResp.Error != nil {
		t.Fatalf("ingest failed with error: %v", ingestResp.Error)
	}

	// 2. Query document via tools/call ask_document
	askReq := RPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`"req-2"`),
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "ask_document",
			"arguments": {
				"query": "Raft consensus replicated log",
				"progressToken": "token-123"
			}
		}`),
	}

	if err := ws.WriteJSON(askReq); err != nil {
		t.Fatalf("failed to write ask request: %v", err)
	}

	// First message should be the progress notification
	var rawFirst map[string]interface{}
	if err := ws.ReadJSON(&rawFirst); err != nil {
		t.Fatalf("failed to read first message: %v", err)
	}

	if rawFirst["method"] != "notifications/progress" {
		t.Fatalf("expected progress notification, got: %+v", rawFirst)
	}

	// Second message should be the final search result
	var askResp RPCResponse
	if err := ws.ReadJSON(&askResp); err != nil {
		t.Fatalf("failed to read ask response: %v", err)
	}

	if askResp.Error != nil {
		t.Fatalf("ask returned error: %v", askResp.Error)
	}

	resultBytes, _ := json.Marshal(askResp.Result)
	var searchResult engine.SearchResult
	if err := json.Unmarshal(resultBytes, &searchResult); err != nil {
		t.Fatalf("failed to unmarshal SearchResult: %v", err)
	}

	if searchResult.Status != "found" {
		t.Fatalf("expected status 'found', got '%s'", searchResult.Status)
	}

	if searchResult.PrimaryMatch == nil {
		t.Fatal("expected non-nil PrimaryMatch")
	}

	// 3. Read resources/read for readr://documents
	resReq := RPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`"req-3"`),
		Method:  "resources/read",
		Params:  json.RawMessage(`{"uri": "readr://documents"}`),
	}

	if err := ws.WriteJSON(resReq); err != nil {
		t.Fatalf("failed to write resources/read: %v", err)
	}

	var resResp RPCResponse
	if err := ws.ReadJSON(&resResp); err != nil {
		t.Fatalf("failed to read resource response: %v", err)
	}

	if resResp.Error != nil {
		t.Fatalf("resources/read error: %v", resResp.Error)
	}
}
