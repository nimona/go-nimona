package functionaltests

import (
	"testing"

	dht "github.com/nimona/go-nimona-kad-dht"
	log "github.com/sirupsen/logrus"
)

func setupTest() (*dht.DHTNode, *dht.DHTNode, *dht.DHTNode) {
	// Start bootstrap node
	peer1 := &dht.Peer{dht.ID("a1"), []string{"127.0.0.1:8889"}}
	peer2 := &dht.Peer{dht.ID("a2"), []string{"127.0.0.1:8890"}}
	peer3 := &dht.Peer{dht.ID("a3"), []string{"127.0.0.1:8891"}}

	rt1 := dht.NewSimpleRoutingTable()
	rt2 := dht.NewSimpleRoutingTable()
	rt3 := dht.NewSimpleRoutingTable()

	rt1.Add(*peer2)
	rt2.Add(*peer1)
	rt3.Add(*peer2)

	node1 := dht.NewDHTNode([]*dht.Peer{peer2}, peer1, rt1, "127.0.0.1:8889")
	node2 := dht.NewDHTNode([]*dht.Peer{peer1}, peer2, rt2, "127.0.0.1:8890")
	node3 := dht.NewDHTNode([]*dht.Peer{peer2}, peer3, rt3, "127.0.0.1:8891")

	return node1, node2, node3
}

// TODO: Create a node factory

func TestServerSendReceiveMessage(t *testing.T) {
	n1, _, _ := setupTest()
	peers, err := n1.Find("a3")
	if err != nil {
		log.Error(err)
		t.Fail()
	}
	if peers[0].ID == "" {
		t.Fail()
	}

	log.Info(peers)
}
