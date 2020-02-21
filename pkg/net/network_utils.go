package net

import (
	"encoding/json"
	"errors"

	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
)

func Write(o object.Object, conn *Connection) error {
	if conn == nil {
		log.DefaultLogger.Info("conn cannot be nil")
		return errors.New("missing conn")
	}

	m := o.ToMap()
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	ra := ""
	if conn.RemotePeerKey != "" {
		ra = conn.RemotePeerKey.String()
	}

	// if os.Getenv("DEBUG_BLOCKS") == "true" {
	log.DefaultLogger.Info(
		"writting to connection",
		log.Any("object", o.ToMap()),
		log.String("local.address", conn.localAddress),
		log.String("remote.address", conn.remoteAddress),
		log.String("remote.fingerprint", ra),
		log.String("direction", "outgoing"),
	)
	// }

	b = append(b, '\n')
	if _, err := conn.conn.Write(b); err != nil {
		return err
	}

	return nil
}

func Read(conn *Connection) (*object.Object, error) {
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

	o := object.FromMap(m)

	ra := ""
	if conn.RemotePeerKey != "" {
		ra = conn.RemotePeerKey.String()
	}

	// if os.Getenv("DEBUG_BLOCKS") == "true" {
	logger.Error(
		"reading from connection",
		log.Any("map", m),
		log.Any("object", o.ToMap()),
		log.String("local.address", conn.localAddress),
		log.String("remote.address", conn.remoteAddress),
		log.String("remote.fingerprint", ra),
		log.String("direction", "incoming"),
	)
	// }

	if !o.GetSignature().IsEmpty() {
		if err := object.Verify(o); err != nil {
			// TODO we should verify, but return an error that doesn't
			// kill the connection
			return &o, nil
		}
	}

	return &o, nil
}
