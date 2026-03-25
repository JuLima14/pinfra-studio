package sandbox

import (
	"testing"
)

func TestAllocate_Sequential(t *testing.T) {
	pa := NewPortAllocator(3000, 3005)

	port1, err := pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port1 != 3000 {
		t.Errorf("expected first allocation to be 3000, got %d", port1)
	}

	port2, err := pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port2 != 3001 {
		t.Errorf("expected second allocation to be 3001, got %d", port2)
	}

	port3, err := pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port3 != 3002 {
		t.Errorf("expected third allocation to be 3002, got %d", port3)
	}
}

func TestRelease_Reuse(t *testing.T) {
	pa := NewPortAllocator(3000, 3001)

	port1, err := pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	port2, err := pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Release port1 so it can be reused
	pa.Release(port1)

	port3, err := pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error allocating after release: %v", err)
	}

	// port3 should reuse the released port (3000)
	if port3 != port1 {
		t.Errorf("expected reuse of released port %d, got %d", port1, port3)
	}

	_ = port2
}

func TestReserve(t *testing.T) {
	pa := NewPortAllocator(3000, 3002)

	// Reserve port 3000 so it's skipped during allocation
	pa.Reserve(3000)

	port, err := pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port == 3000 {
		t.Errorf("expected 3000 to be skipped (reserved), got %d", port)
	}
	if port != 3001 {
		t.Errorf("expected first available to be 3001, got %d", port)
	}
}

func TestAllocate_Exhaustion(t *testing.T) {
	pa := NewPortAllocator(3000, 3001)

	_, err := pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error on first allocation: %v", err)
	}

	_, err = pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error on second allocation: %v", err)
	}

	// All ports exhausted — should return an error
	_, err = pa.Allocate()
	if err == nil {
		t.Error("expected error when all ports are exhausted, got nil")
	}
}

func TestRelease_ThenExhaust(t *testing.T) {
	pa := NewPortAllocator(5000, 5000)

	port, err := pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be exhausted now
	_, err = pa.Allocate()
	if err == nil {
		t.Error("expected exhaustion error")
	}

	// Release and try again
	pa.Release(port)

	port2, err := pa.Allocate()
	if err != nil {
		t.Fatalf("unexpected error after release: %v", err)
	}
	if port2 != port {
		t.Errorf("expected reuse of %d, got %d", port, port2)
	}
}
