package orchestrator_test

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/object"
	"nimona.io/pkg/orchestrator"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
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
// o oh1.E9W237sWo6b9j8C9G43XGYBjBsACq8zH6NvW4cTHxkk4
// m1 oh1.5rA4otS4aA64xuRbDptYsxGeCDS2DWtVSHNXwbDn7d1p
// m2 oh1.AYG83AybocVBuBcWJefS5dK8UPWbQe9r9XqG9HgVc3Fq
// m3 oh1.7BYfVhAgrm2t3pvtAyh1CtHbvqGtcFBjxQ8Q7FuXMWi8
// m4 oh1.2xzVNtS9GLf9iz4t2Ye9rRBYm747xQ1kxHtm23otb8VN
// m5 oh1.EPWPg5K421ZrWMQkzYKeufD22Ndd4ZcBMdejuWNRNGEX
// m6 oh1.2sWcm2YgAo1T8QE1CymyMcpvkBPDzEHPzLpoxozpLRLT
//

var (
	o = object.FromMap(map[string]interface{}{
		"type:s": "foo",
		"data:o": map[string]interface{}{
			"foo:s": "bar",
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
		},
	})

	oh = object.NewHash(o)

	m1 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			oh.String(),
		},
		"stream:s": oh.String(),
		"data:o": map[string]interface{}{
			"foo:s": "bar-m1",
		},
	})

	m2 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			oh.String(),
		},
		"stream:s": oh.String(),
		"data:o": map[string]interface{}{
			"foo:s": "bar-m2",
		},
	})

	m3 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			object.NewHash(m1.ToObject()).String(),
		},
		"stream:s": oh.String(),
		"data:o": map[string]interface{}{
			"foo:s": "bar-m3",
		},
	})

	m4 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			object.NewHash(m2.ToObject()).String(),
		},
		"stream:s": oh.String(),
		"data:o": map[string]interface{}{
			"foo:s": "bar-m4",
		},
	})

	m5 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			object.NewHash(m2.ToObject()).String(),
		},
		"stream:s": oh.String(),
		"data:o": map[string]interface{}{
			"foo:s": "bar-m5",
		},
	})

	m6 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			object.NewHash(m3.ToObject()).String(),
			object.NewHash(m4.ToObject()).String(),
		},
		"stream:s": oh.String(),
		"data:o": map[string]interface{}{
			"foo:s": "bar-m6",
		},
	})
)

