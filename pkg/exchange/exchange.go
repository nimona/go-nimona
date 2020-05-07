package exchange

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"time"

	"github.com/geoah/go-queue"
	"github.com/hashicorp/go-multierror"
	"github.com/patrickmn/go-cache"
	"github.com/zserge/metric"

	"nimona.io/pkg/connmanager"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

var (
	peerType        = new(peer.Peer).GetType()
	dataForwardType = new(DataForward).GetType()
)

const (
	// ErrInvalidRequest when received an invalid request object
	ErrInvalidRequest = errors.Error("invalid request")
	// ErrSendingTimedOut when sending times out
	ErrSendingTimedOut = errors.Error("sending timed out")
)

// nolint: lll
//go:generate $GOBIN/mockery -case underscore -inpkg -name Exchange
//go:generate $GOBIN/genny -in=$GENERATORS/syncmap_named/syncmap.go -out=addresses_generated.go -pkg=exchange gen "KeyType=string ValueType=addressState SyncmapName=addresses"
//go:generate $GOBIN/genny -in=$GENERATORS/syncmap_named/syncmap.go -out=outboxes_generated.go -imp=nimona.io/pkg/crypto -pkg=exchange gen "KeyType=crypto.PublicKey ValueType=outbox SyncmapName=outboxes"
//go:generate $GOBIN/genny -in=$GENERATORS/pubsub/pubsub.go -out=pubsub_envelopes_generated.go -pkg=exchange gen "ObjectType=*Envelope PubSubName=envelope"

// nolint: gochecknoinits
func init() {
	objHandledCounter := metric.NewCounter("2m1s", "15m30s", "1h1m")
	expvar.Publish("nm:exc.obj.received", objHandledCounter)

	objSentCounter := metric.NewCounter("2m1s", "15m30s", "1h1m")
	expvar.Publish("nm:exc.obj.send.success", objSentCounter)

	objRelayedCounter := metric.NewCounter("2m1s", "15m30s", "1h1m")
	expvar.Publish("nm:exc.obj.send.relayed", objRelayedCounter)

	objFailedCounter := metric.NewCounter("2m1s", "15m30s", "1h1m")
	expvar.Publish("nm:exc.obj.send.failed", objFailedCounter)

	objAttemptedCounter := metric.NewCounter("2m1s", "15m30s", "1h1m")
	expvar.Publish("nm:exc.obj.send", objAttemptedCounter)
}

type (
	// Exchange interface for mocking exchange
	Exchange interface {
		Subscribe(
			filters ...EnvelopeFilter,
		) EnvelopeSubscription
		Send(
			ctx context.Context,
			object object.Object,
			recipient peer.LookupOption,
			options ...SendOption,
		) error
		SendToPeer(
			ctx context.Context,
			object object.Object,
			p *peer.Peer,
			options ...SendOption,
		) error
	}
	// echange implements an Exchange
	exchange struct {
		net       net.Network
		connmgr   connmanager.Manager
		keychain  keychain.Keychain
		discover  discovery.PeerStorer
		outboxes  *OutboxesMap
		inboxes   EnvelopePubSub
		blacklist *cache.Cache
	}
	// // Option for creating a new exchange
	// Option func(*exchange)
	// SendOptions (mostly) for Send()
	SendOptions struct {
		LocalDiscovery bool
		Async          bool
	}
	SendOption func(*SendOptions)
	// addressState defines the states of a peer's address
	// current options are:
	// * -1 unconnectable
	// * 0 unknown
	// * 1 connectable
	// * 2 blacklisted
	addressState int
	// outbox holds information about a single peer, its open connection,
	// and the messages for it.
	// the queue should only hold `*outgoingObject`s.
	outbox struct {
		peer  crypto.PublicKey
		queue *queue.Queue
	}
	// outgoingObject holds an object that is about to be sent
	outgoingObject struct {
		context   context.Context
		recipient *peer.Peer
		object    object.Object
		options   *SendOptions
		err       chan error
	}
)

