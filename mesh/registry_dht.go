package mesh

// type DHT struct {
// 	dht *dht.DHT

// 	registry map[string]*peerInfo
// }

// func (r *DHT) Get(ctx context.Context, peerID string) (*PeerInfo, error) {
// 	if exPeerInfo, ok := r.registry[peerID]; ok {
// 		return exPeerInfo.PeerInfo(), nil
// 	}

// 	epi, err := r.fetch(peerID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return epi.PeerInfo(), nil
// }

// func (r *DHT) fetch(peerID string) (*peerInfo, error) {
// 	ctx := context.Background()
// 	labels := map[string]string{
// 		"peer_id": peerID,
// 	}
// 	ch, err := r.dht.Filter(ctx, peerID, labels)
// 	if err != nil {
// 		return nil, err
// 	}

// 	exPeerInfo := &peer{
// 		ID:      peerID,
// 		Records: []net.Record{},
// 	}
// 	for record := range ch {
// 		exPeerInfo.Records = append(exPeerInfo.Records, record)
// 	}
// 	return exPeerInfo, nil
// }

// ---

// func (r *DHT) Resolve(ctx context.Context, peerID string) (string, error) {
// 	ch, err := r.dht.Get(ctx, peerID)
// 	if err != nil {
// 		return "", err
// 	}

// 	peerAddr := <-ch
// 	return peerAddr.GetValue(), nil
// }

// func (r *DHT) Discover(ctx context.Context, peerID, protocol string) ([]*net, error) {
// 	labels := map[string]string{
// 		"peer_id": peerID,
// 	}
// 	if protocol != "" {
// 		labels["protocol"] = protocol
// 	}
// 	ch, err := r.dht.Filter(ctx, peerID, labels)
// 	if err != nil {
// 		return nil, err
// 	}

// 	addrs := []*net.Address{}
// 	for peerAddr := range <-ch {

// 		addr := net.NewAddress(peerAddr)
// 	}
// 	return addrs, nil
// }
