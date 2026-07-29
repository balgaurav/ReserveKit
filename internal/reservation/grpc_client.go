package reservation

import (
	"context"

	inventoryv1 "github.com/balgaurav/ReserveKit/gen/go/inventory/v1"
)

type GRPCClient struct {
	client inventoryv1.InventoryServiceClient
}

func NewGRPCClient(client inventoryv1.InventoryServiceClient) *GRPCClient {
	return &GRPCClient{client: client}
}

func (c *GRPCClient) Availability(ctx context.Context, sku string) (int, error) {
	response, err := c.client.GetAvailability(ctx, &inventoryv1.GetAvailabilityRequest{Sku: sku})
	if err != nil {
		return 0, err
	}
	return int(response.GetAvailableQuantity()), nil
}

func (c *GRPCClient) Hold(ctx context.Context, reservationID, sku string, quantity int) (int, error) {
	response, err := c.client.HoldStock(ctx, &inventoryv1.HoldStockRequest{ReservationId: reservationID, Sku: sku, Quantity: int32(quantity)})
	if err != nil {
		return 0, err
	}
	return int(response.GetRemainingQuantity()), nil
}

func (c *GRPCClient) Release(ctx context.Context, reservationID string) (int, error) {
	response, err := c.client.ReleaseStock(ctx, &inventoryv1.ReleaseStockRequest{ReservationId: reservationID})
	if err != nil {
		return 0, err
	}
	return int(response.GetAvailableQuantity()), nil
}
