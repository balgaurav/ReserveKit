package main

import (
	"log/slog"
	"net/http"
	"os"

	inventoryv1 "github.com/balgaurav/ReserveKit/gen/go/inventory/v1"
	"github.com/balgaurav/ReserveKit/internal/reservation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	port := os.Getenv("RESERVATION_PORT")
	if port == "" {
		port = "8080"
	}

	inventoryAddress := os.Getenv("INVENTORY_GRPC_ADDRESS")
	if inventoryAddress == "" {
		inventoryAddress = "localhost:9090"
	}
	connection, err := grpc.NewClient(inventoryAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("inventory gRPC client failed", "error", err)
		os.Exit(1)
	}
	defer connection.Close()
	service := reservation.NewService(reservation.NewGRPCClient(inventoryv1.NewInventoryServiceClient(connection)))

	slog.Info("reservation API listening", "port", port)
	if err := http.ListenAndServe(":"+port, reservation.NewHTTPHandler(service)); err != nil {
		slog.Error("reservation API stopped", "error", err)
		os.Exit(1)
	}
}
