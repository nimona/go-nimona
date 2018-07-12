package net

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	igd "github.com/emersion/go-upnp-igd"
	"github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/utils"
)

var (
	// ErrAllAddressesFailed for when a peer cannot be dialed
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
	// ErrNotForUs message is not meant for us
	ErrNotForUs = errors.New("message not for us")
)

// EventHandler for net.HandleExtensionEvents
type EventHandler func(event *Message) error

// Messenger interface for mocking messenger
type Messenger interface {
	HandleExtensionEvents(extension string, h EventHandler) error
	Send(ctx context.Context, message *Message) error
	Listen(addr string) (net.Listener, string, error)
}

type messenger struct {
	incoming     chan net.Conn
	outgoing     chan net.Conn
	close        chan bool
	addressBook  PeerManager
	streams      sync.Map
	handlers     map[string]EventHandler
	handlersLock sync.RWMutex
	logger       *zap.Logger
	streamLock   utils.Kmutex
}

// NewMessenger creates a new messenger protocol based on a addressBook
func NewMessenger(addressBook PeerManager) (Messenger, error) {
	ctx := context.Background()

	w := &messenger{
		incoming:    make(chan net.Conn),
		outgoing:    make(chan net.Conn),
		close:       make(chan bool),
		addressBook: addressBook,
		handlers:    map[string]EventHandler{},
		logger:      log.Logger(ctx).Named("messenger"),
		streamLock:  utils.NewKmutex(),
	}

	return w, nil
}

