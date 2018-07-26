package net

// BlockEvent for reporting block metrics
type BlockEvent struct {
	Outgoing     bool
	ContentType  string
	Recipients   int
	PayloadSize  int
	BlockSize int
}

// Collection returns the name of the collection for this event
func (e *BlockEvent) Collection() string {
	return "blocks"
}

// Type returns the content type for this event
func (e *BlockEvent) Type() string {
	return "nimona.telemetry.block"
}