// New creates a exchange on a given network
func New(
	ctx context.Context,
	eb eventbus.Eventbus,
	kc keychain.Keychain,
	n net.Network,
	discover discovery.PeerStorer,
) (Exchange, error) {
	w := &exchange{
		keychain:  kc,
		net:       n,
		discover:  discover,
		outboxes:  NewOutboxesMap(),
		inboxes:   NewEnvelopePubSub(),
		blacklist: cache.New(10*time.Second, 5*time.Minute),
	}

	connmgr, err := connmanager.New(ctx, eb, n, w.handleConnection)
	if err != nil {
		return nil,
			errors.Wrap(
				errors.New("could not construct connection manager"),
				err,
			)
	}
	w.connmgr = connmgr

	// // add local peer to discoverer
	// // TODO this is mostly a hack as discover doesn't have access to local
	// w.discover.Add(w.local.GetSignedPeer(), true)

	logger := log.
		FromContext(ctx).
		Named("exchange").
		With(
			log.String("method", "exchange.New"),
			log.String(
				"local.peer",
				w.keychain.GetPrimaryPeerKey().PublicKey().String(),
			),
		)

	// subscribe to data forward type
	dataForwardSub := w.inboxes.Subscribe(
		FilterByObjectType(dataForwardType),
	)

	go func() {
		if err := w.handleObjects(dataForwardSub); err != nil {
			logger.Error("handling object requests failed", log.Error(err))
		}
	}()

	// subscribe to peers and handle them
	peerSub := w.inboxes.Subscribe(
		FilterByObjectType(peerType),
	)
	go func() {
		if err := w.handlePeers(peerSub); err != nil {
			logger.Error("handling peer failed", log.Error(err))
		}
	}()

	return w, nil
}

func (w *exchange) handleConnection(conn *net.Connection) error {
	expvar.Get("nm:exc.obj.received").(metric.Metric).Add(1)

	if conn == nil {
		return errors.New("missing connection")
	}

	go func() {
		for {
			payload, err := net.Read(conn)
			// TODO split errors into connection or payload
			// ie a payload that cannot be unmarshalled or verified
			// should not kill the connection
			if err != nil {
				log.DefaultLogger.Warn(
					"failed to read from connection",
					log.Error(err),
				)
				return
			}

			log.DefaultLogger.Debug(
				"reading from connection",
				log.String("payload", payload.GetType()),
			)

			w.inboxes.Publish(&Envelope{
				Sender:  conn.RemotePeerKey,
				Payload: *payload,
			})
		}
	}()
	return nil
}

func (w *exchange) getOutbox(recipient crypto.PublicKey) *outbox {
	outbox := &outbox{
		peer:  recipient,
		queue: queue.New(),
	}
	outbox, loaded := w.outboxes.GetOrPut(recipient, outbox)
	if !loaded {
		go w.processOutbox(outbox)
	}
	return outbox
}

func (w *exchange) processOutbox(outbox *outbox) {
	for {
		// dequeue the next item to send
		// TODO figure out what can go wrong here
		v := outbox.queue.Pop()
		req := v.(*outgoingObject)
		// check if the context for this is done
		if err := req.context.Err(); err != nil {
			req.err <- err
			continue
		}
		// make a logger from our req context
		logger := log.FromContext(req.context).With(
			log.String("recipient", req.recipient.PublicKey().String()),
			log.String("object.type", req.object.GetType()),
		)
		// validate req
		if req.recipient == nil {
			logger.Info("missing recipient")
			req.err <- errors.New("missing recipient")
			continue
		}
		// try to send the object
		var lastErr error
		maxAttempts := 1
		for i := 0; i < maxAttempts; i++ {
			logger.Debug("trying to get connection", log.Int("attempt", i+1))

			conn, err := w.connmgr.GetConnection(req.context, req.recipient)
			if err != nil {
				lastErr = err
				continue
			}
			logger.Debug("trying write object", log.Int("attempt", i+1))
			if err := net.Write(req.object, conn); err != nil {
				lastErr = err
				continue
			}
			lastErr = nil
			expvar.Get("nm:exc.obj.send.success").(metric.Metric).Add(1)
			break
		}

		// Only try the relays if we fail to write the object
		if lastErr != nil && req.object.GetType() != "nimona.io/LookupRequest" {
			for _, relayPeer := range req.recipient.Relays {
				df, err := wrapInDataForward(
					req.object,
					req.recipient.PublicKey(),
					relayPeer,
				)
				if err != nil {
					logger.Error(
						"could not create data forward object",
						log.Error(err),
					)
					continue
				}
				logger.Debug(
					"trying relay peer",
					log.String("relay", relayPeer.String()),
					log.String("recipient", req.recipient.PublicKey().String()),
				)
				ctx := context.New(
					context.WithTimeout(time.Second * 5),
				)
				// send the newly wrapped object  to the lookup peer
				err = w.Send(
					ctx,
					df.ToObject(),
					peer.LookupByOwner(relayPeer),
					WithAsync(),
					WithLocalDiscoveryOnly(),
				)
				if err != nil {
					logger.Error(
						"trying relay peer",
						log.String("relay", relayPeer.String()),
						log.String("recipient", req.recipient.PublicKey().String()),
						log.Error(err),
					)
					continue
				}
				// reset error if we managed to send to at least one relay
				lastErr = nil
				expvar.Get("nm:exc.obj.send.relayed").(metric.Metric).Add(1)
			}
			// todo: wait for ack, how??
		}

		if lastErr == nil {
			logger.Debug("wrote object")
		} else {
			expvar.Get("nm:exc.obj.send.failed").(metric.Metric).Add(1)
		}
		req.err <- lastErr
	}
}

