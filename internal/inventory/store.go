package inventory

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrInvalidQuantity     = errors.New("quantity must be greater than zero")
	ErrInsufficientStock   = errors.New("insufficient available stock")
	ErrReservationConflict = errors.New("reservation ID is already associated with another hold")
)

type Hold struct {
	ReservationID string
	SKU           string
	Quantity      int
}

// Store owns available stock and the active holds against it. The mutex keeps a
// hold atomic inside this service boundary; a database transaction replaces it
// when the Postgres adapter is enabled.
type Store struct {
	mu    sync.Mutex
	stock map[string]int
	holds map[string]Hold
}

func NewStore(seed map[string]int) *Store {
	stock := make(map[string]int, len(seed))
	for sku, quantity := range seed {
		stock[sku] = quantity
	}
	return &Store{stock: stock, holds: make(map[string]Hold)}
}

func (s *Store) Availability(sku string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stock[sku]
}

// HoldStock is idempotent for a repeat of the same reservation ID, SKU, and
// quantity. It never decrements stock twice for that retry.
func (s *Store) HoldStock(reservationID, sku string, quantity int) (int, error) {
	if quantity <= 0 {
		return 0, ErrInvalidQuantity
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.holds[reservationID]; ok {
		if existing.SKU != sku || existing.Quantity != quantity {
			return 0, fmt.Errorf("%w: %s", ErrReservationConflict, reservationID)
		}
		return s.stock[sku], nil
	}

	if s.stock[sku] < quantity {
		return s.stock[sku], ErrInsufficientStock
	}

	s.stock[sku] -= quantity
	s.holds[reservationID] = Hold{ReservationID: reservationID, SKU: sku, Quantity: quantity}
	return s.stock[sku], nil
}

// ReleaseStock is idempotent: a duplicate release does not change stock again
// and returns zero because no active hold remains to identify a SKU.
func (s *Store) ReleaseStock(reservationID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	hold, ok := s.holds[reservationID]
	if !ok {
		return 0
	}
	delete(s.holds, reservationID)
	s.stock[hold.SKU] += hold.Quantity
	return s.stock[hold.SKU]
}
