package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/luponetn/readr-go/internal/engine"
)

// Hub manages active WebSocket client connections and message broadcasting.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
	store   *engine.Store
}

// NewHub creates a new WebSocket hub.
func NewHub(store *engine.Store) *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		store:   store,
	}
}

func (h *Hub) register(client *Client) {
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()
	slog.Info("WebSocket client connected", "remoteAddr", client.conn.RemoteAddr().String())
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
	h.mu.Unlock()
	slog.Info("WebSocket client disconnected", "remoteAddr", client.conn.RemoteAddr().String())
}

// BroadcastNotification sends an RPC notification to all connected clients.
func (h *Hub) BroadcastNotification(method string, params interface{}) {
	notification := &RPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(notification)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
		}
	}
}