func TestSync(t *testing.T) {
	dblite := tempSqlite3(t)
	store, err := sqlobjectstore.New(dblite)
	assert.NoError(t, err)

	x := &exchange.MockExchange{}
	subs := []*exchange.MockEnvelopeSubscription{}
	x.On("Subscribe", mock.Anything).
		Return(func(filters ...exchange.EnvelopeFilter) exchange.EnvelopeSubscription {
			sub := &exchange.MockEnvelopeSubscription{}
			subs = append(subs, sub)
			return sub
		})

	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	li, err := peer.NewLocalPeer("", pk)
	assert.NoError(t, err)

	m, err := orchestrator.New(store, x, nil, li)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.NotEmpty(t, subs)

	rkey, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	so, _ := object.NewSignature(rkey, o) // nolint: errcheck
	o = o.SetSignature(so)

	sm1, _ := object.NewSignature(rkey, m1) // nolint: errcheck
	m1 = m1.SetSignature(sm1)

	sm2, _ := object.NewSignature(rkey, m2) // nolint: errcheck
	m2 = m2.SetSignature(sm2)

	sm3, _ := object.NewSignature(rkey, m3) // nolint: errcheck
	m3 = m3.SetSignature(sm3)

	sm4, _ := object.NewSignature(rkey, m4) // nolint: errcheck
	m4 = m4.SetSignature(sm4)

	sm5, _ := object.NewSignature(rkey, m5) // nolint: errcheck
	m5 = m5.SetSignature(sm5)

	sm6, _ := object.NewSignature(rkey, m6) // nolint: errcheck
	m6 = m6.SetSignature(sm6)

	respWith := func(o object.Object) func(args mock.Arguments) {
		return func(args mock.Arguments) {
			for _, s := range subs {
				s.AddNext(&exchange.Envelope{
					Payload: o,
					Sender:  rkey.PublicKey(),
				}, nil)
			}
		}
	}

	// construct event list
	elo := (&stream.Response{
		Stream: oh,
		Children: []object.Hash{
			oh,
			object.NewHash(m1.ToObject()),
			object.NewHash(m2.ToObject()),
			object.NewHash(m3.ToObject()),
			object.NewHash(m4.ToObject()),
			object.NewHash(m5.ToObject()),
			object.NewHash(m6.ToObject()),
		},
	}).ToObject()
	elo = elo.SetOwners([]crypto.PublicKey{
		rkey.PublicKey(),
	})
	sig, err := object.NewSignature(rkey, elo)
	assert.NoError(t, err)

	elo = elo.SetSignature(sig)

	nonce := ""

	// send request
	x.On(
		"Send",
		mock.Anything,
		mock.MatchedBy(
			func(o object.Object) bool {
				if o.GetType() != "nimona.io/stream.Request" {
					return false
				}
				nonce = o.Get("nonce:s").(string)
				return true
			},
		),
		mock.MatchedBy(
			func(opt peer.LookupOption) bool {
				opts := peer.ParseLookupOptions(opt)
				return opts.Lookups[0] == rkey.PublicKey().String()
			},
		),
		mock.Anything,
		mock.Anything,
	).Run(
		respWith(elo),
	).Return(nil)

	o1 := m1.ToObject()
	o2 := m2.ToObject()
	o3 := m3.ToObject()
	o4 := m4.ToObject()
	o5 := m5.ToObject()
	o6 := m6.ToObject()

	ores := &stream.ObjectResponse{
		Nonce:  nonce,
		Stream: oh,
		Objects: []*object.Object{
			&o,
			&o1,
			&o2,
			&o3,
			&o4,
			&o5,
			&o6,
		},
	}

	// request o
	x.On(
		"Send",
		mock.Anything,
		mock.MatchedBy(
			func(o object.Object) bool {
				return o.GetType() == "nimona.io/stream.ObjectRequest"
			},
		),
		mock.MatchedBy(
			func(opt peer.LookupOption) bool {
				opts := peer.ParseLookupOptions(opt)
				return opts.Lookups[0] == rkey.PublicKey().String()
			},
		),
		mock.Anything,
		mock.Anything,
	).Run(
		respWith(ores.ToObject()),
	).Return(nil)

	ctx := context.New(
		context.WithCorrelationID("req1"),
		context.WithTimeout(time.Millisecond*500),
	)
	res, err := m.Sync(
		ctx,
		oh,
		peer.LookupByOwner(rkey.PublicKey()),
	)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Objects, 7)

	assert.Equal(t, jp(o), jp(res.Objects[0]))
	// assert.Equal(t, jp(m1.ToObject()), jp(res.Objects[1]))
	// assert.Equal(t, jp(m2.ToObject()), jp(res.Objects[2]))
	// assert.Equal(t, jp(m3.ToObject()), jp(res.Objects[3]))
	// assert.Equal(t, jp(m4.ToObject()), jp(res.Objects[4]))
	// assert.E	qual(t, jp(m5.ToObject()), jp(res.Objects[5]))
	assert.Equal(t, jp(m6.ToObject()), jp(res.Objects[6]))

	// dos, _ := os.(*graph.Graph).Dump() // nolint
	// dot, _ := graph.Dot(dos)
	// fmt.Println(dot)
}

// jp is a lazy approach to comparing the mess that is unmarshaling json when
// dealing with numbers
func jp(v object.Object) string {
	b, _ := json.MarshalIndent(v.ToMap(), "", "  ") // nolint
	return string(b)
}

func tempSqlite3(t *testing.T) *sql.DB {
	dirPath, err := ioutil.TempDir("", "nimona-store-sql")
	require.NoError(t, err)
	db, err := sql.Open("sqlite3", path.Join(dirPath, "sqlite3.db"))
	require.NoError(t, err)
	return db
}
