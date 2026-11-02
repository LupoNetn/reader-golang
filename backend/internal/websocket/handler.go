package websocket

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/luponetn/readr-go/internal/engine"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 25 * time.Second
	maxMessageSize = 10 * 1024 * 1024 // 10MB to accommodate file uploads
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  65536,
	WriteBufferSize: 65536,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev and frontend integration
	},
}

// ServeWS upgrades the HTTP request to a WebSocket and initializes the client pumps.
func ServeWS(hub *Hub, store *engine.Store, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}

	client := &Client{
		hub:   hub,
		conn:  conn,
		send:  make(chan []byte, 256),
		store: store,
	}

	hub.register(client)

	go client.writePump()
	go client.readPump()
}
