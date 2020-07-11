package streammanager_test

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanagermock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
	"nimona.io/pkg/streammanager"
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
// o oh1.8wKCBdvvSAGmYPsUojqutMJeHJSW9RWsyezWYSDrp61T
// m1 oh1.8vTC4sMVudao8yu9XRqxGxbL5oVzpBnrD1tJEmW5RhsL
// m2 oh1.EJtWxi9SbineBFmvRg2pVqds2ZXYYrBT4bYWjpcqW9wr
// m3 oh1.7yz1xieBqDMzFwKbJEyvzBLZuAncEtVWPzR9sDKhHBfj
// m4 oh1.7sdrcbf5BifrD2bsxD4d4xHaFkr49ZNmLHGBHXQxEtg1
// m5 oh1.7Qh36on7Eq1Yz3M6cyxV3gcA35ujyA22RoajZh67pG9U
// m6 oh1.5MN741kx8Kd9WcUrYz5D5sFk6RLpbx9g8RRpDjqM8S8F
//

var (
	o = object.FromMap(map[string]interface{}{
		"type:s": "foo",
		"content:m": map[string]interface{}{
			"foo:s": "bar",
			"numbers:ai": []int{
				1, 2, 3,
			},
			"strings:as": []string{
				"a", "b", "c",
			},
			"map:m": map[string]interface{}{
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

	oh = o.Hash()

	m1 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			oh.String(),
		},
		"stream:s": oh.String(),
		"content:m": map[string]interface{}{
			"foo:s": "bar-m1",
		},
	})

	m2 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			oh.String(),
		},
		"stream:s": oh.String(),
		"content:m": map[string]interface{}{
			"foo:s": "bar-m2",
		},
	})

	m3 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			m1.ToObject().Hash().String(),
		},
		"stream:s": oh.String(),
		"content:m": map[string]interface{}{
			"foo:s": "bar-m3",
		},
	})

	m4 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			m2.ToObject().Hash().String(),
		},
		"stream:s": oh.String(),
		"content:m": map[string]interface{}{
			"foo:s": "bar-m4",
		},
	})

	m5 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			m2.ToObject().Hash().String(),
		},
		"stream:s": oh.String(),
		"content:m": map[string]interface{}{
			"foo:s": "bar-m5",
		},
	})

	m6 = object.FromMap(map[string]interface{}{
		"parents:as": []string{
			m3.ToObject().Hash().String(),
			m4.ToObject().Hash().String(),
		},
		"stream:s": oh.String(),
		"content:m": map[string]interface{}{
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
		Return(
			func(
				filters ...exchange.EnvelopeFilter,
			) exchange.EnvelopeSubscription {
				sub := &exchange.MockEnvelopeSubscription{}
				subs = append(subs, sub)
				return sub
			},
		)

	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	kc := keychain.New()
	kc.Put(keychain.PrimaryPeerKey, pk)
	kc.Put(keychain.IdentityKey, pk)

	om := objectmanagermock.NewMockRequester(gomock.NewController(t))

	m, err := streammanager.New(store, x, nil, kc, om)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.NotEmpty(t, subs)

	rkey, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	so, _ := object.NewSignature(rkey, o) // nolint: errcheck
	o = o.AddSignature(so)

	sm1, _ := object.NewSignature(rkey, m1) // nolint: errcheck
	m1 = m1.AddSignature(sm1)

	sm2, _ := object.NewSignature(rkey, m2) // nolint: errcheck
	m2 = m2.AddSignature(sm2)

	sm3, _ := object.NewSignature(rkey, m3) // nolint: errcheck
	m3 = m3.AddSignature(sm3)

	sm4, _ := object.NewSignature(rkey, m4) // nolint: errcheck
	m4 = m4.AddSignature(sm4)

	sm5, _ := object.NewSignature(rkey, m5) // nolint: errcheck
	m5 = m5.AddSignature(sm5)

	sm6, _ := object.NewSignature(rkey, m6) // nolint: errcheck
	m6 = m6.AddSignature(sm6)

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
			m1.ToObject().Hash(),
			m2.ToObject().Hash(),
			m3.ToObject().Hash(),
			m4.ToObject().Hash(),
			m5.ToObject().Hash(),
			m6.ToObject().Hash(),
		},
	}).ToObject()
	elo = elo.SetOwners([]crypto.PublicKey{
		rkey.PublicKey(),
	})
	sig, err := object.NewSignature(rkey, elo)
	assert.NoError(t, err)

	elo = elo.AddSignature(sig)

	// nonce := ""

	// send request
	x.On(
		"Send",
		mock.Anything,
		mock.MatchedBy(
			func(o object.Object) bool {
				return o.GetType() == "nimona.io/stream.Request"
			},
		),
		mock.MatchedBy(
			func(p *peer.Peer) bool {
				return p.Owners[0] == rkey.PublicKey()
			},
		),
		mock.Anything,
		mock.Anything,
	).Run(
		respWith(elo),
	).Return(nil)

	om.EXPECT().Request(gomock.Any(), o.Hash(), gomock.Any()).Return(&o, nil)

	o1 := m1.ToObject()
	om.EXPECT().Request(gomock.Any(), o1.Hash(), gomock.Any()).Return(&o1, nil)

	o2 := m2.ToObject()
	om.EXPECT().Request(gomock.Any(), o2.Hash(), gomock.Any()).Return(&o2, nil)

	o3 := m3.ToObject()
	om.EXPECT().Request(gomock.Any(), o3.Hash(), gomock.Any()).Return(&o3, nil)

	o4 := m4.ToObject()
	om.EXPECT().Request(gomock.Any(), o4.Hash(), gomock.Any()).Return(&o4, nil)

	o5 := m5.ToObject()
	om.EXPECT().Request(gomock.Any(), o5.Hash(), gomock.Any()).Return(&o5, nil)

	o6 := m6.ToObject()
	om.EXPECT().Request(gomock.Any(), o6.Hash(), gomock.Any()).Return(&o6, nil)

	ctx := context.New(
		context.WithCorrelationID("req1"),
		context.WithTimeout(time.Millisecond*500),
	)
	res, err := m.Sync(
		ctx,
		oh,
		&peer.Peer{
			Owners: []crypto.PublicKey{
				rkey.PublicKey(),
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Objects, 7)

	assert.Equal(t, jp(o), jp(res.Objects[0]))
	assert.Equal(t, jp(m6.ToObject()), jp(res.Objects[6]))
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
