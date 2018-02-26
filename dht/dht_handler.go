package dht

import (
	"bufio"
	"context"
	"encoding/json"

	fabric "github.com/nimona/go-nimona-fabric"
	"github.com/sirupsen/logrus"
)

func (d *DHT) Name() string {
	return "dht"
}

// Negotiate will be called after all the other protocol have been processed
func (d *DHT) Negotiate(fn fabric.NegotiatorFunc) fabric.NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c fabric.Conn) error {
		return nil
	}
}

// Handle ping requests
func (d *DHT) Handle(fn fabric.HandlerFunc) fabric.HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c fabric.Conn) error {
		sr := bufio.NewReader(c)
		for {
			// read line
			line, err := sr.ReadString('\n')
			if err != nil {
				logrus.WithError(err).Errorf("Could not read")
				return err // TODO(geoah) Return?
			}
			// logrus.WithField("line", line).Debugf("handleStream got line")

			// decode message
			msg := &message{}
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				// logrus.WithError(err).Warnf("Could not decode message")
				return err
			}

			// process message
			if err := d.handleMessage(msg); err != nil {
				// logrus.WithError(err).Warnf("Could not process message")
			}
		}
		// return nil
	}
}
