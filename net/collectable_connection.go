package net

// ConnectionEvent for reporting connection info
type ConnectionEvent struct {
	Outgoing bool
}

// Collection returns the name of the collection for this event
func (e *ConnectionEvent) Collection() string {
	return "connections"
}

// Type returns the content type for this event
func (e *ConnectionEvent) Type() string {
	return "nimona.telemetry.connection"
}
