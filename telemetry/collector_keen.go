package telemetry

import (
	"time"

	keen "gopkg.in/inconshreveable/go-keen.v0"
)

const (
	keenFlushInterval = 10 * time.Second

	keenAPIKey = "234E62D036FA54B637A8B132CC97D7F140DFF01E2D8BE55EDB7C24AF" +
		"86D21A8A0F852D97186D505576D6E89FC1D976E94292F5822512392DBCE22CE19" +
		"9BC1AF31FD5B16648C6FEA8B3334BCB7397098499DF92B4C36C52E53DCFA6B8E4" +
		"BCC3D3"
	keenProjectToken = "5b5092afc9e77c000175c52f"
)

// SetupKeenCollector sets the default collector to keen
func SetupKeenCollector() {
	DefaultCollector = &KeenCollector{
		keenClient: &keen.Client{
			ApiKey:       keenAPIKey,
			ProjectToken: keenProjectToken,
		},
	}
}

// KeenCollector for keen.io
type KeenCollector struct {
	keenClient *keen.Client
}

// Collect sends a event to keen
func (c *KeenCollector) Collect(event Collectable) error {
	return c.keenClient.AddEvent(event.Collection(), event)
}
