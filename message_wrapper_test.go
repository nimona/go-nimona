package nimona

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageWrapper_CBOR(t *testing.T) {
	n0 := &Ping{
		Nonce: "foo",
	}

	b0 := bytes.NewBuffer(nil)
	err := n0.MarshalCBOR(b0)
	require.NoError(t, err)

	m0 := &MessageWrapper{}
	err = m0.UnmarshalCBOR(b0)
	require.NoError(t, err)
	require.Equal(t, "test/ping", m0.Type)
}
