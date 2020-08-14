package subscriptionmanager

import (
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/feed"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/subscription"
)

var (
	subscriptionType = subscription.Subscription{}.GetType()
)

type (
	SubscriptionManager interface {
		Subscribe(
			ctx context.Context,
			subjects []crypto.PublicKey,
			types []string,
			streams []object.Hash,
			expiry time.Time,
			recipient *peer.Peer,
		) error
		GetOwnSubscriptions(
			ctx context.Context,
		) ([]subscription.Subscription, error)
		GetSubscriptionsByType(
			ctx context.Context,
			objectType string,
		) ([]subscription.Subscription, error)
	}
	subscriptionmanager struct {
		keychain      keychain.Keychain
		exchange      exchange.Exchange
		objectstore   objectstore.Store
		objectmanager objectmanager.ObjectManager
	}
)

func New(
	ctx context.Context,
	kc keychain.Keychain,
	xc exchange.Exchange,
	os objectstore.Store,
	sm objectmanager.ObjectManager,
) (SubscriptionManager, error) {
	m := &subscriptionmanager{
		keychain:      kc,
		exchange:      xc,
		objectstore:   os,
		objectmanager: sm,
	}

	// make sure that the hypothetical root for our subscription chain exists
	feedRoot := feed.GetFeedHypotheticalRoot(
		kc.GetPrimaryIdentityKey().PublicKey(),
		subscriptionType,
	)
	if err := m.objectstore.Put(feedRoot.ToObject()); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *subscriptionmanager) Subscribe(
	ctx context.Context,
	subjects []crypto.PublicKey,
	types []string,
	streams []object.Hash,
	expiry time.Time,
	recipient *peer.Peer,
) error {
	ctx = context.FromContext(ctx)
	// find our identity public key
	owners := m.keychain.GetPrimaryIdentityKey().PublicKey()
	// create a new subscription
	sub := subscription.Subscription{
		Metadata: object.Metadata{
			Owner: owners,
		},
		Subjects: subjects,
		Types:    types,
		Streams:  streams,
		Expiry:   expiry.Format(time.RFC3339),
	}
	// store the subscription
	// this will add it to the subscriptions feed as well
	if _, err := m.objectmanager.Put(ctx, sub.ToObject()); err != nil {
		return err
	}
	// and send the subscription to the recipient
	if err := m.exchange.Send(
		ctx,
		sub.ToObject(),
		recipient,
	); err != nil {
		return err
	}
	return nil
}

func (m *subscriptionmanager) GetOwnSubscriptions(
	ctx context.Context,
) ([]subscription.Subscription, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx)
	owner := m.keychain.GetPrimaryIdentityKey().PublicKey()
	// get the subscriptions feed
	subscriptionFeedReader, err := m.objectmanager.RequestStream(
		ctx,
		feed.GetFeedHypotheticalRootHash(
			owner,
			subscriptionType,
		),
	)
	if err != nil {
		return nil, err
	}
	// go through the feed and gather subscriptions that are still there
	subscriptionHashes, err := feed.GetFeedHashes(subscriptionFeedReader)
	if err != nil {
		return nil, err
	}
	// load actual subscriptions
	subs := []subscription.Subscription{}
	for _, hash := range subscriptionHashes {
		obj, err := m.objectstore.Get(hash)
		if err != nil {
			logger.Error(
				"unable to get subscription object",
				log.Error(err),
			)
			continue
		}
		// TODO who is responsible for ignoring expired subscriptions?
		sub := subscription.Subscription{}
		if err := sub.FromObject(obj); err != nil {
			logger.Error(
				"unable to parse subscription from object",
				log.Error(err),
			)
			continue
		}
		ours := false
		subOwner := sub.Metadata.Owner
		// TODO either deal with multiple owners, or wait until we kill them
		if subOwner == owner {
			ours = true
		}
		if ours {
			subs = append(subs, sub)
		}
	}
	return subs, nil
}

func (m *subscriptionmanager) GetSubscriptionsByType(
	ctx context.Context,
	objectType string,
) ([]subscription.Subscription, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx)
	owner := m.keychain.GetPrimaryIdentityKey().PublicKey()
	// get the subscriptions feed
	subscriptionFeedReader, err := m.objectmanager.RequestStream(
		ctx,
		feed.GetFeedHypotheticalRootHash(
			owner,
			subscriptionType,
		),
	)
	if err != nil {
		return nil, err
	}
	// go through the feed and gather subscriptions that are still there
	subscriptionHashes, err := feed.GetFeedHashes(subscriptionFeedReader)
	if err != nil {
		return nil, err
	}
	// load actual subscriptions
	subs := []subscription.Subscription{}
	for _, hash := range subscriptionHashes {
		obj, err := m.objectstore.Get(hash)
		if err != nil {
			logger.Error(
				"unable to get subscription object",
				log.Error(err),
			)
			continue
		}
		// TODO who is responsible for ignoring expired subscriptions?
		sub := subscription.Subscription{}
		if err := sub.FromObject(obj); err != nil {
			logger.Error(
				"unable to parse subscription from object",
				log.Error(err),
			)
			continue
		}
		// TODO should this include own subscriptions?
		for _, subType := range sub.Types {
			if subType == objectType {
				subs = append(subs, sub)
			}
		}
	}
	return subs, nil
}
