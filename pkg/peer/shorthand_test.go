package peer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/crypto"
)

// nolint: lll
func TestShorthand(t *testing.T) {
	tests := []struct {
		name          string
		shorthand     Shorthand
		wantValid     bool
		wantPublicKey crypto.PublicKey
		wantAddresses []string
		wantPeer      *ConnectionInfo
		wantPeerError error
	}{{
		name:          "should pass, valid shorthand",
		shorthand:     "ed25519.2BrcMfGTaVDo2xdbXESsXKNwqA478vrdynJrb4hqtdus@127.0.0.1:18000",
		wantValid:     true,
		wantPublicKey: "ed25519.2BrcMfGTaVDo2xdbXESsXKNwqA478vrdynJrb4hqtdus",
		wantAddresses: []string{"127.0.0.1:18000"},
		wantPeer: &ConnectionInfo{
			PublicKey: crypto.PublicKey("ed25519.2BrcMfGTaVDo2xdbXESsXKNwqA478vrdynJrb4hqtdus"),
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

			gotPeer, gotPeerError := tt.shorthand.ConnectionInfo()
			assert.Equal(t, tt.wantPeer, gotPeer)
			assert.Equal(t, tt.wantPeerError, gotPeerError)
		})
	}
}