func (w *messenger) HandleExtensionEvents(extension string, h EventHandler) error {
	w.handlersLock.Lock()
	defer w.handlersLock.Unlock()
	if _, ok := w.handlers[extension]; ok {
		return errors.New("There already is a handler registered for this extension")
	}
	w.handlers[extension] = h
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

func (w *messenger) HandleIncoming(conn net.Conn) error {
	w.logger.Debug("handling new incoming connection", zap.String("remote", conn.RemoteAddr().String()))
	remotePeerID := ""

	messages := make(chan *Message, 100)

	go func() {
		for message := range messages {
			// TODO check recipients
			// if not-ours {
			// 	if err == ErrNotForUs {
			// 		if err := w.ForwardEnvelope(envelope); err != nil {
			// 			w.logger.Error("could not relay envelope", zap.Error(err))
			// 		}
			// 		continue
			// 	}
			// 	w.logger.Error("could not get message from envelope", zap.Error(err))
			// 	continue
			// }
			// TODO make sure this only happens once
			if message.Headers.ContentType == "net.handshake" {
				handshake := &handshakeMessage{}
				if err := message.DecodePayload(handshake); err != nil {
					w.logger.Error("could not decode handshake", zap.Error(err))
					continue
				}
				remotePeerID = handshake.PeerID
				w.streams.Store(remotePeerID, conn)
				continue
			}
			if err := w.Process(message); err != nil {
				// w.logger.Error("could not process line", zap.Error(err))
				// TODO should we return err?
				w.logger.Error("could not process message", zap.Error(err))
				w.Close(remotePeerID, conn)
				// return err
			}
		}
	}()

	defer close(messages)

	messageDecoder := codec.NewDecoder(conn, &codec.CborHandle{})

	for {
		// timeoutDuration := time.Minute
		// conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		message := &Message{}
		if err := messageDecoder.Decode(message); err != nil {
			// if err != io.EOF {
			w.logger.Error("could not read message", zap.Error(err))
			w.Close(remotePeerID, conn)
			// }
			return err
			// continue
		}
		if err := message.Verify(); err != nil {
			w.logger.Warn("could not verify message", zap.Error(err))
			continue
		}
		messages <- message
	}
}

func (w *messenger) HandleOutgoing(conn net.Conn) error {
	w.logger.Debug("handling new outgoing connection", zap.String("remote", conn.RemoteAddr().String()))

	messages := make(chan *Message, 100)

	go func() {
		for message := range messages {
			// TODO check if message is not for us
			// if err != nil {
			// 	if err == ErrNotForUs {
			// 		if err := w.ForwardEnvelope(envelope); err != nil {
			// 			w.logger.Error("could not relay envelope", zap.Error(err))
			// 		}
			// 		continue
			// 	}
			// 	w.logger.Error("could not get message from envelope", zap.Error(err))
			// 	continue
			// }
			if err := w.Process(message); err != nil {
				// TODO should we return err?
				w.logger.Error("could not process message", zap.Error(err))
				w.Close("", conn)
				// return err
			}
		}
	}()

	defer close(messages)

	msgpackHandler := new(codec.MsgpackHandle)
	msgpackHandler.RawToString = false

	messageDecoder := codec.NewDecoder(conn, &codec.CborHandle{})

	for {
		message := &Message{}
		if err := messageDecoder.Decode(message); err != nil {
			w.logger.Error("could not read message", zap.Error(err))
			w.Close("", conn)
			return err
		}
		messages <- message
	}
}

// func (w *messenger) ForwardMessage(message *Message) error {
// 	messageDecoder := codec.NewDecoder(conn, &codec.CborHandle{})

// 	// decrypt message
// 	r := bytes.NewReader(envelope.Message)
// 	mki, _, err := saltpack.NewDecryptStream(saltpack.SingleVersionValidator(saltpack.CurrentVersion()), r, w.keyring)
// 	if err != nil {
// 		if err == saltpack.ErrNoDecryptionKey {
// 			if len(mki.NamedReceivers) == 0 {
// 				return errors.New("no recipients")
// 			}
// 			for _, receiver := range mki.NamedReceivers {
// 				recipientID := fmt.Sprintf("%x", receiver)
// 				ctx := context.Background() // TODO Fix context
// 				encodedEnvelope := []byte{}
// 				envelopeEncoder := codec.NewEncoderBytes(&encodedEnvelope, msgpackHandler)
// 				if err := envelopeEncoder.Encode(envelope); err != nil {
// 					return err
// 				}
// 				if err := w.writeMessage(ctx, encodedEnvelope, recipientID); err != nil {
// 					return err
// 				}
// 			}
// 			return nil
// 		}
// 		return err
// 	}

// 	return errors.New("message is meant for us, no need to forward")
// }

// Process incoming message
// TODO Process from channel, don't call directly so we don't have to lock handlers
func (w *messenger) Process(message *Message) error {
	w.handlersLock.RLock()
	defer w.handlersLock.RUnlock()
	contentType := message.Headers.ContentType
	var handler EventHandler
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
			"No handler registered for extension",
			zap.String("contentType", contentType),
		)
		return nil
	}

	if err := handler(message); err != nil {
		w.logger.Info(
			"Could not handle event",
			zap.String("contentType", contentType),
			zap.Error(err),
		)
		return errors.New("could not handle")
	}

	return nil
}

func (w *messenger) Send(ctx context.Context, message *Message) error {
	for _, recipient := range message.Headers.Recipients {
		// TODO deal with error
		if err := w.SendOne(ctx, message, recipient); err != nil {
			// TODO log error
			return err
		}
	}
	return nil
}

func (w *messenger) SendOne(ctx context.Context, message *Message, recipient string) error {
	logger := w.logger.With(zap.String("peerID", recipient))
	logger.Debug("sending message", zap.String("contentType", message.Headers.ContentType))

	w.streamLock.Lock(recipient)
	defer w.streamLock.Unlock(recipient)

	w.logger.Debug("getting conn to write message", zap.String("recipient", recipient))
	conn, err := w.GetOrDial(ctx, recipient)
	if err != nil {
		return err
	}

	// try to send the message directly to the recipient
	if err := w.writeMessage(ctx, message, conn); err != nil {
		w.Close(recipient, conn)
		logger.Debug("could not send directly to recipient", zap.Error(err))
	} else {
		return nil
	}

	return ErrAllAddressesFailed
}

func (w *messenger) writeMessage(ctx context.Context, message *Message, rw io.ReadWriter) error {
	signer := w.addressBook.GetLocalPeerInfo()

	if !message.IsSigned() {
		if err := message.Sign(signer); err != nil {
			return err
		}
	}

	messageBytes, err := Marshal(message)
	if err != nil {
		return err
	}

	if _, err := rw.Write(messageBytes); err != nil {
		return err
	}

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
	newConn, err := w.Dial(ctx, peerID)
	if err != nil {
		return nil, err
	}

	w.streams.Store(peerID, newConn)
	return newConn, nil
}

