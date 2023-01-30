package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"nimona.io"
)

func main() {
	publicKey, privateKey, err := nimona.GenerateKey()
	if err != nil {
		panic(fmt.Errorf("error generating key: %w", err))
	}

	ctx, cf := context.WithTimeout(context.Background(), 5*time.Second)
	defer cf()

	transport := &nimona.TransportUTP{}
	// listener, err := transport.Listen(ctx, "0.0.0.0:1013")
	// if err != nil {
	// 	panic(fmt.Errorf("error listening: %w", err))
	// }

	// docDB, err := gorm.Open(
	// 	sqlite.Open("file::memory:?cache=shared"),
	// 	&gorm.Config{},
	// )
	// if err != nil {
	// 	panic(fmt.Errorf("error opening db: %w", err))
	// }

	// docStore, err := nimona.NewDocumentStore(docDB)
	// if err != nil {
	// 	panic(fmt.Errorf("error creating document store: %w", err))
	// }

	sesManager, err := nimona.NewSessionManager(
		transport,
		nil, // listener,
		publicKey,
		privateKey,
	)
	if err != nil {
		panic(fmt.Errorf("error creating session manager: %w", err))
	}

	// peerInfo := &nimona.PeerInfo{
	// 	PublicKey: publicKey,
	// 	Addresses: []nimona.PeerAddr{{
	// 		PublicKey: publicKey,
	// 		Address:   listener.PeerAddr().Address,
	// 		Network:   listener.PeerAddr().Network,
	// 	}},
	// }

	// peerConfig := nimona.NewPeerConfig(
	// 	privateKey,
	// 	publicKey,
	// 	peerInfo,
	// )

	// read peerAddr from args
	peerAddrString := os.Args[1]
	peerAddr, err := nimona.ParsePeerAddr(peerAddrString)
	if err != nil {
		panic(fmt.Errorf("error parsing peer addr: %w", err))
	}

	// Dial the peer
	ses, err := sesManager.Dial(ctx, *peerAddr)
	if err != nil {
		panic(fmt.Errorf("error dialing: %w", err))
	}

	fmt.Println("dialed", peerAddrString, ses.PeerAddr())

	time.Sleep(100 * time.Millisecond)

	res, err := nimona.RequestPing(ctx, ses)
	if err != nil {
		panic(fmt.Errorf("error sending ping: %w", err))
	}

	fmt.Println("response:", res)

	// pingHandler := &nimona.HandlerPing{
	// 	PeerConfig: peerConfig,
	// }

	// sesManager.RegisterHandler(
	// 	"test/ping",
	// 	pingHandler.HandlePingRequest,
	// )

	// nodeConfig := &nimona.NodeConfig{
	// 	Dialer:     transport,
	// 	Listener:   listener,
	// 	PeerConfig: peerConfig,
	// 	// DocumentStore: docStore,
	// }

	// node, err := nimona.NewNode(nodeConfig)
	// if err != nil {
	// 	panic(fmt.Errorf("error creating node: %w", err))
	// }

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprint(w, peerConfig.GetPeerInfo().Addresses[0].String())
	// })

	// defer node.Close()
	// http.ListenAndServe("0.0.0.0:80", nil)
}
