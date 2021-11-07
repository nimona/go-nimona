package mesh

import (
	"time"

	"nimona.io/pkg/peer"
)

type (
	sendOptions struct {
		connectionInfo         *peer.ConnectionInfo
		waitForResponse        interface{}
		waitForResponseTimeout time.Duration
	}
)

func SendWithConnectionInfo(c *peer.ConnectionInfo) func(*sendOptions) {
	return func(w *sendOptions) {
		w.connectionInfo = c
	}
}

func SendWithResponse(
	v interface{},
	t time.Duration,
) func(*sendOptions) {
	return func(w *sendOptions) {
		w.waitForResponse = v
		if t == 0 {
			t = time.Second
		}
		w.waitForResponseTimeout = t
	}
}
