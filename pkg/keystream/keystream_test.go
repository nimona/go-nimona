package keystream

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/did"
	"nimona.io/pkg/object"
	"nimona.io/pkg/tilde"
)

func TestInception_MarshalUnmarshal(t *testing.T) {
	k0, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	i := &Inception{
		Metadata: object.Metadata{
			Timestamp: "foo",
		},
		Version:       Version,
		Key:           k0.PublicKey(),
		NextKeyDigest: "some-digest",
	}

	o, err := object.Marshal(i)
	require.NoError(t, err)

	g := &Inception{}
	err = object.Unmarshal(o, g)
	require.NoError(t, err)
	require.Equal(t, i, g)
}

func Test_FromStream_InceptionRotation(t *testing.T) {
	k0, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	k1, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	k2, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	t0Inception := &Inception{
		Metadata: object.Metadata{
			Sequence: 0,
		},
		Version: Version,
		Key:     k0.PublicKey(),
		DelegatorSeal: &DelegatorSeal{
			Root:     "delegator-root-hash",
			Sequence: 12,
		},
		NextKeyDigest: k1.PublicKey().Hash(),
	}

	t0Rotation := &Rotation{
		Metadata: object.Metadata{
			Parents: object.Parents{
				"*": tilde.DigestArray{
					object.MustMarshal(t0Inception).Hash(),
				},
			},
			Sequence: 1,
		},
		Version:       Version,
		Key:           k1.PublicKey(),
		NextKeyDigest: k2.PublicKey().Hash(),
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
			Version:       Version,
			Sequence:      1,
			Root:          object.MustMarshal(t0Inception).Hash(),
			DelegatorRoot: "delegator-root-hash",
			Delegator: did.DID{
				Method:   did.MethodNimona,
				Identity: "delegator-root-hash",
			},
			ActiveKey:     k1.PublicKey(),
			NextKeyDigest: k2.PublicKey().Hash(),
			RotatedKeys: []crypto.PublicKey{
				k0.PublicKey(),
			},
			latestObject: object.MustMarshal(t0Rotation).Hash(),
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromStream(tt.or)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
