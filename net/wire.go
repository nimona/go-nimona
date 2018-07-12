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

// const SignMessageCBORTag = 98

// // GetCOSEHandle returns a codec.CborHandle with an extension
// // registered for COSE SignMessage as CBOR tag 98
// func GetCOSEHandle() (h *codec.CborHandle) {
// 	h = new(codec.CborHandle)
// 	h.IndefiniteLength = false // no streaming
// 	h.Canonical = true         // sort map keys
// 	h.SignedInteger = true

// 	var cExt cose.Ext
// 	h.SetInterfaceExt(reflect.TypeOf(&Message{}), SignMessageCBORTag, cExt)
// 	return h
// }

// var coseHandler = GetCOSEHandle()
// var coseHandler = cose.GetCOSEHandle()

// EventHandler for net.HandleExtensionEvents
type EventHandler func(event *Message) error

// Wire interface for mocking wire
type Wire interface {
	HandleExtensionEvents(extension string, h EventHandler) error
	Send(ctx context.Context, message *Message) error
	// Pack(payloadType string, payload interface{}, recipient string) ([]byte, error)
	Listen(addr string) (net.Listener, string, error)
}

type wire struct {
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

// NewWire creates a new wire protocol based on a addressBook
func NewWire(addressBook PeerManager) (Wire, error) {
	ctx := context.Background()

	w := &wire{
		incoming:    make(chan net.Conn),
		outgoing:    make(chan net.Conn),
		close:       make(chan bool),
		addressBook: addressBook,
		handlers:    map[string]EventHandler{},
		logger:      log.Logger(ctx).Named("wire"),
		streamLock:  utils.NewKmutex(),
	}

	return w, nil
}

func (w *wire) HandleExtensionEvents(extension string, h EventHandler) error {
	w.handlersLock.Lock()
	defer w.handlersLock.Unlock()
	if _, ok := w.handlers[extension]; ok {
		return errors.New("There already is a handler registered for this extension")
	}
	w.handlers[extension] = h
	return nil
}

func (w *wire) Close(peerID string, conn net.Conn) {
	// ctx := context.Background()

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

	// if peerID != "" {
	// 	go w.GetOrDial(ctx, peerID)
	// }
}

func (w *wire) HandleIncoming(conn net.Conn) error {
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
			if message.Headers.ContentType == "handshake" {
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

func (w *wire) HandleOutgoing(conn net.Conn) error {
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

// func (w *wire) ForwardMessage(message *Message) error {
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

// func (w *wire) DecodeMessage(envelope *Envelope) (*Message, error) {
// 	msgpackHandler := new(codec.MsgpackHandle)
// 	msgpackHandler.RawToString = false

// 	// decrypt message
// 	r := bytes.NewReader(envelope.Message)
// 	_, pr, err := saltpack.NewDecryptStream(saltpack.SingleVersionValidator(saltpack.CurrentVersion()), r, w.keyring)
// 	if err != nil {
// 		if err == saltpack.ErrNoDecryptionKey {
// 			return nil, ErrNotForUs
// 		}
// 		return nil, err
// 	}

// 	body, errRead := ioutil.ReadAll(pr)
// 	if errRead != nil {
// 		return nil, errRead
// 	}

// 	// decode message
// 	message := &Message{}
// 	messageDecoder := codec.NewDecoderBytes(body, msgpackHandler)
// 	if err := messageDecoder.Decode(message); err != nil {
// 		return nil, err
// 	}

// 	return message, nil
// }

// Process incoming message
// TODO Process from channel, don't call directly so we don't have to lock handlers
func (w *wire) Process(message *Message) error {
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

// Unpack an encoded envelope into a message
// TODO Unpack is not currently used, remove?
// func (w *wire) Unpack(encodedEnvelope []byte) (*Message, error) {
// 	// decode envelope
// 	envelope := &Envelope{}
// 	msgpackHandler := new(codec.MsgpackHandle)
// 	msgpackHandler.RawToString = false
// 	envelopeDecoder := codec.NewDecoderBytes(encodedEnvelope, msgpackHandler)
// 	if err := envelopeDecoder.Decode(&envelope); err != nil {
// 		return nil, err
// 	}

// 	return w.DecodeMessage(envelope)
// }

func (w *wire) Send(ctx context.Context, message *Message) error {
	for _, recipient := range message.Headers.Recipients {
		// TODO deal with error
		if err := w.SendOne(ctx, message, recipient); err != nil {
			fmt.Println("ERR SENDING", err)
			return err
		}
	}
	return nil
}

func (w *wire) SendOne(ctx context.Context, message *Message, recipient string) error {
	logger := w.logger.With(zap.String("peerID", recipient))
	logger.Debug("sending message", zap.String("contentType", message.Headers.ContentType))

	w.streamLock.Lock(recipient)
	defer w.streamLock.Unlock(recipient)

	// create message for recipient
	// encodedMessage, err := w.Pack(contentType, payload, recipient)
	// if err != nil {
	// 	return err
	// }

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

	// else, try to find a relay peer and send it to them
	// pi, err := w.addressBook.GetPeerInfo(peerID)
	// if err != nil {
	// 	logger.Debug("could not get relay for recipient", zap.Error(err))
	// 	return err
	// }

	// // go through the addresses and find any relays
	// relayIDs := []string{}
	// for _, address := range pi.Addresses {
	// 	if strings.HasPrefix(address, "relay:") {
	// 		relayIDs = append(relayIDs, strings.Replace(address, "relay:", "", 1))
	// 	}
	// }

	// // if no relays found, fail
	// if len(relayIDs) == 0 {
	// 	return ErrAllAddressesFailed
	// }

	// // so, we need to create message for recipient, but not anonymous
	// encodedMessage, err = w.Pack(contentType, payload, peerID)
	// if err != nil {
	// 	logger.Debug("could not pack message for relay", zap.Error(err))
	// 	return err
	// }

	// // go through the relays and send the message
	// for _, relayID := range relayIDs {
	// 	if err := w.writeMessage(ctx, encodedMessage, relayID); err != nil {
	// 		logger.Debug("could not send to relay", zap.Error(err))
	// 	} else {
	// 		return nil
	// 	}
	// }

	// else fail
	return ErrAllAddressesFailed
}

// Pack a message into something we can send
// func (w *wire) Pack(contentType string, payload interface{}, recipient string) ([]byte, error) {
// 	lpi := w.addressBook.GetLocalPeerInfo()

// 	message := &Message{}
// 	message.Headers.ContentType = contentType
// 	message.Headers.Recipients = []string{recipient}
// 	if err := message.EncodePayload(payload); err != nil {
// 		return nil, err
// 	}

// 	messageBytes, err := Sign(message, lpi)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// payloadBytes := []byte{}
// 	// enc := codec.NewEncoderBytes(&payloadBytes, &codec.CborHandle{})
// 	// if err := enc.Encode(payload); err != nil {
// 	// 	return nil, err
// 	// }

// 	// // TODO get correct alg
// 	// signer, err := cose.NewSignerFromKey(cose.PS256, lpi.GetSecretKey())
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// sig := cose.NewSignature()
// 	// // sig.Headers.Unprotected["kid"] = 1
// 	// sig.Headers.Protected["content-type"] = contentType
// 	// sig.Headers.Protected["alg"] = "PS256"
// 	// sig.Headers.Protected["recipients"] = []string{recipient}

// 	// create a message
// 	// external := []byte("") // optional external data see https://tools.ietf.org/html/rfc8152#section-4.3

// 	// msg := cose.NewSignMessage()
// 	// msg.Payload = payloadBytes
// 	// msg.AddSignature(sig)
// 	// msg.Headers.Protected["foo"] = "bar"

// 	// err = msg.Sign(rand.Reader, nil, []cose.Signer{*signer})
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// msgBytes := []byte{}
// 	// coseEnc := codec.NewEncoderBytes(&msgBytes, coseHandler)
// 	// if err := coseEnc.Encode(msg); err != nil {
// 	// 	return nil, err
// 	// }

// 	return messageBytes, nil
// }

func (w *wire) writeMessage(ctx context.Context, message *Message, rw io.ReadWriter) error {
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

func (w *wire) GetOrDial(ctx context.Context, peerID string) (net.Conn, error) {
	w.logger.Debug("getting conn", zap.String("peer_id", peerID))
	if peerID == "" {
		return nil, errors.New("missing peer id")
	}

	existingConn, ok := w.streams.Load(peerID)
	if ok {
		return existingConn.(net.Conn), nil
	}

	// w.streamLock.Lock(peerID)
	// defer w.streamLock.Unlock(peerID)

	w.logger.Debug("dialing peer", zap.String("peer_id", peerID))
	newConn, err := w.Dial(ctx, peerID)
	if err != nil {
		return nil, err
	}

	w.streams.Store(peerID, newConn)
	return newConn, nil
}

// Dial to a peer and return a net.Conn
func (w *wire) Dial(ctx context.Context, peerID string) (net.Conn, error) {
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
func (w *wire) Listen(addr string) (net.Listener, string, error) {
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
