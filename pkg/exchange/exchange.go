package exchange

import (
	"encoding/json"
	"expvar"
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
	"nimona.io/pkg/log"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

var (
	objectRequestType = new(ObjectRequest).GetType()
	peerType          = new(peer.Peer).GetType()
	dataForwardType   = new(DataForward).GetType()
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
		Request(
			ctx context.Context,
			object object.Hash,
			recipient peer.LookupOption,
			options ...Option,
		) error
		Subscribe(
			filters ...EnvelopeFilter,
		) EnvelopeSubscription
		Send(
			ctx context.Context,
			object object.Object,
			recipient peer.LookupOption,
			options ...Option,
		) error
		SendToPeer(
			ctx context.Context,
			object object.Object,
			p *peer.Peer,
			options ...Option,
		) error
	}
	// echange implements an Exchange
	exchange struct {
		key     crypto.PrivateKey
		net     net.Network
		connmgr connmanager.Manager

		discover discovery.PeerStorer
		local    *peer.LocalPeer

		outboxes *OutboxesMap
		inboxes  EnvelopePubSub

		store     *sqlobjectstore.Store // TODO remove
		blacklist *cache.Cache
	}
	// Options (mostly) for Send()
	Options struct {
		LocalDiscovery bool
		Async          bool
	}
	Option func(*Options)
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
		options   *Options
		err       chan error
	}
)

// New creates a exchange on a given network
func New(
	ctx context.Context,
	key crypto.PrivateKey,
	n net.Network,
	store *sqlobjectstore.Store,
	discover discovery.PeerStorer,
	localInfo *peer.LocalPeer,
	connmgr connmanager.Manager,
) (Exchange, error) {
	w := &exchange{
		key: key,
		net: n,

		discover: discover,
		local:    localInfo,
		connmgr:  connmgr,

		outboxes: NewOutboxesMap(),
		inboxes:  NewEnvelopePubSub(),

		store:     store,
		blacklist: cache.New(10*time.Second, 5*time.Minute),
	}

	// add local peer to discoverer
	// TODO this is mostly a hack as discover doesn't have access to local
	w.discover.Add(w.local.GetSignedPeer(), true)

	logger := log.
		FromContext(ctx).
		Named("exchange").
		With(
			log.String("method", "exchange.New"),
			log.String("local.peer", localInfo.GetPeerPublicKey().String()),
		)

	// subscribe to object requests and handle them
	objectReqSub := w.inboxes.Subscribe(
		FilterByObjectType(dataForwardType),
	)

	// subscribe to data forward type
	dataForwardSub := w.inboxes.Subscribe(
		FilterByObjectType(objectRequestType),
	)

	go func() {
		if err := w.handleObjectRequests(objectReqSub); err != nil {
			logger.Error("handling object requests failed", log.Error(err))
		}
	}()

	go func() {
		if err := w.handleObjectRequests(dataForwardSub); err != nil {
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

	w.connmgr.SetHandler(func(pk crypto.PublicKey, obj object.Object) {
		expvar.Get("nm:exc.obj.received").(metric.Metric).Add(1)

		w.inboxes.Publish(&Envelope{
			Sender:  pk,
			Payload: obj,
		})
	})

	return w, nil
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
			// convert the object from the request to []byte
			// todo: find a way to encrypt it
			payload, err := json.Marshal(req.object.ToMap())
			if err != nil {
				// TODO log and move on?
				lastErr = err
			}

			// Try to send it to one peer
			for _, relayPeer := range req.recipient.Relays {
				logger.Debug(
					"trying relay peer",
					log.String("relay", relayPeer.String()),
					log.String("recipient", req.recipient.PublicKey().String()),
				)
				newReq := DataForward{
					Recipient: req.recipient.PublicKey(),
					Data:      payload,
				}
				ctx := context.New(
					context.WithTimeout(time.Second * 5),
				)
				// send the new wrapped object
				// to the lookup peer
				err := w.Send(
					ctx,
					newReq.ToObject(),
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

// Request an object given its hash from an address
func (w *exchange) Request(
	ctx context.Context,
	hash object.Hash,
	recipient peer.LookupOption,
	options ...Option,
) error {
	req := &ObjectRequest{
		ObjectHash: hash,
	}
	return w.Send(ctx, req.ToObject(), recipient, options...)
}

// Subscribe to incoming objects as envelopes
func (w *exchange) Subscribe(filters ...EnvelopeFilter) EnvelopeSubscription {
	return w.inboxes.Subscribe(filters...)
}

// handleObjectRequests -
func (w *exchange) handleObjectRequests(sub EnvelopeSubscription) error {
	for {
		e, err := sub.Next()
		if err != nil {
			return err
		}

		logger := log.
			FromContext(context.Background()).
			Named("exchange").
			With(
				log.String("method", "exchange.handleObjectRequests"),
				log.String("payload", e.Payload.GetType()),
			)

		// TODO verify signature
		logger.Debug("getting payload")
		switch e.Payload.GetType() {
		case objectRequestType:
			req := &ObjectRequest{}
			if err := req.FromObject(e.Payload); err != nil {
				continue
			}
			res, err := w.store.Get(req.ObjectHash)
			if err != nil {
				continue
			}
			// TODO why the go routine?
			go w.Send( // nolint: errcheck
				context.New(),
				res,
				peer.LookupByOwner(e.Sender),
				WithAsync(),
				WithLocalDiscoveryOnly(),
			)
		case dataForwardType:
			// if we receive a dataForward payload we check if it is meant for
			// this peer, if yes try to decode it, if not forward it
			// we treat this a data because it needs to be encrypted

			// decode object
			fwd := &DataForward{}
			if err := fwd.FromObject(e.Payload); err != nil {
				return err
			}

			m := map[string]interface{}{}
			err := json.Unmarshal(fwd.Data, &m)
			if err != nil {
				return errors.Wrap(errors.Error("could not decode data"), err)
			}

			o := object.FromMap(m)
			// send the object to the intended recipient
			// is the original sender lost this way? do we care?
			if err := w.Send(
				context.Background(),
				o,
				peer.LookupByOwner(fwd.Recipient),
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
func WithLocalDiscoveryOnly() Option {
	return func(opt *Options) {
		opt.LocalDiscovery = true
	}
}

// WithAsync will not wait to actually send the object
func WithAsync() Option {
	return func(opt *Options) {
		opt.Async = true
	}
}

// Send an object to peers resulting from a lookup
func (w *exchange) Send(
	ctx context.Context,
	o object.Object,
	recipient peer.LookupOption,
	options ...Option,
) error {
	ctx = context.FromContext(ctx)
	opts := &Options{}
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

	ownPublicKey := w.local.GetPeerPublicKey()
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
	options ...Option,
) error {
	ctx = context.FromContext(ctx)
	opts := &Options{}
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
