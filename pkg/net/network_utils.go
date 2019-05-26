package net

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func Write(o *object.Object, conn *Connection) error {
	conn.Conn.SetWriteDeadline(time.Now().Add(time.Second)) // nolint: errcheck
	if o == nil {
		log.DefaultLogger.Error("object for fw cannot be nil")
		return errors.New("missing object")
	}

	b, err := json.Marshal(o.ToMap())
	if err != nil {
		return err
	}

	ra := ""
	if conn.RemotePeerKey != nil {
		ra = conn.RemotePeerKey.Fingerprint()
	}

	if os.Getenv("DEBUG_BLOCKS") == "true" {
		b, _ := json.MarshalIndent(o.ToMap(), "", "  ")
		log.DefaultLogger.Info(
			string(b),
			log.String("remote_peer_hash", ra),
			log.String("direction", "outgoing"),
		)
	}

	if _, err := conn.Conn.Write(b); err != nil {
		return err
	}

	SendObjectEvent(
		"outgoing",
		o.GetType(),
		len(b),
	)

	return nil
}

func Read(conn *Connection) (*object.Object, error) {
	logger := log.DefaultLogger

	pDecoder := json.NewDecoder(conn.Conn)
	m := map[string]interface{}{}
	if err := pDecoder.Decode(&m); err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("Recovered while processing", log.Any("r", r))
		}
	}()

	o := &object.Object{}
	if err := o.FromMap(m); err != nil {
		return nil, err
	}

	ra := ""
	if conn.RemotePeerKey != nil {
		ra = conn.RemotePeerKey.Fingerprint()
	}

	if o.GetSignature() != nil {
		if err := crypto.Verify(o); err != nil {
			return nil, err
		}
	}

	// SendObjectEvent(
	// 	"incoming",
	// 	o.GetType(),
	// 	pDecoder.NumBytesRead(),
	// )

	if os.Getenv("DEBUG_BLOCKS") == "true" {
		b, _ := json.MarshalIndent(o.ToMap(), "", "  ")
		logger.Info(
			string(b),
			log.String("remote_peer_hash", ra),
			log.String("direction", "incoming"),
		)
	}

	return o, nil
}
