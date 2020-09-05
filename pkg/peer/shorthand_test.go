package peer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

// nolint: lll
func TestShorthand(t *testing.T) {
	tests := []struct {
		name          string
		shorthand     Shorthand
		wantValid     bool
		wantPublicKey crypto.PublicKey
		wantAddresses []string
		wantPeer      *Peer
		wantPeerError error
	}{{
		name:          "should pass, valid shorthand",
		shorthand:     "ed25519.2BrcMfGTaVDo2xdbXESsXKNwqA478vrdynJrb4hqtdus@127.0.0.1:18000",
		wantValid:     true,
		wantPublicKey: "ed25519.2BrcMfGTaVDo2xdbXESsXKNwqA478vrdynJrb4hqtdus",
		wantAddresses: []string{"127.0.0.1:18000"},
		wantPeer: &Peer{
			Metadata: object.Metadata{
				Owner: crypto.PublicKey("ed25519.2BrcMfGTaVDo2xdbXESsXKNwqA478vrdynJrb4hqtdus"),
			},
			Addresses: []string{
				"127.0.0.1:18000",
			},
		},
	}, {
		name:          "should fail, invalid shorthand",
		shorthand:     "foo",
		wantValid:     false,
		wantPublicKey: "",
		wantAddresses: nil,
		wantPeerError: ErrInvalidShorthand,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid := tt.shorthand.IsValid()
			assert.Equal(t, tt.wantValid, gotValid)

			gotPublicKey := tt.shorthand.PublicKey()
			assert.Equal(t, tt.wantPublicKey, gotPublicKey)

			gotAddresses := tt.shorthand.Addresses()
			assert.Equal(t, tt.wantAddresses, gotAddresses)

			gotPeer, gotPeerError := tt.shorthand.Peer()
			assert.Equal(t, tt.wantPeer, gotPeer)
			assert.Equal(t, tt.wantPeerError, gotPeerError)
		})
	}
}
