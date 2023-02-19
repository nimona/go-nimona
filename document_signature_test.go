package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentSignature_SignVerify(t *testing.T) {
	_, sk, err := GenerateKey()
	require.NoError(t, err)

	doc := &CborFixture{
		String: "foo",
	}

	hash := NewDocumentHash(doc.DocumentMap())
	sig := NewDocumentSignature(sk, hash)
	err = VerifySignature(sig, hash)
	require.NoError(t, err)
}
