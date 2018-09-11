package telemetry

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

// Collectable for metric events
type Collectable interface {
	blocks.Typed
	Collection() string
	Measurements() map[string]interface{}
}

// ConnectionEvent for reporting connection info
type ConnectionEvent struct {
	// Event attributes
	Direction string            `json:"direction"`
	Signature *crypto.Signature `json:"-"`
}

func (ce *ConnectionEvent) GetType() string {
	return "nimona.telemetry.connection"
}

func (ce *ConnectionEvent) GetSignature() *crypto.Signature {
	return ce.Signature
}

func (ce *ConnectionEvent) SetSignature(s *crypto.Signature) {
	ce.Signature = s
}

func (ce *ConnectionEvent) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (ce *ConnectionEvent) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

// Collection returns the string representation of the structure
func (ce *ConnectionEvent) Collection() string {
	return ce.GetType()
}

// Measurements returns a map with all the metrics for the event
func (ce *ConnectionEvent) Measurements() map[string]interface{} {
	return map[string]interface{}{
		"direction": ce.Direction,
	}
}

// BlockEvent for reporting block metrics
type BlockEvent struct {
	// Event attributes
	Direction   string            `json:"direction"`
	ContentType string            `json:"contentType"`
	BlockSize   int               `json:"size"`
	Signature   *crypto.Signature `json:"-"`
}

func (ee *BlockEvent) GetType() string {
	return "nimona.telemetry.block"
}

func (ee *BlockEvent) GetSignature() *crypto.Signature {
	return ee.Signature
}

func (ee *BlockEvent) SetSignature(s *crypto.Signature) {
	ee.Signature = s
}

func (ee *BlockEvent) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (ee *BlockEvent) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

// Collection returns the string representation of the structure
func (ee *BlockEvent) Collection() string {
	return ee.GetType()
}

// Measurements returns a map with all the metrics for the event
func (ee *BlockEvent) Measurements() map[string]interface{} {
	return map[string]interface{}{
		"direction":    ee.Direction,
		"content_type": ee.ContentType,
		"block_size":   ee.BlockSize,
	}
}
