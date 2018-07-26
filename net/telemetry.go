package net

import (
	"fmt"

	"github.com/nimona/go-nimona/telemetry"
)

func init() {
	RegisterContentType("nimona.telemetry.connection", ConnectionEvent{})
	RegisterContentType("nimona.telemetry.block", BlockEvent{})
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
	SendEvent(&ConnectionEvent{
		Outgoing: outgoing,
	})
}

// SendBlockEvent sends a connection event
func SendBlockEvent(outgoing bool, contentType string, recipients, payloadSize, blockSize int) {
	SendEvent(&BlockEvent{
		Outgoing:     outgoing,
		ContentType:  contentType,
		Recipients:   recipients,
		PayloadSize:  payloadSize,
		BlockSize: blockSize,
	})
}
