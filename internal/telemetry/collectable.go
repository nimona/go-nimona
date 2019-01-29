package telemetry

import "nimona.io/pkg/object"

// Collectable for metric events
type Collectable interface {
	Collection() string
	Measurements() map[string]interface{}
	ToObject() *object.Object
}

//go:generate go run nimona.io/tools/objectify -schema nimona.io/telemetry/connection -type ConnectionEvent -in collectable.go -out event_connection_generated.go

// ConnectionEvent for reporting connection info
type ConnectionEvent struct {
	// Event attributes
	Direction string `json:"direction"`
}

// Collection returns the string representation of the structure
func (ce *ConnectionEvent) Collection() string {
	return "nimona.io/telemetry.connection"
}

// Measurements returns a map with all the metrics for the event
func (ce *ConnectionEvent) Measurements() map[string]interface{} {
	return map[string]interface{}{
		"direction": ce.Direction,
	}
}

//go:generate go run nimona.io/tools/objectify -schema nimona.io/telemetry/object -type ObjectEvent -in collectable.go -out event_object_generated.go

// ObjectEvent for reporting object metrics
type ObjectEvent struct {
	// Event attributes
	Direction   string `json:"direction"`
	ContentType string `json:"contentType"`
	ObjectSize  int    `json:"size"`
	// Signature   *crypto.Signature `json:"-"`
}

// Collection returns the string representation of the structure
func (ee *ObjectEvent) Collection() string {
	return "nimona.io/telemetry.object"
}

// Measurements returns a map with all the metrics for the event
func (ee *ObjectEvent) Measurements() map[string]interface{} {
	return map[string]interface{}{
		"direction":    ee.Direction,
		"content_type": ee.ContentType,
		"object_size":   ee.ObjectSize,
	}
}
