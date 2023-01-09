package nimona

import (
	"testing"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/stretchr/testify/require"
)

func TestDocumentSignature_SignVerify(t *testing.T) {
	_, sk, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	doc := &CborFixture{
		String: "foo",
	}

	hash, err := NewDocumentHash(doc)
	require.NoError(t, err)

	sig, err := NewDocumentSignature(sk, hash)
	require.NoError(t, err)

	docBytes, err := doc.MarshalCBORBytes()
	require.NoError(t, err)

	err = VerifyDocument(docBytes, sig)
	require.NoError(t, err)
}
