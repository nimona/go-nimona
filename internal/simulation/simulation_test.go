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
// bNode0 prv bagacnag4afabniaa2p4s2uqgd5yonz3jgmaa6xdhcmior27g7arawskmizawinxoozfmzvcmsqygfqpowo2o6pw72f54vjw4ei7lb6txcmilxwn76q
// bNode0 pub bahwqdag4aeqo45skztkezfbqmla65m5u547n7ul3zktnyir6wd5hoeyqxpm375a
// bNode1 prv bagacnag4afadjifs2n3ol7npnmv7ipddig4sfm4dtpfhgud2e7q2e4q64s3kxdxu7mlszjtgruxmhbzmobimlsrcvigs7bbpfrdu7tcplvzutditsi
// bNode1 pub bahwqdag4aeqpj6yxfstgndjoyodsy4cqyxfcfkqnf6cc6lchj7ge6xltjggrheq
// bNode2 prv bagacnag4afag2oulialyc5cad6iy5eztg4tnu4apeqoocemp4t3lw6orbxh3zvhzh47kjakhohbekc6gybt5kauwafffanr65mpvn6gb6kjnkifebu
// bNode2 pub bahwqdag4aeqpspz6usauo4ociuf4nqdh2ubjmakkka3d52y7k34md4us2uqkidi

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
			"NIMONA_PEER_PRIVATE_KEY=bagacnag4afabniaa2p4s2uqgd5yonz3jgmaa6xdhcmior27g7arawskmizawinxoozfmzvcmsqygfqpowo2o6pw72f54vjw4ei7lb6txcmilxwn76q",
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
			"NIMONA_PEER_PRIVATE_KEY=bagacnag4afadjifs2n3ol7npnmv7ipddig4sfm4dtpfhgud2e7q2e4q64s3kxdxu7mlszjtgruxmhbzmobimlsrcvigs7bbpfrdu7tcplvzutditsi",
			"NIMONA_PEER_BOOTSTRAPS=bahwqdag4aeqo45skztkezfbqmla65m5u547n7ul3zktnyir6wd5hoeyqxpm375a@tcps:" + bNodes[0].Address(),
			"NIMONA_SONAR_PING_PEERS=bahwqdag4aeqpspz6usauo4ociuf4nqdh2ubjmakkka3d52y7k34md4us2uqkidi",
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
			"NIMONA_PEER_PRIVATE_KEY=bagacnag4afag2oulialyc5cad6iy5eztg4tnu4apeqoocemp4t3lw6orbxh3zvhzh47kjakhohbekc6gybt5kauwafffanr65mpvn6gb6kjnkifebu",
			"NIMONA_PEER_BOOTSTRAPS=bahwqdag4aeqo45skztkezfbqmla65m5u547n7ul3zktnyir6wd5hoeyqxpm375a@tcps:" + bNodes[0].Address(),
			"NIMONA_SONAR_PING_PEERS=bahwqdag4aeqpj6yxfstgndjoyodsy4cqyxfcfkqnf6cc6lchj7ge6xltjggrheq",
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
