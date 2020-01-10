package exchange

import (
	"sync"

	"github.com/geoah/go-queue"
	"github.com/hashicorp/go-multierror"

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
)

const (
	// ErrInvalidRequest when received an invalid request object
	ErrInvalidRequest = errors.Error("invalid request")
	// ErrSendingTimedOut when sending times out
	ErrSendingTimedOut = errors.Error("sending timed out")
	// errOutboxForwarded when an object is forewarded to a different outbox
	// this usually happens when an existing connection already existed
	errOutboxForwarded = errors.Error("request has been moved to another outbox")
)

// nolint: lll
//go:generate $GOBIN/mockery -case underscore -inpkg -name Exchange
//go:generate $GOBIN/genny -in=$GENERATORS/syncmap_named/syncmap.go -out=addresses.go -pkg exchange gen "KeyType=string ValueType=addressState SyncmapName=addresses"
//go:generate $GOBIN/genny -in=$GENERATORS/syncmap_named/syncmap.go -out=outboxes.go -pkg exchange gen "KeyType=crypto.PublicKey ValueType=outbox SyncmapName=outboxes"
//go:generate $GOBIN/genny -in=$GENERATORS/pubsub/pubsub.go -out=pubsub_envelopes.go -pkg exchange gen "ObjectType=*Envelope PubSubName=envelope"

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
		SendToAddress(
			ctx context.Context,
			object object.Object,
			address string,
			options ...Option,
		) error
	}
	// echange implements an Exchange
	exchange struct {
		key crypto.PrivateKey
		net net.Network

		discover discovery.PeerStorer
		local    *peer.LocalPeer

		outboxes *OutboxesMap
		inboxes  EnvelopePubSub

		store *sqlobjectstore.Store // TODO remove
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
		peer      crypto.PublicKey
		addresses *AddressesMap
		conn      *net.Connection
		connLock  sync.RWMutex
		queue     *queue.Queue
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
) (
	Exchange,
	error,
) {
	w := &exchange{
		key: key,
		net: n,

		discover: discover,
		local:    localInfo,

		outboxes: NewOutboxesMap(),
		inboxes:  NewEnvelopePubSub(),

		store: store,
	}

	// TODO(superdecimal) we should probably remove .Listen() from here, net
	// should have a function that accepts a connection handler or something.
	incomingConnections, err := w.net.Listen(ctx)
	if err != nil {
		return nil, err
	}

	logger := log.
		FromContext(ctx).
		Named("exchange").
		With(
			log.String("method", "exchange.New"),
			log.String("local.peer", localInfo.GetPeerPublicKey().String()),
		)

	// subscribe to object requests and handle them
	objectReqSub := w.inboxes.Subscribe(
		FilterByObjectType(objectRequestType),
	)
	go func() {
		if err := w.handleObjectRequests(objectReqSub); err != nil {
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

	// handle new incoming connections
	go func() {
		for {
			conn := <-incomingConnections
			go func(conn *net.Connection) {
				if err := w.handleConnection(conn); err != nil {
					logger.Warn("failed to handle connection", log.Error(err))
				}
			}(conn)
		}
	}()

	return w, nil
}

func (w *exchange) getOutbox(peer crypto.PublicKey) *outbox {
	outbox := &outbox{
		peer:      peer,
		addresses: NewAddressesMap(),
		queue:     queue.New(),
	}
	outbox, loaded := w.outboxes.GetOrPut(peer, outbox)
	if !loaded {
		go w.processOutbox(outbox)
	}
	return outbox
}

func (w *exchange) updateOutboxConn(outbox *outbox, conn *net.Connection) {
	outbox.connLock.Lock()
	if outbox.conn != nil {
		outbox.conn.Close() // nolint: errcheck
	}
	outbox.conn = conn
	outbox.connLock.Unlock()
}

func (w *exchange) updateOutboxConnIfEmpty(
	outbox *outbox,
	conn *net.Connection,
) bool {
	outbox.connLock.Lock()
	if outbox.conn == nil {
		outbox.conn = conn
		return true
	}
	outbox.connLock.Unlock()
	return false
}

func (w *exchange) processOutbox(outbox *outbox) {
	getConnection := func(req *outgoingObject) (*net.Connection, error) {
		outbox.connLock.RLock()
		if outbox.conn != nil {
			outbox.connLock.RUnlock()
			return outbox.conn, nil
		}
		outbox.connLock.RUnlock()
		conn, err := w.net.Dial(req.context, req.recipient)
		if err != nil {
			return nil, err
		}
		// if the remote peer doesn't match the one on our outbox
		// this check is mainly done for when the outbox.peer is actually an
		// address, ie tcps:0.0.0.0:0
		// once we have the real remote peer, we should be replacing the outbox
		if conn.RemotePeerKey.String() != outbox.peer.String() {
			// check if we already have an outbox with the new peer
			existingOutbox, outboxExisted := w.outboxes.GetOrPut(
				conn.RemotePeerKey,
				outbox,
			)
			// and if so
			if outboxExisted {
				// enqueue the object on that outbox
				existingOutbox.queue.Append(req)
				// try to update the connection if its gone
				updated := w.updateOutboxConnIfEmpty(existingOutbox, conn)
				// close the connection if we are not using it
				if !updated {
					conn.Close() // nolint: errcheck
					w.updateOutboxConn(outbox, existingOutbox.conn)
				}
				// and finally return errOutboxForwarded so caller knows to exit
				return nil, errOutboxForwarded
			}
		}
		if err := w.handleConnection(conn); err != nil {
			w.updateOutboxConn(outbox, nil)
			log.DefaultLogger.Warn(
				"failed to handle outbox connection",
				log.Error(err),
			)
			return nil, err
		}
		return conn, nil
	}
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
			log.String("object.@type", req.object.GetType()),
		)
		// validate req
		if req.recipient == nil {
			logger.Info("missing recipient")
			req.err <- errors.New("missing recipient")
			continue
		}
		// try to send the object
		var lastErr error
		for i := 0; i < 3; i++ {
			logger.Debug("trying to get connection", log.Int("attempt", i+1))
			conn, err := getConnection(req)
			if err != nil {
				// the object has been forwarded to another outbox
				if err == errOutboxForwarded {
					return
				}
				lastErr = err
				w.updateOutboxConn(outbox, nil)
				continue
			}
			logger.Debug("trying write object", log.Int("attempt", i+1))
			if err := net.Write(req.object, conn); err != nil {
				lastErr = err
				w.updateOutboxConn(outbox, nil)
				continue
			}
			lastErr = nil
			break
		}
		if lastErr == nil {
			logger.Debug("wrote object")
		}
		// errOutboxForwarded are not considered errors here
		// else, we should report back with something
		if lastErr != errOutboxForwarded {
			req.err <- lastErr
		}
		continue
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

func (w *exchange) handleConnection(
	conn *net.Connection,
) error {
	if conn == nil {
		// TODO should this be nil?
		panic(errors.New("missing connection"))
	}

	// get outbox and update the connection to the peer
	outbox := w.getOutbox(conn.RemotePeerKey)
	w.updateOutboxConn(outbox, conn)

	// TODO(geoah) this looks like a hack
	if err := net.Write(
		w.local.GetSignedPeer().ToObject(),
		conn,
	); err != nil {
		return err
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
				w.updateOutboxConn(outbox, nil)
				return
			}

			w.inboxes.Publish(&Envelope{
				Sender:  conn.RemotePeerKey,
				Payload: payload,
			})
		}
	}()

	return nil
}

