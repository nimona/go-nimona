package exchange

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"go.uber.org/zap"

	"nimona.io/internal/errors"
	"nimona.io/internal/log"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/net"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
)

const (
	// ObjectRequestID object attribute
	ObjectRequestID = "_reqID"
)

var (
	// ErrInvalidRequest when received an invalid request object
	ErrInvalidRequest = errors.New("invalid request")
)

// nolint: lll
//go:generate go run github.com/vektra/mockery/cmd/mockery -case underscore -inpkg -name Exchange
//go:generate go run github.com/cheekybits/genny -in=../../../internal/generator/syncmap/syncmap.go -out=syncmap_send_request_generated.go -pkg exchange gen "KeyType=string ValueType=sendRequest"
//go:generate go run github.com/cheekybits/genny -in=../../../internal/generator/syncmap/syncmap.go -out=syncmap_object_request_generated.go -pkg exchange gen "KeyType=string ValueType=ObjectRequest"

type (
	// Exchange interface for mocking exchange
	Exchange interface {
		Request(
			ctx context.Context,
			objectHash string,
			address string,
			options ...Option,
		) error
		Handle(
			contentTypeGlob string,
			handler func(object *Envelope) error,
		) (
			cancelationFunc func(),
			err error,
		)
		Send(
			ctx context.Context,
			object *object.Object,
			address string,
			options ...Option,
		) error
	}
	// echange implements an Exchange
	exchange struct {
		key      *crypto.PrivateKey
		net      net.Network
		manager  *ConnectionManager
		discover discovery.Discoverer
		local    *net.LocalInfo

		outgoing chan *outgoingObject
		incoming chan *incomingObject

		handlers sync.Map
		logger   *zap.Logger

		store        graph.Store
		getRequests  *StringObjectRequestSyncMap
		sendRequests *StringSendRequestSyncMap
	}
	// Options (mostly) for Send()
	Options struct {
		ResponseID     string
		RequestID      string
		Response       chan *Envelope
		LocalDiscovery bool
	}
	Option      func(*Options)
	sendRequest struct {
		out chan *Envelope
	}
	// outgoingObject holds an object that is about to be sent
	outgoingObject struct {
		context   context.Context
		recipient string
		object    *object.Object
		options   *Options
		err       chan error
	}
	// incomingObject holds an object that has just been received
	incomingObject struct {
		conn   *net.Connection
		object *object.Object
	}
	// handler is used for keeping track of handlers and what content
	// types they want to receive
	handler struct {
		contentType glob.Glob
		handler     func(o *Envelope) error
	}
)

// New creates a exchange on a given network
func New(
	key *crypto.PrivateKey,
	n net.Network,
	store graph.Store,
	discover discovery.Discoverer,
	localInfo *net.LocalInfo,
	address string,
) (
	Exchange,
	error,
) {
	ctx := context.Background()

	w := &exchange{
		key:      key,
		net:      n,
		manager:  &ConnectionManager{},
		discover: discover,
		local:    localInfo,

		outgoing: make(chan *outgoingObject, 10),
		incoming: make(chan *incomingObject, 10),

		handlers: sync.Map{},
		logger:   log.Logger(ctx).Named("exchange"),

		store:        store,
		getRequests:  NewStringObjectRequestSyncMap(),
		sendRequests: NewStringSendRequestSyncMap(),
	}

	// TODO(superdecimal) we should probably remove .Listen() from here, net
	// should have a function that accepts a connection handler or something.
	incomingConnections, err := w.net.Listen(ctx, address)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn := <-incomingConnections
			go func(conn *net.Connection) {
				address := "peer:" + conn.RemotePeerKey.Hash
				w.manager.Add(address, conn)
				if err := w.HandleConnection(conn); err != nil {
					w.logger.Warn("failed to handle object", zap.Error(err))
				}
			}(conn)
		}
	}()

	go func() {
		for {
			object := <-w.incoming
			go func(object *incomingObject) {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("Recovered while processing", zap.Any("r", r))
					}
				}()
				if err := w.process(object.object, object.conn); err != nil {
					w.logger.Error("processing object", zap.Error(err))
				}
			}(object)
		}
	}()

	go func() {
		localAddress := w.local.GetPeerInfo().Address()
		for {
			req := <-w.outgoing

			if req.recipient == localAddress {
				req.err <- errors.Error("cannot dial own peer")
			}

			logger := log.Logger(req.context).With(
				zap.String("recipient", req.recipient),
			)

			if req.recipient == "" {
				logger.Info("missing recipient")
				req.err <- errors.New("missing recipient")
				continue
			}

			logger.Debug("trying to send object")

			// try to send the object directly to the recipient
			conn, err := w.getOrDial(req.context, req.recipient, req.options)
			if err != nil {
				logger.Debug("could not get conn to recipient", zap.Error(err))
				req.err <- err
				continue
			}

			if err := net.Write(req.object, conn); err != nil {
				w.manager.Close(req.recipient)
				logger.Debug("could not write to recipient", zap.Error(err))
				req.err <- err
				continue
			}

			// update peer status
			// TODO(geoah) fix status -- not that we are using this
			// w.addressBook.PutPeerStatus(recipientThumbPrint, peer.StatusConnected)
			req.err <- nil
			continue
		}
	}()

	return w, nil
}

