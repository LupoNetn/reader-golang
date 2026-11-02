package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/luponetn/readr-go/internal/engine"
	"github.com/luponetn/readr-go/internal/websocket"
)

func resolveDataPath() string {
	if env := os.Getenv("READR_DATA_PATH"); env != "" {
		return env
	}
	if _, err := os.Stat("data"); err == nil {
		return filepath.Join("data", "store.json")
	}
	if _, err := os.Stat(filepath.Join("backend", "data")); err == nil {
		return filepath.Join("backend", "data", "store.json")
	}
	return filepath.Join("data", "store.json")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting Readr deterministic backend engine...")

	dataPath := resolveDataPath()
	store, err := engine.NewStore(dataPath)
	if err != nil {
		slog.Error("Failed to initialize store", "path", dataPath, "error", err)
		os.Exit(1)
	}

	docCount, chunkCount := store.Stats()
	slog.Info("Store initialized", "dataPath", dataPath, "documents", docCount, "chunks", chunkCount)

	hub := websocket.NewHub(store)

	mux := http.NewServeMux()

	// WebSocket full-duplex endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWS(hub, store, w, r)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		docs, chunks := store.Stats()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"engine":    "readr-go-deterministic",
			"documents": docs,
			"chunks":    chunks,
			"timestamp": time.Now().UTC(),
		})
	})

	// Root welcome / status endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"service": "Readr Deterministic Engine",
			"ws":      "/ws",
			"health":  "/health",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3030"
	}
	addr := fmt.Sprintf(":%s", port)

	server := &http.Server{
		Addr:              addr,
		Handler:           corsMiddleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("Readr server listening", "port", port, "wsUrl", fmt.Sprintf("ws://localhost:%s/ws", port))
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	case sig := <-shutdown:
		slog.Info("Shutdown signal received, shutting down gracefully...", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Graceful shutdown failed, forcing close", "error", err)
			_ = server.Close()
		}
		slog.Info("Server stopped cleanly")
	}
}
