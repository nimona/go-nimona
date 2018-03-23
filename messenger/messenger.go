package messenger

import (
	"bufio"
	"context"
	"encoding/json"
	"regexp"
	"sync"

	"go.uber.org/zap"

	"github.com/nimona/go-nimona/net"
)

// logger "github.com/Financial-Times/go-logger"
// "github.com/sirupsen/logrus"
// sha3 "golang.org/x/crypto/sha3"

// net "github.com/nimona/go-nimona/net"
// telemetry "github.com/nimona/go-telemetry"

type Messenger interface {
	Publish(msg *Message) error
	Subscribe(topics ...string) (chan *Message, error)
	Unsubscribe(chan *Message) error
}

type messenger struct {
	sync.RWMutex

	net net.Net

	incomingQueue chan *Message
	outgoingQueue chan *Message

	subscriptionTopicMatches map[string]*regexp.Regexp
	subscriptions            map[string]map[chan interface{}]bool
	channelSize              int

	logger *zap.Logger
}

func NewMessenger(nn net.Net) (Messenger, error) {
	ctx := context.Background()

	m := &messenger{
		net:                      nn,
		incomingQueue:            make(chan *Message, 100),
		outgoingQueue:            make(chan *Message, 100),
		subscriptionTopicMatches: map[string]*regexp.Regexp{},
		subscriptions:            map[string]map[chan interface{}]bool{},
		channelSize:              100,
		logger:                   net.Logger(ctx),
	}

	go func() {
		for {
			msg := <-m.incomingQueue
			if err := m.Publish(msg); err != nil {
				m.logger.Warn("could not publish incoming message", zap.Error(err))
			}
		}
	}()

	return m, nil
}

// Negotiate will be called after all the other protocol have been processed
func (m *messenger) Negotiate(fn net.NegotiatorFunc) net.NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c net.Conn) error {
		return nil
	}
}

// Handle ping requests
func (m *messenger) Handle(fn net.HandlerFunc) net.HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c net.Conn) error {
		sr := bufio.NewReader(c)
		for {
			// read line
			line, err := sr.ReadString('\n')
			if err != nil {
				return err // TODO Return?
			}

			// decode message
			msg := &Message{}
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				return err
			}

			// process message
			m.incomingQueue <- msg

			// if err := d.handleMessage(msg); err != nil {
			// logrus.WithError(err).Warnf("Could not process message")
			// }
		}
	}
}

func (ps *messenger) Publish(msg *Message) error {
	// ps.net.DialContext(ctx)
	return nil
}

func (ps *messenger) Subscribe(topics ...string) (chan *Message, error) {

	return nil, nil
}

func (ps *messenger) Unsubscribe(chan *Message) error {

	return nil
}

// func New(protocolID string, nn net.Net) (Messenger, error) {
// 	eb := &messenger{
// 		net:             network,
// 		handlers:        []func(hash []byte, msg Message) error{},
// 		handledMessages: []string{},
// 		streams:         map[string]net.Conn{},
// 		queue:           make(chan Envelope, 1000),
// 	}
// 	if err := network.RegisterStreamHandler(protocolID, eb.streamHander); err != nil {
// 		return nil, err
// 	}
// 	go func() {
// 		for {
// 			env := <-eb.queue
// 			if err := eb.send(&env); err != nil {
// 				logger.WithError(err).Errorf("Could not send envelope")
// 			}
// 		}
// 	}()
// 	return eb, nil
// }

// func (eb *messenger) HandleMessage(handler func(hash []byte, msg Message) error) error {
// 	eb.handlers = append(eb.handlers, handler)
// 	return nil
// }

// func (eb *messenger) streamHander(protocolID string, stream io.ReadWriteCloser) error {
// 	sr := bufio.NewReader(stream)
// 	for {
// 		if err := eb.read(protocolID, sr); err != nil {
// 			return err
// 		}
// 	}
// }

// func (eb *messenger) read(protocolID string, sr *bufio.Reader) error {
// 	tfields := map[string]interface{}{
// 		"protocol": protocolID,
// 	}
// 	defer telemetry.Publish("messaging:message:received", tfields)

// 	// read line
// 	// TODO replaces with proper stream decoder
// 	line, err := sr.ReadString('\n')
// 	if err != nil {
// 		logger.WithError(err).Errorf("Could not read")
// 		return err // TODO(geoah) Return?
// 	}

// 	lbs := []byte(line)

// 	tfields["envelope_bytes"] = len(lbs)

// 	// decode envelope
// 	ev := &Envelope{}
// 	if err := json.Unmarshal(lbs, &ev); err != nil {
// 		fmt.Println("Could not decode envelope") // TODO Fix logging
// 		return err
// 	}

// 	// verify hash
// 	// TODO verify hash

// 	// decode message
// 	if err := json.Unmarshal(ev.MessageRaw, &ev.Message); err != nil {
// 		fmt.Println("Could not decode message") // TODO Fix logging
// 		return err                              // TODO Should this fail? Do we even care at this point?
// 	}

// 	prbs := []byte(ev.Message.PayloadRaw)

// 	tfields["type"] = ev.Message.Payload.Type
// 	tfields["payload_bytes"] = len(prbs)

