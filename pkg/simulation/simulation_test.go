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

	"nimona.io/internal/rand"
	"nimona.io/pkg/client"
	"nimona.io/pkg/simulation/node"
	"nimona.io/pkg/stream"
)

func TestSimulation(t *testing.T) {
	// Setup
	env, err := node.NewEnvironment()
	require.NoError(t, err)
	require.NotNil(t, env)

	dockerImage := "nimona:dev"
	if img := os.Getenv("E2E_DOCKER_IMAGE"); img != "" {
		dockerImage = img
	}

	// Setup Nodes
	nodes, err := node.New(
		dockerImage,
		env,
		node.WithName("nimona-e2e-"+rand.String(8)),
		node.WithNodePort(28000),
		node.WithCount(3),
		node.WithEnv([]string{
			"BIND_PRIVATE=true",
			"DEBUG_BLOCKS=true",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, nodes)

	// wait for the containers to settle
	time.Sleep(time.Second * 5)

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

	// create an obj, and attach recipients to policy
	nonce := rand.String(24)
	streamCreated := stream.Created{
		Nonce: nonce,
		Policies: []*stream.Policy{
			&stream.Policy{
				Subjects:  recipients,
				Resources: []string{"*"},
				Action:    "allow",
			},
		},
	}

	err = clients[0].PostObject(streamCreated.ToObject())
	assert.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(len(recipients))

	for _, nd := range nodes[1:] {
		go func(nd *node.Node) {
			loch, _ := nd.Logs()
			for {
				select {
				case ll := <-loch:
					fmt.Println(ll)
					if strings.Contains(ll, nonce) {
						fmt.Println("Node received object")
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

	wg.Wait()

	err = node.Stop(nodes)
	assert.NoError(t, err)

	err = env.Stop()
	assert.NoError(t, err)
}
