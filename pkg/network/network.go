package network

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/geoah/go-queue"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"nimona.io/internal/connmanager"
	"nimona.io/internal/nat"
	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/mockgen -destination=../networkmock/networkmock_generated.go -package=networkmock -source=network.go

var (
	dataForwardType   = new(DataForward).GetType()
	objHandledCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_received_total",
			Help: "Total number of (top level) objects received",
		},
	)
	objSentCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_send_success_total",
			Help: "Total number of (top level) objects sent",
		},
	)
	objRelayedCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_send_relayed_total",
			Help: "Total number of (top level) objects sent via a relay",
		},
	)
	objFailedCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_send_failed_total",
			Help: "Total number of (top level) objects that failed to send",
		},
	)
	objAttemptedCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_send_attempts_total",
			Help: "Total number of (top level) objects attempted to send",
		},
	)
)

const (
	// ErrInvalidRequest when received an invalid request object
	ErrInvalidRequest = errors.Error("invalid request")
	// ErrSendingTimedOut when sending times out
	ErrSendingTimedOut = errors.Error("sending timed out")
	// ErrAlreadySentDuringContext when trying to send to the same peer during
	// this context
	ErrAlreadySentDuringContext = errors.Error("already sent to peer")
)

// nolint: lll
//go:generate $GOBIN/genny -in=$GENERATORS/syncmap_named/syncmap.go -out=outboxes_generated.go -imp=nimona.io/pkg/crypto -pkg=network gen "KeyType=crypto.PublicKey ValueType=outbox SyncmapName=outboxes"
//go:generate $GOBIN/genny -in=$GENERATORS/pubsub/pubsub.go -out=pubsub_envelopes_generated.go -pkg=network gen "ObjectType=*Envelope Name=Envelope name=envelope"

type (
	// Network interface for mocking
	Network interface {
		Subscribe(
			filters ...EnvelopeFilter,
		) EnvelopeSubscription
		Send(
			ctx context.Context,
			object object.Object,
			recipient *peer.Peer,
		) error
		Listen(
			ctx context.Context,
			bindAddress string,
			options ...ListenOption,
		) (net.Listener, error)
		LocalPeer() localpeer.LocalPeer
	}
	// Option for customizing a new Network
	Option func(*network)
	// network implements a Network
	network struct {
		net       net.Network
		connmgr   connmanager.Manager
		localpeer localpeer.LocalPeer
		outboxes  *OutboxesMap
		inboxes   EnvelopePubSub
		deduplist *cache.Cache
	}
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
		err       chan error
	}
)

// New creates a network on a given network
func New(
	ctx context.Context,
	opts ...Option,
) Network {
	w := &network{
		outboxes:  NewOutboxesMap(),
		inboxes:   NewEnvelopePubSub(),
		deduplist: cache.New(10*time.Second, 1*time.Minute),
	}
	for _, opt := range opts {
		opt(w)
	}
	if w.localpeer == nil {
		w.localpeer = localpeer.New()
		k, err := crypto.GenerateEd25519PrivateKey()
		if err != nil {
			panic(err)
		}
		w.localpeer.PutPrimaryPeerKey(k)
	}
	w.net = net.New(w.localpeer)

	logger := log.
		FromContext(ctx).
		Named("network").
		With(
			log.String("method", "network.New"),
		)

	// subscribe to data forward type
	subs := w.inboxes.Subscribe(
		FilterByObjectType(dataForwardType),
	)

	go func() {
		if err := w.handleObjects(subs); err != nil {
			logger.Error("handling object requests failed", log.Error(err))
		}
	}()

	connmgr := connmanager.New(
		ctx,
		w.net,
		w.handleConnection,
	)

	w.connmgr = connmgr
	return w
}

func (w *network) LocalPeer() localpeer.LocalPeer {
	return w.localpeer
}

type (
	ListenOption func(c *listenConfig)
	listenConfig struct {
		net.ListenConfig
		upnp bool
	}
)

func ListenOnLocalIPs(c *listenConfig) {
	c.BindLocal = true
}

