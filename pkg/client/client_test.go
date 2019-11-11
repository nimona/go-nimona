package client

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestClient_Info(t *testing.T) {
	httpClient := http.DefaultClient
	baseURL, err := url.Parse("http://nimona")
	require.NoError(t, err)

	tests := []struct {
		name     string
		respBody interface{}
		respCode int
		assert   func(*testing.T, *InfoResponse)
		wantErr  bool
	}{
		{
			name:     "success",
			respCode: 200,
			respBody: map[string]interface{}{
				"@signature:o": map[string]interface{}{
					"algorithm:s": "OH_ES256",
				},
				"_hash": map[string]interface{}{
					"algorithm:s": "OH1",
				},
				"_fingerprint":  "xxx",
				"_hash.compact": "OH1.xxx",
				"addresses:as": []string{
					"tcps:127.0.0.1:21013",
					"tcps:192.168.1.57:21013",
				},
			},
			assert: func(t *testing.T, r *InfoResponse) {
				require.NotEmpty(t, r.Hash.Algorithm)
				require.NotEmpty(t, r.HashCompact)
				require.NotEmpty(t, r.Fingerprint)
				require.NotEmpty(t, r.Addresses)
			},
			wantErr: false,
		},
		{
			name:     "error on 500",
			respCode: 500,
			respBody: nil,
			assert: func(t *testing.T, r *InfoResponse) {
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
