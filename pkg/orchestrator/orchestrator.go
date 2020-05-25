package orchestrator

import (
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/vburenin/nsync"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
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
			recipient *peer.Peer,
		) (*Graph, error)
		Put(object.Object) error
		Get(
			ctx context.Context,
			root object.Hash,
		) (*Graph, error)
	}
	orchestrator struct {
		store    *sqlobjectstore.Store
		exchange exchange.Exchange
		resolver resolver.Resolver
		syncLock *nsync.NamedMutex
		keychain keychain.Keychain
	}
	Graph struct {
		Objects []object.Object
	}
)

// New constructs a new orchestrator given an object store and exchange
func New(
	st *sqlobjectstore.Store,
	ex exchange.Exchange,
	ds resolver.Resolver,
	kc keychain.Keychain,
) (Orchestrator, error) {
	ctx := context.Background()
	return NewWithContext(
		ctx,
		st,
		ex,
		ds,
		kc,
	)
}

// NewWithContext constructs a new orchestrator given an object store and
// exchange
func NewWithContext(
	ctx context.Context,
	st *sqlobjectstore.Store,
	ex exchange.Exchange,
	ds resolver.Resolver,
	kc keychain.Keychain,
) (Orchestrator, error) {
	logger := log.FromContext(ctx).Named("orchestrator")
	m := &orchestrator{
		store:    st,
		exchange: ex,
		resolver: ds,
		syncLock: nsync.NewNamedMutex(),
		keychain: kc,
	}
	sub := m.exchange.Subscribe(
		exchange.FilterByObjectType("**"),
	)
	go func() {
		if err := m.process(ctx, sub); err != nil {
			logger.Error("processing failed", log.Error(err))
		}
	}()

	go func() {
		// Get all the content types that the local peer supports
		// find all the objects and serve only those objects
		os, err := m.store.Filter()
		if err != nil {
			logger.Error("failed to get objects", log.Error(err))
			return
		}

		// hashes we serve
		hs := []object.Hash{}

		for _, o := range os {
			// ignore peers
			if o.GetType() == "nimona.io/peer.Peer" {
				continue
			}
			// if object is part of a stream, ignore children
			if o.GetStream() != "" && o.GetStream() != o.Hash() {
				continue
			}
			hs = append(hs, o.Hash())
		}

		logger.Info(
			"adding supported object hashes as content",
			log.Any("rootObjectHashes", hs),
		)
	}()

	return m, nil
}

// Process an object
func (m *orchestrator) process(
	ctx context.Context,
	sub exchange.EnvelopeSubscription,
) error {
	for {
		e, err := sub.Next()
		if err != nil {
			return err
		}
		ctx := context.FromContext(ctx)
		logger := log.FromContext(ctx).With(
			log.String("method", "orchestrator.Process"),
			log.String("object._hash", e.Payload.Hash().String()),
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
}

// IsComplete checks if a graph is missing any nodes
func IsComplete(cs []object.Object) bool {
	ms := map[string]bool{}
	cm := map[string]object.Object{}
	for _, c := range cs {
		// k: hash v: object
		cm[c.Hash().String()] = c
	}
	for _, c := range cs {
		// get all the parents of an object
		for _, p := range c.GetParents() {
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
// TODO(geoah) what happened if the graph is not complete? Error or sync?
func (m *orchestrator) Put(o object.Object) error {
	// set parents
	streamHash := o.GetStream()
	os := []object.Object{}
	if !streamHash.IsEmpty() {
		os, _ = m.store.Filter(
			sqlobjectstore.FilterByStreamHash(streamHash),
		)
		if len(os) > 0 {
			parents := stream.GetStreamLeaves(os)
			parentHashes := make([]object.Hash, len(parents))
			for i, p := range parents {
				parentHashes[i] = p.Hash()
			}
			o = o.SetParents(parentHashes)
		}
	}
	o = o.SetOwners(
		m.keychain.ListPublicKeys(keychain.IdentityKey),
	)

	// store the object
	if err := m.store.Put(o); err != nil {
		return err
	}

	// send announcements about new hashes
	announcement := &stream.Announcement{
		Stream: streamHash,
		Objects: []*object.Object{
			&o,
		},
		Owners: m.keychain.ListPublicKeys(keychain.IdentityKey),
	}

	sig, err := object.NewSignature(
		m.keychain.GetPrimaryPeerKey(),
		announcement.ToObject(),
	)
	if err != nil {
		return err
	}

	announcement.Signatures = []object.Signature{sig}

	// figure out who to send it to
	recipients := stream.GetAllowsKeysFromPolicies(os...)

	// send announcement to all recipients
	errs := &multierror.Group{}
	for _, recipient := range recipients {
		recipient := recipient
		errs.Go(func() error {
			ps, err := m.resolver.Lookup(
				context.New(
					context.WithTimeout(time.Second*5),
				),
				resolver.LookupByOwner(recipient),
			)
			if err != nil {
				return err
			}
			for p := range ps {
				if err := m.exchange.Send(
					context.New(),
					announcement.ToObject(),
					p,
				); err != nil {
					return err
				}
			}
			return nil
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
		return nil
	}
	if err != nil {
		return err
	}

	// let's add the objects to the store
	// and also gather up their leaves
	leaves := []object.Hash{}
	for _, o := range req.Objects {
		m.store.Put(*o) // nolint: errcheck
		leaves = append(leaves, o.GetParents()...)
	}

	// then let's go through the leaves
	missingObjects := []object.Hash{}
	for _, leafHash := range leaves {
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
	// TODO reconsider this
	if _, err := m.Sync(
		ctx,
		req.Stream,
		&peer.Peer{
			Owners: []crypto.PublicKey{
				sender,
			},
		},
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
		hs = append(hs, o.Hash())
	}

	res := &stream.Response{
		Stream:   req.Stream,
		Nonce:    req.Nonce,
		Children: hs,
		Owners: []crypto.PublicKey{
			m.keychain.GetPrimaryPeerKey().PublicKey(),
		},
	}
	sig, err := object.NewSignature(
		m.keychain.GetPrimaryPeerKey(),
		req.ToObject(),
	)
	if err != nil {
		return err
	}
	res.Signatures = []object.Signature{sig}

	if err := m.exchange.Send(
		ctx,
		res.ToObject(),
		&peer.Peer{
			Owners: []crypto.PublicKey{
				sender,
			},
		},
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
		Owners: []crypto.PublicKey{
			m.keychain.GetPrimaryPeerKey().PublicKey(),
		},
	}
	for i, obj := range vs {
		obj := obj
		res.Objects[i] = &obj
	}
	sig, err := object.NewSignature(
		m.keychain.GetPrimaryPeerKey(),
		req.ToObject(),
	)
	if err != nil {
		return err
	}
	res.Signatures = []object.Signature{sig}

	if err := m.exchange.Send(
		ctx,
		res.ToObject(),
		&peer.Peer{
			Owners: []crypto.PublicKey{
				sender,
			},
		},
	); err != nil {
		logger.Warn(
			"orchestrator.handleStreamObjectRequest could not send object response",
			log.Error(err),
		)
		return err
	}

	return nil
}
