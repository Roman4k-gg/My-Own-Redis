package storage

import (
	"sync"
	"time"
)

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
	for _, k := range keys {
		if it, ok := s.data[k]; ok {
			if !it.expiresAt.IsZero() && time.Now().After(it.expiresAt) {
				continue
			}

			count++

		}
	}
	return count
}

func (s *Storage) StartGarbageCollector() {
	go func() {
		for {
			time.Sleep(time.Second * 7)

			s.mu.Lock()

			for key, item := range s.data {
				if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
					delete(s.data, key)
				}
			}
			s.mu.Unlock()
		}
	}()
}
