package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("INVENTORY_HTTP_PORT")
	if port == "" {
		port = "8081"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "inventory"})
	})
	slog.Info("inventory service bootstrap listening", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("inventory service stopped", "error", err)
		os.Exit(1)
	}
}
