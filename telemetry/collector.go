package telemetry

// DefaultCollector for holding a global collector
var DefaultCollector Collector

// Collector interface for collecting metrics
type Collector interface {
	Collect(Collectable) error
}
