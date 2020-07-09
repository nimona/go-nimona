package subscriptionmanager

import (
	"time"

	"github.com/hashicorp/go-multierror"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/streammanager"
	"nimona.io/pkg/subscription"
)

type (
	SubscriptionManager interface {
		Subscribe(
			ctx context.Context,
			subjects []crypto.PublicKey,
			types []string,
			streams []object.Hash,
			expiry time.Time,
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
		resolver      resolver.Resolver
		exchange      exchange.Exchange
		objectstore   objectstore.Store
		streammanager streammanager.StreamManager
	}
)

func New(
	ctx context.Context,
	kc keychain.Keychain,
	rs resolver.Resolver,
	xc exchange.Exchange,
	os objectstore.Store,
	sm streammanager.StreamManager,
) (SubscriptionManager, error) {
	m := &subscriptionmanager{
		keychain:      kc,
		resolver:      rs,
		exchange:      xc,
		objectstore:   os,
		streammanager: sm,
	}
	// find our identity public key
	owners := kc.ListPublicKeys(keychain.IdentityKey)
	// make sure that the hypothetical root for our subscription chain exists
	chainRoot := m.hypotheticalRoot(owners)
	if err := m.objectstore.Put(chainRoot); err != nil {
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
) error {
	ctx = context.FromContext(ctx)
	// find our identity public key
	owners := m.keychain.ListPublicKeys(keychain.IdentityKey)
	// create a new subscription
	sub := subscription.Subscription{
		Owners:   owners,
		Subjects: subjects,
		Types:    types,
		Streams:  streams,
		Expiry:   expiry.Format(time.RFC3339),
	}
	// store the subscription
	if err := m.objectstore.Put(sub.ToObject()); err != nil {
		return err
	}
	// find peers who we are subscribing to
	lookupOpts := []resolver.LookupOption{}
	for _, subject := range subjects {
		lookupOpts = append(
			lookupOpts,
			resolver.LookupByCertificateSigner(subject),
		)
	}
	peers, err := m.resolver.Lookup(
		ctx,
		lookupOpts...,
	)
	if err != nil {
		return err
	}
	// and send the subscription to them
	totalSent := 0
	var errs error
	for peer := range peers {
		if err := m.exchange.Send(
			ctx,
			sub.ToObject(),
			peer,
		); err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		totalSent++
	}
	if totalSent == 0 {
		return errors.New("did not find any peers to send subscription")
	}
	// finally, get the hypothetical root for our subscription chain
	chainRoot := m.hypotheticalRoot(owners)
	// and add the subscription to the our chain
	chainEvent := subscription.SubscriptionChainSubscriptionAdded{
		Stream:       chainRoot.Hash(),
		Subscription: sub.ToObject().Hash(),
		Owners:       owners,
	}
	if err := m.streammanager.Put(chainEvent.ToObject()); err != nil {
		return err
	}
	return nil
}

func (m *subscriptionmanager) GetOwnSubscriptions(
	ctx context.Context,
) ([]subscription.Subscription, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx)
	// find our identity public key
	owners := m.keychain.ListPublicKeys(keychain.IdentityKey)
	// get the hypothetical root for our subscription chain
	chainRoot := m.hypotheticalRoot(owners)
	// get the whole stream
	stream, err := m.streammanager.Get(ctx, chainRoot.Hash())
	if err != nil {
		return nil, err
	}
	// replay the stream and gather the subscriptions
	hashes := map[object.Hash]struct{}{}
	typeAdded := subscription.SubscriptionChainSubscriptionAdded{}.GetType()
	typeRemoved := subscription.SubscriptionChainSubscriptionRemoved{}.GetType()
	for _, obj := range stream.Objects {
		switch obj.GetType() {
		case typeAdded:
			event := &subscription.SubscriptionChainSubscriptionAdded{}
			if err := event.FromObject(obj); err != nil {
				logger.Error(
					"unable to parse subscription added from object",
					log.Error(err),
				)
				continue
			}
			hashes[event.Subscription] = struct{}{}
		case typeRemoved:
			event := &subscription.SubscriptionChainSubscriptionRemoved{}
			if err := event.FromObject(obj); err != nil {
				logger.Error(
					"unable to parse subscription removed from object",
					log.Error(err),
				)
				continue
			}
			delete(hashes, event.Subscription)
		}
	}
	// load actual subscriptions
	subs := []subscription.Subscription{}
	for hash := range hashes {
		obj, err := m.objectstore.Get(hash)
		if err != nil {
			logger.Error(
				"unable to get subscription object",
				log.Error(err),
			)
			continue
		}
		sub := subscription.Subscription{}
		if err := sub.FromObject(obj); err != nil {
			logger.Error(
				"unable to parse subscription from object",
				log.Error(err),
			)
			continue
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (m *subscriptionmanager) GetSubscriptionsByType(
	ctx context.Context,
	objectType string,
) ([]subscription.Subscription, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx)
	typeSub := subscription.Subscription{}.GetType()
	objs, err := m.objectstore.GetByType(typeSub)
	if err != nil {
		return nil, err
	}
	subs := []subscription.Subscription{}
	for _, obj := range objs {
		sub := subscription.Subscription{}
		if err := sub.FromObject(obj); err != nil {
			logger.Error(
				"unable to parse subscription from object",
				log.Error(err),
			)
			continue
		}
		for _, t := range sub.Types {
			if t == objectType {
				subs = append(subs, sub)
				break
			}
		}
	}
	return subs, nil
}

func (m *subscriptionmanager) hypotheticalRoot(
	owners []crypto.PublicKey,
) object.Object {
	return subscription.SubscriptionChainCreated{
		Owners: owners,
	}.ToObject()
}
