package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	olog "log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"nimona.io/internal/net"
	"nimona.io/internal/version"
	"nimona.io/pkg/blob"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/sqlobjectstore"

	_ "github.com/arl/statsviz"
)

const (
	peerLookupTime = 5
)

type fileTransfer struct {
	local         localpeer.LocalPeer
	objectmanager objectmanager.ObjectManager
	objectstore   objectstore.Store
	resolver      resolver.Resolver
	listener      net.Listener
	config        *config
}

// nolint: lll
type config struct {
	Peer struct {
		PrivateKey  crypto.PrivateKey `envconfig:"PRIVATE_KEY"`
		BindAddress string            `envconfig:"BIND_ADDRESS" default:"0.0.0.0:0"`
		Bootstraps  []peer.Shorthand  `envconfig:"BOOTSTRAPS"`
	} `envconfig:"PEER"`

	ReceivedFolder string `envconfig:"RECEIVED_FOLDER" default:"received_files"`
	Debug          struct {
		MetricsPort string `envconfig:"METRICS_PORT"`
	} `envconfig:"DEBUG"`
}

type fileUnloaded struct {
	Metadata object.Metadata `nimona:"metadata:m,omitempty"`
	BlobHash object.Hash     `nimona:"blob:r,omitempty"`
}

func (e *fileUnloaded) Type() string {
	return "nimona.io/File"
}

func main() {
	args := os.Args
	if len(args) != 3 {
		fmt.Printf("usage: %s <get/serve> <hash/file>\n", args[0])
		return
	}

	ctx := context.New(
		context.WithCorrelationID("nimona"),
	)
	logger := log.FromContext(ctx).With(
		log.String("build.version", version.Version),
		log.String("build.commit", version.Commit),
		log.String("build.timestamp", version.Date),
	)

	cfg := &config{}
	if err := envconfig.Process("nimona", cfg); err != nil {
		logger.Fatal("error processing config", log.Error(err))
	}

	go func() {
		if cfg.Debug.MetricsPort == "" {
			return
		}

		olog.Println(
			http.ListenAndServe(
				fmt.Sprintf("localhost:%s", cfg.Debug.MetricsPort),
				nil,
			))
	}()

	ft, err := newFileTransfer(ctx, cfg, logger)
	if err != nil {
		logger.Fatal("error initializing", log.Error(err))
	}
	defer ft.close()

	command := args[1]
	param := args[2]

	switch command {
	case "get":
		ft.get(ctx, object.Hash(param))
	case "serve":
		ft.serve(ctx, param)
	default:
		fmt.Println("command not supported")
		return
	}

}

func (ft *fileTransfer) serve(
	ctx context.Context,
	filename string,
) {

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("failed to open file:", err)
		return
	}

	bl, err := blob.ToBlob(f)
	if err != nil {
		fmt.Println("failed to covert to blob", err)
		return
	}

	fl := &File{}
	fl.Name = filename
	fl.Blob = bl

	_, err = ft.objectmanager.Put(ctx, fl.ToObject())
	if err != nil {
		fmt.Println("failed to store file", err)
		return
	}

	blobj, err := ft.objectmanager.Put(ctx, bl.ToObject())
	if err != nil {
		fmt.Println("failed to store blob", err)
		return
	}
	fmt.Println("blob hash:", blobj.Hash())
	fmt.Println("file hash:", fl.ToObject().Hash().String())

	// register for termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// and wait for one
	<-sigs
}

