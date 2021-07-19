package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/crypto"
)

func TestVerify(t *testing.T) {
	testKey0 := mustGenerateKey(t)
	testKey1 := mustGenerateKey(t)

	tests := []struct {
		name    string
		object  *Object
		wantErr bool
	}{{
		name: "should pass, no signature, no owner",
		object: &Object{
			Data: chore.Map{
				"foo:s": chore.String("bar"),
			},
		},
	}, {
		name: "should pass, no owner, with signature",
		object: mustSign(
			t,
			testKey0,
			&Object{
				Data: chore.Map{
					"foo:s": chore.String("bar"),
				},
			},
		),
	}, {
		name: "should fail, with owner, no signature",
		object: &Object{
			Metadata: Metadata{
				Owner: testKey0.PublicKey().DID(),
			},
			Data: chore.Map{
				"foo:s": chore.String("bar"),
			},
		},
		wantErr: true,
	}, {
		name: "should fail, with owner, with wrong signature",
		object: mustSign(
			t,
			testKey1,
			&Object{
				Metadata: Metadata{
					Owner: testKey0.PublicKey().DID(),
				},
				Data: chore.Map{
					"foo:s": chore.String("bar"),
				},
			},
		),
		wantErr: true,
	}, {
		name: "should fail, with owner, invalid signature",
		object: mustSign(
			t,
			testKey1,
			&Object{
				Metadata: Metadata{
					Owner: testKey0.PublicKey().DID(),
					Signature: Signature{
						X: []byte{1, 2, 3},
					},
				},
				Data: chore.Map{
					"foo:s": chore.String("bar"),
				},
			},
		),
		wantErr: true,
	}, {
		name: "should pass, with owner, owner's valid signature",
		object: mustSign(
			t,
			testKey0,
			&Object{
				Metadata: Metadata{
					Owner: testKey0.PublicKey().DID(),
				},
				Data: chore.Map{
					"foo:s": chore.String("bar"),
				},
			},
		),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Verify(tt.object); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mustGenerateKey(t *testing.T) crypto.PrivateKey {
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	return k
}

func mustSign(t *testing.T, k crypto.PrivateKey, o *Object) *Object {
	sig, err := NewSignature(k, o)
	assert.NoError(t, err)
	o.Metadata.Signature = sig
	return o
}
