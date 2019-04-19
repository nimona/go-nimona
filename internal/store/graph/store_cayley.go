package graph

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/kv/bolt" // registers bolt quadstore
	"github.com/cayleygraph/cayley/quad"
	"github.com/cayleygraph/cayley/schema"

	"nimona.io/pkg/object"
)

var (
	predicateDependsOn = quad.IRI("dependsOn")
)

type graphObject struct {
	// rdfType struct{}   `quad:"@type > Object"` // nolint: structcheck
	ID       quad.IRI   `quad:"@id"`
	NodeType string     `quad:"type"`
	Context  string     `quad:"context"`
	Display  string     `quad:"display,optional"`
	Parents  []quad.IRI `quad:"dependsOn < *,optional"`
	Data     string     `quad:"data,optional"`
}

func toGraphObject(v *object.Object) (*graphObject, error) {
	b, err := json.Marshal(v.ToMap())
	if err != nil {
		return nil, err
	}
	nType := "object:root"
	if len(v.GetParents()) > 0 {
		nType = "object"
	}
	o := &graphObject{
		ID:       quad.IRI(v.HashBase58()).Full().Short(),
		NodeType: nType,
		Context:  v.GetType(),
		Parents:  []quad.IRI{},
		Data:     string(b),
	}
	if d, ok := v.GetRaw("@display").(string); ok {
		o.Display = d
	}
	for _, p := range v.GetParents() {
		o.Parents = append(o.Parents, quad.IRI(p).Full().Short())
	}
	return o, nil
}

func fromGraphObject(v *graphObject) (*object.Object, error) {
	o := &object.Object{}
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(v.Data), &m); err != nil {
		return nil, err
	}
	if err := o.FromMap(m); err != nil {
		return nil, err
	}
	if o.HashBase58() != string(v.ID) {
		return nil, fmt.Errorf(
			"expected hash does not match actual, %s, %s ",
			string(v.ID),
			o.HashBase58(),
		)
	}
	return o, nil
}

// NewCayley constructs a new object store given a quad store
func NewCayley(store *cayley.Handle) Store {
	sch := schema.NewConfig()
	return &Cayley{
		store:  store,
		schema: sch,
	}
}

// NewCayleyWithTempStore constructs a new object store with a temp bolt store
// NOTE: Cayley's in-memory store is not thread safe, please don't use it to
// replace this.
func NewCayleyWithTempStore() (Store, error) {
	dir, err := ioutil.TempDir("", "cayley-temp")
	if err != nil {
		return nil, err
	}

	err = graph.InitQuadStore("bolt", dir, nil)
	if err != nil {
		return nil, err
	}

	cs, err := cayley.NewGraph("bolt", dir, nil)
	if err != nil {
		return nil, err
	}

	sch := schema.NewConfig()
	return &Cayley{
		store:  cs,
		schema: sch,
	}, nil
}

// Cayley object store
type Cayley struct {
	store  *cayley.Handle
	schema *schema.Config
}

// Put an object in the store
func (s *Cayley) Put(v *object.Object) error {
	qw := graph.NewWriter(s.store)
	defer func() {
		if err := qw.Flush(); err != nil {
			// TODO(geoah) log error
			fmt.Println("error flushing qw", err)
		}
		if err := qw.Close(); err != nil {
			// TODO(geoah) log error
			fmt.Println("error closing qw", err)
		}
	}()
	g, err := toGraphObject(v)
	if err != nil {
		return err
	}
	if _, err := s.schema.WriteAsQuads(qw, g); err != nil {
		return err
	}
	return nil
}

// Graph returns all objects in a graph given the hash of of its objects
func (s *Cayley) Graph(hash string) ([]*object.Object, error) {
	p := cayley.StartPath(
		s.store,
		quad.IRI(hash),
	).FollowRecursive(
		cayley.
			StartMorphism().
			Or(
				cayley.
					StartMorphism().
					In(predicateDependsOn),
			).
			Or(
				cayley.
					StartMorphism().
					Out(predicateDependsOn),
			),
		0,
		nil,
	)

	gs := []graphObject{}
	err := schema.LoadPathTo(nil, s.store, &gs, p)
	if err != nil {
		return nil, err
	}

	os := []*object.Object{}
	for i := range gs {
		g := gs[i]
		o, gerr := fromGraphObject(&g)
		if gerr != nil {
			return nil, gerr
		}
		os = append(os, o)
	}

	if len(os) == 0 {
		return nil, ErrNotFound
	}

	return os, nil
}

