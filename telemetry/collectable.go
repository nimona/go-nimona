package telemetry

// Collectable for metric events
type Collectable interface {
	Collection() string
	Measurements() map[string]interface{}
}

const (
	connectionEventCollection = "ConnectionEvent"
	envelopeEventCollection   = "EnvelopeEvent"
)

// ConnectionEvent for reporting connection info
type ConnectionEvent struct {
	// Event attributes
	Outgoing bool
}

func (cl *ConnectionEvent) Collection() string {
	return connectionEventCollection
}

func (cl *ConnectionEvent) Measurements() map[string]interface{} {

	return map[string]interface{}{
		"outgoing": cl.Outgoing,
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

func (cl *EnvelopeEvent) Collection() string {
	return envelopeEventCollection
}

func (cl *EnvelopeEvent) Measurements() map[string]interface{} {

	return map[string]interface{}{
		"outgoing":      cl.Outgoing,
		"content_type":  cl.ContentType,
		"recipients":    cl.Recipients,
		"payload_size":  cl.PayloadSize,
		"envelope_size": cl.EnvelopeSize,
	}
}