func (ft *fileTransfer) findAndRequest(
	ctx context.Context,
	hash object.Hash,
) (
	*object.Object,
	error,
) {
	peersCh, err := ft.resolver.Lookup(ctx, resolver.LookupByContentHash(hash))
	if err != nil {
		return nil, err
	}

	peerFound := &peer.Peer{}

	select {
	case peerFound = <-peersCh:
	case <-ctx.Done():
		return nil, errors.New("context")
	case <-time.After(peerLookupTime * time.Second):
		break
	}

	obj, err := ft.objectmanager.Request(ctx, hash, peerFound, true)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (ft *fileTransfer) get(
	ctx context.Context,
	hash object.Hash,
) {
	fmt.Println("getting file:", hash)

	obj, err := ft.findAndRequest(ctx, hash)
	if err != nil {
		fmt.Println("failed to request file: ", err)
		return
	}

	fl := &File{}
	if err := fl.FromObject(obj); err != nil {
		fmt.Println("object not of type file: ", err)
		return
	}

	flun := &fileUnloaded{}

	err = object.Decode(obj, flun)
	bmgr := blob.NewRequester(
		ctx,
		blob.WithObjectManager(ft.objectmanager),
		blob.WithResolver(ft.resolver),
	)
	if err != nil {
		fmt.Println("failed to decode:", err)
		return
	}

	fmt.Println("getting blob:", flun.BlobHash)
	bl, err := bmgr.Request(ctx, flun.BlobHash)
	if err != nil {
		fmt.Println("failed to request file:", err)
	}

	_ = os.MkdirAll(ft.config.ReceivedFolder, os.ModePerm)
	f, err := os.Create(filepath.Join(ft.config.ReceivedFolder, fl.Name))
	if err != nil {
		fmt.Println("failed to create file:", err)
		return
	}

	fmt.Println("writing file:", fl.Name)
	r := blob.FromBlob(bl)
	bf := bufio.NewReader(r)
	if _, err := io.Copy(f, bf); err != nil {
		fmt.Println("failed to write to file:", err)
	}

	fmt.Println("done")

}

func (ft *fileTransfer) close() {
	if ft.listener != nil {
		ft.listener.Close()
	}
}
func newFileTransfer(
	ctx context.Context,
	cfg *config,
	logger log.Logger,
) (*fileTransfer, error) {
	ft := &fileTransfer{}
	ft.config = cfg
	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.PutPrimaryPeerKey(cfg.Peer.PrivateKey)
	local.PutContentTypes(
		new(File).Type(),
		new(blob.Blob).Type(),
		new(blob.Chunk).Type(),
	)
	ft.local = local

	// construct new network
	net := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

	if cfg.Peer.BindAddress != "" {
		// start listening
		lis, err := net.Listen(
			ctx,
			cfg.Peer.BindAddress,
			network.ListenOnLocalIPs,
			network.ListenOnPrivateIPs,
		)
		if err != nil {
			logger.Fatal("error while listening", log.Error(err))
		}
		ft.listener = lis
	}

	// make sure we have some bootstrap peers to start with
	if len(cfg.Peer.Bootstraps) == 0 {
		cfg.Peer.Bootstraps = []peer.Shorthand{
			"ed25519.CJi6yjjXuNBFDoYYPrp697d6RmpXeW8ZUZPmEce9AgEc@tcps:asimov.bootstrap.nimona.io:22581",
			"ed25519.6fVWVAK2DVGxBhtVBvzNWNKBWk9S83aQrAqGJfrxr75o@tcps:egan.bootstrap.nimona.io:22581",
			"ed25519.7q7YpmPNQmvSCEBWW8ENw8XV8MHzETLostJTYKeaRTcL@tcps:sloan.bootstrap.nimona.io:22581",
		}
	}

	// convert shorthands into peers
	bootstrapPeers := []*peer.Peer{}
	for _, s := range cfg.Peer.Bootstraps {
		bootstrapPeer, err := s.Peer()
		if err != nil {
			logger.Fatal("error parsing bootstrap peer", log.Error(err))
		}
		bootstrapPeers = append(bootstrapPeers, bootstrapPeer)
	}

	// construct new resolver
	res := resolver.New(
		ctx,
		net,
		resolver.WithBoostrapPeers(bootstrapPeers),
	)
	ft.resolver = res

	logger = logger.With(
		log.String("peer.privateKey", local.GetPrimaryPeerKey().String()),
		log.String("peer.publicKey", local.GetPrimaryPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", local.GetAddresses()),
	)

	// construct object store
	db, err := sql.Open("sqlite3", "file_transfer.db")
	if err != nil {
		logger.Fatal("error opening sql file", log.Error(err))
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		logger.Fatal("error starting sql store", log.Error(err))
	}
	ft.objectstore = str

	// construct manager
	man := objectmanager.New(
		ctx,
		net,
		res,
		str,
	)
	ft.objectmanager = man

	return ft, nil

}
