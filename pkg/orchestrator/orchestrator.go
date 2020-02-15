package orchestrator

import (
	"github.com/hashicorp/go-multierror"
	"github.com/vburenin/nsync"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

var (
	streamRequestType        = new(stream.Request).GetType()
	streamResponseType       = new(stream.Response).GetType()
	streamObjectRequestType  = new(stream.ObjectRequest).GetType()
	streamObjectResponseType = new(stream.ObjectResponse).GetType()
	streamAnnouncementType   = new(stream.Announcement).GetType()
)

type (
	// Orchestrator is responsible of keeping streams and their underlying
	// graphs up to date
	Orchestrator interface {
		Sync(
			ctx context.Context,
			stream object.Hash,
			recipients peer.LookupOption,
			options ...exchange.Option,
		) (*Graph, error)
		Put(object.Object) error
		Get(
			ctx context.Context,
			root object.Hash,
		) (*Graph, error)
	}
	orchestrator struct {
		store     *sqlobjectstore.Store
		exchange  exchange.Exchange
		discovery discovery.Discoverer
		syncLock  *nsync.NamedMutex
		localInfo *peer.LocalPeer
	}
	Graph struct {
		Objects []object.Object
	}
)

// New constructs a new orchestrator given an object store and exchange
func New(
	store *sqlobjectstore.Store,
	exchange exchange.Exchange,
	discovery discovery.Discoverer,
	localInfo *peer.LocalPeer,
) (Orchestrator, error) {
	ctx := context.Background()
	return NewWithContext(
		ctx,
		store,
		exchange,
		discovery,
		localInfo,
	)
}

// NewWithContext constructs a new orchestrator given an object store and exchange
func NewWithContext(
	ctx context.Context,
	store *sqlobjectstore.Store,
	exc exchange.Exchange,
	discovery discovery.Discoverer,
	localInfo *peer.LocalPeer,
) (Orchestrator, error) {
	logger := log.FromContext(ctx).Named("orchestrator")
	m := &orchestrator{
		store:     store,
		exchange:  exc,
		discovery: discovery,
		syncLock:  nsync.NewNamedMutex(),
		localInfo: localInfo,
	}
	sub := m.exchange.Subscribe(
		exchange.FilterByObjectType("**"),
	)
	go func() {
		if err := m.process(ctx, sub); err != nil {
			logger.Error("processing failed", log.Error(err))
		}
	}()

	// Get all the content types that the local peer supports
	// find all the objects and serve only those objects
	// TODO which objects do we need to serve?
	contentTypes := m.localInfo.GetContentTypes()

	supportedObjects, err := m.store.Filter(sqlobjectstore.FilterByObjectType(contentTypes...))
	if err != nil {
		logger.Error("failed to get objects", log.Error(err))
	} else {
		// serve all the object hashes
		supportedHashes := make([]object.Hash, len(supportedObjects))

		for i, sobj := range supportedObjects {
			supportedHashes[i] = object.NewHash(sobj)
		}

		logger.Info(
			"adding supported object hashes as content",
			log.Any("rootObjectHashes", supportedHashes),
		)
		m.localInfo.AddContentHash(supportedHashes...)
	}

	return m, nil
}

// Process an object
func (m *orchestrator) process(ctx context.Context, sub exchange.EnvelopeSubscription) error {
	for {
		e, err := sub.Next()
		if err != nil {
			return err
		}
		ctx := context.FromContext(ctx)
		logger := log.FromContext(ctx).With(
			log.String("method", "orchestrator.Process"),
			log.String("object._hash", object.NewHash(e.Payload).String()),
			log.String("object.type", e.Payload.GetType()),
		)
		logger.Debug("handling object")

		o := e.Payload
		switch o.GetType() {
		case streamRequestType:
			v := &stream.Request{}
			if err := v.FromObject(o); err != nil {
				return err
			}
			if err := m.handleStreamRequest(
				ctx,
				e.Sender,
				v,
			); err != nil {
				logger.Warn(
					"could not handle graph request object",
					log.Error(err),
				)
			}
		case streamObjectRequestType:
			v := &stream.ObjectRequest{}
			if err := v.FromObject(o); err != nil {
				return err
			}
			if err := m.handleStreamObjectRequest(
				ctx,
				e.Sender,
				v,
			); err != nil {
				logger.Warn(
					"could not handle graph request object",
					log.Error(err),
				)
			}
		case streamAnnouncementType:
			v := &stream.Announcement{}
			if err := v.FromObject(o); err != nil {
				return err
			}
			if err := m.handleStreamAnnouncement(
				ctx,
				e.Sender,
				v,
			); err != nil {
				logger.Warn(
					"could not handle graph announcement object",
					log.Error(err),
				)
			}

		}
	}

	return nil
}

// IsComplete checks if a graph is missing any nodes
func IsComplete(cs []object.Object) bool {
	ms := map[string]bool{}
	cm := map[string]object.Object{}
	for _, c := range cs {
		// k: hash v: object
		cm[object.NewHash(c).String()] = c
	}
	for _, c := range cs {
		// get all the parents of an object
		for _, p := range stream.GetParents(c) {
			h := p.String()
			// check if that hash exists in the map
			if _, ok := cm[h]; ok {
				continue
			}
			// if missing add the entry to the map
			ms[h] = true
		}
	}
	return len(ms) == 0
}

// Put stores a given object
// TODO(geoah) what happend if the graph is not complete? Error or sync?
func (m *orchestrator) Put(o object.Object) error {
	// set parents
	streamHash := stream.GetStream(o)
	if !streamHash.IsEmpty() {
		os, _ := m.store.Filter(
			sqlobjectstore.FilterByStreamHash(streamHash),
		)
		if len(os) > 0 {
			parents := stream.GetStreamLeaves(os)
			parentHashes := make([]string, len(parents))
			for i, p := range parents {
				parentHashes[i] = object.NewHash(p).String()
			}
			o.Set("@parents:as", parentHashes)
		}
	}
	o.Set("@owners:as", []interface{}{m.localInfo.GetIdentityPublicKey().String()})

	h := object.NewHash(o)

	// store the object
	if err := m.store.Put(o); err != nil {
		return err
	}

	// // get all the objects that are part of the same graph
	// os, err := m.store.Filter(
	// 	sqlobjectstore.FilterByStreamHash(h),
	// )
	// if err != nil {
	// 	return errors.Wrap(
	// 		errors.Error("could not retrieve graph"),
	// 		err,
	// 	)
	// }

	// if !IsComplete(os) {
	// 	return errors.Wrap(
	// 		errors.New("cannot store object"),
	// 		ErrIncompleteGraph,
	// 	)
	// }

	// start publishing new content hashes
	m.localInfo.AddContentHash(h)

	// find leaves
	os, err := m.store.Filter(
		sqlobjectstore.FilterByStreamHash(streamHash),
	)
	if err != nil {
		return err
	}
	leaves := stream.GetStreamLeaves(os)
	leafHashes := make([]object.Hash, len(leaves))
	for i, p := range leaves {
		leafHashes[i] = object.NewHash(p)
	}

	// send announcements about new hashes
	announcement := &stream.Announcement{
		Stream: streamHash,
		Leaves: leafHashes,
		Header: object.Header{
			Owners: []crypto.PublicKey{
				m.localInfo.GetIdentityPublicKey(),
			},
		},
	}

	sig, err := object.NewSignature(
		m.localInfo.GetPeerPrivateKey(),
		announcement.ToObject(),
	)
	if err != nil {
		return err
	}

	announcement.Header.Signature = sig

	// figure out who to send it to
	recipients := stream.GetAllowsKeysFromPolicies(os...)

	// send announcement to all recipients
	errs := &multierror.Group{}
	for _, recipient := range recipients {
		recipient := recipient
		errs.Go(func() error {
			return m.exchange.Send(
				context.New(),
				announcement.ToObject(),
				peer.LookupByOwner(recipient),
				exchange.WithAsync(),
				exchange.WithLocalDiscoveryOnly(),
			)
		})
	}

	return errs.Wait().ErrorOrNil()
}

// Get returns a complete and ordered graph given any node of the graph.
func (m *orchestrator) Get(
	ctx context.Context,
	root object.Hash,
) (*Graph, error) {
	os, err := m.store.Filter(sqlobjectstore.FilterByStreamHash(root))
	if err != nil {
		return nil, errors.Wrap(
			errors.Error("could not retrieve graph"),
			err,
		)
	}

	if !IsComplete(os) {
		return nil, ErrIncompleteGraph
	}

	g := &Graph{
		Objects: os,
	}

	return g, nil
}

func (m *orchestrator) handleStreamAnnouncement(
	ctx context.Context,
	sender crypto.PublicKey,
	req *stream.Announcement,
) error {
	// first let's check if we care about this announcement
	_, err := m.store.Get(req.Stream)
	// if we don't have the root, we probably don't care about this
	if errors.CausedBy(err, sqlobjectstore.ErrNotFound) {
		if len(req.Leaves) == 1 && req.Leaves[0] == req.Stream {
		}
		return nil
	}
	if err != nil {
		return err
	}

	// then let's go through the leaves
	missingObjects := []object.Hash{}
	for _, leafHash := range req.Leaves {
		// see if we already have each of them
		_, err := m.store.Get(leafHash)
		// and if not, request them
		if errors.CausedBy(err, sqlobjectstore.ErrNotFound) {
			missingObjects = append(missingObjects, leafHash)
			continue
		}
		if err != nil {
			return err
		}
		// else we already have it
	}

	if len(missingObjects) == 0 {
		return nil
	}

	// if there are leaves missing, sync
	if _, err := m.Sync(
		ctx,
		req.Stream,
		peer.LookupByOwner(sender),
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithAsync(),
	); err != nil {
		return err
	}

	return nil
}

func (m *orchestrator) handleStreamRequest(
	ctx context.Context,
	sender crypto.PublicKey,
	req *stream.Request,
) error {
	// TODO check if policy allows requested to retrieve the object
	logger := log.FromContext(ctx)

	// get the entire graph for this stream
	vs, err := m.store.Filter(
		sqlobjectstore.FilterByStreamHash(req.Stream),
	)
	if err != nil {
		return err
	}

	// get only the object hashes
	hs := []object.Hash{}
	for _, o := range vs {
		hs = append(hs, object.NewHash(o))
	}

	res := &stream.Response{
		Stream:   req.Stream,
		Nonce:    req.Nonce,
		Children: hs,
		Header: object.Header{
			Owners: []crypto.PublicKey{
				m.localInfo.GetIdentityPublicKey(),
			},
		},
	}
	sig, err := object.NewSignature(
		m.localInfo.GetPeerPrivateKey(),
		req.ToObject(),
	)
	if err != nil {
		return err
	}
	res.Header.Signature = sig

	if err := m.exchange.Send(
		ctx,
		res.ToObject(),
		peer.LookupByOwner(sender),
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithAsync(),
	); err != nil {
		logger.Warn(
			"orchestrator.handleStreamRequest could not send response",
			log.Error(err),
		)
		return err
	}

	return nil
}

func (m *orchestrator) handleStreamObjectRequest(
	ctx context.Context,
	sender crypto.PublicKey,
	req *stream.ObjectRequest,
) error {
	// TODO check if policy allows requested to retrieve the object
	logger := log.FromContext(ctx)

	// create filter to get all requested objects
	filters := []sqlobjectstore.LookupOption{}
	for _, hash := range req.Objects {
		filters = append(filters, sqlobjectstore.FilterByHash(hash))
	}

	// get the objects
	vs, err := m.store.Filter(filters...)
	if err != nil {
		return err
	}

	// construct object response
	res := &stream.ObjectResponse{
		Stream:  req.Stream,
		Nonce:   req.Nonce,
		Objects: make([]*object.Object, len(vs)),
		Header: object.Header{
			Owners: []crypto.PublicKey{
				m.localInfo.GetIdentityPublicKey(),
			},
		},
	}
	for i, obj := range vs {
		obj := obj
		res.Objects[i] = &obj
	}
	sig, err := object.NewSignature(
		m.localInfo.GetPeerPrivateKey(),
		req.ToObject(),
	)
	if err != nil {
		return err
	}
	res.Header.Signature = sig

	if err := m.exchange.Send(
		ctx,
		res.ToObject(),
		peer.LookupByOwner(sender),
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithAsync(),
	); err != nil {
		logger.Warn(
			"orchestrator.handleStreamObjectRequest could not send object response",
			log.Error(err),
		)
		return err
	}

	return nil
}
