package shared

import (
	"maps"
	"sync"
)

type SharedKey[T any] struct {
	name string
}

// NewKey creates a new typed key with the given name.
func NewKey[T any](name string) SharedKey[T] {
	return SharedKey[T]{name: name}
}

// Name returns the underlying string key.
func (k SharedKey[T]) Name() string {
    return k.name
}

// SharedData is a generic, goroutine-safe key-value store that all jobs
// can read from and write to. Jobs should use the typed helpers when possible
// and hold the lock for the shortest time necessary.
type SharedData struct {
    mu   sync.RWMutex
    data map[string]any
}

// NewSharedData creates an initialised SharedData store.
func NewSharedData(initial map[string]any) *SharedData {
    d := make(map[string]any, len(initial))
    maps.Copy(d, initial)
    return &SharedData{data: d}
}

// Get returns a value by key. ok is false when the key doesn't exist.
func Get[T any](s *SharedData, key SharedKey[T]) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var zero T
	v, ok := s.data[key.name]
	if !ok {
		return zero, false
	}
	typed, ok := v.(T)
	if !ok {
		return zero, false
	}
	return typed, true
}

// Set writes a typed value under a typed key.
func Set[T any](s *SharedData, key SharedKey[T], value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key.name] = value
}

// Update atomically reads and replaces the value under a typed key.
func Update[T any](s *SharedData, key SharedKey[T], fn func(current T, exists bool) T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, exists := s.data[key.name]
	var typed T
	if exists {
		typed, _ = current.(T)
	}
	s.data[key.name] = fn(typed, exists)
}

// Delete removes a key from the store.
func Delete[T any](s *SharedData, key SharedKey[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key.name)
}

// Snapshot returns a shallow copy of all raw data.
func (s *SharedData) Snapshot() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]any, len(s.data))
	maps.Copy(out, s.data)
	return out
}
