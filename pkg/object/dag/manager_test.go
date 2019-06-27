package dag_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"encoding/json"

	"nimona.io/internal/store/graph"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/dag"
	"nimona.io/pkg/object/exchange"
	"nimona.io/pkg/object/mutation"
	"nimona.io/pkg/object/subscription"
)

//
//    o
//   / \
//  m1  m2
//  |   | \
//  m3  m4 |
//   \ /   |
//    m6   m5
//     \  /
//      s1
//
// o Ffsa2mABctpZ1rTpguU1N65GaDVMnbMHW3sLvJ3cAVri
// m1 EnFp6PUXJd7UckwkMpzFD9iVPGwRnYJEVc5ADtjHF7rj
// m2 7oPoh9GC5wRt8xXUFCBpySYYS6V5pbz9PzxexPYi33et
// m3 FkAHo36tu1zUqiX1eBwUq8AWUcRDgjRcDtKTPz31YiBa
// m4 AU1qfwJEAmxCZgRSVkW9FX4ZQBzXeE8nEgQtMoXKpUmP
// m5 GXLmHnazQGu6obWWgsi1dgPRwjHQ4xxVj6oYpZTvYk4V
// m6 FMSQefm7qnspxSH13zXkCX2nhvC7CJL2e1BfX5SM3t7S
//

var (
	o = object.FromMap(map[string]interface{}{
		"@ctx:s": "foo",
		"foo:s":  "bar",
		"numbers:ai": []int{
			1, 2, 3,
		},
		"strings:as": []string{
			"a", "b", "c",
		},
		"map:o": map[string]interface{}{
			"nested-foo:s": "bar",
			"nested-numbers:ai": []interface{}{
				1, 2, 3,
			},
			"nested-strings:as": []interface{}{
				"a", "b", "c",
			},
		},
	})

	m1 = &mutation.Mutation{
		Parents: []string{
			o.HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAssign,
				Cursor:    []string{"foo:s"},
				Value:     "not-bar",
			},
		},
	}

	m2 = &mutation.Mutation{
		Parents: []string{
			o.HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAppend,
				Cursor:    []string{"numbers:ai"},
				Value:     4,
			},
		},
	}

	m3 = &mutation.Mutation{
		Parents: []string{
			m1.ToObject().HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAppend,
				Cursor:    []string{"strings:as"},
				Value:     "d",
			},
		},
	}

	m4 = &mutation.Mutation{
		Parents: []string{
			m2.ToObject().HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAssign,
				Cursor:    []string{"map:o", "nested-foo:s"},
				Value:     "not-nested-bar",
			},
		},
	}

	m5 = &mutation.Mutation{
		Parents: []string{
			m2.ToObject().HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAppend,
				Cursor:    []string{"map:o", "nested-numbers:ai"},
				Value:     9,
			},
		},
	}

	m6 = &mutation.Mutation{
		Parents: []string{
			m3.ToObject().HashBase58(),
			m4.ToObject().HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAppend,
				Cursor:    []string{"map:o", "nested-strings:as"},
				Value:     "z",
			},
		},
	}

	s1 = &subscription.Subscription{
		Subscriber: "foo",
		Parents: []string{
			m5.ToObject().HashBase58(),
			m6.ToObject().HashBase58(),
		},
	}
)

// func TestRequestMissing(t *testing.T) {
// 	os, err := graph.NewCayleyWithTempStore()
// 	assert.NoError(t, err)

// 	x := &exchange.MockExchange{}

// 	var handler func(*object.Object) error
// 	x.On("Handle", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
// 		handler = args[1].(func(*object.Object) error)
// 	}).Return(nil, nil)

// 	m, err := dag.New(os, x, nil)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, m)
// 	assert.NotNil(t, handler)

// 	rHashes := []string{}
// 	x.On("LookupAndRequest", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
// 		rHashes = append(rHashes, args[1].(string))
// 	}).Return(nil)

// 	// receive m1, request o
// 	rHashes = []string{}
// 	err = handler(m1.ToObject())
// 	assert.NoError(t, err)
// 	assert.Equal(t, sortStrings([]string{
// 		o.ToObject().HashBase58(),
// 	}), sortStrings(rHashes))

// 	// receive m3, request o
// 	rHashes = []string{}
// 	err = handler(m3.ToObject())
// 	assert.NoError(t, err)
// 	assert.Equal(t, sortStrings([]string{
// 		o.ToObject().HashBase58(),
// 	}), sortStrings(rHashes))

// 	// receive m6, request o, m4
// 	rHashes = []string{}
// 	err = handler(m6.ToObject())
// 	assert.NoError(t, err)
// 	assert.Equal(t, sortStrings([]string{
// 		o.ToObject().HashBase58(),
// 		m4.ToObject().HashBase58(),
// 	}), sortStrings(rHashes))

