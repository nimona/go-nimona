package stream

import (
	"fmt"
	"sync"

	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"

	"github.com/ghodss/yaml"
)

type (
	controller struct {
		lock sync.RWMutex
		// services
		network     network.Network
		objectStore *sqlobjectstore.Store
		// dag graph
		graph *Graph[tilde.Digest, object.Metadata]
		// state, not thread safe
		streamInfo *StreamInfo
	}
)

func NewController(
	network network.Network,
	objectStore *sqlobjectstore.Store,
) Controller {
	c := &controller{
		graph:       NewGraph[tilde.Digest, object.Metadata](),
		network:     network,
		objectStore: objectStore,
		streamInfo:  NewStreamInfo(),
	}
	return c
}

// Insert an event to the stream.
// Can either accept an Object, or anything that can be marshalled into one.
// This method will make any necessary changes to the object to make it valid.
// - Will set the object's root if it's not set
// - Will set the object's parents if they are not set
// - Will set the object's sequence if it's not set
// Returns the updated object's hash.
func (s *controller) Insert(v interface{}) (tilde.Digest, error) {
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
			return tilde.EmptyDigest, fmt.Errorf("failed to marshal object: %w", err)
		}
	}

	// verify that the object has the basic metadata
	if o.Type == "" {
		return tilde.EmptyDigest, fmt.Errorf("object type is required")
	}

	// TODO: verify that the object is not already in the graph

	// if this is the first object we're applying and it doesn't have parents,
	// set it as the root
	if s.streamInfo.RootObject == nil && o.Metadata.Root.IsEmpty() {
		h := o.Hash()
		// set the initial stream info
		s.streamInfo.RootType = o.Type
		s.streamInfo.RootObject = o
		s.streamInfo.RootDigest = h
		// add the root to the graph
		s.graph.Add(h, o.Metadata, nil)
		// store the object
		fmt.Println("------ " + o.Hash().String() + ":")
		print(o)
		err := s.objectStore.Put(o)
		if err != nil {
			return tilde.EmptyDigest, fmt.Errorf("failed to store object: %w", err)
		}
		return h, nil
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

	// add it to the graph
	s.graph.Add(o.Hash(), o.Metadata, o.Metadata.Parents.All())

	// store the object
	err := s.objectStore.Put(o)
	if err != nil {
		return tilde.EmptyDigest, fmt.Errorf("failed to store object: %w", err)
	}

	fmt.Println("\n------ " + o.Hash().String() + ":")
	print(o)

	// add the object to the metadata list
	oi := GetObjectInfo(o)
	s.streamInfo.Objects[oi.Digest] = oi

	return o.Hash(), nil
}

// Apply an event to the stream.
// Can either accept an Object, or anything that can be marshalled into one.
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

	// check if this is the first object we're applying
	if s.streamInfo.RootDigest.IsEmpty() && o.Metadata.Root.IsEmpty() {
		// store the object
		err := s.objectStore.Put(o)
		if err != nil {
			return fmt.Errorf("failed to store object: %w", err)
		}
		// update stream info
		s.streamInfo.RootType = o.Type
		s.streamInfo.RootDigest = digest
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

	return nil
}

func (s *controller) GetStreamInfo() StreamInfo {
	// TODO lock and copy
	return *s.streamInfo
}

func (s *controller) GetObjectDigests() ([]tilde.Digest, error) {
	// TODO lock
	return s.graph.TopologicalSort()
}

func (s *controller) GetStreamRoot() tilde.Digest {
	return tilde.EmptyDigest
}

func print(o *object.Object) {
	m, err := o.MarshalMap()
	if err != nil {
		panic(err)
	}
	y, err := yaml.Marshal(m)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(y))
}
