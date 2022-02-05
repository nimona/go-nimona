package stream

import (
	"fmt"
	"sync"

	"github.com/ghodss/yaml"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
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

// Apply an event to the stream.
// Can either accept an Object, or anything that can be marshalled into one.
func (s *controller) Apply(v interface{}) (tilde.Digest, error) {
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

	// gather the object's parents
	pm := map[tilde.Digest]struct{}{}
	for _, ps := range o.Metadata.Parents {
		for _, p := range ps {
			pm[p] = struct{}{}
		}
	}
	ps := []tilde.Digest{}
	for p := range pm {
		ps = append(ps, p)
	}

	// verify or set the object's sequence
	if o.Metadata.Sequence == 0 {
		// calculate number of total parent nodes in all paths
		tp := len(ps)
		for _, p := range ps {
			tp += s.graph.countToRoot(p)
		}
		o.Metadata.Sequence = uint64(tp)
	}

	// add it to the graph
	s.graph.Add(o.Hash(), o.Metadata, ps)

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
