package net

import (
	"encoding/json"
	"errors"
	"os"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
)

func Write(o object.Object, conn *Connection) error {
	if conn == nil {
		log.DefaultLogger.Error("conn cannot be nil")
		return errors.New("missing conn")
	}

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
		ra = conn.RemotePeerKey.Fingerprint().String()
	}

	if os.Getenv("DEBUG_BLOCKS") == "true" {
		log.DefaultLogger.Info(
			"writting to connection",
			log.Any("object", o.ToMap()),
			log.String("local.address", conn.localAddress),
			log.String("remote.address", conn.remoteAddress),
			log.String("remote.fingerprint", ra),
			log.String("direction", "outgoing"),
		)
	}

	b = append(b, '\n')
	if _, err := conn.conn.Write(b); err != nil {
		return err
	}

	return nil
}

func Read(conn *Connection) (object.Object, error) {
	logger := log.DefaultLogger

	r := <-conn.lines
	m := map[string]interface{}{}
	if err := json.Unmarshal(r, &m); err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("Recovered while processing", log.Any("r", r))
		}
	}()

	o := object.Object{}
	if err := o.FromMap(m); err != nil {
		return nil, err
	}

	ra := ""
	if conn.RemotePeerKey != nil {
		ra = conn.RemotePeerKey.Fingerprint().String()
	}

	if o.GetSignature() != nil {
		if err := crypto.Verify(o); err != nil {
			return nil, err
		}
	}

	if os.Getenv("DEBUG_BLOCKS") == "true" {
		logger.Info(
			"reading from connection",
			log.Any("object", o.ToMap()),
			log.String("local.address", conn.localAddress),
			log.String("remote.address", conn.remoteAddress),
			log.String("remote.fingerprint", ra),
			log.String("direction", "incoming"),
		)
	}

	return o, nil
}
