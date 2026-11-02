package websocket

import "encoding/json"

// Standard JSON-RPC 2.0 error codes
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// RPCRequest represents an incoming JSON-RPC 2.0 request frame.
type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// RPCResponse represents an outgoing JSON-RPC 2.0 response frame.
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError captures standard JSON-RPC 2.0 error fields.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RPCNotification represents a server-to-client notification with no ID.
type RPCNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// ProgressNotificationParams conforms to MCP notifications/progress schema.
type ProgressNotificationParams struct {
	ProgressToken interface{} `json:"progressToken,omitempty"`
	Progress      int         `json:"progress"`
	Message       string      `json:"message"`
}

// ToolCallParams defines parameters received for "tools/call" method.
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// AskDocumentArgs captures arguments for the "ask_document" tool call.
type AskDocumentArgs struct {
	Query         string      `json:"query"`
	ProgressToken interface{} `json:"progressToken,omitempty"`
}

// IngestDocumentArgs captures arguments for the "ingest_document" tool call.
type IngestDocumentArgs struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	FileType string `json:"fileType"`
}

// DeleteDocumentArgs captures arguments for the "delete_document" tool call.
type DeleteDocumentArgs struct {
	DocumentID string `json:"documentId"`
}

// ResourceReadParams defines parameters received for "resources/read" method.
type ResourceReadParams struct {
	URI string `json:"uri"`
}

// ResourceContent matches standard MCP resource read payload.
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

// ResourceReadResult wraps resource content array.
type ResourceReadResult struct {
	Contents []ResourceContent `json:"contents"`
}

// NewSuccessResponse constructs a standard JSON-RPC 2.0 success frame.
func NewSuccessResponse(id json.RawMessage, result interface{}) *RPCResponse {
	return &RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse constructs a standard JSON-RPC 2.0 error frame.
func NewErrorResponse(id json.RawMessage, code int, message string, data interface{}) *RPCResponse {
	return &RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// NewProgressNotification constructs a streaming progress notification.
func NewProgressNotification(token interface{}, progress int, message string) *RPCNotification {
	return &RPCNotification{
		JSONRPC: "2.0",
		Method:  "notifications/progress",
		Params: ProgressNotificationParams{
			ProgressToken: token,
			Progress:      progress,
			Message:       message,
		},
	}
}
