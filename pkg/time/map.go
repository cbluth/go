package time

import (
	"sync"
	"time"
)

type (
	// TMap
	TMap[T comparable] interface {
		Size() int
		Add(T) time.Time
		Capacity(...int) int
		Get(T) (time.Time, bool)
		Delete(T) (time.Time, bool)
		Newest() (T, time.Time, bool)
		Oldest() (T, time.Time, bool)
		Each(func(T, time.Time) error) error
	}

	// tMap
	tMap[T comparable] struct {
		capacity int
		policy   DropPolicy
		mutex    sync.RWMutex
		M        map[T]time.Time `json:"map"`
	}
	// DropPolicy
	DropPolicy int
)

const (
	unknown DropPolicy = iota
	DontDrop
	DropRandom
	DropOldest
	DropNewest
)

// Add
func (s *tMap[T]) Add(v T) time.Time {
	t := time.Now().UTC()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.M[v] = t
	s.drop()
	return t
}

// drop
func (s *tMap[T]) drop() {
	// if size is bigger than cap, trim
	size := len(s.M)
	capacity := s.capacity
	if (size > capacity) && (capacity != 0) {
		for i := 0; i < (size - capacity); i++ {
			switch s.policy {
			case DropOldest:
				v, _, ok := s.oldest()
				if ok {
					s.delete(v)
				}
			case DropNewest:
				v, _, ok := s.newest()
				if ok {
					s.delete(v)
				}
			case DropRandom:
				for v := range s.M {
					delete(s.M, v)
					break
				}
			}
		}
	}
}

// Get
func (s *tMap[T]) Get(v T) (time.Time, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	t, ok := s.M[v]
	return t, ok
}

// Delete
func (s *tMap[T]) Delete(v T) (time.Time, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.delete(v)
}

// delete
func (s *tMap[T]) delete(v T) (time.Time, bool) {
	t, ok := s.M[v]
	delete(s.M, v)
	return t, ok
}

// Oldest
func (s *tMap[T]) Oldest() (v T, t time.Time, ok bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.oldest()
}

// oldest
func (s *tMap[T]) oldest() (v T, t time.Time, ok bool) {
	t = time.Now().UTC()
	s.each(
		func(tc T, tt time.Time) error {
			if tt.Before(t) {
				t = tt
				v = tc
				ok = true
			}
			return nil
		},
	)
	return v, t, ok
}

// Newest
func (s *tMap[T]) Newest() (v T, t time.Time, ok bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.newest()
}

// newest
func (s *tMap[T]) newest() (v T, t time.Time, ok bool) {
	t = time.Time{}
	s.each(
		func(tc T, tt time.Time) error {
			if tt.After(t) {
				t = tt
				v = tc
				ok = true
			}
			return nil
		},
	)
	return v, t, ok
}

// Each
func (s *tMap[T]) Each(f func(T, time.Time) error) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.each(f)
}

// each
func (s *tMap[T]) each(f func(T, time.Time) error) error {
	for element, addedAt := range s.M {
		err := f(element, addedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// Size
func (s *tMap[T]) Size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.M)
}

// Capacity
func (s *tMap[T]) Capacity(c ...int) int {
	switch len(c) {
	case 0:
		s.mutex.RLock()
		defer s.mutex.RUnlock()
		return s.capacity
	case 1:
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.capacity = c[0]
		return s.capacity
	default:
		return -1
	}
}

// NewTMap
func NewTMap[T comparable](capacity int, policy DropPolicy) TMap[T] {
	return &tMap[T]{
		policy:   policy,
		capacity: capacity,
		M:        make(map[T]time.Time),
	}
}
