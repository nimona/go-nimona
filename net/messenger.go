package net

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/utils"
)

func init() {
	RegisterContentType("net.handshake", HandshakeEnvelope{})
}

var (
	// ErrAllAddressesFailed for when a peer cannot be dialed
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
	// ErrNotForUs envelope is not meant for us
	ErrNotForUs = errors.New("envelope not for us")
)

// EnvelopeHandler to handle incoming envelopes
type EnvelopeHandler func(event *Envelope) error

// Messenger interface for mocking messenger
type Messenger interface {
	Handle(contentType string, h EnvelopeHandler) error
	Send(ctx context.Context, envelope *Envelope, recipients ...string) error
	Listen(ctx context.Context, addrress string) (net.Listener, error)
}

type messenger struct {
	network Networker
	// listener net.Listener

	addressBook PeerManager

	incoming  chan net.Conn
	outgoing  chan net.Conn
	envelopes chan *Envelope
	close     chan bool

	streams  sync.Map
	handlers map[string]EnvelopeHandler
	// handlersLock sync.RWMutex
	logger     *zap.Logger
	streamLock utils.Kmutex
}

// NewMessenger creates a messenger on a given network
func NewMessenger(addressBook *AddressBook) (Messenger, error) {
	ctx := context.Background()

	network, err := NewNetwork(addressBook)
	if err != nil {
		return nil, err
	}

	m := &messenger{
		network:     network,
		addressBook: addressBook,

		incoming:  make(chan net.Conn),
		outgoing:  make(chan net.Conn),
		envelopes: make(chan *Envelope, 100),
		close:     make(chan bool),

		handlers:   map[string]EnvelopeHandler{},
		logger:     log.Logger(ctx).Named("messenger"),
		streamLock: utils.NewKmutex(),
	}

	return m, nil
}

func (w *messenger) Handle(contentType string, h EnvelopeHandler) error {
	if _, ok := w.handlers[contentType]; ok {
		return errors.New("There already is a handler registered for this contentType")
	}
	w.handlers[contentType] = h
	return nil
}

func (w *messenger) Close(peerID string, conn net.Conn) {
	if conn != nil {
		conn.Close()
	}
	w.streams.Range(func(k, v interface{}) bool {
		if k.(string) == peerID {
			w.streams.Delete(k)
		}
		if v.(net.Conn) == conn {
			w.streams.Delete(k)
		}
		return true
	})
}

func (w *messenger) HandleConnection(conn net.Conn) error {
	w.logger.Debug("handling new connection", zap.String("remote", conn.RemoteAddr().String()))

	envelopeDecoder := codec.NewDecoder(conn, getCborHandler())
	for {
		envelope := &Envelope{}
		if err := envelopeDecoder.Decode(envelope); err != nil {
			w.logger.Error("could not read envelope", zap.Error(err))
			w.Close("", conn)
			return err
		}

		w.logger.Debug("handling envelope", zap.Any("envelope", envelope))

		if err := envelope.Verify(); err != nil {
			w.logger.Warn("could not verify envelope", zap.Error(err))
			return err
		}

		if envelope.Type == "net.handshake" {
			payload, ok := envelope.Payload.(HandshakeEnvelope)
			if !ok {
				continue
			}
			if err := w.addressBook.PutPeerInfoFromEnvelope(payload.PeerInfo); err != nil {
				// TODO handle error
				continue
			}
			w.streams.Store(envelope.Headers.Signer, conn)
			continue
		}

		contentType := envelope.Type
		var handler EnvelopeHandler
		ok := false
		for handlerContentType, hn := range w.handlers {
			if !strings.HasPrefix(contentType, handlerContentType) {
				continue
			}
			ok = true
			handler = hn
			break
		}

		if !ok {
			w.logger.Info(
				"No handler registered for contentType",
				zap.String("contentType", contentType),
			)
			return nil
		}

		if err := handler(envelope); err != nil {
			w.logger.Info(
				"Could not handle event",
				zap.String("contentType", contentType),
				zap.Error(err),
			)
			continue
		}
	}
}

