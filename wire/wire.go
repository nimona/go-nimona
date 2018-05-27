package wire

import (
	"bufio"
	"context"
	"encoding/json"
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
	"github.com/keybase/saltpack"
	"github.com/keybase/saltpack/basic"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/mesh"
)

var (
	// ErrAllAddressesFailed for when a peer cannot be dialed
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
)

// EventHandler for wire.HandleExtensionEvents
type EventHandler func(event *Message) error

// Wire interface for mocking wire
type Wire interface {
	HandleExtensionEvents(extension string, h EventHandler) error
	Send(ctx context.Context, extension, payloadType string, payload interface{}, to []string) error
	Listen(addr string) (net.Listener, string, error)
}

type wire struct {
	accepted chan net.Conn
	close    chan bool
	keyring  *basic.Keyring
	registry mesh.Registry
	streams  map[string]io.ReadWriteCloser
	handlers map[string]EventHandler
	logger   *zap.Logger
}

// NewWire creates a new wire protocol based on a registry
func NewWire(reg mesh.Registry) (Wire, error) {
	ctx := context.Background()

	w := &wire{
		accepted: make(chan net.Conn),
		close:    make(chan bool),
		keyring:  mesh.Keyring,
		registry: reg,
		streams:  map[string]io.ReadWriteCloser{},
		handlers: map[string]EventHandler{},
		logger:   log.Logger(ctx).Named("wire"),
	}

	return w, nil
}

func (w *wire) HandleExtensionEvents(extension string, h EventHandler) error {
	if _, ok := w.handlers[extension]; ok {
		return errors.New("There already is a handler registered for this extension")
	}
	w.handlers[extension] = h
	return nil
}

func (w *wire) Handle(conn net.Conn) error {
	w.logger.Debug("Handling new connection")
	handshakeComplete := false
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		w.logger.Debug("Got line", zap.String("line", line))
		if !handshakeComplete {
			remotePeerID, err := w.ProcessHandshake(line)
			if err != nil {
				fmt.Println("Could not process handshake", err)
				return err
			}
			// TODO Lock
			w.streams[remotePeerID] = conn
			handshakeComplete = true
			continue
		}
		if err := w.Process(line); err != nil {
			w.logger.Debug("Could not process line", zap.Error(err))
			// return err
		}
	}

	return nil
}

func (w *wire) ProcessHandshake(encryptedBody string) (string, error) {
	msg, err := w.DecodeMessage(encryptedBody)
	if err != nil {
		return "", err
	}

	handshake := &handshakeMessage{}
	if err := msg.DecodePayload(&handshake); err != nil {
		return "", err
	}

	return handshake.PeerID, nil
}

func (w *wire) Process(encryptedBody string) error {
	msg, err := w.DecodeMessage(encryptedBody)
	if err != nil {
		return err
	}

	hn, ok := w.handlers[msg.Extension]
	if !ok {
		w.logger.Info(
			"No handler registered for extension",
			zap.String("extension", msg.Extension),
		)
		return errors.New("no handler")
	}

	if err := hn(msg); err != nil {
		w.logger.Info(
			"Could not handle event",
			zap.String("extension", msg.Extension),
			zap.String("payload_type", msg.PayloadType),
			zap.Error(err),
		)
		return errors.New("could not handle")
	}

	return nil
}

func (w *wire) DecodeMessage(encryptedBody string) (*Message, error) {
	mki, body, _, err := saltpack.Dearmor62DecryptOpen(saltpack.CheckKnownMajorVersion, encryptedBody, w.keyring)
	if err != nil {
		if err == saltpack.ErrNoDecryptionKey {
			fmt.Println("NAMED", mki.NamedReceivers)
			fmt.Println("RECIE", mki.ReceiverKey)
			if len(mki.NamedReceivers) > 0 {
				for _, receiver := range mki.NamedReceivers {
					recipientID := fmt.Sprintf("%x", receiver)
					fmt.Println("Trying to FW message to", recipientID)
					ctx := context.Background() // TODO Fix context
					if fwErr := w.sendMessage(ctx, encryptedBody+"\n", recipientID); fwErr != nil {
						fmt.Println("Could not relay message", fwErr)
						return nil, fwErr // TODO Should this return an error?
					}
				}
			}
			return nil, errors.New("Message was not for us")
		}
		return nil, err
	}

	msg := &Message{}
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		w.logger.Info(
			"Could not unmarshal  into type",
			zap.String("extension", msg.Extension),
			zap.String("payload_type", msg.PayloadType),
			zap.Error(err),
		)
		return nil, errors.New("could not unmarshal into type")
	}

	return msg, nil
}

func (w *wire) Send(ctx context.Context, extension, payloadType string, payload interface{}, recipients []string) error {
	if len(recipients) == 0 {
		return nil
	}

	pks := []saltpack.BoxPublicKey{}
	for _, rec := range recipients {
		pi, err := w.registry.GetPeerInfo(rec)
		if err != nil {
			return err
		}
		pks = append(pks, pi.GetPublicKey())
	}

	msg := &messageOut{
		Extension:   extension,
		PayloadType: payloadType,
		Payload:     payload,
	}

	encodedMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	fmt.Println("Sending message", string(encodedMsg))

	sk := w.registry.GetLocalPeerInfo().GetSecretKey()
	ecryptedMsg, err := saltpack.EncryptArmor62Seal(saltpack.CurrentVersion(), encodedMsg, sk, pks, "WIRE")
	if err != nil {
		return err
	}

	for _, rec := range recipients {
		if err := w.sendMessage(ctx, ecryptedMsg, rec); err != nil {
			fmt.Println("Could not send to peer", rec, err)
		}
	}
	return nil
}

