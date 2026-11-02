package websocket

import "encoding/json"

func (c *Client) handleToolsList(id json.RawMessage) {
	tools := []map[string]interface{}{
		{
			"name":        "ask_document",
			"description": "Performs deterministic mathematical BM25 ranking over ingested source passages.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":         map[string]interface{}{"type": "string"},
					"progressToken": map[string]interface{}{"type": "string"},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "ingest_document",
			"description": "Parses and ingests a document (CSV, TXT, MD, JSON) into deterministic chunks.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"filename": map[string]interface{}{"type": "string"},
					"content":  map[string]interface{}{"type": "string"},
					"fileType": map[string]interface{}{"type": "string"},
				},
				"required": []string{"filename", "content"},
			},
		},
		{
			"name":        "delete_document",
			"description": "Removes a document and cascades deletion across all its passages.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"documentId": map[string]interface{}{"type": "string"},
				},
				"required": []string{"documentId"},
			},
		},
	}

	c.sendResponse(NewSuccessResponse(id, map[string]interface{}{"tools": tools}))
}

func (c *Client) handleResourcesList(id json.RawMessage) {
	resources := []map[string]interface{}{
		{
			"uri":         "readr://documents",
			"name":        "Ingested Documents",
			"description": "List of all ingested document metadata records in the store.",
			"mimeType":    "application/json",
		},
		{
			"uri":         "readr://history",
			"name":        "Query History",
			"description": "List of recent search queries and their deterministic results.",
			"mimeType":    "application/json",
		},
	}

	c.sendResponse(NewSuccessResponse(id, map[string]interface{}{"resources": resources}))
}