// Request an object given its hash from an address
func (w *exchange) Request(
	ctx context.Context,
	hash string,
	address string,
	options ...Option,
) error {
	req := &ObjectRequest{
		ObjectHash: hash,
	}
	return w.Send(ctx, req.ToObject(), address, options...)
}

// Handle allows registering a callback function to handle incoming objects
func (w *exchange) Handle(
	typePatern string,
	h func(o *Envelope) error,
) (
	func(),
	error,
) {
	g, err := glob.Compile(typePatern, '.', '/', '#')
	if err != nil {
		return nil, err
	}
	hID := net.RandStringBytesMaskImprSrc(8)
	w.handlers.Store(hID, &handler{
		contentType: g,
		handler:     h,
	})
	r := func() {
		w.handlers.Delete(hID)
	}
	return r, nil
}

func (w *exchange) HandleConnection(
	conn *net.Connection,
) error {
	w.logger.Debug(
		"handling new connection",
		zap.String("remote", "peer:"+conn.RemotePeerKey.Hash),
	)
	for {
		// TODO use decoder
		obj, err := net.Read(conn)
		if err != nil {
			return err
		}

		w.incoming <- &incomingObject{
			conn:   conn,
			object: obj,
		}
	}
}

// Process incoming object
func (w *exchange) process(
	o *object.Object,
	conn *net.Connection,
) error {
	reqID := ""
	if id, ok := o.GetRaw(ObjectRequestID).(string); ok {
		reqID = id
	}

	logger := w.logger.With(
		zap.String("local_peer", w.key.PublicKey.Hash),
		zap.String("remote_peer", conn.RemotePeerKey.Hash),
		zap.String("request_id", reqID),
		zap.String("object.type", o.GetType()),
	)

	logger.Info("processing object")

	// TODO verify signature
	switch o.GetType() {
	case ObjectRequestType:
		req := &ObjectRequest{}
		if err := req.FromObject(o); err != nil {
			return err
		}
		logger = logger.With(
			zap.String("requested_hash", req.ObjectHash),
			zap.String("recipient", conn.RemotePeerKey.Hash),
		)
		logger.Info("got object request")
		res, err := w.store.Get(req.ObjectHash)
		if err != nil {
			return errors.Wrap(
				errors.Error("could not retrieve object"),
				err,
			)
		}
		if reqID != "" {
			res.SetRaw(ObjectRequestID, reqID)
		}
		cerr := make(chan error, 1)
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		defer cf()
		w.outgoing <- &outgoingObject{
			context:   ctx,
			recipient: "peer:" + conn.RemotePeerKey.Hash,
			object:    res,
			err:       cerr,
		}
		err = <-cerr
		if err != nil {
			logger.Warn("could not send response", zap.Error(err))
		} else {
			logger.Info("sent response")
		}
		return err

	case ObjectForwardRequestType:
		v := &ObjectForwardRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		w.logger.Info("got forwarded message", zap.String("recipient", v.Recipient))
		cerr := make(chan error, 1)
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		defer cf()
		w.outgoing <- &outgoingObject{
			context:   ctx,
			recipient: v.Recipient,
			object:    v.FwObject,
			err:       cerr,
		}
		return <-cerr
	}

	// check if this is in response to a Send() WithResponse()
	if reqID != "" {
		logger.Info("got object with request id, returning to req channel")
		if rs, ok := w.sendRequests.Get(reqID); ok {
			rs.out <- &Envelope{
				Sender:  conn.RemotePeerKey,
				Payload: o,
			}
		}
	}

	ct := o.GetType()
	w.handlers.Range(func(_, v interface{}) bool {
		logger.Debug("publishing object to handler")
		h := v.(*handler)
		if !h.contentType.Match(ct) {
			return true
		}
		go func(h *handler, object interface{}) {
			defer func() {
				if r := recover(); r != nil {
					w.logger.Error("Recovered while handling", zap.Any("r", r))
				}
			}()
			if err := h.handler(&Envelope{
				Sender:  conn.RemotePeerKey,
				Payload: o,
			}); err != nil {
				w.logger.Info(
					"Could not handle event",
					zap.String("contentType", ct),
					zap.Error(err),
				)
			}
		}(h, o)
		return true
	})

	return nil
}

// WithResponse will send any responses to the object being sent on the given
// channel.
// WARNING: the channel MUST NOT be closed before the context has either timed
// out, or been canceled.
func WithResponse(reqID string, out chan *Envelope) Option {
	if reqID == "" {
		reqID = net.RandStringBytesMaskImprSrc(12)
	}
	return func(opt *Options) {
		opt.RequestID = reqID
		opt.Response = out
	}
}

// AsResponse will send the object as a response.
func AsResponse(reqID string) Option {
	return func(opt *Options) {
		opt.ResponseID = reqID
	}
}

