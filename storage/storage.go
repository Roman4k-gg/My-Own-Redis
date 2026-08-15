package storage

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

// errNotInteger is returned by Incr when the stored value is not an integer.
var errNotInteger = errors.New("ERR value is not an integer or out of range")

type Item struct {
	value     string
	expiresAt time.Time
}

type Storage struct {
	mu   sync.RWMutex
	data map[string]Item
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]Item),
	}
}

func (s *Storage) Get(key string) (string, bool) {
	s.mu.RLock()
	it, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return "", false
	}
	if !it.expiresAt.IsZero() && time.Now().After(it.expiresAt) {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return "", false
	}

	return it.value, true
}

func (s *Storage) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var expires time.Time
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	s.data[key] = Item{
		value:     value,
		expiresAt: expires,
	}
}

// Incr atomically increments the integer value of a key by 1.
// If the key does not exist, it is initialized to 0 before incrementing.
// Returns an error if the stored value is not a valid integer.
func (s *Storage) Incr(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := s.data[key]

	// Treat expired keys as non-existent.
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		delete(s.data, key)
		item = Item{}
	}

	val := item.value
	if val == "" {
		val = "0"
	}

	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, errNotInteger
	}
	n++
	item.value = strconv.FormatInt(n, 10)
	s.data[key] = item
	return n, nil
}

func (s *Storage) Delete(keys []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, k := range keys {
		if _, ok := s.data[k]; ok {
			delete(s.data, k)
			count++
		}
	}
	return count
}

func (s *Storage) Exists(keys []string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	now := time.Now()
	for _, k := range keys {
		if it, ok := s.data[k]; ok {
			if !it.expiresAt.IsZero() && now.After(it.expiresAt) {
				continue
			}
			count++
		}
	}
	return count
}

// StartGarbageCollector starts a background goroutine that periodically removes
// expired keys. It stops cleanly when done is closed — no goroutine leaks.
//
// Two-phase locking algorithm:
//   - Phase 1 (RLock): scan for expired keys without blocking concurrent reads/writes.
//   - Phase 2 (WLock): delete only the collected keys — minimal lock hold time.
func (s *Storage) StartGarbageCollector(done <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(7 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.collectExpired()
			case <-done:
				return
			}
		}
	}()
}

// collectExpired performs one two-phase GC sweep.
func (s *Storage) collectExpired() {
	now := time.Now()

	// Phase 1: collect expired key names under RLock.
	// Concurrent GET/SET operations are NOT blocked during this phase.
	s.mu.RLock()
	var toDelete []string
	for key, item := range s.data {
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			toDelete = append(toDelete, key)
		}
	}
	s.mu.RUnlock()

	if len(toDelete) == 0 {
		return
	}

	// Phase 2: delete under WLock only for the minimum required time.
	// Re-check each key: it may have been refreshed between phase 1 and phase 2.
	s.mu.Lock()
	for _, key := range toDelete {
		if item, ok := s.data[key]; ok {
			if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
				delete(s.data, key)
			}
		}
	}
	s.mu.Unlock()
}
