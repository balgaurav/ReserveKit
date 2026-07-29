package reservation

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPReservationLifecycle(t *testing.T) {
	handler := NewHTTPHandler(NewService(&fakeInventory{}))
	create := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewBufferString(`{"sku":"keyboard","quantity":2,"ttlSeconds":60}`))
	create.Header.Set("Idempotency-Key", "checkout-123")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, create)
	if response.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", response.Code, response.Body.String())
	}
	if response.Header().Get("X-Correlation-ID") == "" {
		t.Fatal("missing correlation ID response header")
	}
}

func TestHTTPRejectsMissingIdempotencyKey(t *testing.T) {
	handler := NewHTTPHandler(NewService(&fakeInventory{}))
	request := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewBufferString(`{"sku":"keyboard","quantity":1,"ttlSeconds":60}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}
