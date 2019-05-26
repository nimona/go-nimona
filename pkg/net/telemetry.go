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
	// nolint: errcheck
	telemetry.SendEvent(context.Background(), &telemetry.ConnectionEvent{
		Direction: direction,
	})
}

// SendObjectEvent sends a connection event
func SendObjectEvent(direction string, contentType string, objectSize int) {
	if os.Getenv("TELEMETRY") != "client" {
		return
	}
	if strings.Contains(contentType, "telemetry") {
		return
	}
	// nolint: errcheck
	go telemetry.SendEvent(context.Background(), &telemetry.ObjectEvent{
		Direction:   direction,
		ContentType: contentType,
		ObjectSize:  objectSize,
	})
}
