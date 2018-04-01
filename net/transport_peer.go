package net

// import (
// 	"context"
// )

// // NewPeerTransport returns a new Peer transport
// func NewPeerTransport(nn Net, dn DHT, peerID string) Transport {
// 	return &PeerTransport{
// 		peerID: peerID,
// 		net:    nn,
// 		dht:    dn,
// 	}
// }

// // PeerTransport transport
// type PeerTransport struct {
// 	peerID string
// 	dht    DHT
// 	net    Net
// }

// // DialContext attemps to dial to the peer with the given addr
// func (t *PeerTransport) DialContext(ctx context.Context, addr *Address) (context.Context, Conn, error) {
// 	pcaddr, err := t.dht.Filter(ctx, addr.CurrentParams(), map[string]string{
// 		"protocol": "peer",
// 	})
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	paddr := <-pcaddr

// 	// TODO loop addresses
// 	addr.Pop()
// 	naddr := paddr.GetValue() + "/" + addr.RemainingString()
// 	return t.net.DialContext(ctx, naddr)
// }

// // CanDial checks if address can be dialed by this transport
// func (t *PeerTransport) CanDial(addr *Address) (bool, error) {
// 	if addr.CurrentProtocol() != "peer" {
// 		return false, nil
// 	}

// 	return true, nil
// }

// // Listen handles the transports
// func (t *PeerTransport) Listen(ctx context.Context, handler HandlerFunc) error {
// 	// logger := Logger(ctx)
// 	// labels := map[string]string{
// 	// 	"protocol": "peer",
// 	// }
// 	// go func() {
// 	// 	for {
// 	// 		<-time.After(15 * time.Second)
// 	// 		addrs := t.net.GetAddresses()
// 	// 		// logger.Debug("Updating addresses", zap.Strings("addresses", addrs))
// 	// 		for _, addr := range addrs {
// 	// 			t.dht.Put(context.Background(), t.peerID, addr, labels)
// 	// 		}
// 	// 	}
// 	// }()
// 	return nil
// }

// // Addresses returns the addresses the transport is listening to
// func (t *PeerTransport) GetAddresses() []string {
// 	// TODO return peer address
// 	return []string{
// 		"peer:" + t.peerID,
// 	}
// }
