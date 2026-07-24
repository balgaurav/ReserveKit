package inventory

import (
	"context"
	"errors"

	inventoryv1 "github.com/balgaurav/ReserveKit/gen/go/inventory/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer adapts the inventory domain to the generated protobuf contract.
type GRPCServer struct {
	inventoryv1.UnimplementedInventoryServiceServer
	store *Store
}

func NewGRPCServer(store *Store) *GRPCServer {
	return &GRPCServer{store: store}
}

func (s *GRPCServer) GetAvailability(_ context.Context, request *inventoryv1.GetAvailabilityRequest) (*inventoryv1.GetAvailabilityResponse, error) {
	if request.GetSku() == "" {
		return nil, status.Error(codes.InvalidArgument, "sku is required")
	}
	return &inventoryv1.GetAvailabilityResponse{
		Sku:               request.GetSku(),
		AvailableQuantity: int32(s.store.Availability(request.GetSku())),
	}, nil
}

func (s *GRPCServer) HoldStock(_ context.Context, request *inventoryv1.HoldStockRequest) (*inventoryv1.HoldStockResponse, error) {
	if request.GetReservationId() == "" || request.GetSku() == "" {
		return nil, status.Error(codes.InvalidArgument, "reservation_id and sku are required")
	}
	remaining, err := s.store.HoldStock(request.GetReservationId(), request.GetSku(), int(request.GetQuantity()))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidQuantity):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, ErrInsufficientStock):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		case errors.Is(err, ErrReservationConflict):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, "could not hold inventory")
		}
	}
	return &inventoryv1.HoldStockResponse{RemainingQuantity: int32(remaining)}, nil
}

func (s *GRPCServer) ReleaseStock(_ context.Context, request *inventoryv1.ReleaseStockRequest) (*inventoryv1.ReleaseStockResponse, error) {
	if request.GetReservationId() == "" {
		return nil, status.Error(codes.InvalidArgument, "reservation_id is required")
	}
	return &inventoryv1.ReleaseStockResponse{AvailableQuantity: int32(s.store.ReleaseStock(request.GetReservationId()))}, nil
}
