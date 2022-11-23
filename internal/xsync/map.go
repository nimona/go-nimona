package xsync

import "sync"

// Map wraps sync.Map to add generics support.
type Map[K comparable, V any] struct {
	m sync.Map
}

// NewMap returns a new Map.
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{}
}

// Load returns the value stored in the map for a key, or nil if no
// value is present. The ok result indicates whether value was found
// in the map.
func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return
	}
	return v.(V), ok
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value. The loaded
// result is true if the value was loaded, false if stored.
func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	v, loaded := m.m.LoadOrStore(key, value)
	return v.(V), loaded
}

// Store sets the value for a key.
func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

// Delete deletes the value for a key.
func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value interface{}) bool {
		return f(key.(K), value.(V))
	})
}
