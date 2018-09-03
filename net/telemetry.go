package net

import (
	"fmt"

	blocks "github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/telemetry"
)

func init() {
	blocks.RegisterContentType("nimona.telemetry.connection",
		telemetry.ConnectionEvent{})
	blocks.RegisterContentType("nimona.telemetry.block",
		telemetry.BlockEvent{})
}

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
	SendEvent(&telemetry.ConnectionEvent{
		Outgoing: outgoing,
	})
}

// SendBlockEvent sends a connection event
func SendBlockEvent(outgoing bool, contentType string, recipients,
	payloadSize, blockSize int) {
	SendEvent(&telemetry.BlockEvent{
		Outgoing:    outgoing,
		ContentType: contentType,
		Recipients:  recipients,
		PayloadSize: payloadSize,
		BlockSize:   blockSize,
	})
}
