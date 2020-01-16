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
	"nimona.io/pkg/hash"
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
		"@type:s": "foo",
		"foo:s":   "bar",
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

	oh = hash.New(o)

	m1 = object.FromMap(map[string]interface{}{
		"@parents:as": []interface{}{
			oh,
		},
		"@stream:s": oh,
		"foo:s":     "bar-m1",
	})

	m2 = object.FromMap(map[string]interface{}{
		"@parents:as": []interface{}{
			oh,
		},
		"@stream:s": oh,
		"foo:s":     "bar-m2",
	})

	m3 = object.FromMap(map[string]interface{}{
		"@parents:as": []interface{}{
			hash.New(m1.ToObject()),
		},
		"@stream:s": oh,
		"foo:s":     "bar-m3",
	})

	m4 = object.FromMap(map[string]interface{}{
		"@parents:as": []interface{}{
			hash.New(m2.ToObject()),
		},
		"@stream:s": oh,
		"foo:s":     "bar-m4",
	})

	m5 = object.FromMap(map[string]interface{}{
		"@parents:as": []interface{}{
			hash.New(m2.ToObject()),
		},
		"@stream:s": oh,
		"foo:s":     "bar-m5",
	})

	m6 = object.FromMap(map[string]interface{}{
		"@parents:as": []interface{}{
			hash.New(m3.ToObject()),
			hash.New(m4.ToObject()),
		},
		"@stream:s": oh,
		"foo:s":     "bar-m6",
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

	crypto.Sign(o, rkey)  // nolint: errcheck
	crypto.Sign(m1, rkey) // nolint: errcheck
	crypto.Sign(m2, rkey) // nolint: errcheck
	crypto.Sign(m3, rkey) // nolint: errcheck
	crypto.Sign(m4, rkey) // nolint: errcheck
	crypto.Sign(m5, rkey) // nolint: errcheck
	crypto.Sign(m6, rkey) // nolint: errcheck

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
			hash.New(m1.ToObject()),
			hash.New(m2.ToObject()),
			hash.New(m3.ToObject()),
			hash.New(m4.ToObject()),
			hash.New(m5.ToObject()),
			hash.New(m6.ToObject()),
		},
		Identity: rkey.PublicKey(),
	}).ToObject()

	err = crypto.Sign(elo, rkey)
	assert.NoError(t, err)

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
		peer.LookupByKey(rkey.PublicKey()),
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
