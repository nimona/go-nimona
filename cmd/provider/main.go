package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"nimona.io"
)

type Config struct {
	StoreSqliteDSN      string `envconfig:"STORE_SQLITE_DSN" default:"file::memory:?cache=shared"`
	PeerPublicKey       string `envconfig:"PEER_PUBLIC_KEY"`
	PeerPrivateKey      string `envconfig:"PEER_PRIVATE_KEY"`
	ListenUTPPort       string `envconfig:"LISTEN_UTP_PORT" default:"0"`
	ListenUTPPublicHost string `envconfig:"LISTEN_UTP_PUBLIC_HOST"`
	ListenHTTPPort      string `envconfig:"LISTEN_HTTP_PORT" default:"8080"`
}

// example usage.
// server: STORE_SQLITE_DSN=provider.sqlite LISTEN_PORT=1013 go run main.go
// client: go run main.go nimona://peer:addr:utp:127.0.0.1:1013

func main() {
	cfg := &Config{}
	err := envconfig.Process("nimona", cfg)
	if err != nil {
		panic(fmt.Errorf("error parsing config: %w", err))
	}

	var publicKey nimona.PublicKey
	var privateKey nimona.PrivateKey

	if cfg.PeerPublicKey == "" || cfg.PeerPrivateKey == "" {
		publicKey, privateKey, err = nimona.GenerateKey()
		if err != nil {
			panic(fmt.Errorf("error generating key: %w", err))
		}
	} else {
		publicKey, err = nimona.ParsePublicKey(cfg.PeerPublicKey)
		if err != nil {
			panic(fmt.Errorf("error parsing public key: %w", err))
		}
		privateKey, err = nimona.ParsePrivateKey(cfg.PeerPrivateKey)
		if err != nil {
			panic(fmt.Errorf("error parsing private key: %w", err))
		}
	}

	ctx := context.Background()
	transport := &nimona.TransportUTP{
		PublicAddress: cfg.ListenUTPPublicHost,
		PublicKey:     publicKey,
	}
	listener, err := transport.Listen(ctx, "0.0.0.0:"+cfg.ListenUTPPort)
	if err != nil {
		panic(fmt.Errorf("error listening: %w", err))
	}

	docDB, err := gorm.Open(
		sqlite.Open(cfg.StoreSqliteDSN),
		&gorm.Config{},
	)
	if err != nil {
		panic(fmt.Errorf("error opening db: %w", err))
	}

	docStore, err := nimona.NewDocumentStore(docDB)
	if err != nil {
		panic(fmt.Errorf("error creating document store: %w", err))
	}

	sesManager, err := nimona.NewSessionManager(
		transport,
		listener,
		publicKey,
		privateKey,
	)
	if err != nil {
		panic(fmt.Errorf("error creating session manager: %w", err))
	}

	nimona.HandlePingRequest(sesManager)
	nimona.HandleDocumentGraphRequest(sesManager, docStore)
	nimona.HandleDocumentRequest(sesManager, docStore)
	nimona.HandleDocumentStoreRequest(sesManager, docStore)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sesManager.PeerAddr().String())
	})

	// if we have a peer addr in args, dial it instead of listening
	if len(os.Args) > 1 {
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

		fmt.Println("ping response:", res.Nonce)
		return
	}

	fmt.Println("Listening on:")
	fmt.Println("-", sesManager.PeerAddr())
	fmt.Println("- http://0.0.0.0:" + cfg.ListenHTTPPort)

	http.ListenAndServe("0.0.0.0:"+cfg.ListenHTTPPort, nil) // nolint: errcheck
}
