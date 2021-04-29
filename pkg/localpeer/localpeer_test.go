package localpeer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/rand"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func TestLocalPeer(t *testing.T) {
	k1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	k2, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	lp := New()

	lp.SetPeerKey(k1)
	assert.Equal(t, k1, lp.GetPeerKey())

	csr := &object.CertificateRequest{
		Metadata: object.Metadata{
			Owner: k2.PublicKey(),
		},
		Nonce:      rand.String(8),
		VendorName: "foo",
	}
	csr.Metadata.Signature, err = object.NewSignature(k2, csr.ToObject())
	require.NoError(t, err)

	csrRes, err := object.NewCertificate(k2, *csr, true, "bar")
	require.NoError(t, err)

	lp.SetPeerCertificate(csrRes)
	assert.Equal(t, k2.PublicKey(), lp.GetIdentityPublicKey())

	ch1 := object.CID("f01")
	ch2 := object.CID("f02")

	lp.RegisterCIDs(ch1)
	assert.ElementsMatch(t, []object.CID{ch1}, lp.GetCIDs())

	lp.RegisterCIDs(ch1, ch2)
	assert.ElementsMatch(t, []object.CID{ch1, ch2}, lp.GetCIDs())
}

func TestEventUpdates(t *testing.T) {
	lp := New()

	c, cf := lp.ListenForUpdates()
	defer cf()

	go func() {
		time.Sleep(10 * time.Millisecond)
		lp.SetPeerCertificate(&object.CertificateResponse{})
	}()

	e := <-c
	assert.Equal(t, EventIdentityKeyUpdated, e)
}
