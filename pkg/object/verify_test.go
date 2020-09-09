package object

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
)

func TestVerify(t *testing.T) {
	testKey0 := mustGenerateKey(t)
	testKey1 := mustGenerateKey(t)
	testKey2 := mustGenerateKey(t)

	tests := []struct {
		name    string
		object  Object
		wantErr bool
	}{{
		name: "should pass, no signature, no owner",
		object: new(Object).
			Set("foo:s", "bar"),
	}, {
		name: "should pass, no owner, with signature",
		object: mustSign(
			t,
			testKey0,
			new(Object).
				Set("foo:s", "bar"),
		),
	}, {
		name: "should fail, with owner, no signature",
		object: new(Object).
			Set("foo:s", "bar").
			SetOwner(testKey0.PublicKey()),
		wantErr: true,
	}, {
		name: "should fail, with owner, with wrong signature",
		object: mustSign(
			t,
			testKey1,
			new(Object).
				Set("foo:s", "bar").
				SetOwner(testKey0.PublicKey()),
		),
		wantErr: true,
	}, {
		name: "should fail, with owner, invalid signature",
		object: mustSign(
			t,
			testKey1,
			new(Object).
				Set("foo:s", "bar").
				SetOwner(testKey0.PublicKey()),
		).set("metadata:m/_signature:m/x:d", Bytes([]byte{1, 2, 3})),
		wantErr: true,
	}, {
		name: "should pass, with owner, owner's valid signature",
		object: mustSign(
			t,
			testKey0,
			new(Object).
				Set("foo:s", "bar").
				SetOwner(testKey0.PublicKey()),
		),
	}, {
		name: "should pass, with owner, other valid signature, with certificate",
		object: mustSignWithCertificate(
			t,
			testKey0,
			testKey1,
			new(Object).
				Set("foo:s", "bar").
				SetOwner(testKey0.PublicKey()),
		),
	}, {
		name: "should fail, with owner, with invalid certificate signature",
		object: mustSignWithCertificate(
			t,
			testKey0,
			testKey1,
			new(Object).
				Set("foo:s", "bar").
				SetOwner(testKey0.PublicKey()),
		).set(
			"metadata:m/_signature:m/certificate:m/metadata:m/_signature:m/x:d",
			Bytes([]byte{1, 2, 3}),
		),
		wantErr: true,
	}, {
		name: "should fail, with owner, invalid certificate signer",
		object: func() Object {
			o := mustSignWithCertificate(
				t,
				testKey2,
				testKey1,
				new(Object).
					Set("foo:s", "bar").
					SetOwner(testKey0.PublicKey()),
			)
			// resign with random key
			sig, err := NewSignature(testKey2, o)
			assert.NoError(t, err)
			sig.Certificate = o.GetSignature().Certificate
			return o.SetSignature(sig)
		}(),
		wantErr: true,
	}, {
		name: "should fail, with owner, wrong signature, valid certificate",
		object: func() Object {
			o := mustSignWithCertificate(
				t,
				testKey0,
				testKey1,
				new(Object).
					Set("foo:s", "bar").
					SetOwner(testKey0.PublicKey()),
			)
			sig, err := NewSignature(testKey2, o)
			assert.NoError(t, err)
			sig.Certificate = o.GetSignature().Certificate
			return o.SetSignature(sig)
		}(),
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(Dump(tt.object))
			if err := Verify(tt.object); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mustSignWithCertificate(
	t *testing.T,
	owner crypto.PrivateKey,
	signer crypto.PrivateKey,
	o Object,
) Object {
	c, err := NewCertificate(signer.PublicKey(), owner)
	require.NoError(t, err)
	sig, err := NewSignature(signer, o)
	assert.NoError(t, err)
	sig.Certificate = c
	return o.SetSignature(sig)
}

func mustGenerateKey(t *testing.T) crypto.PrivateKey {
	k, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	return k
}

func mustSign(t *testing.T, k crypto.PrivateKey, o Object) Object {
	sig, err := NewSignature(k, o)
	assert.NoError(t, err)
	return o.SetSignature(sig)
}