// Dial to a peer and return a net.Conn
func (w *messenger) Dial(ctx context.Context, peerID string) (net.Conn, error) {
	peerInfo, err := w.addressBook.GetPeerInfo(peerID)
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	for _, addr := range peerInfo.Addresses {
		if !strings.HasPrefix(addr, "tcp:") {
			continue
		}
		addr = strings.Replace(addr, "tcp:", "", 1)
		dialer := net.Dialer{Timeout: time.Second * 5}
		newConn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			// TODO(superdecimal) Address can be black-listed maybe?
			continue
		}
		conn = newConn
		break
	}

	if conn == nil {
		// TODO(superdecimal) Mark peer as non-connectable directly
		return nil, ErrAllAddressesFailed
	}

	// TODO move after handshake
	// handle outgoing connections
	w.outgoing <- conn

	// store conn for reuse
	w.streams.Store(peerID, conn)

	w.logger.Debug("writing handshake")

	// handshake so the other side knows who we are
	handshake := &handshakeMessage{
		PeerID: w.addressBook.GetLocalPeerInfo().ID,
	}

	handshakeMessage := &Message{
		Headers: Headers{
			ContentType: "net.handshake",
			Recipients:  []string{peerID},
		},
	}
	if err := handshakeMessage.EncodePayload(handshake); err != nil {
		return nil, err
	}

	if err := w.writeMessage(ctx, handshakeMessage, conn); err != nil {
		return nil, err
	}

	// TODO(superdecimal) Mark peer as connectable directly

	return conn, nil
}

// Listen on an address
// TODO do we need to return a listener?
func (w *messenger) Listen(addr string) (net.Listener, string, error) {
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", err
	}
	port := tcpListener.Addr().(*net.TCPAddr).Port

	newAddresses := make(chan string, 100)
	devices := make(chan igd.Device)
	go func() {
		for device := range devices {
			upnp := true
			upnpFlag := os.Getenv("UPNP")
			if upnpFlag != "" {
				upnp, _ = strconv.ParseBool(upnpFlag)
			}
			if !upnp {
				continue
			}
			externalAddress, err := device.GetExternalIPAddress()
			if err != nil {
				w.logger.Error("could not get external ip", zap.Error(err))
				continue
			}
			desc := "nimona"
			ttl := time.Hour * 24 * 365
			if _, err := device.AddPortMapping(igd.TCP, port, port, desc, ttl); err != nil {
				w.logger.Error("could not add port mapping", zap.Error(err))
			} else {
				newAddresses <- fmt.Sprintf("tcp:%s:%d", externalAddress.String(), port)
			}
		}
		close(newAddresses)
	}()

	go func() {
		if err := igd.Discover(devices, 5*time.Second); err != nil {
			close(newAddresses)
			w.logger.Error("could not discover devices", zap.Error(err))
		}

		addresses := GetAddresses(tcpListener)
		for newAddress := range newAddresses {
			addresses = append(addresses, newAddress)
		}

		// TODO Replace with actual relay peer ids
		addresses = append(addresses, "relay:7730b73e34ae2e3ad92235aefc7ee0366736602f96785e6f35e8b710923b4562")

		localPeerInfo := w.addressBook.GetLocalPeerInfo()
		localPeerInfo.Addresses = addresses
		w.addressBook.PutLocalPeerInfo(localPeerInfo)
	}()

	closed := false

	go func() {
		for {
			select {
			case conn := <-w.incoming:
				go w.HandleIncoming(conn)
			case conn := <-w.outgoing:
				go w.HandleOutgoing(conn)
			case <-w.close:
				closed = true
				w.logger.Debug("connection closed")
				tcpListener.Close()
			}
		}
	}()

	go func() {
		w.logger.Debug("accepting connections", zap.String("address", tcpListener.Addr().String()))
		for {
			conn, err := tcpListener.Accept()
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

	return tcpListener, tcpListener.Addr().String(), nil
}
