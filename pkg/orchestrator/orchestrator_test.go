package orchestrator_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/store/graph"
	"nimona.io/internal/store/kv"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/object"
	"nimona.io/pkg/orchestrator"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/stream"
)

//
//    o
//   / \
//  m1  m2
//  |   | \
//  m3  m4 |
//   \ /   |
//    m6   m5
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

	m1 = object.FromMap(map[string]interface{}{
		"@parents:as": []string{
			o.Hash().String(),
		},
		"@root:s": o.Hash().String(),
		"foo:s":   "bar-m1",
	})

	m2 = object.FromMap(map[string]interface{}{
		"@parents:as": []string{
			o.Hash().String(),
		},
		"@root:s": o.Hash().String(),
		"foo:s":   "bar-m2",
	})

	m3 = object.FromMap(map[string]interface{}{
		"@parents:as": []string{
			m1.ToObject().Hash().String(),
		},
		"@root:s": o.Hash().String(),
		"foo:s":   "bar-m3",
	})

	m4 = object.FromMap(map[string]interface{}{
		"@parents:as": []string{
			m2.ToObject().Hash().String(),
		},
		"@root:s": o.Hash().String(),
		"foo:s":   "bar-m4",
	})

	m5 = object.FromMap(map[string]interface{}{
		"@parents:as": []string{
			m2.ToObject().Hash().String(),
		},
		"@root:s": o.Hash().String(),
		"foo:s":   "bar-m5",
	})

	m6 = object.FromMap(map[string]interface{}{
		"@parents:as": []string{
			m3.ToObject().Hash().String(),
			m4.ToObject().Hash().String(),
		},
		"@root:s": o.Hash().String(),
		"foo:s":   "bar-m6",
	})
)

func TestSync(t *testing.T) {
	kv := kv.NewMemory()
	os := graph.New(kv)

	x := &exchange.MockExchange{}

	var handler func(*exchange.Envelope) error
	x.On("Handle", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		handler = args[1].(func(*exchange.Envelope) error)
	}).Return(nil, nil)

	pk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	li, err := peer.NewLocalPeer("", pk)
	assert.NoError(t, err)

	m, err := orchestrator.New(os, x, nil, li)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.NotNil(t, handler)

	rkey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	respWith := func(o object.Object) func(args mock.Arguments) {
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
		respWith((&stream.EventListCreated{
			Events: []*object.Hash{
				o.Hash(),
				m1.ToObject().Hash(),
				m2.ToObject().Hash(),
				m3.ToObject().Hash(),
				m4.ToObject().Hash(),
				m5.ToObject().Hash(),
				m6.ToObject().Hash(),
			},
		}).ToObject()),
	).Return(nil)

	// request o
	for _, i := range []object.Object{
		o,
		m1.ToObject(),
		m2.ToObject(),
		m3.ToObject(),
		m4.ToObject(),
		m5.ToObject(),
		m6.ToObject(),
	} {
		x.On(
			"Request",
			mock.Anything,
			i.Hash(),
			"peer:"+rkey.PublicKey.Fingerprint().String(),
			mock.Anything,
		).Run(
			respWith(i),
		).Return(nil)
	}

	ctx := context.Background()
	res, err := m.Sync(
		ctx,
		[]*object.Hash{
			o.Hash(),
		},
		[]string{
			"peer:" + rkey.PublicKey.Fingerprint().String(),
		},
	)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Objects, 7)

	assert.Equal(t, jp(o), jp(res.Objects[0]))
	// assert.Equal(t, jp(m1.ToObject()), jp(res.Objects[1]))
	// assert.Equal(t, jp(m2.ToObject()), jp(res.Objects[2]))
	// assert.Equal(t, jp(m3.ToObject()), jp(res.Objects[3]))
	// assert.Equal(t, jp(m4.ToObject()), jp(res.Objects[4]))
	// assert.E	qual(t, jp(m5.ToObject()), jp(res.Objects[5]))
	assert.Equal(t, jp(m6.ToObject()), jp(res.Objects[6]))

	dos, _ := os.(*graph.Graph).Dump() // nolint
	dot, _ := graph.Dot(dos)
	fmt.Println(dot)
}

// jp is a lazy approach to comparing the mess that is unmarshaling json when
// dealing with numbers
func jp(v object.Object) string {
	b, _ := json.MarshalIndent(v.ToMap(), "", "  ") // nolint
	return string(b)
}
