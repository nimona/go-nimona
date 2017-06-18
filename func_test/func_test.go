package functional_test

import (
	"testing"

	dht "github.com/nimona/go-nimona-kad-dht"
)

func setupTest() {
	// Start bootstrap node
}
func TestServerSendReceiveMessage(t *testing.T) {

	pr := &dht.Peer{dht.ID("adf"), []string{"localhost:8889"}}
	bp := []*dht.Peer{pr}
	rtd := &dht.RoutingTableSimple{}

	dhtNode := dht.NewDHTNode(bp, pr, rtd)
	dhtNode.Find()
}
