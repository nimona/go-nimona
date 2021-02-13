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
// bNode0 prv bagacmacaukn7qjqmryxlatyqfmdqwhjnpmp7xpwku2bpfdjrmwqzognl53dfz7ie2itebh4qosflq7vlrs62tjomad2ak7xrgjjm6wlohmlphwi
// bNode0 pub bahwqcabalt6qjurgicpza5ekxb7kxdf5vgs4yahuav7pcmssz5mw4oyw6pmq
// bNode1 prv bagacmaca4u5keys3owvgynw2rjmeohit5ipxuc765ysxivkrycgng7v5zp3hffixsjm6dd76z5cwp32iy34tvtulw2jjb5stcqlxywkdqj2llyq
// bNode1 pub bahwqcabaokkrpesz4gh75t2fm7xurrxzhlhixnussd3fgfaxprmuhatuwxra
// bNode2 prv bagacmacar6wn4q3rn6owbihzuhf442wc6rxngrbkpziwdarf7whs4g5eo6qmfek3ahjt2ekijerpchktxldlnhpbyac5d5tz7nnlv6v5nmljrci
// bNode2 pub bahwqcabaykivwaothuiuqsjc6eovhowgw2o6dqaf2h3ht622xl5l22ywtceq

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
			"NIMONA_PEER_PRIVATE_KEY=bagacmacaukn7qjqmryxlatyqfmdqwhjnpmp7xpwku2bpfdjrmwqzognl53dfz7ie2itebh4qosflq7vlrs62tjomad2ak7xrgjjm6wlohmlphwi",
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
			"NIMONA_PEER_PRIVATE_KEY=bagacmaca4u5keys3owvgynw2rjmeohit5ipxuc765ysxivkrycgng7v5zp3hffixsjm6dd76z5cwp32iy34tvtulw2jjb5stcqlxywkdqj2llyq",
			"NIMONA_PEER_BOOTSTRAPS=bahwqcabalt6qjurgicpza5ekxb7kxdf5vgs4yahuav7pcmssz5mw4oyw6pmq@tcps:" + bNodes[0].Address(),
			"NIMONA_SONAR_PING_PEERS=bahwqcabaykivwaothuiuqsjc6eovhowgw2o6dqaf2h3ht622xl5l22ywtceq",
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
			"NIMONA_PEER_PRIVATE_KEY=bagacmacar6wn4q3rn6owbihzuhf442wc6rxngrbkpziwdarf7whs4g5eo6qmfek3ahjt2ekijerpchktxldlnhpbyac5d5tz7nnlv6v5nmljrci",
			"NIMONA_PEER_BOOTSTRAPS=bahwqcabalt6qjurgicpza5ekxb7kxdf5vgs4yahuav7pcmssz5mw4oyw6pmq@tcps:" + bNodes[0].Address(),
			"NIMONA_SONAR_PING_PEERS=bahwqcabaokkrpesz4gh75t2fm7xurrxzhlhixnussd3fgfaxprmuhatuwxra",
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
