package net

import "nimona.io/go/primitives"

// BlockEvent for reporting block metrics
type BlockEvent struct {
	Outgoing    bool
	ContentType string
	Recipients  int
	PayloadSize int
	BlockSize   int
}

// Collection returns the name of the collection for this event
func (e *BlockEvent) Collection() string {
	return "blocks"
}

// Type returns the content type for this event
func (e *BlockEvent) GetType() string {
	return "nimona.telemetry.block"
}

func (e *BlockEvent) GetSignature() *primitives.Signature {
	// no signature
	return nil
}

func (e *BlockEvent) SetSignature(s *primitives.Signature) {
	// no signature
}

func (e *BlockEvent) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (e *BlockEvent) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
