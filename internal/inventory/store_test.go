package inventory

import (
	"errors"
	"testing"
)

func TestHoldStockIsIdempotent(t *testing.T) {
	store := NewStore(map[string]int{"keyboard": 5})

	remaining, err := store.HoldStock("res-1", "keyboard", 2)
	if err != nil || remaining != 3 {
		t.Fatalf("first hold = (%d, %v), want (3, nil)", remaining, err)
	}

	remaining, err = store.HoldStock("res-1", "keyboard", 2)
	if err != nil || remaining != 3 {
		t.Fatalf("retry hold = (%d, %v), want (3, nil)", remaining, err)
	}
	if got := store.Availability("keyboard"); got != 3 {
		t.Fatalf("availability = %d, want 3", got)
	}
}

func TestHoldStockRejectsInsufficientInventory(t *testing.T) {
	store := NewStore(map[string]int{"keyboard": 1})
	remaining, err := store.HoldStock("res-1", "keyboard", 2)
	if !errors.Is(err, ErrInsufficientStock) || remaining != 1 {
		t.Fatalf("hold = (%d, %v), want (1, insufficient stock)", remaining, err)
	}
}

func TestReleaseStockIsIdempotent(t *testing.T) {
	store := NewStore(map[string]int{"keyboard": 5})
	if _, err := store.HoldStock("res-1", "keyboard", 2); err != nil {
		t.Fatal(err)
	}
	if got := store.ReleaseStock("res-1"); got != 5 {
		t.Fatalf("release = %d, want 5", got)
	}
	if got := store.ReleaseStock("res-1"); got != 0 {
		t.Fatalf("duplicate release = %d, want 0", got)
	}
}
