package reservation

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryv1 "github.com/balgaurav/ReserveKit/gen/go/inventory/v1"
	"github.com/balgaurav/ReserveKit/internal/inventory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestHTTPToGRPCReservationFlow(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	inventoryv1.RegisterInventoryServiceServer(grpcServer, inventory.NewGRPCServer(inventory.NewStore(map[string]int{"keyboard": 5})))
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(grpcServer.Stop)

	connection, err := grpc.NewClient("passthrough:///inventory", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })

	handler := NewHTTPHandler(NewService(NewGRPCClient(inventoryv1.NewInventoryServiceClient(connection))))
	request := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewBufferString(`{"sku":"keyboard","quantity":2,"ttlSeconds":60}`))
	request.Header.Set("Idempotency-Key", "integration-checkout")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", response.Code, response.Body.String())
	}

	availabilityRequest := httptest.NewRequest(http.MethodGet, "/api/inventory?sku=keyboard", nil)
	availabilityResponse := httptest.NewRecorder()
	handler.ServeHTTP(availabilityResponse, availabilityRequest)
	var availability struct {
		AvailableQuantity int `json:"availableQuantity"`
	}
	if err := json.NewDecoder(availabilityResponse.Body).Decode(&availability); err != nil {
		t.Fatal(err)
	}
	if availability.AvailableQuantity != 3 {
		t.Fatalf("availability = %d, want 3 after hold", availability.AvailableQuantity)
	}
}
