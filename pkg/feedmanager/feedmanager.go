package feedmanager

import (
	"fmt"
	"sync"
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/feed"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/value"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
)

type (
	FeedManager interface {
		RegisterFeed(rootType string, eventTypes ...string) error
	}
	feedManager struct {
		localpeer         localpeer.LocalPeer
		resolver          resolver.Resolver
		objectstore       objectstore.Store
		objectmanager     objectmanager.ObjectManager
		contentTypesMutex sync.RWMutex
		contentTypes      map[string][]string // map[rootType][]eventTypes
	}
)

func New(
	ctx context.Context,
	lpr localpeer.LocalPeer,
	res resolver.Resolver,
	str objectstore.Store,
	man objectmanager.ObjectManager,
) (*feedManager, error) {
	m := &feedManager{
		localpeer:     lpr,
		resolver:      res,
		objectstore:   str,
		objectmanager: man,
		contentTypes: map[string][]string{
			"stream:nimona.io/schema/relationship": {
				"event:nimona.io/schema/relationship.Added",
				"event:nimona.io/schema/relationship.Removed",
			},
		},
	}

	// if there is no primary identity key set, we should wait for one to be
	// set before initializing the manager
	if m.localpeer.GetIdentityPublicKey().IsEmpty() {
		localPeerUpdates, lpDone := m.localpeer.ListenForUpdates()
		go func() {
			for {
				defer lpDone()
				update := <-localPeerUpdates
				if update == localpeer.EventIdentityUpdated {
					// TODO should we panic?
					m.initialize(ctx) // nolint: errcheck
					return
				}
			}
		}()
	} else if err := m.initialize(ctx); err != nil {
		return m, fmt.Errorf("error initializing feed manager, %w", err)
	}

	return m, nil
}

func (m *feedManager) initialize(ctx context.Context) error {
	subs := m.objectmanager.Subscribe()
	go func() {
		m.handleObjects(subs) // nolint: errcheck
	}()

	if err := m.createFeedsForRegisteredTypes(ctx); err != nil {
		return err
	}

	return nil
}

func (m *feedManager) createFeedsForRegisteredTypes(ctx context.Context) error {
	if m.localpeer.GetIdentityPublicKey().IsEmpty() {
		return nil
	}

	m.contentTypesMutex.RLock()
	defer m.contentTypesMutex.RUnlock()

	for streamType, eventTypes := range m.contentTypes {
		if err := m.createFeed(ctx, streamType, eventTypes); err != nil {
			return fmt.Errorf("error registering feed, %w", err)
		}
	}
	return nil
}

func (m *feedManager) createFeed(
	ctx context.Context,
	streamType string,
	eventTypes []string,
) error {
	ownPeer := m.localpeer.GetPeerKey().PublicKey()

	// create a feed to the given type
	feedRoot := GetFeedRoot(
		m.localpeer.GetIdentityPublicKey(),
		streamType,
	)
	feedRootObj, err := feedRoot.MarshalObject()
	if err != nil {
		return err
	}
	feedRootCID := feedRootObj.CID()

	// check if it exists
	_, err = m.objectstore.Get(feedRootCID)
	if err == nil {
		return nil
	}

	// and if not, store it
	_, err = m.objectmanager.Put(ctx, feedRootObj)
	if err != nil {
		return fmt.Errorf("error trying to store feed root, %w", err)
	}

	err = m.objectstore.Pin(feedRootCID)
	if err != nil {
		return fmt.Errorf("error trying to pin feed root, %w", err)
	}

	// add a subscription to the feed's stream
	err = m.objectmanager.AddStreamSubscription(ctx, feedRootCID)
	if err != nil {
		return fmt.Errorf("error trying to subscribe to feed, %w", err)
	}

	// subscribe to stream updates and fetch any objects that have been added
	sub := m.objectmanager.Subscribe(
		objectmanager.FilterByStreamCID(feedRootCID),
		objectmanager.FilterByObjectType(new(feed.Added).Type()),
	)

	go func() {
		for {
			obj, err := sub.Read()
			if errors.Is(err, object.ErrReaderDone) {
				return
			}
			if err != nil {
				// TODO log
				return
			}

			feedAdded := &feed.Added{}
			if err := feedAdded.UnmarshalObject(obj); err != nil {
				// TODO log
				continue
			}

			if feedAdded.Metadata.Owner.Equals(ownPeer) {
				continue
			}

			for _, objCID := range feedAdded.ObjectCID {
				// check if we already have this object
				if _, err := m.objectstore.Get(objCID); err == nil {
					continue
				}
				peers, err := m.resolver.Lookup(
					context.New(
						context.WithParent(ctx),
						context.WithTimeout(time.Second),
					),
					// TODO lookup by something better?
					// we were using `resolver.LookupByCID(objCID)` but it was
					// not really working as we don't seem to be publishing
					// stream events.
					// TODO we should at least be caching possible peers
					resolver.LookupByCID(feedRootCID),
				)
				if err != nil {
					// TODO log
					continue
				}
				for _, connInfo := range peers {
					o, err := m.objectmanager.Request(
						context.New(
							context.WithParent(ctx),
							context.WithTimeout(time.Second),
						),
						objCID,
						connInfo,
					)
					if err != nil {
						continue
					}
					if _, err := m.objectmanager.Put(
						ctx,
						o,
					); err != nil {
						continue
					}
					break
				}
			}
		}
	}()

	// find other providers
	peers, err := m.resolver.Lookup(
		context.New(
			context.WithParent(ctx),
			context.WithTimeout(time.Second*5),
		),
		resolver.LookupByCID(feedRootCID),
	)
	if err != nil {
		return fmt.Errorf("error looking for other feed providers, %w", err)
	}

	// and sync with each of them
	for _, connInfo := range peers {
		r, err := m.objectmanager.RequestStream(
			context.New(
				context.WithParent(ctx),
				context.WithTimeout(time.Second*5),
			),
			feedRootCID,
			connInfo,
		)
		if err != nil {
			continue
		}
		r.Close()
	}

	return nil
}

