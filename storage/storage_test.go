package storage

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// --- Set / Get ---

func TestSetAndGet(t *testing.T) {
	s := NewStorage()
	s.Set("key", "value", 0)

	got, ok := s.Get("key")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if got != "value" {
		t.Fatalf("expected 'value', got %q", got)
	}
}

func TestGetMissingKey(t *testing.T) {
	s := NewStorage()
	_, ok := s.Get("missing")
	if ok {
		t.Fatal("expected key to be absent")
	}
}

func TestOverwrite(t *testing.T) {
	s := NewStorage()
	s.Set("k", "v1", 0)
	s.Set("k", "v2", 0)

	got, _ := s.Get("k")
	if got != "v2" {
		t.Fatalf("expected 'v2', got %q", got)
	}
}

// --- TTL / Expiry ---

func TestSetWithTTLExpires(t *testing.T) {
	s := NewStorage()
	s.Set("k", "v", 50*time.Millisecond)

	// Key should be visible right after set.
	if _, ok := s.Get("k"); !ok {
		t.Fatal("key should exist immediately after Set with TTL")
	}

	time.Sleep(100 * time.Millisecond)

	_, ok := s.Get("k")
	if ok {
		t.Fatal("key should have expired")
	}
}

func TestSetWithZeroTTLNeverExpires(t *testing.T) {
	s := NewStorage()
	s.Set("k", "v", 0)
	time.Sleep(20 * time.Millisecond)
	if _, ok := s.Get("k"); !ok {
		t.Fatal("key with zero TTL should never expire")
	}
}

// --- Delete ---

func TestDelete(t *testing.T) {
	s := NewStorage()
	s.Set("a", "1", 0)
	s.Set("b", "2", 0)

	count := s.Delete([]string{"a", "b", "nonexistent"})
	if count != 2 {
		t.Fatalf("expected 2 deleted, got %d", count)
	}
	if _, ok := s.Get("a"); ok {
		t.Fatal("key 'a' should be deleted")
	}
}

// --- Exists ---

func TestExists(t *testing.T) {
	s := NewStorage()
	s.Set("x", "1", 0)
	s.Set("y", "2", 50*time.Millisecond)

	if n := s.Exists([]string{"x", "y", "z"}); n != 2 {
		t.Fatalf("expected 2 existing keys, got %d", n)
	}

	time.Sleep(100 * time.Millisecond)

	// "y" should have expired
	if n := s.Exists([]string{"x", "y"}); n != 1 {
		t.Fatalf("expected 1 key after TTL expiry, got %d", n)
	}
}

// --- INCR ---

func TestIncrNewKey(t *testing.T) {
	s := NewStorage()
	n, err := s.Incr("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
}

func TestIncrExistingKey(t *testing.T) {
	s := NewStorage()
	s.Set("counter", "41", 0)
	n, err := s.Incr("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 42 {
		t.Fatalf("expected 42, got %d", n)
	}
}

func TestIncrNonIntegerValue(t *testing.T) {
	s := NewStorage()
	s.Set("k", "not_a_number", 0)
	_, err := s.Incr("k")
	if err == nil {
		t.Fatal("expected error for non-integer value")
	}
}

func TestIncrNegativeValue(t *testing.T) {
	s := NewStorage()
	s.Set("k", "-5", 0)
	n, err := s.Incr("k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != -4 {
		t.Fatalf("expected -4, got %d", n)
	}
}

// --- GC ---

func TestGarbageCollector(t *testing.T) {
	s := NewStorage()
	done := make(chan struct{})
	// Override interval is not possible without changing the signature,
	// so we call collectExpired() directly to test the logic.

	s.Set("expired", "v", 1*time.Nanosecond)
	time.Sleep(5 * time.Millisecond) // ensure TTL has passed

	s.collectExpired()

	if _, ok := s.Get("expired"); ok {
		t.Fatal("expected key to be collected by GC")
	}

	// Ensure the goroutine stops cleanly.
	s.StartGarbageCollector(done)
	close(done)
}

// --- Concurrency (race detector) ---

// TestConcurrentReadWrite hammers the storage with concurrent goroutines.
// Run with: go test -race ./storage/...
func TestConcurrentReadWrite(t *testing.T) {
	s := NewStorage()
	const goroutines = 200
	const opsPerGoroutine = 500

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Writers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				s.Set(fmt.Sprintf("key:%d", id), fmt.Sprintf("val:%d", j), 0)
			}
		}(i)
	}

	// Readers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				s.Get(fmt.Sprintf("key:%d", id))
			}
		}(i)
	}

	// Deleters
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				s.Delete([]string{fmt.Sprintf("key:%d", id)})
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentIncr ensures INCR is race-free under parallel access.
func TestConcurrentIncr(t *testing.T) {
	s := NewStorage()
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if _, err := s.Incr("shared_counter"); err != nil {
				t.Errorf("Incr failed: %v", err)
			}
		}()
	}
	wg.Wait()

	val, ok := s.Get("shared_counter")
	if !ok {
		t.Fatal("shared_counter should exist")
	}
	// Value must be exactly 100 (each goroutine incremented once).
	if val != "100" {
		t.Fatalf("expected '100', got %q (possible race condition)", val)
	}
}
