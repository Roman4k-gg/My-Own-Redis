package storage

import (
	"sync"
)

type Storage struct {
	mu sync.RWMutex
	data map[string]string
}

func NewStorage() *Storage{
	return &Storage{
		data: make(map[string]string),
	}
}

func (s *Storage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *Storage) Set(key, value string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
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
		if _, ok := s.data[k]; ok {
			count++
		}
	}
	return count
}
