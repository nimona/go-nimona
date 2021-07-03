package keri

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func TestInception_MarshalUnmarshal(t *testing.T) {
	k0, err := crypto.NewEd25519PrivateKey(crypto.IdentityKey)
	require.NoError(t, err)

	i := &Inception{
		Metadata: object.Metadata{
			Datetime: "foo",
		},
		Version: Version,
		// Prefix:        "prefix",
		// Sequence: 2,
		Key:           k0.PublicKey(),
		NextKeyDigest: "some-digest",
		Config: []*Config{{
			Trait: "trait-a",
		}, {
			Trait: "trait-b",
		}},
		// Seals: []*Seal{{
		// 	Root: "root-a",
		// }, {
		// 	Root: "root-b",
		// }},
		// DelegatorSeal: &Seal{
		// 	Root: "root-c",
		// },
	}

	o, err := object.Marshal(i)
	require.NoError(t, err)

	b, _ := json.MarshalIndent(o, "", "  ")
	fmt.Println(string(b))

	g := &Inception{}
	err = object.Unmarshal(o, g)
	require.NoError(t, err)
	require.Equal(t, i, g)
}

func TestCreateState(t *testing.T) {
	k0, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	k1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	t0Inception := &Inception{
		Metadata: object.Metadata{},
		Version:  Version,
		Key:      k0.PublicKey(),
		// NextKeyDigest: hash(k0.PublicKey()),
	}

	t0Rotation := &Rotation{
		Metadata: object.Metadata{
			Parents: object.Parents{
				"*": chore.HashArray{
					object.MustMarshal(t0Inception).Hash(),
				},
			},
		},
		Version: Version,
		Key:     k1.PublicKey(),
		// NextKeyDigest: hash(k0.PublicKey()),
	}

	tests := []struct {
		name string
		or   object.ReadCloser

		want    *State
		wantErr bool
	}{{
		name: "small aggregate, ok",
		or: object.NewReadCloserFromObjects(
			[]*object.Object{
				object.MustMarshal(t0Inception),
				object.MustMarshal(t0Rotation),
			},
		),
		want: &State{
			Version:   Version,
			RootHash:  object.MustMarshal(t0Inception).Hash(),
			ActiveKey: k1.PublicKey(),
			RotatedKeys: []crypto.PublicKey{
				k0.PublicKey(),
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateState(tt.or)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
