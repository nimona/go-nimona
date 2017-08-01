package dht

// func TestFindPeersNear(t *testing.T) {
// 	peer1 := net.Peer{ID: "a1", Addresses: []string{"127.0.0.1:8889"}}
// 	peer2 := net.Peer{ID: "a2", Addresses: []string{"127.0.0.1:8890"}}
// 	peer3 := net.Peer{ID: "a3", Addresses: []string{"127.0.0.1:8891"}}
// 	peer4 := net.Peer{ID: "a4", Addresses: []string{"127.0.0.1:8821"}}
// 	peer5 := net.Peer{ID: "a5", Addresses: []string{"127.0.0.1:8841"}}
// 	peer6 := net.Peer{ID: "a6", Addresses: []string{"127.0.0.1:8861"}}

// 	rt := NewSimpleRoutingTable()
// 	rt.Add(peer2)
// 	rt.Add(peer3)
// 	rt.Add(peer4)
// 	rt.Add(peer5)
// 	rt.Add(peer6)

// 	nnet, _ := net.NewTCPNetwork(&peer1)

// 	node, _ := NewDHTNode([]net.Peer{peer2}, peer1, rt, nnet)
// 	peers, err := node.findPeersNear(peer6.ID, 3)
// 	if err != nil {
// 		fmt.Println(err)
// 		t.Fail()
// 	}

// 	if len(peers) == 0 {
// 		t.Fail()
// 	}

// 	if peers[0].ID != peer6.ID {
// 		t.Fail()
// 	}
// }
