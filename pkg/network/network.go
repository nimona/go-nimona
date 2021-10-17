package network

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"nimona.io/internal/nat"
	"nimona.io/internal/net"
	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
)

//go:generate mockgen -destination=../networkmock/networkmock_generated.go -package=networkmock -source=network.go

var (
	objHandledCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_received_total",
			Help: "Total number of (top level) objects received",
		},
	)
	objSendSuccessCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_send_success_total",
			Help: "Total number of (top level) objects sent",
		},
	)
	objSendRelayedCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_send_relayed_total",
			Help: "Total number of (top level) objects sent via a relay",
		},
	)
	objSendFailedCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_send_failed_total",
			Help: "Total number of (top level) objects that failed to send",
		},
	)
	objRelayedSuccessCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_relayed_success_total",
			Help: "Total number of objects relayed on behalf of others",
		},
	)
	objRelayedFailedCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_exchange_object_relayed_failed_total",
			Help: "Total number of objects failed to relay on behalf of others",
		},
	)
)

// nolint: lll
//go:generate genny -in=$GENERATORS/syncmap_named/syncmap.go -out=outboxes_generated.go -imp=nimona.io/pkg/crypto -pkg=network gen "KeyType=string ValueType=outbox SyncmapName=outboxes"
//go:generate genny -in=$GENERATORS/pubsub/pubsub.go -out=pubsub_envelopes_generated.go -pkg=network gen "ObjectType=*Envelope Name=Envelope name=envelope"
//go:generate genny -in=$GENERATORS/synclist/synclist.go -out=resolvers_generated.go -imp=nimona.io/pkg/object -pkg=network gen "KeyType=Resolver"
//go:generate genny -in=$GENERATORS/synclist/synclist.go -out=addresses_generated.go -imp=nimona.io/pkg/peer -pkg=network gen "KeyType=string"

type (
	Resolver interface {
		LookupPeer(
			ctx context.Context,
			publicKey crypto.PublicKey,
		) (*peer.ConnectionInfo, error)
	}
	// Network interface for mocking
	Network interface {
		Subscribe(
			filters ...EnvelopeFilter,
		) EnvelopeSubscription
		SubscribeOnce(
			ctx context.Context,
			filters ...EnvelopeFilter,
		) (*Envelope, error)
		Send(
			ctx context.Context,
			object *object.Object,
			publicKey crypto.PublicKey,
			sendOptions ...SendOption,
		) error
		Listen(
			ctx context.Context,
			bindAddress string,
			options ...ListenOption,
		) (net.Listener, error)
		RegisterResolver(
			resolver Resolver,
		)
		GetAddresses() []string
		RegisterAddresses(...string)
		GetRelays() []*peer.ConnectionInfo
		RegisterRelays(...*peer.ConnectionInfo)
		GetPeerKey() crypto.PrivateKey
		GetConnectionInfo() *peer.ConnectionInfo
		Close() error
	}
	// Option for customizing New
	Option func(*network)
	// SendOption for customizing a Send method
	SendOption func(*sendOptions)
	// network implements a Network
	network struct {
		net        net.Network
		peerKey    crypto.PrivateKey
		inboxes    EnvelopePubSub
		deduplist  *cache.Cache
		resolvers  *ResolverSyncList
		addresses  *StringSyncList
		relayLock  sync.RWMutex
		relays     []*peer.ConnectionInfo
		closeFns   []closeFn
		closeMutex sync.Mutex
	}
	// closeFn are functions that will be called during the network's Close
	closeFn func() error
)

