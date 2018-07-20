package telemetry

// Collectable for metric events
type Collectable interface {
	Collection() string
	// Measurements() map[string]interface{}
}
