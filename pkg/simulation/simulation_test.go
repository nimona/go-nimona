package simulation

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/fixtures"
	"nimona.io/internal/rand"
	"nimona.io/pkg/client"
	"nimona.io/pkg/object"
	"nimona.io/pkg/simulation/node"
)

func TestSimulation(t *testing.T) {
	// create new env
	env, err := node.NewEnvironment()
	require.NoError(t, err)
	require.NotNil(t, env)

	// stop env when we are done
	defer func() {
		err := env.Stop()
		assert.NoError(t, err)
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
		node.WithPortMapping(28000, 27000),
		node.WithCount(1),
		node.WithEnv([]string{
			"NIMONA_ALIAS=nimona-e2e-bootstrap",
			"BIND_PRIVATE=true",
			"DEBUG_BLOCKS=true",
			"LOG_LEVEL=debug",
			"NIMONA_API_HOST=0.0.0.0",
			"NIMONA_API_PORT=28000",
			"NIMONA_PEER_BOOTSTRAP_ADDRESSES=",
			"XNODE=NODE-BOOTSTRAP",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, bNodes)

	// stop bootstrap nodes when we are done
	defer func() {
		err := node.Stop(bNodes)
		assert.NoError(t, err)
	}()

	// wait for the containers to settle
	time.Sleep(time.Second * 15)

	// create clients for all nodes
	bClients := make([]*client.Client, len(bNodes))
	for i, node := range bNodes {
		baseURL := fmt.Sprintf("http://%s", node.Address())
		c, err := client.New(baseURL)
		require.NoError(t, err)
		bClients[i] = c
	}

	// gather bootstrap addresses
	bootstrapAddresses := make([]string, len(bClients))
	for i, bClient := range bClients {
		res, err := bClient.Info()
		require.NoError(t, err)
		require.NotNil(t, res)
		bootstrapAddresses[i] = res.Addresses[0]
	}

	// setup normal nodes
	nodes := make([]*node.Node, 3)
	for i := range nodes {
		ns, err := node.New(
			dockerImage,
			env,
			node.WithName(fmt.Sprintf("nimona-e2e-%d", i)),
			node.WithPortMapping(28000, 28000+i),
			node.WithCount(1),
			node.WithEnv([]string{
				"BIND_PRIVATE=true",
				"DEBUG_BLOCKS=true",
				"LOG_LEVEL=debug",
				"NIMONA_API_HOST=0.0.0.0",
				"NIMONA_API_PORT=28000",
				fmt.Sprintf("NIMONA_ALIAS=nimona-e2e-node-%d", i),
				fmt.Sprintf("XNODE=NODE%d", i),
				"NIMONA_PEER_BOOTSTRAP_ADDRESSES=" +
					strings.Join(bootstrapAddresses, ","),
			}),
		)
		require.NoError(t, err)
		require.NotNil(t, nodes)
		nodes[i] = ns[0]
	}

	// stop normal nodes when we are done
	defer func() {
		err := node.Stop(nodes)
		assert.NoError(t, err)
	}()

	// wait for the containers to settle
	time.Sleep(time.Second * 30)

	// create clients for all nodes
	clients := make([]*client.Client, len(nodes))
	for i, node := range nodes {
		baseURL := fmt.Sprintf("http://%s", node.Address())
		c, err := client.New(baseURL)
		require.NoError(t, err)
		clients[i] = c
	}

	// gather recipients to send the obj to
	recipients := []string{}
	// skip first client as it's the one we'll be sending from
	for _, c := range clients[1:] {
		info, err := c.Info()
		assert.NoError(t, err)
		recipients = append(recipients, info.Fingerprint)
	}

	fmt.Printf("\n\n\n")

	bInfo, err := bClients[0].Info()
	require.NoError(t, err)
	fmt.Printf("> bootstrap 0 - %s - %v\n", bInfo.Fingerprint, bInfo.Addresses)

	for i, c := range clients {
		info, err := c.Info()
		require.NoError(t, err)
		fmt.Printf("> peer %d - %s - %v\n", i, info.Fingerprint, info.Addresses)
	}

	fmt.Printf("\n\n\n")

	// create an obj, and attach recipients to policy
	nonce := rand.String(24) + "xnonce"
	streamCreated := fixtures.TestStream{
		Header: object.Header{
			Policy: object.Policy{
				Subjects:  recipients,
				Resources: []string{"*"},
				Actions:   []string{"read"},
				Effect:    "allow",
			},
		},
		Nonce: nonce,
	}

	err = clients[0].PostObject(streamCreated.ToObject())
	require.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(len(recipients))

	for _, nd := range nodes[1:] {
		go func(nd *node.Node) {
			loch, _ := nd.Logs()
			for {
				select {
				case ll := <-loch:
					if strings.Contains(ll, nonce) {
						fmt.Println("Node logged nonce")
						wg.Done()
						return
					}
				case <-time.After(time.Second * 20):
					t.Log("node didn't get obj in time")
					t.Fail()
					wg.Done()
					return
				}
			}
		}(nd)
	}

	for _, nd := range append(bNodes, nodes...) {
		go func(nd *node.Node) {
			loch, _ := nd.Logs()
			for ll := range loch {
				fmt.Println(ll)
			}
		}(nd)
	}

	wg.Wait()
}
