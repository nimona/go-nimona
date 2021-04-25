package feedmanager

import (
	"fmt"
	"strings"
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/feed"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
)

var (
	ErrDone    = errors.Error("done")
	ErrTimeout = errors.Error("request timed out")
)

type (
	FeedManager struct {
		localpeer     localpeer.LocalPeer
		resolver      resolver.Resolver
		objectstore   objectstore.Store
		objectmanager objectmanager.ObjectManager
	}
)

func New(
	ctx context.Context,
	lpr localpeer.LocalPeer,
	res resolver.Resolver,
	str objectstore.Store,
	man objectmanager.ObjectManager,
) error {
	m := &FeedManager{
		localpeer:     lpr,
		resolver:      res,
		objectstore:   str,
		objectmanager: man,
	}

	// if there is no primary identity key set, we should wait for one to be
	// set before initializing the manager
	if m.localpeer.GetIdentityPublicKey().IsEmpty() {
		localPeerUpdates, lpDone := m.localpeer.ListenForUpdates()
		go func() {
			for {
				defer lpDone()
				update := <-localPeerUpdates
				if update == localpeer.EventPrimaryIdentityKeyUpdated {
					// TODO should we panic?
					m.initialize(ctx) // nolint: errcheck
					return
				}
			}
		}()
	} else if err := m.initialize(ctx); err != nil {
		return fmt.Errorf("error initializing feed manager, %w", err)
	}

	return nil
}

func (m *FeedManager) initialize(ctx context.Context) error {
	m.localpeer.PutContentTypes(
		"stream:nimona.io/feed",
		"event:nimona.io/feed.Added",
		"event:nimona.io/feed.Removed",
		"nimona.io/stream.Policy",
		"nimona.io/stream.Request",
		"nimona.io/stream.Response",
		"nimona.io/stream.Announcement",
		"nimona.io/stream.Subscription",
		"stream:nimona.io/schema/relationship",
		"event:nimona.io/schema/relationship.Added",
		"event:nimona.io/schema/relationship.Removed",
	)
	subs := m.objectmanager.Subscribe()
	go func() {
		m.handleObjects(subs) // nolint: errcheck
	}()

	if err := m.createFeedsForRegisteredTypes(ctx); err != nil {
		return err
	}

	localPeerUpdates, _ := m.localpeer.ListenForUpdates()
	go func() {
		for {
			update := <-localPeerUpdates
			if update != localpeer.EventContentTypesUpdated {
				continue
			}
			m.createFeedsForRegisteredTypes(ctx) // nolint: errcheck
		}
	}()

	return nil
}

func (m *FeedManager) createFeedsForRegisteredTypes(ctx context.Context) error {
	registeredTypes := m.localpeer.GetContentTypes()
	for _, registeredType := range registeredTypes {
		switch registeredType {
		case "stream:nimona.io/feed",
			"event:nimona.io/feed.Added",
			"event:nimona.io/feed.Removed",
			"nimona.io/stream.Policy",
			"nimona.io/stream.Request",
			"nimona.io/stream.Response",
			"nimona.io/stream.Announcement",
			"nimona.io/stream.Subscription",
			"nimona.io/Request",
			"nimona.io/Response":
			continue
		}
		if err := m.createFeed(ctx, registeredType); err != nil {
			return fmt.Errorf("error registering feed, %w", err)
		}
	}
	return nil
}

func (m *FeedManager) createFeed(
	ctx context.Context,
	registeredType string,
) error {
	ownPeer := m.localpeer.GetPrimaryPeerKey().PublicKey()

	// create a feed to the given type
	feedRoot := GetFeedRoot(
		m.localpeer.GetIdentityPublicKey(),
		getTypeForFeed(registeredType),
	)
	feedRootObj := feedRoot.ToObject()

	// and store it
	_, err := m.objectmanager.Put(ctx, feedRootObj)
	if err != nil {
		return fmt.Errorf("error trying to store feed root, %w", err)
	}

	// TODO we should also pin the feed

	// add a subscription to the feed's stream
	err = m.objectmanager.AddStreamSubscription(ctx, feedRootObj.CID())
	if err != nil {
		return fmt.Errorf("error trying to subscribe to feed, %w", err)
	}

	// subscribe to stream updates and fetch any objects that have been added
	sub := m.objectmanager.Subscribe(
		objectmanager.FilterByStreamCID(feedRootObj.CID()),
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
			if err := feedAdded.FromObject(obj); err != nil {
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
					resolver.LookupByCID(feedRootObj.CID()),
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
		resolver.LookupByCID(feedRootObj.CID()),
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
			feedRootObj.CID(),
			connInfo,
		)
		if err != nil {
			continue
		}
		r.Close()
	}

	return nil
}

func (m *FeedManager) handleObjects(
	sub objectmanager.ObjectSubscription,
) error {
	identityKey := m.localpeer.GetIdentityPublicKey()
	peerKey := m.localpeer.GetPrimaryPeerKey().PublicKey()
	for {
		obj, err := sub.Read()
		if errors.Is(err, object.ErrReaderDone) {
			return nil
		}
		if err != nil {
			return err
		}

		switch obj.Type {
		case "stream:nimona.io/feed",
			"event:nimona.io/feed.Added",
			"event:nimona.io/feed.Removed",
			"nimona.io/stream.Policy",
			"nimona.io/stream.Request",
			"nimona.io/stream.Response",
			"nimona.io/stream.Announcement",
			"nimona.io/stream.Subscription":
			continue
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

		if !m.isRegisteredContentType(objType) {
			continue
		}

		logger.Debug("handling object")

		// add to feed
		// TODO check if identity key exists, this will not work without one
		feedStreamCID := GetFeedRoot(
			identityKey,
			getTypeForFeed(objType),
		).ToObject().CID()
		feedEvent := feed.Added{
			Metadata: object.Metadata{
				Stream: feedStreamCID,
				Owner:  peerKey,
			},
			ObjectCID: []object.CID{
				objCID,
			},
		}
		if _, err := m.objectmanager.Put(ctx, feedEvent.ToObject()); err != nil {
			// TODO log
			continue
		}
		// _, err = m.objectstore.Get(feedStreamCID)
		// if err != nil { // && !errors.Is(err, objectstore.ErrNotFound) {
		// 	// TODO log
		// 	continue
		// }
		// feedEvent.Metadata.Parents = object.Parents{
		// 	"*": []object.CID{
		// 		feedStreamCID,
		// 	},
		// }
		// if !errors.Is(err, objectstore.ErrNotFound) {
		// 	leaves, err := m.objectstore.GetStreamLeaves(feedStreamCID)
		// 	if err == nil {
		// 		object.SortCIDs(leaves)
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
// func (m *FeedManager) RequestFeed(
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

func (m *FeedManager) isRegisteredContentType(
	contentType string,
) bool {
	contentTypes := m.localpeer.GetContentTypes()
	for _, ct := range contentTypes {
		if contentType == ct {
			return true
		}
	}
	return false
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

func getTypeForFeed(objectType string) string {
	pt := object.ParseType(objectType)
	return strings.TrimLeft(pt.Namespace+"/"+pt.Object, "/")
}
