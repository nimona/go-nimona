package simulation

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/rand"
	"nimona.io/internal/simulation/node"
)

// nolint: lll
func TestSimulation(t *testing.T) {
	// create new env
	env, err := node.NewEnvironment()
	require.NoError(t, err)
	require.NotNil(t, env)

	// stop env when we are done
	defer func() {
		time.Sleep(time.Second)
		env.Stop() // nolint: errcheck
	}()

	// figure out which docker image we need
	dockerImage := "nimona:dev"
	if img := os.Getenv("E2E_DOCKER_IMAGE"); img != "" {
		dockerImage = img
	}

	// setup bootstrap nodes
	bNodes, err := node.New(
		dockerImage,
		env,
		node.WithName("nimona-e2e-bootstrap-"+rand.String(8)),
		node.WithPortMapping(17000, 17000),
		node.WithCount(1),
		node.WithEnv([]string{
			"NIMONA_LOG_LEVEL=error",
			"NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17000",
			"NIMONA_PEER_PRIVATE_KEY=ed25519.prv.Jf3xha8ZqEnFv9T9UDcN41nFFfZpc9MY4tzUnpgGHx8ZwKQ6uXX6PGY1nHLQAKhPiFtV4YEqMsCd5vjkdRyC5nJ",
			"NIMONA_SONAR_PING_PEERS=ed25519.3ykKbHUoHE8Sa9P6ckzrsXzGw3HC9iV4vnTcNrrcBBmP,ed25519.9CA3BuLzPrxHAHET8zicCtTku5zAaPsA6WRFp4PRARx2",
		}),
		node.WithEntrypoint([]string{
			"/sonar",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, bNodes)

	// wait for the containers to settle
	time.Sleep(time.Second * 15)

	// setup node 1
	nodes := bNodes
	newNodes, err := node.New(
		dockerImage,
		env,
		node.WithName("nimona-e2e-1"),
		node.WithPortMapping(17001, 17001),
		node.WithCount(1),
		node.WithEnv([]string{
			"NIMONA_LOG_LEVEL=error",
			"NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17001",
			"NIMONA_PEER_PRIVATE_KEY=ed25519.prv.2bAdgQxfcJsGRMccgMXkGSPQt396g77KKq8y6fEeQbxpnPqS5Ujh1DXTNU539wW5ispS1McLyKjrJDsgxYKneyCZ",
			"NIMONA_PEER_BOOTSTRAPS=ed25519.J9AfT7J2SbXen83NuyVEQ7UkpCaLJbnw41nLrR82HnSW@tcps:" + bNodes[0].Address(),
			"NIMONA_SONAR_PING_PEERS=ed25519.J9AfT7J2SbXen83NuyVEQ7UkpCaLJbnw41nLrR82HnSW,ed25519.9CA3BuLzPrxHAHET8zicCtTku5zAaPsA6WRFp4PRARx2",
		}),
		node.WithEntrypoint([]string{
			"/sonar",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, newNodes)
	nodes = append(nodes, newNodes[0])

	// setup node 2
	newNodes, err = node.New(
		dockerImage,
		env,
		node.WithName("nimona-e2e-2"),
		node.WithPortMapping(17002, 17002),
		node.WithCount(1),
		node.WithEnv([]string{
			"NIMONA_LOG_LEVEL=error",
			"NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17002",
			"NIMONA_PEER_PRIVATE_KEY=ed25519.prv.4i1anFeotM4TKnjLsJFLgwERtq4rD5yaR6AQ5HuChgNBfzrApXpQYA8WT83bMSc8CLj76LbJfdSKn3HiKmSpn25U",
			"NIMONA_PEER_BOOTSTRAPS=ed25519.J9AfT7J2SbXen83NuyVEQ7UkpCaLJbnw41nLrR82HnSW@tcps:" + bNodes[0].Address(),
			"NIMONA_SONAR_PING_PEERS=ed25519.J9AfT7J2SbXen83NuyVEQ7UkpCaLJbnw41nLrR82HnSW,ed25519.3ykKbHUoHE8Sa9P6ckzrsXzGw3HC9iV4vnTcNrrcBBmP",
		}),
		node.WithEntrypoint([]string{
			"/sonar",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, newNodes)
	nodes = append(nodes, newNodes[0])

	// stop all nodes when we are done
	defer func() {
		node.Stop(nodes)   // nolint: errcheck
		node.Delete(nodes) // nolint: errcheck
	}()

	wgSent := &sync.WaitGroup{}
	wgReceived := &sync.WaitGroup{}

	wgSent.Add(len(nodes))
	wgReceived.Add(len(nodes))

	for _, nd := range nodes {
		go func(nd *node.Node) {
			loch, _ := nd.Logs()
			for {
				select {
				case ll := <-loch:
					if strings.Contains(ll, "all pings sent") {
						fmt.Println("log: all pings sent")
						wgSent.Done()
					}
					if strings.Contains(ll, "all pings received") {
						fmt.Println("log: all pings received")
						wgReceived.Done()
					}
				case <-time.After(time.Minute):
					t.Log("didn't see expected logs in time")
					t.Fail()
					wgSent.Done()
					wgReceived.Done()
					return
				}
			}
		}(nd)
	}

	for _, nd := range nodes {
		go func(nd *node.Node) {
			loch, _ := nd.Logs()
			for ll := range loch {
				fmt.Println(ll)
			}
		}(nd)
	}

	wgSent.Wait()
	wgReceived.Wait()
}
