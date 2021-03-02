package net

import (
	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
)

var (
	ErrInvalidSignature = errors.Error("invalid signature")
	ErrConnectionClosed = errors.Error("connection closed")
)

func Write(o *object.Object, conn *Connection) error {
	if conn == nil {
		log.DefaultLogger.Info("conn cannot be nil")
		return errors.New("missing conn")
	}

	ra := ""
	if conn.RemotePeerKey != "" {
		ra = conn.RemotePeerKey.String()
	}

	log.DefaultLogger.Debug(
		"writing to connection",
		log.Any("object", o.ToMap()),
		log.String("local.address", conn.localAddress),
		log.String("remote.address", conn.remoteAddress),
		log.String("remote.fingerprint", ra),
		log.String("direction", "outgoing"),
	)

	if err := conn.encoder.Encode(o); err != nil {
		return errors.Wrap(errors.New("error marshaling object"), err)
	}

	return nil
}

func Read(conn *Connection) (*object.Object, error) {
	logger := log.DefaultLogger

	defer func() {
		if r := recover(); r != nil {
			logger.Error(
				"recovered from panic during read",
				log.Any("r", r),
				log.Stack(),
			)
		}
	}()

	o := &object.Object{}
	err := conn.decoder.Decode(o)
	if err != nil {
		return nil, err
	}

	logger.Debug(
		"reading from connection",
		log.Any("object", o),
		log.String("local.address", conn.localAddress),
		log.String("remote.address", conn.remoteAddress),
		log.String("remote.publicKey", conn.RemotePeerKey.String()),
		log.String("direction", "incoming"),
	)

	if !o.Metadata.Signature.IsEmpty() {
		if err := object.Verify(o); err != nil {
			return o, errors.Wrap(ErrInvalidSignature, err)
		}
	}

	return o, nil
}