// Process incoming envelope
func (w *messenger) Process(envelope *Envelope) error {
	if err := envelope.Verify(); err != nil {
		w.logger.Warn("could not verify envelope", zap.Error(err))
		return err
	}

	eb, _ := Marshal(envelope)
	tb, _ := Marshal(envelope.Payload)
	SendEnvelopeEvent(
		false,
		envelope.Type,
		len(envelope.Headers.Recipients),
		len(tb),
		len(eb),
	)

	contentType := envelope.Type
	var handler EnvelopeHandler
	ok := false
	for handlerContentType, hn := range w.handlers {
		if !strings.HasPrefix(contentType, handlerContentType) {
			continue
		}
		ok = true
		handler = hn
		break
	}

	if !ok {
		w.logger.Info(
			"No handler registered for contentType",
			zap.String("contentType", contentType),
		)
		return nil
	}

	if err := handler(envelope); err != nil {
		w.logger.Info(
			"Could not handle event",
			zap.String("contentType", contentType),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (w *messenger) Send(ctx context.Context, envelope *Envelope, recipients ...string) error {
	signer := w.addressBook.GetLocalPeerInfo()

	if !envelope.IsSigned() {
		if err := envelope.Sign(signer); err != nil {
			return err
		}
	}

	if len(recipients) == 0 {
		recipients = envelope.Headers.Recipients
	}

	for _, recipient := range recipients {
		// TODO deal with error
		if err := w.sendOne(ctx, envelope, recipient); err != nil {
			// TODO log error
			return err
		}
	}

	return nil
}

func (w *messenger) sendOne(ctx context.Context, envelope *Envelope, recipient string) error {
	logger := w.logger.With(zap.String("peerID", recipient))

	w.streamLock.Lock(recipient)
	defer w.streamLock.Unlock(recipient)

	w.logger.Debug("getting conn to write envelope", zap.String("recipient", recipient))
	conn, err := w.GetOrDial(ctx, recipient)
	if err != nil {
		return err
	}

	// TODO this seems messy
	// try to send the envelope directly to the recipient
	if err := w.writeEnvelope(ctx, envelope, conn); err != nil {
		w.Close(recipient, conn)
		logger.Debug("could not send directly to recipient", zap.Error(err))
	} else {
		return nil
	}

	return ErrAllAddressesFailed
}

func (w *messenger) writeEnvelope(ctx context.Context, envelope *Envelope, rw io.ReadWriter) error {
	signer := w.addressBook.GetLocalPeerInfo()

	if !envelope.IsSigned() {
		if err := envelope.Sign(signer); err != nil {
			return err
		}
	}

	envelopeBytes, err := Marshal(envelope)
	if err != nil {
		return err
	}

	if _, err := rw.Write(envelopeBytes); err != nil {
		return err
	}

	tb, _ := Marshal(envelope.Payload)
	SendEnvelopeEvent(
		true,
		envelope.Type,
		len(envelope.Headers.Recipients),
		len(tb),
		len(envelopeBytes),
	)

	w.logger.Debug("writing envelope", zap.Any("envelope", envelope))

	return nil
}

func (w *messenger) GetOrDial(ctx context.Context, peerID string) (net.Conn, error) {
	w.logger.Debug("getting conn", zap.String("peer_id", peerID))
	if peerID == "" {
		return nil, errors.New("missing peer id")
	}

	existingConn, ok := w.streams.Load(peerID)
	if ok {
		return existingConn.(net.Conn), nil
	}

	w.logger.Debug("dialing peer", zap.String("peer_id", peerID))
	conn, err := w.network.Dial(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// TODO move after handshake
	// handle outgoing connections
	w.outgoing <- conn

	// store conn for reuse
	w.streams.Store(peerID, conn)

	w.logger.Debug("writing handshake")

	// handshake so the other side knows who we are
	signer := w.addressBook.GetLocalPeerInfo()
	handshakeEnvelope := &Envelope{
		Type: "net.handshake",
		Headers: Headers{
			Recipients: []string{peerID},
		},
		Payload: HandshakeEnvelope{
			PeerInfo: w.addressBook.GetLocalPeerInfo().Envelope(),
		},
	}

	// TODO move someplace common with send()
	if err := handshakeEnvelope.Sign(signer); err != nil {
		return nil, err
	}

	if err := w.writeEnvelope(ctx, handshakeEnvelope, conn); err != nil {
		return nil, err
	}

	w.addressBook.PutPeerStatus(peerID, Connected)

	return conn, nil
}

// Listen on an address
// TODO do we need to return a listener?
func (w *messenger) Listen(ctx context.Context, addr string) (net.Listener, error) {
	listener, err := w.network.Listen(ctx, addr)
	if err != nil {
		return nil, err
	}

	closed := false

	go func() {
		for {
			select {
			case envelope := <-w.envelopes:
				if err := w.Process(envelope); err != nil {
					w.logger.Warn("failed to process envelope", zap.Error(err))
				}
			case conn := <-w.incoming:
				go func() {
					if err := w.HandleConnection(conn); err != nil {
						w.logger.Warn("failed to handle envelope", zap.Error(err))
					}
				}()
			case conn := <-w.outgoing:
				go func() {
					if err := w.HandleConnection(conn); err != nil {
						w.logger.Warn("failed to handle envelope", zap.Error(err))
					}
				}()
			case <-w.close:
				closed = true
				w.logger.Debug("connection closed")
				listener.Close()
			}
		}
	}()

	go func() {
		w.logger.Debug("accepting connections", zap.String("address", listener.Addr().String()))
		for {
			conn, err := listener.Accept()
			w.logger.Debug("connection accepted")
			if err != nil {
				if closed {
					return
				}
				w.logger.Error("could not accept", zap.Error(err))
				// TODO check conn is still alive and return
				return
			}
			w.incoming <- conn
		}
	}()

	return listener, nil
}
