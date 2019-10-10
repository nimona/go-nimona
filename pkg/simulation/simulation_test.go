package simulation

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/rand"
	"nimona.io/pkg/simulation/node"
)

func TestSimulation(t *testing.T) {
	// Setup
	env, err := node.NewEnvironment()
	require.NoError(t, err)
	require.NotNil(t, env)

	dockerImage := "docker.io/nimona/nimona:v0.5.0-alpha"
	if img := os.Getenv("E2E_DOCKER_IMAGE"); img != "" {
		dockerImage = img
	}

	// Setup Nodes
	nodes, err := node.New(
		dockerImage,
		env,
		node.WithName("nimona-e2e-"+rand.String(8)),
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
					fmt.Println(ll)
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