// 	// receive m4, request o, m2
// 	rHashes = []string{}
// 	err = handler(m4.ToObject())
// 	assert.NoError(t, err)
// 	assert.Equal(t, sortStrings([]string{
// 		o.ToObject().HashBase58(),
// 		m2.ToObject().HashBase58(),
// 	}), sortStrings(rHashes))

// 	// receive o, request m2
// 	rHashes = []string{}
// 	err = handler(o.ToObject())
// 	assert.NoError(t, err)
// 	assert.Equal(t, sortStrings([]string{
// 		m2.ToObject().HashBase58(),
// 	}), sortStrings(rHashes))

// 	// receive m5, request m2
// 	rHashes = []string{}
// 	err = handler(m5.ToObject())
// 	assert.NoError(t, err)
// 	assert.Equal(t, sortStrings([]string{
// 		m2.ToObject().HashBase58(),
// 	}), sortStrings(rHashes))

// 	// receive m2, request nothing
// 	rHashes = []string{}
// 	err = handler(m2.ToObject())
// 	assert.NoError(t, err)
// 	assert.Equal(t, sortStrings([]string{}), sortStrings(rHashes))

// 	os.(*graph.Cayley).Dump() // nolint
// }

func TestSync(t *testing.T) {
	os, err := graph.NewCayleyWithTempStore()
	assert.NoError(t, err)

	x := &exchange.MockExchange{}

	var handler func(*exchange.Envelope) error
	x.On("Handle", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		handler = args[1].(func(*exchange.Envelope) error)
	}).Return(nil, nil)

	pk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	li, err := net.NewLocalInfo("", pk)
	assert.NoError(t, err)

	m, err := dag.New(os, x, nil, li)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.NotNil(t, handler)

	rkey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// rHashes := []string{}
	// x.On("LookupAndRequest", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
	// 	rHashes = append(rHashes, args[1].(string))
	// }).Return(nil)

	respWith := func(o *object.Object) func(args mock.Arguments) {
		return func(args mock.Arguments) {
			opt := &exchange.Options{}
			args[3].(exchange.Option)(opt)
			opt.Response <- &exchange.Envelope{
				Payload: o,
				Sender:  rkey.PublicKey,
			}
		}
	}

	// send request
	x.On(
		"Send",
		mock.Anything,
		mock.Anything,
		"peer:"+rkey.PublicKey.Fingerprint().String(),
		mock.Anything,
	).Run(
		respWith(dag.ObjectGraphResponse{
			ObjectHashes: []string{
				o.HashBase58(),
				m1.ToObject().HashBase58(),
				m2.ToObject().HashBase58(),
				m3.ToObject().HashBase58(),
				m4.ToObject().HashBase58(),
				m5.ToObject().HashBase58(),
				m6.ToObject().HashBase58(),
				s1.ToObject().HashBase58(),
			},
		}.ToObject()),
	).Return(nil)

	// request o
	for _, o := range []*object.Object{
		o,
		m1.ToObject(),
		m2.ToObject(),
		m3.ToObject(),
		m4.ToObject(),
		m5.ToObject(),
		m6.ToObject(),
		s1.ToObject(),
	} {
		x.On(
			"Request",
			mock.Anything,
			o.HashBase58(),
			"peer:"+rkey.PublicKey.Fingerprint().String(),
			mock.Anything,
		).Run(
			respWith(o),
		).Return(nil)
	}

	ctx := context.Background()
	res, err := m.Sync(
		ctx,
		[]string{
			o.HashBase58(),
		},
		[]string{
			"peer:" + rkey.PublicKey.Fingerprint().String(),
		},
	)

	assert.Equal(t, jp(o), jp(res.Objects[0]))
	assert.Equal(t, jp(m1.ToObject()), jp(res.Objects[1]))
	assert.Equal(t, jp(m2.ToObject()), jp(res.Objects[2]))
	assert.Equal(t, jp(m3.ToObject()), jp(res.Objects[3]))
	assert.Equal(t, jp(m4.ToObject()), jp(res.Objects[4]))
	assert.Equal(t, jp(m5.ToObject()), jp(res.Objects[5]))
	assert.Equal(t, jp(m6.ToObject()), jp(res.Objects[6]))
	assert.Equal(t, jp(s1.ToObject()), jp(res.Objects[7]))
	assert.NoError(t, err)

	os.(*graph.Cayley).Dump() // nolint
}

// jp is a lazy approach to comparing the mess that is unmarshaling json when
// dealing with numbers
func jp(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ") // nolint
	return string(b)
}
