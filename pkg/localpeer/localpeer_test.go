package localpeer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
)

func TestLocalPeer(t *testing.T) {
	k1, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	// k2, err := crypto.NewEd25519PrivateKey()
	// require.NoError(t, err)

	lp := New()

	lp.SetPeerKey(k1)
	assert.Equal(t, k1, lp.GetPeerKey())

	// TODO(geoah): fix identity

	// csr := &object.CertificateRequest{
	// 	Metadata: object.Metadata{
	// 		Owner: k2.PublicKey(),
	// 	},
	// 	Nonce:      rand.String(8),
	// 	VendorName: "foo",
	// }
	// csr.Metadata.Signature, err = object.NewSignature(
	// 	k2,
	// 	object.MustMarshal(csr),
	// )
	// require.NoError(t, err)

	// csrRes, err := object.NewCertificate(k2, *csr, true, "bar")
	// require.NoError(t, err)

	// lp.SetPeerCertificate(csrRes)
	// assert.Equal(t, k2.PublicKey(), lp.GetIdentityPublicKey())
}

// TODO(geoah): fix identity
// func TestEventUpdates(t *testing.T) {
// 	lp := New()

// 	c, cf := lp.ListenForUpdates()
// 	defer cf()

// 	go func() {
// 		time.Sleep(10 * time.Millisecond)
// 		lp.SetPeerCertificate(&object.CertificateResponse{})
// 	}()

// 	e := <-c
// 	assert.Equal(t, EventIdentityUpdated, e)
// }
