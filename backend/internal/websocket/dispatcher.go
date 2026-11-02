package websocket

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/luponetn/readr-go/internal/engine"
	"github.com/luponetn/readr-go/internal/parsers"
)

func (c *Client) handleMessage(raw []byte) {
	var req RPCRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		c.sendResponse(NewErrorResponse(nil, CodeParseError, "Parse error: invalid JSON-RPC payload", err.Error()))
		return
	}

	if req.JSONRPC != "2.0" && req.JSONRPC != "" {
		c.sendResponse(NewErrorResponse(req.ID, CodeInvalidRequest, "Invalid Request: jsonrpc must be '2.0'", nil))
		return
	}

	switch req.Method {
	case "tools/call":
		c.handleToolCall(req.ID, req.Params)
	case "resources/read":
		c.handleResourceRead(req.ID, req.Params)
	case "tools/list":
		c.handleToolsList(req.ID)
	case "resources/list":
		c.handleResourcesList(req.ID)
	case "ping":
		c.sendResponse(NewSuccessResponse(req.ID, map[string]string{"status": "pong"}))
	default:
		c.sendResponse(NewErrorResponse(req.ID, CodeMethodNotFound, fmt.Sprintf("Method '%s' not found", req.Method), nil))
	}
}

func (c *Client) handleToolCall(id json.RawMessage, paramsRaw json.RawMessage) {
	var params ToolCallParams
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		c.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Invalid params: failed to parse ToolCallParams", err.Error()))
		return
	}

	switch params.Name {
	case "ask_document":
		c.handleAskDocument(id, params.Arguments)
	case "ingest_document":
		c.handleIngestDocument(id, params.Arguments)
	case "delete_document":
		c.handleDeleteDocument(id, params.Arguments)
	default:
		c.sendResponse(NewErrorResponse(id, CodeMethodNotFound, fmt.Sprintf("Tool '%s' not supported", params.Name), nil))
	}
}

func (c *Client) handleAskDocument(id json.RawMessage, argsRaw json.RawMessage) {
	var args AskDocumentArgs
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		c.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Invalid arguments for ask_document", err.Error()))
		return
	}

	if args.Query == "" {
		c.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Query cannot be empty", nil))
		return
	}

	// Send streaming progress notification frame immediately
	c.sendNotification(NewProgressNotification(args.ProgressToken, 50, "Scanning chunks..."))

	// Execute deterministic BM25 search
	chunks := c.store.GetAllChunks()
	searchResult := engine.Search(chunks, args.Query)

	// Record query in history
	historyItem := engine.ChatHistoryItem{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Query:     args.Query,
		Result:    searchResult,
	}
	_ = c.store.AddHistory(historyItem)

	// Send final JSON-RPC response
	c.sendResponse(NewSuccessResponse(id, searchResult))
}

func (c *Client) handleIngestDocument(id json.RawMessage, argsRaw json.RawMessage) {
	var args IngestDocumentArgs
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		c.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Invalid arguments for ingest_document", err.Error()))
		return
	}

	if args.Filename == "" {
		c.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Filename is required", nil))
		return
	}

	if args.Content == "" {
		c.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Content is required", nil))
		return
	}

	parsedText, err := parsers.ParseContent(args.Filename, args.FileType, args.Content)
	if err != nil {
		c.sendResponse(NewErrorResponse(id, CodeInternalError, fmt.Sprintf("Failed to parse content: %v", err), nil))
		return
	}

	docID := uuid.New().String()
	chunks := engine.ChunkText(parsedText, docID, args.Filename)

	meta := engine.DocumentMeta{
		ID:         docID,
		Filename:   args.Filename,
		FileType:   args.FileType,
		UploadedAt: time.Now(),
		ChunkCount: len(chunks),
		TotalChars: len(parsedText),
	}

	if err := c.store.Ingest(meta, chunks); err != nil {
		c.sendResponse(NewErrorResponse(id, CodeInternalError, fmt.Sprintf("Failed to store document: %v", err), nil))
		return
	}

	slog.Info("Document ingested successfully", "filename", args.Filename, "docID", docID, "chunks", len(chunks))

	c.sendResponse(NewSuccessResponse(id, map[string]interface{}{
		"document":   meta,
		"chunkCount": len(chunks),
		"message":    fmt.Sprintf("Successfully ingested %s with %d chunks", args.Filename, len(chunks)),
	}))
}

func (c *Client) handleDeleteDocument(id json.RawMessage, argsRaw json.RawMessage) {
	var args DeleteDocumentArgs
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		c.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Invalid arguments for delete_document", err.Error()))
		return
	}

	if args.DocumentID == "" {
		c.sendResponse(NewErrorResponse(id, CodeInvalidParams, "documentId is required", nil))
		return
	}

	if err := c.store.Delete(args.DocumentID); err != nil {
		c.sendResponse(NewErrorResponse(id, CodeInternalError, fmt.Sprintf("Failed to delete document: %v", err), nil))
		return
	}

	slog.Info("Document deleted", "docID", args.DocumentID)

	c.sendResponse(NewSuccessResponse(id, map[string]interface{}{
		"success":    true,
		"documentId": args.DocumentID,
	}))
}

func (c *Client) handleResourceRead(id json.RawMessage, paramsRaw json.RawMessage) {
	var params ResourceReadParams
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		c.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Invalid params for resources/read", err.Error()))
		return
	}

	switch params.URI {
	case "readr://documents":
		docs := c.store.GetDocuments()
		docsJSON, _ := json.Marshal(docs)
		c.sendResponse(NewSuccessResponse(id, map[string]interface{}{
			"documents": docs,
			"contents": []ResourceContent{
				{
					URI:      "readr://documents",
					MimeType: "application/json",
					Text:     string(docsJSON),
				},
			},
		}))

	case "readr://history":
		history := c.store.GetHistory()
		historyJSON, _ := json.Marshal(history)
		c.sendResponse(NewSuccessResponse(id, map[string]interface{}{
			"history": history,
			"contents": []ResourceContent{
				{
					URI:      "readr://history",
					MimeType: "application/json",
					Text:     string(historyJSON),
				},
			},
		}))

	default:
		c.sendResponse(NewErrorResponse(id, CodeInvalidParams, fmt.Sprintf("Resource URI '%s' not recognized", params.URI), nil))
	}
}
