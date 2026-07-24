package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("RESERVATION_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "reservation-api"})
	})

	slog.Info("reservation API listening", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("reservation API stopped", "error", err)
		os.Exit(1)
	}
}