func (m *feedManager) RegisterFeed(
	rootType string,
	eventTypes ...string,
) error {
	m.contentTypesMutex.Lock()
	m.contentTypes[rootType] = eventTypes
	m.contentTypesMutex.Unlock()
	return m.createFeedsForRegisteredTypes(context.New())
}

func (m *feedManager) handleObjects(
	sub objectmanager.ObjectSubscription,
) error {
	identityKey := m.localpeer.GetIdentityPublicKey()
	peerKey := m.localpeer.GetPeerKey().PublicKey()
	for {
		obj, err := sub.Read()
		if errors.Is(err, object.ErrReaderDone) {
			return nil
		}
		if err != nil {
			return err
		}

		ctx := context.Background()
		logger := log.
			FromContext(ctx).
			Named("feedmanager").
			With(
				log.String("method", "feedmanager.handleObjects"),
				log.String("payload.type", obj.Type),
				log.String("payload.cid", obj.CID().String()),
			)

		objType := obj.Type
		objCID := obj.CID()

		streamType, registered := m.isRegisteredContentType(objType)
		if !registered {
			continue
		}

		logger.Debug("handling registered feed type")

		// add to feed
		// TODO check if identity key exists, this will not work without one
		feedStream := GetFeedRoot(
			identityKey,
			streamType,
		)
		feedStreamObj, err := feedStream.MarshalObject()
		if err != nil {
			continue
		}
		feedStreamCID := feedStreamObj.CID()
		feedEvent := feed.Added{
			Metadata: object.Metadata{
				Stream: feedStreamCID,
				Owner:  peerKey,
			},
			ObjectCID: []value.CID{
				objCID,
			},
		}
		feedEventObj, err := feedEvent.MarshalObject()
		if err != nil {
			continue
		}
		if _, err := m.objectmanager.Put(ctx, feedEventObj); err != nil {
			logger.Warn("error storing feed event", log.Error(err))
			continue
		}
		// _, err = m.objectstore.Get(feedStreamCID)
		// if err != nil { // && !errors.Is(err, objectstore.ErrNotFound) {
		// 	// TODO log
		// 	continue
		// }
		// feedEvent.Metadata.Parents = object.Parents{
		// 	"*": []value.CID{
		// 		feedStreamCID,
		// 	},
		// }
		// if !errors.Is(err, objectstore.ErrNotFound) {
		// 	leaves, err := m.objectstore.GetStreamLeaves(feedStreamCID)
		// 	if err == nil {
		// 		cid.SortCIDs(leaves)
		// 		feedEvent.Metadata.Parents["*"] = leaves
		// 	}
		// }

		// if err := m.objectstore.Put(feedEvent.ToObject()); err != nil {
		// 	// TODO log
		// 	continue
		// }
	}
}

// TODO add support for multiple recipients
// TODO this currently needs to be storing objects for it to work.
// func (m *feedManager) RequestFeed(
// 	ctx context.Context,
// 	objectType string,
// 	recipients ...*peer.ConnectionInfo,
// ) (object.ReadCloser, error) {
// 	feedRoot := GetFeedRoot(
// 		m.localpeer.GetIdentityPublicKey().PublicKey(),
// 		getTypeForFeed(objectType),
// 	)

// 	feedRootCID := feedRoot.ToObject().CID()
// 	if len(recipients) == 0 {
// 		return m.objectstore.GetByStream(feedRootCID)
// 	}

// 	// TODO support more than 1 recipient
// 	if len(recipients) > 1 {
// 		panic(errors.Error("currently only a single recipient is supported"))
// 	}

// 	return m.objectstore.GetByStream(feedRootCID)
// }

func (m *feedManager) isRegisteredContentType(
	contentType string,
) (string, bool) {
	m.contentTypesMutex.RLock()
	defer m.contentTypesMutex.RUnlock()

	for streamType, eventTypes := range m.contentTypes {
		if streamType == contentType {
			return streamType, true
		}
		for _, eventType := range eventTypes {
			if eventType == contentType {
				return streamType, true
			}
		}
	}
	return "", false
}

func GetFeedRoot(
	owner crypto.PublicKey,
	feedType string,
) *feed.FeedStreamRoot {
	return &feed.FeedStreamRoot{
		ObjectType: feedType,
		Metadata: object.Metadata{
			Owner: owner,
		},
	}
}
