package reservation

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(service *Service) http.Handler {
	handler := &HTTPHandler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.health)
	mux.HandleFunc("GET /api/inventory", handler.availability)
	mux.HandleFunc("POST /api/reservations", handler.create)
	mux.HandleFunc("GET /api/reservations/{id}", handler.get)
	mux.HandleFunc("POST /api/reservations/{id}/confirm", handler.confirm)
	mux.HandleFunc("POST /api/reservations/{id}/release", handler.release)
	return correlationID(mux)
}

func (h *HTTPHandler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "reservation-api"})
}

func (h *HTTPHandler) availability(w http.ResponseWriter, request *http.Request) {
	sku := strings.TrimSpace(request.URL.Query().Get("sku"))
	quantity, err := h.service.Availability(request.Context(), sku)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sku": sku, "availableQuantity": quantity})
}

func (h *HTTPHandler) create(w http.ResponseWriter, request *http.Request) {
	var input struct {
		SKU        string `json:"sku"`
		Quantity   int    `json:"quantity"`
		TTLSeconds int    `json:"ttlSeconds"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, request.Body, 1<<20)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if input.TTLSeconds == 0 {
		input.TTLSeconds = 900
	}
	if input.TTLSeconds < 30 || input.TTLSeconds > 3600 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ttlSeconds must be between 30 and 3600"})
		return
	}
	created, err := h.service.Create(request.Context(), CreateInput{
		ID: newID(), SKU: strings.TrimSpace(input.SKU), Quantity: input.Quantity,
		IdempotencyKey: strings.TrimSpace(request.Header.Get("Idempotency-Key")),
		ExpiresAt:      time.Now().Add(time.Duration(input.TTLSeconds) * time.Second),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *HTTPHandler) get(w http.ResponseWriter, request *http.Request) {
	reservation, err := h.service.Get(request.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reservation)
}

func (h *HTTPHandler) confirm(w http.ResponseWriter, request *http.Request) {
	reservation, err := h.service.Confirm(request.Context(), request.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reservation)
}

func (h *HTTPHandler) release(w http.ResponseWriter, request *http.Request) {
	reservation, err := h.service.Release(request.Context(), request.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reservation)
}

func writeError(w http.ResponseWriter, err error) {
	statusCode := http.StatusBadGateway
	switch {
	case errors.Is(err, ErrInvalidInput):
		statusCode = http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, ErrInvalidTransition):
		statusCode = http.StatusConflict
	}
	writeJSON(w, statusCode, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}

func correlationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		id := request.Header.Get("X-Correlation-ID")
		if id == "" {
			id = newID()
		}
		w.Header().Set("X-Correlation-ID", id)
		next.ServeHTTP(w, request)
	})
}

func newID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(bytes)
}