func ListenOnPrivateIPs(c *listenConfig) {
	c.BindPrivate = true
}

func ListenOnIPV6(c *listenConfig) {
	c.BindIPV6 = true
}

func ListenOnExternalPort(c *listenConfig) {
	c.upnp = true
}

func (w *network) Listen(
	ctx context.Context,
	bindAddress string,
	options ...ListenOption,
) (net.Listener, error) {
	listenConfig := &listenConfig{}
	for _, o := range options {
		o(listenConfig)
	}
	listener, err := w.net.Listen(
		ctx,
		bindAddress,
		&listenConfig.ListenConfig,
	)
	if err != nil {
		return nil, err
	}
	// TODO consider if we should be erroring if there are no addresses, but
	// a upnp port was provided in the config options.
	if len(listener.Addresses()) == 0 {
		return listener, nil
	}
	localPort, err := strconv.ParseInt(
		strings.Split(listener.Addresses()[0], ":")[2], 10, 64,
	)
	if err != nil || localPort == 0 {
		return nil, errors.Wrap(
			errors.New("unable to get port from address"),
			err,
		)
	}
	externalAddress, _, err := nat.MapExternalPort(int(localPort))
	if err != nil {
		// TODO return error or simply log it?
		return listener, nil
	}
	w.localpeer.PutAddresses(externalAddress)
	return listener, nil
}

