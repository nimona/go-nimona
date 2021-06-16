package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"nimona.io/internal/net"
	"nimona.io/pkg/blob"
	"nimona.io/pkg/chore"
	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/version"
)

type fileTransfer struct {
	local         localpeer.LocalPeer
	objectmanager objectmanager.ObjectManager
	objectstore   objectstore.Store
	blobmanager   blob.Manager
	resolver      resolver.Resolver
	listener      net.Listener
	config        *comboConf
}

type Config struct {
	ReceivedFolder string `envconfig:"RECEIVED_FOLDER" default:"received_files"`
}

type comboConf struct {
	fconf *Config
	nconf *config.Config
}

type fileUnloaded struct {
	Metadata object.Metadata `nimona:"@metadata:m,omitempty"`
	BlobHash chore.Hash      `nimona:"blob:r,omitempty"`
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

	cfg := &Config{}
	ncfg, err := config.New(
		config.WithExtraConfig("FILES", cfg),
	)

	cconf := &comboConf{
		fconf: cfg,
		nconf: ncfg,
	}
	if err != nil {
		logger.Fatal("error parsing config", log.Error(err))
	}

	ft, err := newFileTransfer(ctx, cconf, logger)
	if err != nil {
		logger.Fatal("error initializing", log.Error(err))
	}
	defer ft.close()

	command := args[1]
	param := args[2]

	switch command {
	case "get":
		ft.get(ctx, chore.Hash(param))
	case "serve":
		ft.serve(ctx, param)
	default:
		fmt.Println("command not supported")
		return
	}
}

func (ft *fileTransfer) serve(
	ctx context.Context,
	filePath string,
) {
	fileName := filepath.Base(filePath)

	start := time.Now()

	blobUnl, err := ft.blobmanager.ImportFromFile(ctx, filePath)
	if err != nil {
		fmt.Println("failed to import blob", err)
		return
	}

	blobUnlo, err := blobUnl.MarshalObject()
	if err != nil {
		fmt.Println("failed to marshal blob", err)
		return
	}

	fl := &File{
		Name: fileName,
		Blob: blobUnlo.Hash(),
	}

	flo, err := fl.MarshalObject()
	if err != nil {
		fmt.Println("failed to marshal file", err)
		return
	}

	if _, err := ft.objectmanager.Put(ctx, flo); err != nil {
		fmt.Println("failed to store blob", err)
		return
	}
	fmt.Println(">> imported in", time.Now().Sub(start))
	fmt.Println(">> blob hash:", blobUnlo.Hash())
	fmt.Println(">> file hash:", flo.Hash())

	// os.Exit(1)
	// register for termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// and wait for one
	<-sigs
}

func (ft *fileTransfer) findAndRequest(
	ctx context.Context,
	hash chore.Hash,
) (
	*object.Object,
	error,
) {
	peers, err := ft.resolver.Lookup(ctx, resolver.LookupByHash(hash))
	if err != nil {
		return nil, err
	}

	if len(peers) == 0 {
		return nil, errors.Error("no providers found")
	}

	obj, err := ft.objectmanager.Request(ctx, hash, peers[0])
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (ft *fileTransfer) get(
	ctx context.Context,
	hash chore.Hash,
) {
	fmt.Println("getting file:", hash)

	obj, err := ft.findAndRequest(ctx, hash)
	if err != nil {
		fmt.Println("failed to request file: ", err)
		return
	}

	fl := &File{}
	if err := fl.UnmarshalObject(obj); err != nil {
		fmt.Println("object not of type file: ", err)
		return
	}

	flun := &fileUnloaded{
		Metadata: obj.Metadata,
		BlobHash: chore.Hash(obj.Data["blob"].(chore.String)),
	}

	fmt.Println("getting blob:", flun.BlobHash)
	_, ch, err := ft.blobmanager.Request(ctx, flun.BlobHash)
	if err != nil {
		fmt.Println("failed to request file:", err)
	}

	_ = os.MkdirAll(ft.config.fconf.ReceivedFolder, os.ModePerm)
	f, err := os.Create(filepath.Join(ft.config.fconf.ReceivedFolder, fl.Name))
	if err != nil {
		fmt.Println("failed to create file:", err)
		return
	}

	fmt.Println("writing file:", fl.Name)
	r := blob.NewReader(ch)
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
	cfg *comboConf,
	logger log.Logger,
) (*fileTransfer, error) {
	ft := &fileTransfer{}
	ft.config = cfg
	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.SetPeerKey(cfg.nconf.Peer.PrivateKey)
	ft.local = local

	// construct new network
	net := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

	if cfg.nconf.Peer.BindAddress != "" {
		// start listening
		lis, err := net.Listen(
			ctx,
			cfg.nconf.Peer.BindAddress,
			network.ListenOnLocalIPs,
			network.ListenOnPrivateIPs,
		)
		if err != nil {
			logger.Fatal("error while listening", log.Error(err))
		}
		ft.listener = lis
	}

	// make sure we have some bootstrap peers to start with
	if len(cfg.nconf.Peer.Bootstraps) == 0 {
		cfg.nconf.Peer.Bootstraps = []peer.Shorthand{
			"bahwqdag4aeqewwlutsgr7kv2iaqsrnppbdcmyykpckqn5uaqczae6fergklclea@tcps:asimov.bootstrap.nimona.io:22581",
			"bahwqdag4aeqomor45il7jjxlox7y5aj6cigawcljgsfftytwf6ulrpfqtiuzsya@tcps:egan.bootstrap.nimona.io:22581",
			"bahwqdag4aeqm5gkdk7dlbzke6wgc7rkm67cnqiv2jctfoxoo3vjmbdpjt5qi6za@tcps:sloan.bootstrap.nimona.io:22581",
		}
	}

	// convert shorthands into peers
	bootstrapPeers := []*peer.ConnectionInfo{}
	for _, s := range cfg.nconf.Peer.Bootstraps {
		bootstrapPeer, err := s.GetConnectionInfo()
		if err != nil {
			logger.Fatal("error parsing bootstrap peer", log.Error(err))
		}
		bootstrapPeers = append(bootstrapPeers, bootstrapPeer)
	}

	// add bootstrap peers as relays
	net.RegisterRelays(bootstrapPeers...)

	logger = logger.With(
		log.String("peer.publicKey", local.GetPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", net.GetAddresses()),
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

	// construct new resolver
	res := resolver.New(
		ctx,
		net,
		str,
		resolver.WithBoostrapPeers(bootstrapPeers...),
	)
	ft.resolver = res

	// construct object manager
	man := objectmanager.New(
		ctx,
		net,
		res,
		str,
	)
	ft.objectmanager = man

	// construct blob manager
	bm := blob.NewManager(
		ctx,
		blob.WithObjectManager(man),
		blob.WithResolver(res),
	)
	ft.blobmanager = bm

	return ft, nil
}
