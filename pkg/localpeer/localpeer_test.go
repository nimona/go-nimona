package localpeer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/rand"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestLocalPeer(t *testing.T) {
	k1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	k2, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	lp := New()

	lp.PutPrimaryPeerKey(k1)
	assert.Equal(t, k1, lp.GetPrimaryPeerKey())

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

	lp.PutPeerCertificate(csrRes)
	assert.Equal(t, k2.PublicKey(), lp.GetIdentityPublicKey())

	ch1 := object.CID("f01")
	ch2 := object.CID("f02")

	lp.PutCIDs(ch1)
	assert.ElementsMatch(t, []object.CID{ch1}, lp.GetCIDs())

	lp.PutCIDs(ch1, ch2)
	assert.ElementsMatch(t, []object.CID{ch1, ch2}, lp.GetCIDs())

	lp.PutRelays(&peer.ConnectionInfo{
		PublicKey: k1.PublicKey(),
	})
	assert.ElementsMatch(t, []*peer.ConnectionInfo{{
		PublicKey: k1.PublicKey(),
	}}, lp.GetRelays())

	c1 := &object.Certificate{
		Nonce: rand.String(6),
	}
	lp.PutCertificate(c1)
	assert.Len(t, lp.GetCertificates(), 1)
	assert.ElementsMatch(t, []*object.Certificate{c1}, lp.GetCertificates())

	a1 := "foo"
	a2 := "foo2"
	lp.PutAddresses(a1, a2)
	assert.ElementsMatch(t, []string{a1, a2}, lp.GetAddresses())

	lp.PutContentTypes(a1, a2)
	assert.ElementsMatch(t, []string{a1, a2}, lp.GetContentTypes())

	ci := lp.ConnectionInfo()
	e := &peer.ConnectionInfo{
		PublicKey: k1.PublicKey(),
		Addresses: []string{"foo", "foo2"},
		Relays: []*peer.ConnectionInfo{{
			PublicKey: k1.PublicKey(),
		}},
		ObjectFormats: []string{
			"json",
		},
	}
	assert.Equal(t, ci, e)
}

func TestEventUpdates(t *testing.T) {
	lp := New()

	c, cf := lp.ListenForUpdates()
	defer cf()

	go func() {
		time.Sleep(10 * time.Millisecond)
		lp.PutAddresses("a", "b")
	}()

	e := <-c
	assert.Equal(t, EventAddressesUpdated, e)
}
