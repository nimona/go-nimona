package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/rand"
	"nimona.io/pkg/chore"
	"nimona.io/pkg/crypto"
)

func TestVerify(t *testing.T) {
	testKey0 := mustGenerateKey(t)
	testKey1 := mustGenerateKey(t)
	testKey2 := mustGenerateKey(t)

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
				Owner: testKey0.PublicKey(),
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
					Owner: testKey0.PublicKey(),
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
					Owner: testKey0.PublicKey(),
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
					Owner: testKey0.PublicKey(),
				},
				Data: chore.Map{
					"foo:s": chore.String("bar"),
				},
			},
		),
	}, {
		name: "should pass, with owner, other valid signature, with certificate",
		object: mustSignWithCertificate(
			t,
			testKey0,
			testKey1,
			&Object{
				Type: "",
				Metadata: Metadata{
					Owner: testKey0.PublicKey(),
				},
				Data: chore.Map{
					"foo:s": chore.String("bar"),
				},
			},
		),
	}, {
		name: "should fail, with owner, with invalid certificate signature",
		object: func() *Object {
			o := mustSignWithCertificate(
				t,
				testKey0,
				testKey1,
				&Object{
					Type: "",
					Metadata: Metadata{
						Owner: testKey0.PublicKey(),
					},
					Data: chore.Map{
						"foo:s": chore.String("bar"),
					},
				},
			)
			sig := []byte{1, 2, 3}
			o.Metadata.Signature.Certificate.Metadata.Signature.X = sig
			return o
		}(),
		wantErr: true,
	}, {
		name: "should fail, with owner, invalid certificate signer",
		object: func() *Object {
			o := mustSignWithCertificate(
				t,
				testKey2,
				testKey1,
				&Object{
					Type: "",
					Metadata: Metadata{
						Owner: testKey0.PublicKey(),
					},
					Data: chore.Map{
						"foo:s": chore.String("bar"),
					},
				},
			)
			// resign with random key
			sig, err := NewSignature(testKey2, o)
			assert.NoError(t, err)
			sig.Certificate = o.Metadata.Signature.Certificate
			o.Metadata.Signature = sig
			return o
		}(),
		wantErr: true,
	}, {
		name: "should fail, with owner, wrong signature, valid certificate",
		object: func() *Object {
			o := mustSignWithCertificate(
				t,
				testKey0,
				testKey1,
				&Object{
					Type: "",
					Metadata: Metadata{
						Owner: testKey0.PublicKey(),
					},
					Data: chore.Map{
						"foo:s": chore.String("bar"),
					},
				},
			)
			sig, err := NewSignature(testKey2, o)
			assert.NoError(t, err)
			sig.Certificate = o.Metadata.Signature.Certificate
			o.Metadata.Signature = sig
			return o
		}(),
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Verify(tt.object); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mustSignWithCertificate(
	t *testing.T,
	identityKey crypto.PrivateKey,
	peerKey crypto.PrivateKey,
	o *Object,
) *Object {
	req := &CertificateRequest{
		Metadata: Metadata{
			Owner: peerKey.PublicKey(),
		},
		Nonce:                  rand.String(5),
		VendorName:             "vendor",
		ApplicationName:        "app-name",
		ApplicationDescription: "app-descr",
		ApplicationURL:         "https://foo",
		Permissions: []CertificatePermission{{
			Types:   []string{"*"},
			Actions: []string{"*"},
		}},
	}
	reso, err := Marshal(req)
	require.NoError(t, err)
	resSig, err := NewSignature(peerKey, reso)
	require.NoError(t, err)
	req.Metadata.Signature = resSig
	res, err := NewCertificate(identityKey, *req, true, "")
	require.NoError(t, err)
	sig, err := NewSignature(peerKey, o)
	assert.NoError(t, err)
	sig.Certificate = &res.Certificate
	o.Metadata.Signature = sig
	return o
}

func mustGenerateKey(t *testing.T) crypto.PrivateKey {
	k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	return k
}

func mustSign(t *testing.T, k crypto.PrivateKey, o *Object) *Object {
	sig, err := NewSignature(k, o)
	assert.NoError(t, err)
	o.Metadata.Signature = sig
	return o
}
