package telemetry

var DefaultCollector Collector

// Collector interface for collecting metrics
type Collector interface {
	Collect(Collectable) error
}
