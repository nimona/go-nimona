package net

// EnvelopeEvent for reporting envelope metrics
type EnvelopeEvent struct {
	Outgoing     bool
	ContentType  string
	Recipients   int
	PayloadSize  int
	EnvelopeSize int
}

// Collection returns the name of the collection for this event
func (e *EnvelopeEvent) Collection() string {
	return "envelopes"
}

// Type returns the content type for this event
func (e *EnvelopeEvent) Type() string {
	return "nimona.telemetry.envelope"
}