// 	// decode payload
// 	if err := json.Unmarshal(prbs, &ev.Message.Payload); err != nil {
// 		fmt.Println("Could not decode payload") // TODO Fix logging
// 		return nil                              // TODO Should this fail? Do we even care at this point?
// 	}

// 	// get creator
// 	// TODO Make verification optional ena re-enable
// 	// creator, err := eb.network.GetPeer(ev.Message.Payload.Creator)
// 	// if err != nil {
// 	// 	// TODO attempt to retrieve creator and retry?
// 	// 	fmt.Println("Unknown creator")
// 	// 	return err
// 	// }

// 	// verify signature
// 	// valid, err := creator.Verify(ev.Message.PayloadRaw, ev.Message.Signature)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// if valid == false {
// 	// 	return errors.New("Invalid signature") // TODO Better error
// 	// }

// 	// sent message to handlers
// 	eb.handle(ev.Hash, ev.Message)

// 	return nil
// }

// func (eb *messenger) getStream(peerID string) (net.Conn, error) {
// 	if stream, ok := eb.streams[peerID]; ok && stream != nil {
// 		// TODO Check if stream is still ok
// 		logger.Debugf("Found stream for peer %s", peerID)
// 		return stream, nil
// 	}
// 	addr := peerID + "/" + eb.protocolID
// 	logrus.WithField("addr", addr).Infof("Dialing peer for messabus.getStream")
// 	stream, err := eb.net.Dial(addr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	logger.Debugf("Created new stream for peer %s", peerID)
// 	eb.streams[peerID] = stream
// 	return stream, nil
// }

// // send an event to a peer
// func (eb *messenger) send(ev *Envelope) error {
// 	tfields := map[string]interface{}{
// 		"protocol":       eb.protocolID,
// 		"envelope_bytes": len(ev.MessageRaw),
// 		"payload_bytes":  len(ev.Message.PayloadRaw),
// 	}
// 	defer telemetry.Publish("messaging:message:sent", tfields)

// 	stream, err := eb.getStream(ev.Recipipent)
// 	if err != nil {
// 		return err
// 	}
// 	evbs, _ := json.Marshal(ev)
// 	evbs = append(evbs, '\n')
// 	if _, err := stream.Write(evbs); err != nil {
// 		if err == syscall.EPIPE || err.Error() == "broken pipe" {
// 			logrus.WithError(err).Warnf("Writing to stream failed, removing stream")
// 			delete(eb.streams, ev.Recipipent)
// 			// put envelope back to the queue
// 			eb.queue <- *ev
// 			tfields["error"] = "broken_pipe"
// 		} else {
// 			tfields["error"] = "generic"
// 		}
// 		logrus.WithError(err).WithField("evbs", string(evbs)).Errorf("Writing to stream failed")
// 		return err
// 	}

// 	// logger.WithField("evbs", string(evbs)).Debugf("Wrote envelope to peer %s (%d bytes)", peerID, n)
// 	return nil
// }

// // Send takes an Event and will send it to its intended recipients
// func (eb *messenger) Send(payload *Payload, peerIDs []string) error {
// 	// encode payload
// 	bpay, err := json.Marshal(payload)
// 	if err != nil {
// 		return err
// 	}

// 	// sign message
// 	// TODO Make signing optional and readd
// 	// spay, err := eb.peer.Sign(bpay)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// create signed message
// 	msg := &Message{
// 		PayloadRaw: bpay,
// 		// Signature:  spay,
// 	}

// 	// encode message
// 	bmsg, err := json.Marshal(msg)
// 	if err != nil {
// 		return err
// 	}

// 	// hash bmsg
// 	shash := sha3.Sum256(bmsg)

// 	// create envelope
// 	ev := &Envelope{
// 		MessageRaw: bmsg,
// 		Hash:       shash[:],
// 	}

// 	// go through all recipiends
// 	for _, pid := range peerIDs {
// 		// don't send messagebus to ourselves
// 		if pid == string(eb.net.GetLocalPeer().ID) {
// 			continue
// 		}

// 		// logger.
// 		// 	WithField("type", payload.Type).
// 		// 	WithField("data", string(payload.Data)).
// 		// 	Infof("Sending message to %s", pid)

// 		// send the event to the peer
// 		ev.Recipipent = pid
// 		eb.queue <- *ev
// 	}

// 	return nil
// }

// // Handle an incoming event
// func (eb *messenger) handle(hash []byte, msg Message) error {
// 	// TODO event checking and adding are not thread safe

// 	// fmt.Println("Handling message", string(msg.Payload.Type), string(msg.Payload.Data))

// 	// check if we have already handled this event
// 	// shash := fmt.Sprintf("%x", hash)
// 	// for _, msgHash := range eb.handledMessages {
// 	// 	if msgHash == shash {
// 	// 		return nil
// 	// 	}
// 	// }

// 	// else add to handled messagebus
// 	// eb.handledMessages = append(eb.handledMessages, shash)

// 	// and trigger handlers
// 	for _, h := range eb.handlers {
// 		h(hash, msg)
// 	}

// 	return nil
// }