// handleObjectRequests -
func (w *exchange) handleObjectRequests(subscription EnvelopeSubscription) error {
	for {
		e, err := subscription.Next()
		if err != nil {
			return err
		}
		// TODO verify signature
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
			go w.Send( // nolint: errcheck
				context.New(),
				res,
				peer.LookupByKey(e.Sender),
				WithLocalDiscoveryOnly(),
			)
		}
	}
	return nil
}

// handlePeers -
func (w *exchange) handlePeers(subscription EnvelopeSubscription) error {
	for {
		e, err := subscription.Next()
		if err != nil {
			return err
		}
		// TODO verify peer
		switch e.Payload.GetType() {
		case peerType:
			p := &peer.Peer{}
			if err := p.FromObject(e.Payload); err != nil {
				continue
			}
			w.discover.Add(p, false)
		}
	}
	return nil
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
	oo object.Object,
	recipient peer.LookupOption,
	options ...Option,
) error {
	ctx = context.FromContext(ctx)
	opts := &Options{}
	for _, option := range options {
		option(opts)
	}

	o := object.Copy(oo) // TODO do we really need to copy?

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

	errs := &multierror.Group{}
	for _, p := range ps {
		p := p
		errs.Go(func() error {
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
		})
	}
	return errs.Wait().ErrorOrNil()
}

// SendToAddress an object to an address
func (w *exchange) SendToAddress(
	ctx context.Context,
	oo object.Object,
	address string,
	options ...Option,
) error {
	ctx = context.FromContext(ctx)
	opts := &Options{}
	for _, option := range options {
		option(opts)
	}

	o := object.Copy(oo) // TODO do we really need to copy?

	outbox := w.getOutbox(crypto.PublicKey(address))
	errRecv := make(chan error, 1)
	req := &outgoingObject{
		context: ctx,
		recipient: &peer.Peer{
			Addresses: []string{address},
		},
		object:  o,
		options: opts,
		err:     errRecv,
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
