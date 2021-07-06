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
// bNode0 prv zrv4RST6eWXy5hHdRFLcnp9RQnMXHrERGeTRtRdABQUAUpK1ZgTczfY3UVDgoKuGsiZZ8WXVeX1xgy2TCFHXJgXCDDf
// bNode0 pub z6Mkoggr1hwTs93tie6AdpJAC96zv4z49ZpnKcq2S2KzEvyB
// bNode1 prv zrv5ifrwiak5ScRYZor8eajoAiKfgq31B7NFmxt94UR2ixi61cXr1PbR5TTZHgeR6xpdbchzgzGdf4R1BhFprasXz4m
// bNode1 pub z6MkrwtZfhcPcMgZ8py8Dz1UEXgZV1DgmB4Ryhwt42k6oQ8V
// bNode2 prv zrv1YwP9RV2p6R2zZEeaK1FeRn3BRvycyQebaWxUefBBWrgrW6AQCa3N2cNfghqJeL66AHH3wv7nrN5KrBnxVJkHYrv
// bNode2 pub z6Mkp9rzAnzcEhQ54YaiV68Mtb15qrkMKrMHgG1uRok4s4vn

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
		node.WithPortMapping(17000, 17000),
		node.WithName("nimona-e2e-bootstrap-"+rand.String(4)),
		node.WithCount(1),
		node.WithEnv([]string{
			"NIMONA_UPNP_DISABLE=true",
			"NIMONA_LOG_LEVEL=error",
			"NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17000",
			"NIMONA_PEER_PRIVATE_KEY=zrv4RST6eWXy5hHdRFLcnp9RQnMXHrERGeTRtRdABQUAUpK1ZgTczfY3UVDgoKuGsiZZ8WXVeX1xgy2TCFHXJgXCDDf",
		}),
		node.WithEntrypoint([]string{
			"/bootstrap",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, bNodes)

	// wait for the containers to settle
	time.Sleep(time.Second * 5)

	// setup node 1
	nodes := bNodes
	newNodes, err := node.New(
		dockerImage,
		env,
		node.WithPortMapping(17001, 17001),
		node.WithName("nimona-e2e-1-"+rand.String(4)),
		node.WithCount(1),
		node.WithEnv([]string{
			"NIMONA_UPNP_DISABLE=true",
			"NIMONA_LOG_LEVEL=error",
			"NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17001",
			"NIMONA_PEER_PRIVATE_KEY=zrv5ifrwiak5ScRYZor8eajoAiKfgq31B7NFmxt94UR2ixi61cXr1PbR5TTZHgeR6xpdbchzgzGdf4R1BhFprasXz4m",
			"NIMONA_PEER_BOOTSTRAPS=z6Mkoggr1hwTs93tie6AdpJAC96zv4z49ZpnKcq2S2KzEvyB@tcps:" + bNodes[0].Address(),
			"NIMONA_SONAR_PING_PEERS=z6Mkp9rzAnzcEhQ54YaiV68Mtb15qrkMKrMHgG1uRok4s4vn",
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
		node.WithPortMapping(17002, 17002),
		node.WithName("nimona-e2e-2-"+rand.String(4)),
		node.WithCount(1),
		node.WithEnv([]string{
			"NIMONA_UPNP_DISABLE=true",
			"NIMONA_LOG_LEVEL=error",
			"NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17002",
			"NIMONA_PEER_PRIVATE_KEY=zrv1YwP9RV2p6R2zZEeaK1FeRn3BRvycyQebaWxUefBBWrgrW6AQCa3N2cNfghqJeL66AHH3wv7nrN5KrBnxVJkHYrv",
			"NIMONA_PEER_BOOTSTRAPS=z6Mkoggr1hwTs93tie6AdpJAC96zv4z49ZpnKcq2S2KzEvyB@tcps:" + bNodes[0].Address(),
			"NIMONA_SONAR_PING_PEERS=z6MkrwtZfhcPcMgZ8py8Dz1UEXgZV1DgmB4Ryhwt42k6oQ8V",
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

	wgSent.Add(len(nodes) - 1)
	wgReceived.Add(len(nodes) - 1)

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