// New creates a network on a given network
func New(
	ctx context.Context,
	opts ...Option,
) Network {
	w := &network{
		inboxes:    NewEnvelopePubSub(),
		deduplist:  cache.New(10*time.Second, 1*time.Minute),
		resolvers:  &ResolverSyncList{},
		addresses:  &StringSyncList{},
		relays:     []*peer.ConnectionInfo{},
		closeFns:   []closeFn{},
		closeMutex: sync.Mutex{},
	}
	for _, opt := range opts {
		opt(w)
	}
	if w.peerKey.IsEmpty() {
		k, err := crypto.NewEd25519PrivateKey()
		if err != nil {
			panic(err)
		}
		w.peerKey = k
	}
	w.net = net.New(w.peerKey)

	// subscribe to data forward type
	subs := w.inboxes.Subscribe(
		FilterByObjectType(
			DataForwardRequestType,
			DataForwardEnvelopeType,
		),
	)

	go w.handleObjects(subs)

	w.net.RegisterConnectionHandler(w.handleConnection)

	return w
}

func (w *network) GetPeerKey() crypto.PrivateKey {
	return w.peerKey
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

func (w *network) RegisterResolver(
	resolver Resolver,
) {
	w.resolvers.Put(resolver)
}

func (w *network) lookup(
	ctx context.Context,
	publicKey crypto.PublicKey,
) (*peer.ConnectionInfo, error) {
	resolvers := w.resolvers.List()
	if len(resolvers) == 0 {
		return nil, errors.Error("no resolvers")
	}
	var errs error
	for _, r := range resolvers {
		c, err := r.LookupPeer(ctx, publicKey)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		if c != nil {
			return c, nil
		}
	}
	return nil, errs
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
	logger := log.FromContext(ctx).With(
		log.String("method", "network.Listen"),
		log.String("bindAddress", bindAddress),
		log.Any("listenConfig", listenConfig),
	)
	listener, err := w.net.Listen(
		ctx,
		bindAddress,
		&listenConfig.ListenConfig,
	)
	if err != nil {
		return nil, err
	}

	w.closeMutex.Lock()
	w.closeFns = append(w.closeFns, listener.Close)
	w.closeMutex.Unlock()

	w.RegisterAddresses(listener.Addresses()...)

	// TODO consider if we should be erroring if there are no addresses, but
	// a upnp port was provided in the config options.
	if len(listener.Addresses()) == 0 {
		return listener, nil
	}
	localPort, err := strconv.ParseInt(
		strings.Split(listener.Addresses()[0], ":")[2], 10, 64,
	)
	if err != nil || localPort == 0 {
		return nil, errors.Merge(
			errors.Error("unable to get port from address"),
			err,
		)
	}
	if listenConfig.upnp {
		externalAddress, rm, err := nat.MapExternalPort(
			context.Background(),
			int(localPort),
		)
		if err != nil {
			// TODO return error or simply log it?
			logger.Warn(
				"unable to create port mapping",
				log.Error(err),
			)
			return listener, nil
		}
		w.closeMutex.Lock()
		w.closeFns = append(w.closeFns, func() error {
			rm()
			return nil
		})
		w.closeMutex.Unlock()
		logger.Info(
			"created port mapping",
			log.Strings("internalAddress", listener.Addresses()),
			log.Int("internalPort", int(localPort)),
			log.String("externalAddress", externalAddress),
		)
		w.RegisterAddresses(externalAddress)
	}
	return listener, nil
}

func (w *network) handleConnection(
	conn net.Connection,
) {
	go func() {
		remotePeerKey := conn.RemotePeerKey()
		reader := conn.Read(context.Background())
		for {
			payload, err := reader.Read()
			// TODO split errors into connection or payload
			// ie a payload that cannot be unmarshalled or verified
			// should not kill the connection
			if err != nil {
				// nolint: gocritic
				switch err {
				case net.ErrInvalidSignature:
					log.DefaultLogger.Warn(
						"error reading from connection, non fatal",
						log.String(
							"remote.publicKey",
							remotePeerKey.String(),
						),
						log.Error(err),
					)
					continue
				}
				log.DefaultLogger.Warn(
					"error reading from connection, handler returning",
					log.String(
						"remote.publicKey",
						remotePeerKey.String(),
					),
					log.Error(err),
				)
				return
			}

			log.DefaultLogger.Debug(
				"reading from connection",
				log.String("payload", payload.Type),
			)

			fmt.Println("++", remotePeerKey, payload.Type)

			w.inboxes.Publish(&Envelope{
				Sender:  remotePeerKey,
				Payload: payload,
			})
		}
	}()
}

// func (w *network) processOutbox(outbox *outbox) {
// 	var (
// 		conn     *net.Connection
// 		connInfo *peer.ConnectionInfo
// 		// lastRelay *peer.ConnectionInfo
// 	)
// 	// sendViaRelay:
// 	sendViaRelay := func(
// 		ctx context.Context,
// 		relayConnInfo *peer.ConnectionInfo,
// 		recipient crypto.PublicKey,
// 		obj *object.Object,
// 	) error {
// 		// wrap object
// 		df, err := w.wrapInDataForward(
// 			obj,
// 			recipient,
// 		)
// 		if err != nil {
// 			return err
// 		}
// 		// send the newly wrapped object to the relay
// 		dfo, err := object.Marshal(df)
// 		if err != nil {
// 			return err
// 		}
// 		err = w.Send(
// 			ctx,
// 			dfo,
// 			relayConnInfo.PublicKey,
// 			SendWithConnectionInfo(relayConnInfo),
// 		)
// 		if err != nil {
// 			return err
// 		}
// 		resSub := w.Subscribe(
// 			FilterByObjectType(DataForwardResponseType),
// 			FilterByRequestID(df.RequestID),
// 		)
// 		var resObj *object.Object
// 		c := resSub.Channel()
// 		// TODO arbitrary timeout
// 		t := time.NewTimer(time.Second)
// 		select {
// 		case env := <-c:
// 			resObj = env.Payload
// 		case <-t.C:
// 		case <-ctx.Done():
// 		}
// 		if resObj == nil {
// 			return errors.Error("didn't get a data forward response in time")
// 		}
// 		res := &DataForwardResponse{}
// 		if err := object.Unmarshal(resObj, res); err != nil {
// 			return err
// 		}
// 		if !res.Success {
// 			return fmt.Errorf(
// 				"relay %v wasn't able to delivery object",
// 				relayConnInfo.Addresses,
// 			)
// 		}
// 		return nil
// 	}
// 	// processRequest:
// 	// - if there is already an established connection:
// 	//   - try to write to it
// 	// - if we have been given a connInfo:
// 	//   - try to establish a connection and try to write to it
// 	// - if there is an existing connInfo:
// 	//   - try to establish a connection and try to write to it
// 	// - try to lookup the peer's connInfo via a resolver and try to write to it
// 	// - try to send via relay
// 	processRequest := func(req *outgoingObject) error {
// 		var errs error
// 		// if we have a connection
// 		if conn != nil {
// 			// attempt to send the object
// 			err := net.Write(req.object, conn)
// 			if err == nil {
// 				objSendSuccessCounter.Inc()
// 				return nil
// 			}
// 			// if that fails, close and remove connection
// 			w.connmgr.CloseConnection(
// 				conn,
// 			)
// 			conn = nil
// 		}
// 		// decide which connInfo to use
// 		switch {
// 		case connInfo == nil && req.connectionInfo != nil:
// 			connInfo = req.connectionInfo
// 		case connInfo != nil && req.connectionInfo != nil:
// 			if req.connectionInfo.Version >= connInfo.Version {
// 				connInfo = req.connectionInfo
// 			}
// 		case connInfo == nil && req.connectionInfo == nil:
// 			connInfo = &peer.ConnectionInfo{
// 				PublicKey: req.recipient,
// 			}
// 		}
// 		// try to get establish a connection
// 		newConn, err := w.connmgr.GetConnection(
// 			req.context,
// 			connInfo,
// 		)
// 		if err == nil {
// 			// attempt to send the object
// 			err := net.Write(req.object, newConn)
// 			if err == nil {
// 				// update conn and connInfo and return
// 				conn = newConn
// 				objSendSuccessCounter.Inc()
// 				return nil
// 			}
// 			// if that fails, close and remove connection
// 			errs = multierror.Append(errs, err)
// 			w.connmgr.CloseConnection(
// 				conn,
// 			)
// 		}
// 		// try to lookup the peer's connInfo via a resolver
// 		if req.connectionInfo == nil {
// 			newConnInfo, err := w.lookup(req.context, req.recipient)
// 			if err == nil && newConnInfo != nil {
// 				// use this connInfo from now on
// 				connInfo = newConnInfo
// 				// try to get a connection
// 				newConn, err := w.connmgr.GetConnection(
// 					req.context,
// 					connInfo,
// 				)
// 				if err == nil {
// 					// attempt to send the object
// 					err := net.Write(req.object, newConn)
// 					if err == nil {
// 						// update conn and connInfo and return
// 						conn = newConn
// 						objSendSuccessCounter.Inc()
// 						return nil
// 					}
// 					// if that fails, close and remove connection
// 					errs = multierror.Append(errs, err)
// 					w.connmgr.CloseConnection(
// 						conn,
// 					)
// 				}
// 			}
// 		}
// 		// try to send via relay
// 		if len(connInfo.Relays) == 0 {
// 			return errors.Error("all addresses failed")
// 		}
// 		// TODO use lastRelay first
// 		for _, relay := range connInfo.Relays {
// 			err := sendViaRelay(
// 				req.context,
// 				relay,
// 				req.recipient,
// 				req.object,
// 			)
// 			if err == nil {
// 				objSendRelayedCounter.Inc()
// 				return nil
// 			}
// 			errs = multierror.Append(errs, err)
// 		}
// 		return fmt.Errorf("all addresses failed, all relays failed, %w", errs)
// 	}
// 	for {
// 		// dequeue the next item to send
// 		// TODO figure out what can go wrong here
// 		v := outbox.queue.Pop()
// 		req := v.(*outgoingObject)
// 		// check if the context for this is done
// 		if err := req.context.Err(); err != nil {
// 			req.err <- err
// 			continue
// 		}
// 		// make a logger from our req context
// 		logger := log.FromContext(req.context).With(
// 			log.String("recipient", req.recipient.String()),
// 			log.String("object.type", req.object.Type),
// 		)
// 		if err := processRequest(req); err != nil {
// 			logger.Error("error sending object", log.Error(err))
// 			objSendFailedCounter.Inc()
// 			req.err <- err
// 			continue
// 		}
// 		logger.Info("sent object")
// 		req.err <- nil
// 	}
// }

// Subscribe to incoming objects as envelopes
func (w *network) Subscribe(filters ...EnvelopeFilter) EnvelopeSubscription {
	return w.inboxes.Subscribe(filters...)
}

// SubscribeOnce will wait for the next envelope matching the given filters
// and return it or error if the context is done first.
func (w *network) SubscribeOnce(
	ctx context.Context,
	filters ...EnvelopeFilter,
) (*Envelope, error) {
	s := w.inboxes.Subscribe(filters...)
	select {
	case <-ctx.Done():
		s.Cancel() // TODO verify that Cancel() is non blocking
		return nil, ctx.Err()
	case e := <-s.Channel():
		return e, nil
	}
}

// handleObjects -
func (w *network) handleObjects(sub EnvelopeSubscription) {
	for {
		e, err := sub.Next()
		if err != nil {
			return
		}

		objHandledCounter.Inc()

		logger := log.
			FromContext(context.Background()).
			Named("network").
			With(
				log.String("method", "network.handleObjects"),
				log.String("payload", e.Payload.Type),
			)

		// TODO verify signature
		logger.Debug("handling object")
		// nolint: gocritic // don't care about singleCaseSwitch here
		switch e.Payload.Type {
		case DataForwardRequestType:
			// forward requests are just decoded to get the recipient and their
			// payload is sent to them
			fwd := &DataForwardRequest{}
			if err := object.Unmarshal(e.Payload, fwd); err != nil {
				logger.Warn(
					"error decoding DataForwardRequest",
					log.Error(err),
				)
				continue
			}

			// the way we create the peer is a hack to make sure that we only
			// try to send this to an existing connection and not bothering
			// with dialing the peer.
			err := w.Send(
				context.New(
					context.WithTimeout(time.Second),
				),
				fwd.Payload,
				fwd.Recipient,
			)

			if err != nil {
				objRelayedFailedCounter.Inc()
			} else {
				objRelayedSuccessCounter.Inc()
			}

			df := &DataForwardResponse{
				Metadata: object.Metadata{
					Owner: w.peerKey.PublicKey().DID(),
				},
				RequestID: fwd.RequestID,
				Success:   err == nil,
			}
			dfo, err := object.Marshal(df)
			if err != nil {
				logger.Warn(
					"error marshaling DataForwardResponse",
					log.String("requestID", fwd.RequestID),
					log.Error(err),
				)
				continue
			}
			if resErr := w.Send(
				context.New(
					context.WithTimeout(time.Second),
				),
				dfo,
				e.Sender,
			); resErr != nil {
				logger.Warn(
					"error sending DataForwardResponse",
					log.String("requestID", fwd.RequestID),
					log.Error(err),
				)
				continue
			}

			if err != nil {
				logger.Warn(
					"error sending DataForwardEnvelope",
					log.String("requestID", fwd.RequestID),
					log.Error(err),
				)
				continue
			}

		case DataForwardEnvelopeType:
			// envelopes contain relayed objects, so we decode them and publish
			// them to our inboxes
			fwd := &DataForwardEnvelope{}
			if err := object.Unmarshal(e.Payload, fwd); err != nil {
				logger.Warn(
					"error decoding DataForwardEnvelope",
					log.Error(err),
				)
				continue
			}

			// if the data are encrypted we should first decrypt them
			if !fwd.Sender.IsEmpty() {
				ss, err := crypto.CalculateSharedKey(
					w.peerKey,
					fwd.Sender,
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
			o := &object.Object{}
			err := json.Unmarshal(fwd.Data, o)
			if err != nil {
				logger.Warn(
					"error decoding DataForwardEnvelope's payload",
					log.Error(err),
				)
				continue
			}

			logger.Info(
				"got relayed object",
				log.String("sender", fwd.Sender.String()),
				log.String("relay", e.Sender.String()),
				log.String("payload.type", o.Type),
				log.String("data", string(fwd.Data)),
			)

			w.inboxes.Publish(&Envelope{
				Sender:  fwd.Sender,
				Payload: o,
			})
			continue
		}
	}
}

// Send an object to the given peer.
// Before sending, we'll go through the root object as well as any embedded
func (w *network) Send(
	ctx context.Context,
	o *object.Object,
	p crypto.PublicKey,
	opts ...SendOption,
) error {
	if p.Equals(w.peerKey.PublicKey()) {
		return ErrCannotSendToSelf
	}

	opt := &sendOptions{}
	for _, r := range opts {
		r(opt)
	}

	dedupKey := ctx.CorrelationID() + p.String() + o.Hash().String()
	if _, ok := w.deduplist.Get(dedupKey); ok {
		return ErrAlreadySentDuringContext
	}

	ctx = context.FromContext(ctx)

	var err error
	if k := w.peerKey; !k.IsEmpty() {
		// TODO(geoah) we should be passing the certificates to signAll
		err = object.SignDeep(k, o)
		if err != nil {
			return err
		}
	}

	var rSub EnvelopeSubscription
	if opt.waitForResponse != nil {
		rIDVal, ok := o.Data["requestID"]
		if !ok {
			return errors.Error("cannot wait for response without a request id")
		}
		rID, ok := rIDVal.(tilde.String)
		if !ok {
			return errors.Error("cannot wait for response with an invalid request id")
		}
		if rID == "" {
			return errors.Error("cannot wait for response with empty request id")
		}
		rSub = w.Subscribe(
			FilterByRequestID(string(rID)),
		)
	}

	ci := opt.connectionInfo
	if ci == nil {
		ci = &peer.ConnectionInfo{
			PublicKey: p,
		}
	}

	var c net.Connection
	c, err = w.net.Dial(ctx, ci)
	if err != nil {
		cp, err := w.lookup(ctx, p)
		if err != nil {
			return fmt.Errorf("error dialing peer: %w", err)
		}
		c, err = w.net.Dial(ctx, cp)
		if err != nil {
			return fmt.Errorf("error dialing peer: %w", err)
		}
	}

	err = c.Write(ctx, o)
	if err != nil {
		return fmt.Errorf("error writing object: %w", err)
	}

	w.deduplist.Set(dedupKey, struct{}{}, cache.DefaultExpiration)

	if rSub == nil {
		return nil
	}

	rT := time.NewTimer(opt.waitForResponseTimeout)
	select {
	case <-rT.C:
		// TODO should we return an error if no response came?
		return nil
	case e := <-rSub.Channel():
		if err := object.Unmarshal(e.Payload, opt.waitForResponse); err != nil {
			return errors.Merge(
				ErrUnableToUnmarshalIntoResponse,
				err,
			)
		}
	}
	return nil
}

func encrypt(data []byte, key []byte) ([]byte, error) {
	block, _ := aes.NewCipher(key)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(crand.Reader, nonce); err != nil {
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
	o *object.Object,
	recipient crypto.PublicKey,
) (*DataForwardRequest, error) {
	// marshal payload
	payload, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	// create an ephemeral key pair, and calculate the shared key
	ek, ss, err := crypto.NewSharedKey(
		w.peerKey,
		recipient,
	)
	if err != nil {
		return nil, err
	}
	// encrypt payload
	ep, err := encrypt(payload, ss)
	if err != nil {
		return nil, err
	}
	// create data forward envelope
	dfe := &DataForwardEnvelope{
		Metadata: object.Metadata{
			Owner: ek.PublicKey().DID(),
		},
		Sender: ek.PublicKey(),
		Data:   ep,
	}
	dfeo, err := object.Marshal(dfe)
	if err != nil {
		return nil, err
	}
	err = object.Sign(ek, dfeo)
	if err != nil {
		return nil, err
	}
	// and wrap it in a request
	dfr := &DataForwardRequest{
		Metadata: object.Metadata{
			Owner: ek.PublicKey().DID(),
		},
		RequestID: rand.String(8),
		Recipient: recipient,
		Payload:   dfeo,
	}
	dfro, err := object.Marshal(dfr)
	if err != nil {
		return nil, err
	}
	err = object.Sign(ek, dfro)
	if err != nil {
		return nil, err
	}
	dfr.Metadata.Signature = dfro.Metadata.Signature
	// else return
	return dfr, nil
}

func (w *network) GetAddresses() []string {
	as := w.addresses.List()
	sort.Strings(as)
	return as
}

func (w *network) RegisterAddresses(addresses ...string) {
	for _, h := range addresses {
		w.addresses.Put(h)
	}
}

func (w *network) GetConnectionInfo() *peer.ConnectionInfo {
	return &peer.ConnectionInfo{
		PublicKey: w.peerKey.PublicKey(),
		Addresses: w.GetAddresses(),
		Relays:    w.GetRelays(),
		ObjectFormats: []string{
			"json",
		},
	}
}

func (w *network) GetRelays() []*peer.ConnectionInfo {
	w.relayLock.RLock()
	defer w.relayLock.RUnlock()
	return w.relays
}

func (w *network) RegisterRelays(relays ...*peer.ConnectionInfo) {
	w.relayLock.Lock()
	defer w.relayLock.Unlock()
	w.relays = append(w.relays, relays...)
}

// Close all listeners created in this network as well as remove all nat
// mappings created
func (w *network) Close() error {
	w.closeMutex.Lock()
	defer w.closeMutex.Unlock()
	var errs error
	for _, fn := range w.closeFns {
		err := fn()
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}
