package telemetry

// Collectable for metric events
type Collectable interface {
	Collection() string
	Measurements() map[string]interface{}
}

const (
	connectionEventCollection = "nimona.telemetry.connection"
	envelopeEventCollection   = "nimona.telemetry.envelope"
)

// ConnectionEvent for reporting connection info
type ConnectionEvent struct {
	// Event attributes
	Outgoing bool
}

// Collection returns the string representation of the structure
func (ce *ConnectionEvent) Collection() string {
	return connectionEventCollection
}

// Measurements returns a map with all the metrics for the event
func (ce *ConnectionEvent) Measurements() map[string]interface{} {

	return map[string]interface{}{
		"outgoing": ce.Outgoing,
	}
}

// EnvelopeEvent for reporting envelope metrics
type EnvelopeEvent struct {
	// Event attributes
	Outgoing     bool
	ContentType  string
	Recipients   int
	PayloadSize  int
	EnvelopeSize int
}

// Collection returns the string representation of the structure
func (ee *EnvelopeEvent) Collection() string {
	return envelopeEventCollection
}

// Measurements returns a map with all the metrics for the event
func (ee *EnvelopeEvent) Measurements() map[string]interface{} {

	return map[string]interface{}{
		"outgoing":      ee.Outgoing,
		"content_type":  ee.ContentType,
		"recipients":    ee.Recipients,
		"payload_size":  ee.PayloadSize,
		"envelope_size": ee.EnvelopeSize,
	}
}
