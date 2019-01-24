package net

import (
	"context"
	"fmt"
	"os"
	"strings"

	"nimona.io/internal/telemetry"
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
	direction := "incoming"
	if outgoing {
		direction = "outgoing"
	}
	telemetry.SendEvent(context.Background(), &telemetry.ConnectionEvent{
		Direction: direction,
	})
}

// SendBlockEvent sends a connection event
func SendBlockEvent(direction string, contentType string, blockSize int) {
	if os.Getenv("TELEMETRY") != "client" {
		return
	}
	if strings.Contains(contentType, "telemetry") {
		return
	}
	go telemetry.SendEvent(context.Background(), &telemetry.BlockEvent{
		Direction:   direction,
		ContentType: contentType,
		BlockSize:   blockSize,
	})
}
