package telemetry

import "errors"

// DefaultCollector for holding a global collector
var DefaultCollector Collector

// ErrStorageInit is returned when a storage backend fails to initialize
var ErrStorageInit = errors.New("Error initializing storage")

// Collector interface for collecting metrics
type Collector interface {
	Collect(Collectable) error
}
