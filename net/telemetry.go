package net

import (
	"context"

	"github.com/nimona/go-nimona/telemetry"
)

// SendConnectionEvent sends a connection event
func SendConnectionEvent(outgoing bool) {
	telemetry.SendEvent(context.Background(), &telemetry.ConnectionEvent{
		Outgoing: outgoing,
	})
}

// SendBlockEvent sends a connection event
func SendBlockEvent(outgoing bool, contentType string, recipients,
	payloadSize, blockSize int) {
	telemetry.SendEvent(context.Background(), &telemetry.BlockEvent{
		Outgoing:    outgoing,
		ContentType: contentType,
		Recipients:  recipients,
		PayloadSize: payloadSize,
		BlockSize:   blockSize,
	})
}
