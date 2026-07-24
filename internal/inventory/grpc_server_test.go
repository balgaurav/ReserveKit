package inventory

import (
	"context"
	"net"
	"testing"

	inventoryv1 "github.com/balgaurav/ReserveKit/gen/go/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCServerHoldsInventory(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	inventoryv1.RegisterInventoryServiceServer(server, NewGRPCServer(NewStore(map[string]int{"keyboard": 4})))
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	connection, err := grpc.NewClient("passthrough:///inventory", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })

	client := inventoryv1.NewInventoryServiceClient(connection)
	response, err := client.HoldStock(context.Background(), &inventoryv1.HoldStockRequest{ReservationId: "res-1", Sku: "keyboard", Quantity: 2})
	if err != nil {
		t.Fatal(err)
	}
	if response.GetRemainingQuantity() != 2 {
		t.Fatalf("remaining quantity = %d, want 2", response.GetRemainingQuantity())
	}
}