// WithLocalDiscoveryOnly will only use local discovery to resolve addresses.
func WithLocalDiscoveryOnly() Option {
	return func(opt *Options) {
		opt.LocalDiscovery = true
	}
}

// Send an object to an address
func (w *exchange) Send(
	ctx context.Context,
	o *object.Object,
	address string,
	options ...Option,
) error {
	logger := log.Logger(ctx)

	opts := &Options{}
	for _, option := range options {
		option(opts)
	}

	if opts.ResponseID != "" {
		o.SetRaw(ObjectRequestID, opts.ResponseID)
	}

	if opts.RequestID != "" {
		o.SetRaw(ObjectRequestID, opts.RequestID)
		sr := &sendRequest{
			out: opts.Response,
		}
		w.sendRequests.Put(opts.RequestID, sr)
		go func() {
			select {
			case <-ctx.Done():
				w.sendRequests.Delete(opts.RequestID)
			}
		}()
	}

	switch getAddressType(address) {
	case "peer":
		err := w.sendDirectlyToPeer(ctx, o, address, opts)
		if err == nil {
			return nil
		}
		logger.Debug("could not send directly to peer", zap.Error(err))

		err = w.sendViaRelayToPeer(ctx, o, address, opts)
		if err == nil {
			return nil
		}
		logger.Debug("could not via relay to peer", zap.Error(err))

		return net.ErrAllAddressesFailed

	default:
		err := w.sendDirectlyToPeer(ctx, o, address, opts)
		if err != nil {
			return errors.New("sending directly to address failed")
		}

	}

	return nil
}

func (w *exchange) sendDirectlyToPeer(
	ctx context.Context,
	o *object.Object,
	address string,
	options *Options,
) error {
	cerr := make(chan error, 1)
	w.outgoing <- &outgoingObject{
		context:   ctx,
		recipient: address,
		object:    o,
		options:   options,
		err:       cerr,
	}
	return <-cerr
}

func (w *exchange) sendViaRelayToPeer(
	ctx context.Context,
	o *object.Object,
	address string,
	options *Options,
) error {
	// if recipient is a peer, we might have some relay addresses
	if !strings.HasPrefix(address, "peer:") {
		return net.ErrAllAddressesFailed
	}

	recipient := strings.Replace(address, "peer:", "", 1)
	if recipient == w.key.PublicKey.Hash {
		// TODO(geoah) error or nil?
		return errors.New("cannot send obj to self")
	}

	q := &peer.PeerInfoRequest{
		SignerKeyHash: recipient,
	}
	peers, err := w.discover.Discover(ctx, q)
	if err != nil {
		return err
	}

	if len(peers) == 0 {
		return net.ErrNoAddresses
	}

	peer := peers[0]

	// TODO(geoah) make sure fw object is signed
	fw := &ObjectForwardRequest{
		Recipient: address,
		FwObject:  o,
	}

	fwo := fw.ToObject()
	if err := crypto.Sign(fwo, w.key); err != nil {
		return err
	}

	for _, address := range peer.Addresses {
		if strings.HasPrefix(address, "relay:") {
			w.logger.
				Debug("found relay address",
					zap.String("peer", recipient),
					zap.String("address", address),
				)
			relayAddress := strings.Replace(address, "relay:", "", 1)
			// TODO this is an ugly hack
			if strings.Contains(address, w.key.PublicKey.Hash) {
				continue
			}
			cerr := make(chan error, 1)
			w.outgoing <- &outgoingObject{
				context:   ctx,
				recipient: relayAddress,
				object:    fwo,
				err:       cerr,
				options:   options,
			}
			if err := <-cerr; err == nil {
				return nil
			}
		}
	}

	return net.ErrAllAddressesFailed
}

func (w *exchange) getOrDial(
	ctx context.Context,
	address string,
	options *Options,
) (
	*net.Connection,
	error,
) {
	w.logger.Debug("looking for existing connection",
		zap.String("address", address),
	)
	if address == "" {
		return nil, errors.New("missing address")
	}

	existingConn, err := w.manager.Get(address)
	if err == nil {
		w.logger.Debug("found existing connection",
			zap.String("address", address),
		)
		return existingConn, nil
	}

	netOpts := []net.Option{}
	if options.LocalDiscovery {
		netOpts = append(netOpts, net.WithLocalDiscoveryOnly())
	}

	conn, err := w.net.Dial(ctx, address, netOpts...)
	if err != nil {
		// w.manager.Close(address)
		return nil, errors.Wrap(err, errors.New("dialing failed"))
	}

	go func() {
		if err := w.HandleConnection(conn); err != nil {
			w.logger.Warn("failed to handle object", zap.Error(err))
		}
	}()

	w.manager.Add(address, conn)

	return conn, nil
}

func getAddressType(addr string) string {
	return strings.Split(addr, ":")[0]
}

func getAddressValue(addr string) (string, error) {
	ps := strings.Split(addr, ":")
	if len(ps) != 2 {
		return "", errors.New("invalid address")
	}
	return ps[1], nil
}
