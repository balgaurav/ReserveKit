package reservation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInvalidInput      = errors.New("invalid reservation input")
	ErrNotFound          = errors.New("reservation not found")
	ErrInvalidTransition = errors.New("invalid reservation state transition")
)

type Status string

const (
	Pending   Status = "pending"
	Confirmed Status = "confirmed"
	Released  Status = "released"
	Expired   Status = "expired"
)

// Client is a narrow adapter around the generated grpc-go client. It keeps
// reservation business rules testable without a network transport.
type Client interface {
	Hold(context.Context, string, string, int) (int, error)
	Release(context.Context, string) (int, error)
}

type Reservation struct {
	ID             string    `json:"id"`
	SKU            string    `json:"sku"`
	Quantity       int       `json:"quantity"`
	Status         Status    `json:"status"`
	IdempotencyKey string    `json:"-"`
	ExpiresAt      time.Time `json:"expiresAt"`
	CreatedAt      time.Time `json:"createdAt"`
}

type CreateInput struct {
	ID             string
	SKU            string
	Quantity       int
	IdempotencyKey string
	ExpiresAt      time.Time
}

type Service struct {
	mu           sync.Mutex
	client       Client
	clock        func() time.Time
	reservations map[string]Reservation
	keys         map[string]string
}

func NewService(client Client) *Service {
	return &Service{client: client, clock: time.Now, reservations: map[string]Reservation{}, keys: map[string]string{}}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Reservation, error) {
	if input.ID == "" || input.SKU == "" || input.Quantity <= 0 || input.IdempotencyKey == "" || !input.ExpiresAt.After(s.clock()) {
		return Reservation{}, ErrInvalidInput
	}

	s.mu.Lock()
	if existingID, ok := s.keys[input.IdempotencyKey]; ok {
		existing := s.reservations[existingID]
		s.mu.Unlock()
		return existing, nil
	}
	s.mu.Unlock()

	if _, err := s.client.Hold(ctx, input.ID, input.SKU, input.Quantity); err != nil {
		return Reservation{}, err
	}

	reservation := Reservation{ID: input.ID, SKU: input.SKU, Quantity: input.Quantity, Status: Pending, IdempotencyKey: input.IdempotencyKey, ExpiresAt: input.ExpiresAt, CreatedAt: s.clock()}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existingID, ok := s.keys[input.IdempotencyKey]; ok {
		return s.reservations[existingID], nil
	}
	s.reservations[reservation.ID] = reservation
	s.keys[reservation.IdempotencyKey] = reservation.ID
	return reservation, nil
}

func (s *Service) Confirm(id string) (Reservation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	reservation, ok := s.reservations[id]
	if !ok {
		return Reservation{}, ErrNotFound
	}
	if reservation.Status != Pending || !reservation.ExpiresAt.After(s.clock()) {
		if reservation.Status == Pending {
			reservation.Status = Expired
			s.reservations[id] = reservation
		}
		return Reservation{}, fmt.Errorf("%w: %s", ErrInvalidTransition, reservation.Status)
	}
	reservation.Status = Confirmed
	s.reservations[id] = reservation
	return reservation, nil
}

func (s *Service) Release(ctx context.Context, id string) (Reservation, error) {
	s.mu.Lock()
	reservation, ok := s.reservations[id]
	if !ok {
		s.mu.Unlock()
		return Reservation{}, ErrNotFound
	}
	if reservation.Status == Released {
		s.mu.Unlock()
		return reservation, nil
	}
	if reservation.Status != Pending {
		s.mu.Unlock()
		return Reservation{}, fmt.Errorf("%w: %s", ErrInvalidTransition, reservation.Status)
	}
	s.mu.Unlock()

	if _, err := s.client.Release(ctx, id); err != nil {
		return Reservation{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	reservation = s.reservations[id]
	reservation.Status = Released
	s.reservations[id] = reservation
	return reservation, nil
}

func (s *Service) Get(id string) (Reservation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	reservation, ok := s.reservations[id]
	if !ok {
		return Reservation{}, ErrNotFound
	}
	return reservation, nil
}
