package stream

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"

	"nimona.io/pkg/context"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/tilde"
)

type (
	syncStrategyTopographical struct {
		network      network.Network
		resolver     resolver.Resolver
		store        objectstore.Store
		newRequestID func() string
	}
)

func NewTopographicalSyncStrategy(
	network network.Network,
	resolver resolver.Resolver,
	store objectstore.Store,
) *syncStrategyTopographical {
	return &syncStrategyTopographical{
		network:  network,
		resolver: resolver,
		store:    store,
		newRequestID: func() string {
			return fmt.Sprintf("%d", time.Now().UnixNano())
		},
	}
}

func (f *syncStrategyTopographical) Serve(
	ctx context.Context,
	manager Manager,
) {
	go f.handleRequests(ctx, manager)
	go f.handleAnnouncements(ctx, manager)
}

func (f *syncStrategyTopographical) handleRequests(
	ctx context.Context,
	manager Manager,
) {
	sub := f.network.Subscribe(
		network.FilterByObjectType(RequestLinearType),
	)
	for {
		env, err := sub.Next()
		if err != nil {
			return
		}

		req := &RequestLinear{}
		err = object.Unmarshal(env.Payload, req)
		if err != nil {
			continue
		}

		if req.RequestID == "" {
			continue
		}

		if req.RootHash.IsEmpty() {
			continue
		}

		respond := func(leaves []tilde.Digest, total int) {
			res := &Response{
				// TODO add metadata
				RequestID: req.RequestID,
				RootHash:  req.RootHash,
				Leaves:    leaves,
				Total:     int64(total),
			}
			obj, err := object.Marshal(res)
			if err != nil {
				return
			}
			err = f.network.Send(
				context.New(),
				obj,
				env.Sender,
			)
			if err != nil {
				return
			}
		}

		ctrl, err := manager.GetController(req.RootHash)
		if err != nil {
			respond(nil, 0)
			continue
		}

		leaves, err := ctrl.GetDigests()
		if err != nil {
			respond(nil, 0)
			continue
		}

		respond(leaves, len(leaves))
	}
}

// TODO: move to manager?
func (f *syncStrategyTopographical) handleAnnouncements(
	ctx context.Context,
	manager Manager,
) {
	sub := f.network.Subscribe(
		network.FilterByObjectType(AnnouncementType),
	)
	for {
		env, err := sub.Next()
		if err != nil {
			return
		}

		announcement := &Announcement{}
		err = object.Unmarshal(env.Payload, announcement)
		if err != nil {
			continue
		}

		ctrl, err := manager.GetOrCreateController(announcement.StreamHash)
		if err != nil {
			continue
		}

		missing := []tilde.Digest{}
		for _, d := range announcement.ObjectHashes {
			if ctrl.ContainsDigest(d) {
				continue
			}
			missing = append(missing, d)
		}

		if len(missing) == 0 {
			continue
		}

		// TODO: improve always having to sync
		sync := true

		// sync := false
		// for _, d := range missing {
		// 	req := &object.Request{
		// 		RequestID:  f.newRequestID(),
		// 		ObjectHash: d,
		// 	}
		// 	res := &object.Response{}
		// 	err := f.network.Send(
		// 		context.New(
		// 			context.WithTimeout(time.Second*2),
		// 		),
		// 		object.MustMarshal(req),
		// 		env.Sender,
		// 		network.SendWithResponse(res, time.Second*2),
		// 	)
		// 	if err != nil {
		// 		continue
		// 	}
		// 	if !res.Found {
		// 		continue
		// 	}
		// 	err = ctrl.Apply(res.Object)
		// 	if err != nil {
		// 		sync = true
		// 		break
		// 	}
		// }
		if sync {
			_, err := f.Fetch(ctx, ctrl, announcement.StreamHash)
			if err != nil {
				continue
			}
		}
	}
}

func (f *syncStrategyTopographical) Fetch(
	ctx context.Context,
	ctrl Controller,
	// TODO: remove streamroot, all controllers now have one
	streamRoot tilde.Digest,
) (int, error) {
	// HACK: strategies need rework to not need the private controller.
	// the reason Fetch expects the interface is because of the strategy
	// integration test.
	controller := ctrl.(*controller)

	// lock the controller
	controller.graph.lock.RLock()

	// check that the controller matches the given root
	if !controller.streamInfo.RootDigest.Equal(streamRoot) {
		// remove the lock and return
		controller.graph.lock.RUnlock()
		return 0, fmt.Errorf("controller's root does not match")
	}

	// keep a copy of the existing digests in the graph
	currentDigests := map[tilde.Digest]struct{}{}
	for d := range controller.graph.nodes {
		currentDigests[d] = struct{}{}
	}

	// remove lock
	controller.graph.lock.RUnlock()

	// find providers for stream
	providers, err := f.resolver.LookupByContent(
		ctx,
		streamRoot,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to lookup providers: %w", err)
	}

	// keep track of the number of objects fetched
	objectsFetched := 0

	// ask providers for first 100 objects of stream
	var errs error
	for _, provider := range providers {
		// keep track of missing objects
		missing := map[tilde.Digest]struct{}{}
		// keep track of progress
		limit := int64(100)
		skip := int64(0)
		// start asking provider about this stream
		for {
			res := &Response{}
			err := f.network.Send(
				ctx,
				object.MustMarshal(&RequestLinear{
					Metadata:  object.Metadata{},
					RequestID: f.newRequestID(),
					RootHash:  streamRoot,
					Limit:     limit,
					Skip:      skip,
				}),
				provider.PublicKey.DID(),
				network.SendWithConnectionInfo(provider),
				network.SendWithResponse(res, time.Second),
			)
			if err != nil || res == nil {
				errs = multierror.Append(errs, err)
				break
			}
			for _, digest := range res.Leaves {
				if _, exists := currentDigests[digest]; exists {
					continue
				}
				if _, exists := missing[digest]; exists {
					continue
				}
				missing[digest] = struct{}{}
			}
			if skip+limit >= res.Total {
				break
			}
			if int64(len(res.Leaves)) >= res.Total {
				break
			}
			skip += limit
		}
		for digest := range missing {
			// get object from provider
			res := &object.Response{}
			err := f.network.Send(
				ctx,
				object.MustMarshal(&object.Request{
					Metadata:   object.Metadata{},
					RequestID:  f.newRequestID(),
					ObjectHash: digest,
				}),
				provider.PublicKey.DID(),
				network.SendWithConnectionInfo(provider),
				network.SendWithResponse(res, time.Second),
			)
			if err != nil {
				errs = multierror.Append(errs, err)
				continue
			}
			// add object to graph
			err = controller.Apply(res.Object)
			if err != nil {
				errs = multierror.Append(errs, err)
				// TODO break? can we recover from this?
				continue
			}
			// add the digest to the list of digests we already have
			currentDigests[digest] = struct{}{}
			// increment the number of objects fetched
			objectsFetched++
		}
	}
	return objectsFetched, errs
}