// Get returns an object given its hash
func (s *Cayley) Get(hash string) (*object.Object, error) {
	p := cayley.StartPath(
		s.store,
		quad.IRI(hash),
	)

	g := graphObject{}
	err := schema.LoadPathTo(nil, s.store, &g, p)
	if err != nil {
		return nil, err
	}

	o, gerr := fromGraphObject(&g)
	if gerr != nil {
		return nil, gerr
	}

	return o, err
}

// Children returns the all the children of an object, sorted
func (s *Cayley) Children(hash string) ([]*object.Object, error) {
	p := cayley.StartPath(
		s.store,
		quad.IRI(hash),
	).Or(
		cayley.StartPath(
			s.store,
			quad.IRI(hash),
		).FollowRecursive(
			predicateDependsOn,
			0,
			nil,
		),
	)

	gs := []graphObject{}
	err := schema.LoadPathTo(nil, s.store, &gs, p)
	if err != nil {
		return nil, err
	}

	os := []*object.Object{}
	for i := range gs {
		g := gs[i]
		o, gerr := fromGraphObject(&g)
		if gerr != nil {
			return nil, gerr
		}
		os = append(os, o)
	}

	return os, err
}

// Head returns the root object given another object in the graph.
func (s *Cayley) Head(hash string) (*object.Object, error) {
	// TODO(geoah) Figure out how to use cayley to get the graph's head
	p := cayley.StartPath(
		s.store,
		quad.IRI(hash),
	).FollowRecursive(
		cayley.
			StartMorphism().
			In(predicateDependsOn),
		0,
		nil,
	)

	gs := []graphObject{}
	err := schema.LoadPathTo(nil, s.store, &gs, p)
	if err != nil {
		return nil, err
	}

	if len(gs) == 0 {
		return nil, ErrNotFound
	}

	return fromGraphObject(&gs[len(gs)-1])
}

// Heads returns all the objects that do not have any parents
func (s *Cayley) Heads() ([]*object.Object, error) {
	p := cayley.StartPath(
		s.store,
	).Has(
		quad.IRI("type"),
		quad.String("object:root"),
	)

	gs := []graphObject{}
	err := schema.LoadPathTo(nil, s.store, &gs, p)
	if err != nil {
		return nil, err
	}

	os := []*object.Object{}
	for i := range gs {
		g := gs[i]
		o, gerr := fromGraphObject(&g)
		if gerr != nil {
			return nil, gerr
		}
		os = append(os, o)
	}

	return os, nil
}

// Dump returns all objects
func (s *Cayley) Dump() ([]graphObject, error) {
	p := cayley.StartPath(
		s.store,
	)

	gs := []graphObject{}
	err := schema.LoadPathTo(nil, s.store, &gs, p)
	if err != nil {
		return nil, err
	}

	fmt.Println(dot(gs))

	return gs, nil
}

// Tails returns all the leaf nodes starting from an object
func (s *Cayley) Tails(hash string) ([]*object.Object, error) {
	// TODO(geoah) Figure out how to use cayley to get the graph's leaves
	p := cayley.StartPath(
		s.store,
		quad.IRI(hash),
	).FollowRecursive(
		predicateDependsOn,
		0,
		nil,
	)

	gs := []graphObject{}
	err := schema.LoadPathTo(nil, s.store, &gs, p)
	if err != nil {
		return nil, err
	}

	// figure out tail nodes
	ts := map[string]bool{}
	for i := range gs {
		g := gs[i]
		if _, ok := ts[g.ID.String()]; !ok {
			ts[g.ID.String()] = true
		}
		for _, p := range g.Parents {
			ts[p.String()] = false
		}
	}

	os := []*object.Object{}
	for i := range gs {
		g := gs[i]
		if !ts[g.ID.String()] {
			continue
		}
		o, gerr := fromGraphObject(&g)
		if gerr != nil {
			return nil, gerr
		}
		os = append(os, o)
	}

	return os, nil
}
