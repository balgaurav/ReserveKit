package main

import (
	"log/slog"
	"net"
	"os"

	inventoryv1 "github.com/balgaurav/ReserveKit/gen/go/inventory/v1"
	"github.com/balgaurav/ReserveKit/internal/inventory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	port := os.Getenv("INVENTORY_GRPC_PORT")
	if port == "" {
		port = "9090"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("inventory listener failed", "error", err)
		os.Exit(1)
	}
	server := grpc.NewServer()
	inventoryv1.RegisterInventoryServiceServer(server, inventory.NewGRPCServer(inventory.NewStore(map[string]int{
		"keyboard": 12,
		"monitor":  8,
		"webcam":   24,
	})))
	healthServer := health.NewServer()
	healthv1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("reservekit.inventory.v1.InventoryService", healthv1.HealthCheckResponse_SERVING)
	slog.Info("inventory gRPC service listening", "port", port)
	if err := server.Serve(listener); err != nil {
		slog.Error("inventory service stopped", "error", err)
		os.Exit(1)
	}
}
