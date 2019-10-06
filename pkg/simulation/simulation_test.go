package simulation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nimona.io/pkg/simulation/node"
)

func TestSimulation(t *testing.T) {
	// Setup
	env, err := node.NewEnvironment()
	require.NoError(t, err)
	require.NotNil(t, env)

	// Setup Nodes
	nodes, err := node.New(
		"docker.io/nimona/nimona:v0.5.0-alpha",
		env,
		node.WithName("NimTest"),
		node.WithNodePort(8000),
		node.WithCount(5),
	)
	require.NoError(t, err)
	require.NotNil(t, nodes)

	done := make(chan bool)

	for _, nd := range nodes {
		loch, errCh := nd.Logs()
		lineCounter := 0

		go func() {
			for {
				select {
				case ll := <-loch:
					assert.NotEmpty(t, ll)
					lineCounter++
				case <-time.After(1 * time.Second):
					done <- true
					return
				}
			}
		}()

		<-done
		assert.Empty(t, errCh)
		assert.NotZero(t, lineCounter)
	}

	// Teardown
	err = node.Stop(nodes)
	require.NoError(t, err)
	err = env.Stop()
	require.NoError(t, err)
}
