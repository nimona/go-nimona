package daemon

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
)

func TestNew_ThingsAreThere(t *testing.T) {
	d, err := New(
		context.New(),
		WithConfigOptions(
			config.WithDefaultPath(t.TempDir()),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, d.Config())
	require.NotNil(t, d.Network())
	require.NotNil(t, d.Resolver())
	require.NotNil(t, d.ObjectStore())
	require.NotNil(t, d.ObjectManager())
}
