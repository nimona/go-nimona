package net

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func Write(o *object.Object, conn *Connection) error {
	conn.Conn.SetWriteDeadline(time.Now().Add(time.Second))
	if o == nil {
		log.DefaultLogger.Error("object for fw cannot be nil")
		return errors.New("missing object")
	}

	b, err := object.Marshal(o)
	if err != nil {
		return err
	}

	ra := ""
	if conn.RemotePeerKey != nil {
		ra = conn.RemotePeerKey.Hash
	}

	if os.Getenv("DEBUG_BLOCKS") == "true" {
		b, _ := json.MarshalIndent(o.ToMap(), "", "  ")
		log.DefaultLogger.Info(
			string(b),
			zap.String("remote_peer_hash", ra),
			zap.String("direction", "outgoing"),
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

	pDecoder := codec.NewDecoder(conn.Conn, object.RawCborHandler())
	r := &codec.Raw{}
	if err := pDecoder.Decode(r); err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("Recovered while processing", zap.Any("r", r))
		}
	}()

	o, err := object.FromBytes([]byte(*r))
	if err != nil {
		return nil, err
	}

	ra := ""
	if conn.RemotePeerKey != nil {
		ra = conn.RemotePeerKey.Hash
	}

	if o.GetSignature() != nil {
		if err := crypto.Verify(o); err != nil {
			return nil, err
		}
	}

	SendObjectEvent(
		"incoming",
		o.GetType(),
		pDecoder.NumBytesRead(),
	)
	if os.Getenv("DEBUG_BLOCKS") == "true" {
		b, _ := json.MarshalIndent(o.ToMap(), "", "  ")
		logger.Info(
			string(b),
			zap.String("remote_peer_hash", ra),
			zap.String("direction", "incoming"),
		)
	}
	return o, nil
}