func (w *wire) sendMessage(ctx context.Context, msg string, recipient string) error {
	logger := w.logger.With(zap.String("peerID", recipient))

	peerInfo, err := w.registry.GetPeerInfo(recipient)
	if err != nil {
		return err
	}

	stream, ok := w.streams[recipient]
	if !ok || stream == nil {
		conn, err := w.Dial(ctx, recipient)
		if err != nil {
			logger.Info("could not dial to peer, attempting to relay", zap.Error(err))
			// return err
			for _, address := range peerInfo.Addresses {
				if strings.HasPrefix(address, "relay:") {
					relayPeerID := strings.Replace(address, "relay:", "", 1)
					logger.Info("trying to relay message", zap.String("relay", relayPeerID))
					if relErr := w.sendMessage(ctx, msg, relayPeerID); relErr != nil {
						logger.Info("could not FW to relay", zap.Error(relErr))
						continue
					}
					logger.Info("relayed message via", zap.String("relay", relayPeerID))
					return nil
				}
			}
			logger.Info("could not dial to peer, failed to relay", zap.Error(err))
			return err
		}

		w.streams[recipient] = conn
		stream = conn
		handshake := &handshakeMessage{
			PeerID: w.registry.GetLocalPeerInfo().ID,
		}
		if err := w.Send(ctx, "wire", "handshake", handshake, []string{recipient}); err != nil {
			// TODO close and remove stream
			return err
		}
	}

	if _, err := stream.Write([]byte(msg)); err != nil {
		// TODO Lock
		logger.Warn("could not write outgoing message", zap.Error(err))
		delete(w.streams, recipient)
		stream.Close()
		return err
	}

	logger.Debug("Wrote message")
	return nil
}

func (w *wire) Dial(ctx context.Context, peerID string) (net.Conn, error) {
	fmt.Println("Dial()ing ", peerID)
	peerInfo, err := w.registry.GetPeerInfo(peerID)
	if err != nil {
		return nil, err
	}

	// addresses := peerInfo.Addresses
	// for _, address := range addresses {
	// 	if strings.HasPrefix(address, "relay:") {
	// 		relayPeerID := strings.Replace(address, "relay:", "", 1)
	// 		relayPeer, err := w.registry.GetPeerInfo(relayPeerID)
	// 		if err != nil {
	// 			continue
	// 		}
	// 		addresses = append(addresses, relayPeer.Addresses...)
	// 	}
	// }

	var conn net.Conn
	for _, addr := range peerInfo.Addresses {
		if !strings.HasPrefix(addr, "tcp:") {
			continue
		}
		addr = strings.Replace(addr, "tcp:", "", 1)
		// fmt.Println("dial dialing new conn to", addr)
		dialer := net.Dialer{Timeout: time.Second * 5}
		newConn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			// TODO blacklist address for a bit
			// TODO hold error maybe?
			// return nil, err
			fmt.Println("Coult not dial addr", addr, err)
			continue
		}
		conn = newConn
		break
	}
	if conn == nil {
		return nil, ErrAllAddressesFailed
	}

	w.accepted <- conn

	// handshake so the other side knows who we are

	if err != nil {
		if err := conn.Close(); err != nil {
			fmt.Println("could not close connection after failure to select")
		}
		fmt.Println("error selecting", err)
		return nil, err
	}

	return conn, nil
}

// TODO do we need to return a listener?
func (w *wire) Listen(addr string) (net.Listener, string, error) {
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", err
	}
	port := tcpListener.Addr().(*net.TCPAddr).Port

	lock := sync.Mutex{}
	addresses := GetAddresses(tcpListener)

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
				fmt.Println("could not get external ip")
				continue
			}
			desc := "nimona"
			ttl := time.Hour * 24 * 365
			if _, err := device.AddPortMapping(igd.TCP, port, port, desc, ttl); err != nil {
				fmt.Println("could not add port mapping", err)
			} else {
				lock.Lock()
				addresses = append(addresses, fmt.Sprintf("tcp:%s:%d", externalAddress.String(), port))
				lock.Unlock()
			}
		}
	}()

	if err := igd.Discover(devices, 5*time.Second); err != nil {
		fmt.Println("could not discover devices")
	}

	addresses = append(addresses, "relay:7730b73e34ae2e3ad92235aefc7ee0366736602f96785e6f35e8b710923b4562")
	w.registry.GetLocalPeerInfo().UpdateAddresses(addresses)

	go func() {
		for {
			conn := <-w.accepted
			go w.Handle(conn)
		}
	}()

	closed := false

	go func() {
		closed = true
		<-w.close
		// fmt.Println("Closing")
		tcpListener.Close()
	}()

	go func() {
		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				if closed {
					return
				}
				fmt.Println("Error accepting: ", err.Error())
				// TODO check conn is still alive and return
				return
			}
			w.accepted <- conn
		}
	}()

	return nil, tcpListener.Addr().String(), nil
}
