package core

import (
	"fmt"
	"sync"
)

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
    for k, v := range initial {
        d[k] = v
    }
    return &SharedData{data: d}
}

// Get returns a value by key. ok is false when the key doesn't exist.
func (s *SharedData) Get(key string) (any, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    v, ok := s.data[key]
    return v, ok
}

// Set writes a value under key, creating it if it doesn't exist.
func (s *SharedData) Set(key string, value any) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
}

// Delete removes a key from the store.
func (s *SharedData) Delete(key string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.data, key)
}

// Update atomically reads the current value and replaces it using the provided
// function. This prevents lost-update races between a Get and a Set.
func (s *SharedData) Update(key string, fn func(current any, exists bool) any) {
    s.mu.Lock()
    defer s.mu.Unlock()
    current, exists := s.data[key]
    s.data[key] = fn(current, exists)
}

// GetString is a typed helper that returns a string value or an error.
func (s *SharedData) GetString(key string) (string, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    v, ok := s.data[key]
    if !ok {
        return "", fmt.Errorf("key %q not found", key)
    }
    str, ok := v.(string)
    if !ok {
        return "", fmt.Errorf("key %q is %T, not string", key, v)
    }
    return str, nil
}

// GetInt is a typed helper that returns an int value or an error.
func (s *SharedData) GetInt(key string) (int, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    v, ok := s.data[key]
    if !ok {
        return 0, fmt.Errorf("key %q not found", key)
    }
    n, ok := v.(int)
    if !ok {
        return 0, fmt.Errorf("key %q is %T, not int", key, v)
    }
    return n, nil
}

// Snapshot returns a shallow copy of all data under a read lock.
func (s *SharedData) Snapshot() map[string]any {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make(map[string]any, len(s.data))
    for k, v := range s.data {
        out[k] = v
    }
    return out
}
