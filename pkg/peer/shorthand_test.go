package peer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// nolint: lll
func TestShorthand(t *testing.T) {
	tests := []struct {
		name          string
		shorthand     Shorthand
		wantValid     bool
		wantPublicKey string
		wantAddresses []string
		wantPeerError error
	}{{
		name:          "should pass, valid shorthand",
		shorthand:     "z6MktxdZRoFTVasDPBYTqYWwWgspFGWYSzhP3r8aNH8pppeh@127.0.0.1:18000",
		wantValid:     true,
		wantPublicKey: "z6MktxdZRoFTVasDPBYTqYWwWgspFGWYSzhP3r8aNH8pppeh",
		wantAddresses: []string{"127.0.0.1:18000"},
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

			gotPeer, gotPeerError := tt.shorthand.GetConnectionInfo()
			if tt.wantValid {
				assert.Equal(t, tt.wantAddresses, gotPeer.Addresses)
				assert.Equal(t, tt.wantPublicKey, gotPeer.PublicKey.String())
			}
			assert.Equal(t, tt.wantPeerError, gotPeerError)
		})
	}
}
