package telemetry

import "github.com/nimona/go-nimona/blocks"

// Collectable for metric events
type Collectable interface {
	Collection() string
	Measurements() map[string]interface{}
}

const (
	connectionEventCollection = "nimona.telemetry.connection"
	blockEventCollection      = "nimona.telemetry.block"
)

// ConnectionEvent for reporting connection info
type ConnectionEvent struct {
	// Event attributes
	Outgoing  bool
	Signature *blocks.Signature `nimona:",signature" json:"signature"`
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

// BlockEvent for reporting block metrics
type BlockEvent struct {
	// Event attributes
	Outgoing    bool
	ContentType string
	Recipients  int
	PayloadSize int
	BlockSize   int
	Signature   *blocks.Signature `nimona:",signature" json:"signature"`
}

// Collection returns the string representation of the structure
func (ee *BlockEvent) Collection() string {
	return blockEventCollection
}

// Measurements returns a map with all the metrics for the event
func (ee *BlockEvent) Measurements() map[string]interface{} {

	return map[string]interface{}{
		"outgoing":     ee.Outgoing,
		"content_type": ee.ContentType,
		"recipients":   ee.Recipients,
		"payload_size": ee.PayloadSize,
		"block_size":   ee.BlockSize,
	}
}
