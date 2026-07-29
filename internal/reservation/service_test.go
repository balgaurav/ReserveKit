package reservation

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeInventory struct {
	holds    int
	releases int
}

func (f *fakeInventory) Availability(_ context.Context, _ string) (int, error) {
	return 5, nil
}
func (f *fakeInventory) Hold(_ context.Context, _, _ string, _ int) (int, error) {
	f.holds++
	return 3, nil
}
func (f *fakeInventory) Release(_ context.Context, _ string) (int, error) {
	f.releases++
	return 5, nil
}

func TestCreateIsIdempotent(t *testing.T) {
	inventory := &fakeInventory{}
	service := NewService(inventory)
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	service.clock = func() time.Time { return now }
	input := CreateInput{ID: "res-1", SKU: "keyboard", Quantity: 2, IdempotencyKey: "key-1", ExpiresAt: now.Add(15 * time.Minute)}

	first, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Create(context.Background(), CreateInput{ID: "res-2", SKU: "keyboard", Quantity: 2, IdempotencyKey: "key-1", ExpiresAt: now.Add(15 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || inventory.holds != 1 {
		t.Fatalf("idempotency result = (%s, holds %d), want same ID and one hold", second.ID, inventory.holds)
	}
}

func TestConfirmRejectsExpiredReservation(t *testing.T) {
	service := NewService(&fakeInventory{})
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	service.clock = func() time.Time { return now }
	reservation, err := service.Create(context.Background(), CreateInput{ID: "res-1", SKU: "keyboard", Quantity: 2, IdempotencyKey: "key-1", ExpiresAt: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	service.clock = func() time.Time { return now.Add(2 * time.Minute) }
	_, err = service.Confirm(context.Background(), reservation.ID)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("confirm error = %v, want invalid transition", err)
	}
}

func TestExpiredConfirmationReleasesInventory(t *testing.T) {
	inventory := &fakeInventory{}
	service := NewService(inventory)
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	service.clock = func() time.Time { return now }
	reservation, err := service.Create(context.Background(), CreateInput{ID: "res-1", SKU: "keyboard", Quantity: 2, IdempotencyKey: "key-1", ExpiresAt: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	service.clock = func() time.Time { return now.Add(2 * time.Minute) }
	_, _ = service.Confirm(context.Background(), reservation.ID)
	if inventory.releases != 1 {
		t.Fatalf("release calls = %d, want 1", inventory.releases)
	}
}

func TestReleaseIsIdempotent(t *testing.T) {
	inventory := &fakeInventory{}
	service := NewService(inventory)
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	service.clock = func() time.Time { return now }
	reservation, err := service.Create(context.Background(), CreateInput{ID: "res-1", SKU: "keyboard", Quantity: 2, IdempotencyKey: "key-1", ExpiresAt: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Release(context.Background(), reservation.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Release(context.Background(), reservation.ID); err != nil {
		t.Fatal(err)
	}
	if inventory.releases != 1 {
		t.Fatalf("release calls = %d, want 1", inventory.releases)
	}
}