// Subscribe to incoming objects as envelopes
func (w *exchange) Subscribe(filters ...EnvelopeFilter) EnvelopeSubscription {
	return w.inboxes.Subscribe(filters...)
}

// handleObjects -
func (w *exchange) handleObjects(sub EnvelopeSubscription) error {
	for {
		e, err := sub.Next()
		if err != nil {
			return err
		}

		logger := log.
			FromContext(context.Background()).
			Named("exchange").
			With(
				log.String("method", "exchange.handleObjects"),
				log.String("payload", e.Payload.GetType()),
			)

		// TODO verify signature
		logger.Debug("getting payload")
		// nolint: gocritic // don't care about singleCaseSwitch here
		switch e.Payload.GetType() {
		case dataForwardType:
			// if we receive a dataForward payload we check if it is meant for
			// this peer, if yes try to decode it, if not forward it
			// we treat this a data because it needs to be encrypted

			// decode object
			fwd := &DataForward{}
			if err := fwd.FromObject(e.Payload); err != nil {
				return err
			}

			// if the data are encrypted we should first decrupt them
			if !fwd.Ephermeral.IsEmpty() {
				ss, err := crypto.CalculateSharedKey(
					w.keychain.GetPrimaryPeerKey(),
					fwd.Ephermeral,
				)
				if err != nil {
					continue
				}
				fwd.Data, err = decrypt(fwd.Data, ss)
				if err != nil {
					continue
				}
			}

			// unmarshal payload
			m := map[string]interface{}{}
			err := json.Unmarshal(fwd.Data, &m)
			if err != nil {
				return errors.Wrap(errors.Error("could not decode data"), err)
			}

			// convert it into an object
			o := object.FromMap(m)

			// and if this is not a dataforward then we assume it is for us
			if o.GetType() != dataForwardType {
				if len(o.GetSignatures()) == 0 {
					fmt.Println("___", o.GetType())
					logger.Error("forwarded object has no signature")
					continue
				}
				w.inboxes.Publish(&Envelope{
					Sender:  o.GetSignatures()[0].Signer,
					Payload: o,
				})
				continue
			}

			// if this is a dataforward and for some reason we are the
			// recipient, handle it
			nfwd := &DataForward{}
			if err := nfwd.FromObject(o); err != nil {
				return errors.Wrap(errors.Error("could not parse nfwd"), err)
			}

			pk := w.keychain.GetPrimaryPeerKey().PublicKey()
			if nfwd.Recipient.Equals(pk) {
				w.inboxes.Publish(&Envelope{
					Sender:  o.GetSignatures()[0].Signer,
					Payload: o,
				})
				continue
			}

			// send the object to the intended recipient
			// is the original sender lost this way? do we care?
			if err := w.Send(
				context.Background(),
				o,
				peer.LookupByOwner(nfwd.Recipient),
				WithAsync(),
				WithLocalDiscoveryOnly(),
			); err != nil {
				return errors.Wrap(
					errors.Error("could not send object"),
					err,
				)
			}
		}
	}
}

