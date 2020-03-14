package client

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"nimona.io/pkg/peer"
)

func TestClient_Info(t *testing.T) {
	httpClient := http.DefaultClient
	baseURL, err := url.Parse("http://nimona")
	require.NoError(t, err)

	tests := []struct {
		name     string
		respBody interface{}
		respCode int
		assert   func(*testing.T, *peer.Peer)
		wantErr  bool
	}{
		{
			name:     "success",
			respCode: 200,
			respBody: map[string]interface{}{
				"_signatures:ao": []interface{}{
					map[string]interface{}{
						"algorithm:s": "OH_ES256",
					},
				},
				"_hash:s": "oh1.xxx",
				"owners:as": []string{
					"foo",
				},
				"data:o": map[string]interface{}{
					"addresses:as": []string{
						"tcps:127.0.0.1:21013",
						"tcps:192.168.1.57:21013",
					},
				},
			},
			assert: func(t *testing.T, r *peer.Peer) {
				require.NotEmpty(t, r.PublicKey().String())
				require.NotEmpty(t, r.Addresses)
			},
			wantErr: false,
		},
		{
			name:     "error on 500",
			respCode: 500,
			respBody: nil,
			assert: func(t *testing.T, r *peer.Peer) {
				require.Nil(t, r)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL:    baseURL,
				httpClient: httpClient,
			}
			gock.New("http://nimona").
				Get("/api/v1/local").
				Reply(tt.respCode).
				JSON(tt.respBody)
			got, err := c.Info()
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Client.Info() error = %v, wantErr %v",
					err, tt.wantErr,
				)
			}
			tt.assert(t, got)
			gock.Off()
		})
	}
}
