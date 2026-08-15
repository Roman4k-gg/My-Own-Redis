package storage

import (
	"fmt"
	"testing"
)

func BenchmarkSet(b *testing.B) {
	s := NewStorage()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			s.Set(fmt.Sprintf("key:%d", i), "value", 0)
			i++
		}
	})
}

func BenchmarkGet(b *testing.B) {
	s := NewStorage()
	// Pre-populate.
	for i := 0; i < 10_000; i++ {
		s.Set(fmt.Sprintf("key:%d", i), "value", 0)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			s.Get(fmt.Sprintf("key:%d", i%10_000))
			i++
		}
	})
}

func BenchmarkSetGet(b *testing.B) {
	s := NewStorage()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key:%d", i%1000)
			if i%2 == 0 {
				s.Set(key, "v", 0)
			} else {
				s.Get(key)
			}
			i++
		}
	})
}

func BenchmarkIncr(b *testing.B) {
	s := NewStorage()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Incr("counter") //nolint:errcheck
		}
	})
}

func BenchmarkDelete(b *testing.B) {
	s := NewStorage()
	keys := make([]string, 1000)
	for i := range keys {
		key := fmt.Sprintf("key:%d", i)
		keys[i] = key
		s.Set(key, "v", 0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Re-populate every 1000 iterations to keep the benchmark meaningful.
		if i%1000 == 0 {
			for _, k := range keys {
				s.Set(k, "v", 0)
			}
		}
		s.Delete([]string{fmt.Sprintf("key:%d", i%1000)})
	}
}
