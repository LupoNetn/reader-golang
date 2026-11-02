package websocket

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/luponetn/readr-go/internal/engine"
)

// Client represents a single full-duplex WebSocket connection.
type Client struct {
	hub   *Hub
	conn  *websocket.Conn
	send  chan []byte
	store *engine.Store
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("WebSocket unexpected close", "error", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				return
			}

			// Flush any queued messages in send channel
			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write([]byte{'\n'}); err != nil {
					return
				}
				if _, err := w.Write(<-c.send); err != nil {
					return
				}
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendNotification(notification *RPCNotification) {
	data, err := json.Marshal(notification)
	if err != nil {
		slog.Error("failed to marshal notification", "error", err)
		return
	}
	select {
	case c.send <- data:
	default:
		slog.Warn("client send buffer full, dropped notification")
	}
}

func (c *Client) sendResponse(response *RPCResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		slog.Error("failed to marshal rpc response", "error", err)
		return
	}
	select {
	case c.send <- data:
	default:
		slog.Warn("client send buffer full, dropped response")
	}
}

