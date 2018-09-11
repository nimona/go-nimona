package net

import (
	"context"
	"fmt"

	"nimona.io/go/telemetry"
)

// SendEvent sends an event
func SendEvent(event telemetry.Collectable) {
	if telemetry.DefaultCollector == nil {
		return
	}
	// TODO Wrap or extend event with version and other static information
	// TODO properly log error
	if err := telemetry.DefaultCollector.Collect(event); err != nil {
		fmt.Println("telemetry error", err)
	}
}

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
