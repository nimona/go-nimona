package telemetry

import (
	"nimona.io/go/primitives"
)

// Collectable for metric events
type Collectable interface {
	Block() *primitives.Block
	FromBlock(*primitives.Block)
	Collection() string
	Measurements() map[string]interface{}
}

// ConnectionEvent for reporting connection info
type ConnectionEvent struct {
	// Event attributes
	Direction string `json:"direction"`
}

func (ce *ConnectionEvent) FromBlock(block *primitives.Block) {
	ce.Direction = block.Payload["direction"].(string)
}

func (ce *ConnectionEvent) Block() *primitives.Block {
	// TODO(geoah) sign
	return &primitives.Block{
		Type: "nimona.io/telemetry.connection",
		Payload: map[string]interface{}{
			"direction": ce.Direction,
		},
	}
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

// BlockEvent for reporting block metrics
type BlockEvent struct {
	// Event attributes
	Direction   string `json:"direction"`
	ContentType string `json:"contentType"`
	BlockSize   int    `json:"size"`
	// Signature   *primitives.Signature `json:"-"`
}

func (ee *BlockEvent) FromBlock(block *primitives.Block) {
	ee.BlockSize = int(block.Payload["size"].(uint64))
	ee.Direction = block.Payload["direction"].(string)
	ee.ContentType = block.Payload["contentType"].(string)

}

func (ee *BlockEvent) Block() *primitives.Block {
	// TODO sign?
	return &primitives.Block{
		Type: "nimona.io/telemetry.block",
		Payload: map[string]interface{}{
			"direction":   ee.Direction,
			"contentType": ee.ContentType,
			"size":        ee.BlockSize,
		},
		// Signature: ee.Signature,
	}
}

// Collection returns the string representation of the structure
func (ee *BlockEvent) Collection() string {
	return "nimona.io/telemetry.block"
}

// Measurements returns a map with all the metrics for the event
func (ee *BlockEvent) Measurements() map[string]interface{} {
	return map[string]interface{}{
		"direction":    ee.Direction,
		"content_type": ee.ContentType,
		"block_size":   ee.BlockSize,
	}
}
