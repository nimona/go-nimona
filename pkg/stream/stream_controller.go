package stream

import (
	"fmt"
	"sync"
	"time"

	"github.com/Code-Hex/go-generics-cache/policy/simple"

	"nimona.io/pkg/context"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

var ErrInvalidRoot = fmt.Errorf("root object doesn't match stream's hash")

type (
	controller struct {
		lock sync.RWMutex
		// services
		network     network.Network
		objectStore *sqlobjectstore.Store
		// dag graph
		graph *Graph[tilde.Digest, object.Metadata]
		// state, not thread safe
		streamInfo *Info
		// subscriptions
		subscriptions *simple.Cache[peer.ID, bool]
	}
)

func NewController(
	cid tilde.Digest,
	network network.Network,
	objectStore *sqlobjectstore.Store,
) Controller {
	c := &controller{
		graph:         NewGraph[tilde.Digest, object.Metadata](),
		network:       network,
		objectStore:   objectStore,
		streamInfo:    NewInfo(),
		subscriptions: simple.NewCache[peer.ID, bool](),
	}
	c.streamInfo.RootDigest = cid
	return c
}

// Insert an event to the stream.
// Can either accept an Object, or anything that can be marshaled into one.
// This method will make any necessary changes to the object to make it valid.
// - Will set the object's root if it's not set
// - Will set the object's parents if they are not set
// - Will set the object's sequence if it's not set
// Returns the updated object's hash.
func (s *controller) Insert(v interface{}) (tilde.Digest, error) {
	var o *object.Object
	switch vv := v.(type) {
	case *object.Object:
		o = vv
	default:
		var err error
		o, err = object.Marshal(o)
		if err != nil {
			return tilde.EmptyDigest, fmt.Errorf("failed to marshal object: %w", err)
		}
	}

	// verify that the object has the basic metadata
	if o.Type == "" {
		return tilde.EmptyDigest, fmt.Errorf("object type is required")
	}

	// if the object has no root, set it to the stream root
	if o.Metadata.Root.IsEmpty() && s.streamInfo.RootObject == nil {
		err := s.Apply(o)
		if err != nil {
			return tilde.EmptyDigest, fmt.Errorf("failed to apply object: %w", err)
		}
		return o.Hash(), nil
	}

	// verify or set the object's root
	if o.Metadata.Root.IsEmpty() {
		o.Metadata.Root = s.streamInfo.RootDigest
	} else if !o.Metadata.Root.Equal(s.streamInfo.RootDigest) {
		return tilde.EmptyDigest, fmt.Errorf("roots don't match")
	}

	// verify or set the object's parents
	if len(o.Metadata.Parents) == 0 {
		ps := s.graph.GetLeaves()
		if len(ps) > 0 {
			o.Metadata.Parents = object.Parents{
				"*": ps,
			}
		}
	}

	// gather all nodes until the graph's root
	pns := map[tilde.Digest]struct{}{}
	for _, pn := range o.Metadata.Parents.All() {
		pns[pn] = struct{}{}
		ps := s.graph.nodesToRoot(pn)
		for _, p := range ps {
			pns[p] = struct{}{}
		}
	}

	// verify or set the object's sequence
	if o.Metadata.Sequence == 0 {
		o.Metadata.Sequence = uint64(len(pns))
	}

	// get the object's hash
	h := o.Hash()

	// apply the event
	err := s.Apply(o)
	if err != nil {
		return tilde.EmptyDigest, fmt.Errorf("failed to apply object: %w", err)
	}

	// TODO: figure out how to move announcements to first-time applies

	// announce the event to subscribers
	subscribers := s.subscriptions.Keys()
	if len(subscribers) == 0 {
		return h, nil
	}

	announcement := &Announcement{
		Metadata: object.Metadata{
			Owner: s.network.GetPeerID(),
		},
		StreamHash:   s.streamInfo.RootDigest,
		ObjectHashes: []tilde.Digest{h},
	}

	announcementObject, err := object.Marshal(announcement)
	if err != nil {
		return tilde.EmptyDigest,
			fmt.Errorf("failed to marshal announcement: %w", err)
	}

	for _, sub := range subscribers {
		err := s.network.Send(
			context.New(
				context.WithTimeout(time.Second*2),
			),
			announcementObject,
			sub,
		)
		if err != nil {
			// TODO: log error
			continue
		}
	}

	return h, nil
}

// Apply an event to the stream.
// Can either accept an Object, or anything that can be marshaled into one.
func (s *controller) Apply(v interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var o *object.Object
	switch vv := v.(type) {
	case *object.Object:
		o = vv
	default:
		var err error
		o, err = object.Marshal(o)
		if err != nil {
			return fmt.Errorf("failed to marshal object: %w", err)
		}
	}

	// verify that the object has the basic metadata
	if o.Type == "" {
		return fmt.Errorf("object type is required")
	}

	// verify that the object has not been applied already
	digest := o.Hash()
	_, ok := s.streamInfo.Objects[digest]
	if ok {
		return nil
	}

	// check if we're applying the root object
	if o.Metadata.Root.IsEmpty() && s.streamInfo.RootObject == nil {
		h := o.Hash()
		// check the object's hash against the stream root
		if !h.Equal(s.streamInfo.RootDigest) {
			return ErrInvalidRoot
		}
		// store the object
		err := s.objectStore.Put(o)
		if err != nil {
			return fmt.Errorf("failed to store object: %w", err)
		}
		// update stream info
		s.streamInfo.RootType = o.Type
		s.streamInfo.RootObject = o
		// add the root to the graph
		s.graph.Add(digest, o.Metadata, nil)
		// add the object to the metadata list
		oi := GetObjectInfo(o)
		s.streamInfo.Objects[digest] = oi
		// return
		return nil
	}

	// verify the object's root
	if !o.Metadata.Root.Equal(s.streamInfo.RootDigest) {
		return fmt.Errorf("roots don't match")
	}

	// verify the object's parents
	if len(o.Metadata.Parents) == 0 {
		return fmt.Errorf("object has no parents")
	}

	// verify the object's sequence
	if o.Metadata.Sequence == 0 {
		return fmt.Errorf("object has no sequence")
	}

	// store the object
	err := s.objectStore.Put(o)
	if err != nil {
		return fmt.Errorf("failed to store object: %w", err)
	}

	// add it to the graph
	s.graph.Add(digest, o.Metadata, o.Metadata.Parents.All())

	// add the object to the metadata list
	oi := GetObjectInfo(o)
	s.streamInfo.Objects[oi.Digest] = oi

	// handle special objects
	switch o.Type {
	case SubscriptionType:
		// add the subscription to the subscriptions list
		sub := &Subscription{}
		err := object.Unmarshal(o, sub)
		// in case of error, just move on
		if err != nil {
			break
		}
		if sub.Metadata.Owner.IsEmpty() {
			break
		}
		s.subscriptions.Set(sub.Metadata.Owner, true)
	}

	return nil
}

func (s *controller) GetStreamInfo() Info {
	// TODO lock and copy
	return *s.streamInfo
}

func (s *controller) GetObjectDigests() ([]tilde.Digest, error) {
	// TODO lock
	return s.graph.TopologicalSort()
}

func (s *controller) GetStreamRoot() tilde.Digest {
	return s.streamInfo.RootDigest
}

func (s *controller) GetDigests() ([]tilde.Digest, error) {
	return s.graph.TopologicalSort()
}

func (s *controller) GetReader(ctx context.Context) (object.ReadCloser, error) {
	os := make(chan *object.Object)
	er := make(chan error)
	cl := make(chan struct{})
	or := object.NewReadCloser(ctx, os, er, cl)
	ds, err := s.graph.TopologicalSort()
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(os)
		defer close(er)
		defer close(cl)
		for _, d := range ds {
			o, err := s.objectStore.Get(d)
			if err != nil {
				er <- err
				return
			}
			select {
			case os <- o:
			case <-ctx.Done():
				return
			case <-cl:
				return
			}
		}
	}()
	return or, nil
}

func (s *controller) GetSubscribers() ([]peer.ID, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.subscriptions.Keys(), nil
}

func (s *controller) ContainsDigest(cid tilde.Digest) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.graph.Contains(cid)
}
