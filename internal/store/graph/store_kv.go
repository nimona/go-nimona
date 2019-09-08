package graph

import (
	"encoding/json"
	"fmt"
	"strings"

	"nimona.io/internal/errors"
	"nimona.io/internal/store/kv"
	"nimona.io/pkg/object"
)

// New constructs a new graph store using a kv backend
func New(store kv.Store) Store {
	return &Graph{
		store: store,
	}
}

// Graph object store
type Graph struct {
	store kv.Store
}

// Put an object in the store
func (s *Graph) Put(v object.Object) error {
	value, err := json.MarshalIndent(v.ToMap(), "", "    ")
	if err != nil {
		return errors.Wrap(err, errors.New("could not marshal object"))
	}

	key := v.HashBase58()
	if err := s.store.Store(key, value); err != nil {
		return errors.Wrap(err, errors.New("could not persist object"))
	}
	if root := v.GetRoot(); root != "" {
		key = root + "." + key
		if err := s.store.Store(key, value); err != nil {
			return errors.Wrap(err, errors.New("could not persist object"))
		}
	}

	return nil
}

// Graph returns all objects in a graph given the hash of of its objects
func (s *Graph) Graph(hash string) ([]object.Object, error) {
	ohs, err := s.store.Scan(hash)
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not find objects"))
	}

	os := make([]object.Object, len(ohs))
	for i, oh := range ohs {
		o, err := s.Get(oh)
		if err != nil {
			return nil, errors.Wrap(err, errors.New("could not get object"))
		}
		os[i] = o
	}

	oos := topographicalSortObjects(os)
	return oos, nil
}

// Get returns an object given its hash
func (s *Graph) Get(hash string) (object.Object, error) {
	value, err := s.store.Get(hash)
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not find object"))
	}

	m := map[string]interface{}{}
	if err := json.Unmarshal(value, &m); err != nil {
		return nil, errors.Wrap(err, errors.New("could not unmarshal object"))
	}

	o := object.New()
	if err := o.FromMap(m); err != nil {
		return nil, errors.Wrap(err, errors.New("could not convert to object"))
	}

	return o, nil
}

// Heads returns all the objects that do not have any parents
func (s *Graph) Heads() ([]object.Object, error) {
	dohs, err := s.store.List()
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not find objects"))
	}

	ohs := []string{}
	for _, doh := range dohs {
		if strings.Contains(doh, ".") {
			continue
		}
		ohs = append(ohs, doh)
	}

	os := make([]object.Object, len(ohs))
	for i, oh := range ohs {
		o, err := s.Get(oh)
		if err != nil {
			return nil, errors.Wrap(err, errors.New("could not get object"))
		}
		os[i] = o
	}

	return os, nil
}

// Dump returns all objects
func (s *Graph) Dump() ([]object.Object, error) {
	ohs, err := s.store.List()
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not find objects"))
	}

	os := make([]object.Object, len(ohs))
	for i, oh := range ohs {
		o, err := s.Get(oh)
		if err != nil {
			return nil, errors.Wrap(err, errors.New("could not get object"))
		}
		os[i] = o
	}

	dot, _ := Dot(os)
	fmt.Println(dot)

	return os, nil
}