// handlePeers -
func (w *exchange) handlePeers(subscription EnvelopeSubscription) error {
	for {
		e, err := subscription.Next()
		if err != nil {
			return err
		}
		// TODO verify peer
		if e.Payload.GetType() == peerType {
			p := &peer.Peer{}
			if err := p.FromObject(e.Payload); err != nil {
				continue
			}
			w.discover.Add(p, false)
		}
	}
}

// WithLocalDiscoveryOnly will only use local discovery to resolve addresses.
func WithLocalDiscoveryOnly() SendOption {
	return func(opt *SendOptions) {
		opt.LocalDiscovery = true
	}
}

// WithAsync will not wait to actually send the object
func WithAsync() SendOption {
	return func(opt *SendOptions) {
		opt.Async = true
	}
}

// Send an object to peers resulting from a lookup
func (w *exchange) Send(
	ctx context.Context,
	o object.Object,
	recipient peer.LookupOption,
	options ...SendOption,
) error {
	ctx = context.FromContext(ctx)
	opts := &SendOptions{}
	for _, option := range options {
		option(opts)
	}

	lookupOpts := []peer.LookupOption{
		recipient,
	}
	if opts.LocalDiscovery {
		lookupOpts = append(lookupOpts, peer.LookupOnlyLocal())
	}
	ps, err := w.discover.Lookup(ctx, lookupOpts...)
	if err != nil {
		return err
	}

	done := map[crypto.PublicKey]bool{}

	ownPublicKey := w.keychain.GetPrimaryPeerKey().PublicKey()
	errs := &multierror.Group{}
	for p := range ps {
		// check if the peer is done
		if _, ok := done[p.PublicKey()]; ok {
			continue
		}
		// mark peer as done
		done[p.PublicKey()] = true
		// check this is not our own peer
		if p.PublicKey().Equals(ownPublicKey) {
			continue
		}
		p := p
		errs.Go(func() error {
			return w.SendToPeer(
				ctx,
				o,
				p,
				options...,
			)
		})
	}

	return errs.Wait().ErrorOrNil()
}

// SendToPeer an object to an address
func (w *exchange) SendToPeer(
	ctx context.Context,
	o object.Object,
	p *peer.Peer,
	options ...SendOption,
) error {
	ctx = context.FromContext(ctx)
	opts := &SendOptions{}
	for _, option := range options {
		option(opts)
	}

	expvar.Get("nm:exc.obj.send").(metric.Metric).Add(1)

	outbox := w.getOutbox(p.PublicKey())
	errRecv := make(chan error, 1)
	req := &outgoingObject{
		context:   ctx,
		recipient: p,
		object:    o,
		options:   opts,
		err:       errRecv,
	}
	outbox.queue.Append(req)
	if opts.Async {
		return nil
	}
	select {
	case <-ctx.Done():
		return ErrSendingTimedOut
	case err := <-errRecv:
		return err
	}
}

func encrypt(data []byte, key []byte) ([]byte, error) {
	block, _ := aes.NewCipher(key)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func wrapInDataForward(
	o object.Object,
	rks ...crypto.PublicKey,
) (*DataForward, error) {
	if len(rks) == 0 {
		return nil, errors.New("missing recipients")
	}
	// marshal payload
	payload, err := json.Marshal(o.ToMap())
	if err != nil {
		return nil, err
	}
	// create an ephemeral key pair, and calculate the shared key
	ek, ss, err := crypto.NewEphemeralSharedKey(rks[0])
	if err != nil {
		return nil, err
	}
	// encrypt payload
	ep, err := encrypt(payload, ss)
	if err != nil {
		return nil, err
	}
	// create data forward object
	df := &DataForward{
		Recipient:  rks[0],
		Ephermeral: *ek,
		Data:       ep,
	}
	// if there are more than one recipients, wrap for the next
	if len(rks) > 1 {
		return wrapInDataForward(df.ToObject(), rks[1:]...)
	}
	// else return
	return df, nil
}
