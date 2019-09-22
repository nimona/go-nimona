package simulation

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nimona.io/internal/simulation/node"
)

func TestSimulation(t *testing.T) {
	// Setup
	env, err := node.NewEnvironment()
	require.NoError(t, err)
	require.NotNil(t, env)

	// Setup Nodes
	nodes, err := node.New(
		"docker.io/nimona/nimona:v0.4.2",
		env,
		node.WithName("NimTest"),
		node.WithNodePort(8000),
		node.WithCount(10),
	)
	require.NoError(t, err)
	require.NotNil(t, nodes)

	// Teardown
	err = node.Stop(nodes)
	require.NoError(t, err)
	err = env.Stop()
	require.NoError(t, err)
}