func (w *network) handleConnection(conn *net.Connection) error {
	objHandledCounter.Inc()

	if conn == nil {
		return errors.New("missing connection")
	}

	go func() {
		defer func() {
			conn.Close() // nolint: errcheck
			if r := recover(); r != nil {
				log.DefaultLogger.Error(
					"recovered from panic, closed conn",
					log.Any("r", r),
					log.Stack(),
				)
			}
		}()
		for {
			payload, err := net.Read(conn)
			// TODO split errors into connection or payload
			// ie a payload that cannot be unmarshalled or verified
			// should not kill the connection
			if err != nil {
				if err == net.ErrInvalidSignature {
					log.DefaultLogger.Warn(
						"error reading from connection",
						log.Error(err),
						log.String("hash", payload.Hash().String()),
					)
					continue
				}
				log.DefaultLogger.Warn(
					"error reading from connection, closing connection",
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

func (w *network) getOutbox(recipient crypto.PublicKey) *outbox {
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

func (w *network) processOutbox(outbox *outbox) {
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
			objSentCounter.Inc()
			break
		}

		// Only try the relays if we fail to write the object
		if lastErr != nil && req.object.GetType() != "nimona.io/LookupRequest" {
			for _, relayPeer := range req.recipient.Relays {
				df, err := w.wrapInDataForward(
					req.object,
					req.recipient.PublicKey(),
					relayPeer.PublicKey(),
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
					log.String("relay", relayPeer.PublicKey().String()),
					log.String("recipient", req.recipient.PublicKey().String()),
				)
				ctx := context.New(
					context.WithTimeout(time.Second * 5),
				)
				// send the newly wrapped object  to the lookup peer
				err = w.Send(
					ctx,
					df.ToObject(),
					relayPeer,
				)
				if err != nil {
					logger.Error(
						"trying relay peer",
						log.String("relay", relayPeer.PublicKey().String()),
						log.String("recipient", req.recipient.PublicKey().String()),
						log.Error(err),
					)
					continue
				}
				// reset error if we managed to send to at least one relay
				lastErr = nil
				objRelayedCounter.Inc()
			}
			// todo: wait for ack, how??
		}

		if lastErr == nil {
			logger.Debug("wrote object")
		} else {
			objFailedCounter.Inc()
		}
		req.err <- lastErr
	}
}

// Subscribe to incoming objects as envelopes
func (w *network) Subscribe(filters ...EnvelopeFilter) EnvelopeSubscription {
	return w.inboxes.Subscribe(filters...)
}

// handleObjects -
func (w *network) handleObjects(sub EnvelopeSubscription) error {
	for {
		e, err := sub.Next()
		if err != nil {
			return err
		}

		logger := log.
			FromContext(context.Background()).
			Named("network").
			With(
				log.String("method", "network.handleObjects"),
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
					w.localpeer.GetPrimaryPeerKey(),
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
				if o.GetSignature().IsEmpty() {
					logger.Error("forwarded object has no signature")
					continue
				}
				w.inboxes.Publish(&Envelope{
					Sender:  o.GetSignature().Signer,
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

			pk := w.localpeer.GetPrimaryPeerKey().PublicKey()
			if nfwd.Recipient.Equals(pk) {
				w.inboxes.Publish(&Envelope{
					Sender:  o.GetSignature().Signer,
					Payload: o,
				})
				continue
			}

			// send the object to the intended recipient.
			// the way we create the peer is a hack to make sure that we only
			// try to send this to an existing connection and not bothering
			// with dialing the peer.
			if err := w.Send(
				context.Background(),
				o,
				&peer.Peer{
					Metadata: object.Metadata{
						Owner: nfwd.Recipient,
					},
				},
			); err != nil {
				return errors.Wrap(
					errors.Error("could not send object"),
					err,
				)
			}
		}
	}
}

// Send an object to the given peer.
// Before sending, we'll go through the root object as well as any embedded
func (w *network) Send(
	ctx context.Context,
	o object.Object,
	p *peer.Peer,
) error {
	dedupKey := ctx.CorrelationID() + p.PublicKey().String() + o.Hash().String()
	if _, ok := w.deduplist.Get(dedupKey); ok {
		return ErrAlreadySentDuringContext
	}

	ctx = context.FromContext(ctx)

	var err error
	if k := w.localpeer.GetPrimaryPeerKey(); !k.IsEmpty() {
		o, err = signAll(k, o)
		if err != nil {
			return err
		}
	}

	if k := w.localpeer.GetPrimaryIdentityKey(); !k.IsEmpty() {
		o, err = signAll(k, o)
		if err != nil {
			return err
		}
	}

	objAttemptedCounter.Inc()

	outbox := w.getOutbox(p.PublicKey())
	errRecv := make(chan error, 1)
	req := &outgoingObject{
		context:   ctx,
		recipient: p,
		object:    o,
		err:       errRecv,
	}
	outbox.queue.Append(req)
	select {
	case <-ctx.Done():
		return ErrSendingTimedOut
	case err := <-errRecv:
		w.deduplist.Set(dedupKey, struct{}{}, cache.DefaultExpiration)
		return err
	}
}

func signAll(k crypto.PrivateKey, o object.Object) (object.Object, error) {
	os := map[string]object.Object{}
	object.Traverse(o.Raw(), func(path string, v object.Value) bool {
		if !v.IsMap() {
			return true
		}
		m := v.(object.Map)
		t := m.Value("type:s")
		if t == nil || !t.IsString() || t.(object.String) == "" {
			return true
		}
		os[path] = object.Object(m)
		return true
	})
	for path, obj := range os {
		if obj.GetOwner() != k.PublicKey() {
			continue
		}
		sig, err := object.NewSignature(k, obj)
		if err != nil {
			return object.Object{}, err
		}

		signedObj := obj.SetSignature(sig)
		if path == "" {
			o = signedObj
		} else {
			o = object.Object(
				o.Raw().Set(path, signedObj.Raw()),
			)
		}
	}
	return o, nil
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

func (w *network) wrapInDataForward(
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
		Metadata: object.Metadata{
			Owner: ek.PublicKey(),
		},
		Recipient:  rks[0],
		Ephermeral: ek.PublicKey(), // TODO to we need this attribute?
		Data:       ep,
	}
	sig, err := object.NewSignature(*ek, df.ToObject())
	if err != nil {
		return nil, err
	}
	df.Metadata.Signature = sig
	// if there are more than one recipients, wrap for the next
	if len(rks) > 1 {
		return w.wrapInDataForward(df.ToObject(), rks[1:]...)
	}
	// else return
	return df, nil
}
