package containers_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/simulation/containers"
)

func TestCreateContainer(t *testing.T) {
	tests := []struct {
		name          string
		image         string
		containerName string
		networkName   string
	}{
		{
			name:          "HappyPath",
			image:         "docker.io/library/alpine:3.10",
			containerName: "testnode",
			// todo randomize name
			networkName: "testnet",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			nn, err := containers.NewNetwork(ctx, tt.networkName)
			require.NoError(t, err)
			require.NotNil(t, nn)

			defer func() {
				err = nn.Remove(ctx)
				require.NoError(t, err)
			}()

			c1, err := containers.New(
				ctx,
				tt.image,
				tt.containerName,
				tt.networkName,
				nil,
				[]string{"uname", "-a"},
			)
			require.NoError(t, err)
			require.NotNil(t, c1)

			defer func() {
				err := c1.Stop(ctx)
				require.NoError(t, err)
			}()

			logReader, err := c1.Logs(ctx)
			defer logReader.Close() // nolint
			require.NoError(t, err)

			logData, err := ioutil.ReadAll(logReader)
			require.NoError(t, err)
			require.NotEmpty(t, logData)
		})
	}
}
