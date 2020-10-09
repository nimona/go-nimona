package main

import (
	"database/sql"
	"fmt"
	olog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

type fileTransfer struct {
	local         localpeer.LocalPeer
	objectmanager objectmanager.ObjectManager
	objectstore   objectstore.Store
	resolver      resolver.Resolver
	listener      net.Listener
}

// nolint: lll
type config struct {
	Peer struct {
		PrivateKey  crypto.PrivateKey `envconfig:"PRIVATE_KEY"`
		BindAddress string            `envconfig:"BIND_ADDRESS" default:"0.0.0.0:0"`
		Bootstraps  []peer.Shorthand  `envconfig:"BOOTSTRAPS"`
	} `envconfig:"PEER"`

	Debug struct {
		MetricsPort string `envconfig:"METRICS_PORT"`
	} `envconfig:"DEBUG"`
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
		ft.get(ctx, object.Hash(param), logger)
	case "serve":
		ft.serve(ctx, param, logger)
	default:
		fmt.Println("command not supported")
		return
	}

	// register for termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// and wait for one
	<-sigs

}

func (ft *fileTransfer) serve(
	ctx context.Context,
	filename string,
	logger log.Logger,
) {

	f, err := os.Open(filename)
	if err != nil {
		logger.Fatal("failed to open file", log.Error(err))
	}

	bl, err := blob.ToBlob(f)
	if err != nil {
		logger.Fatal("failed to covert to blob", log.Error(err))
	}

	obj, err := ft.objectmanager.Put(ctx, bl.ToObject())
	if err != nil {
		logger.Fatal("failed to store blob", log.Error(err))
	}

	logger.Info("blob sharing",
		log.String("hash", bl.ToObject().Hash().String()),
		log.String("obj_hash", obj.Hash().String()),
	)
}

func (ft *fileTransfer) get(
	ctx context.Context,
	hash object.Hash,
	logger log.Logger,
) {
	logger.Info("getting blob")
	bmgr := blob.NewRequester(
		ctx,
		blob.WithObjectManager(ft.objectmanager),
		blob.WithResolver(ft.resolver),
	)

	bl, err := bmgr.Request(ctx, hash)
	if err != nil {
		logger.Fatal("failed to request blob", log.Error(err))
	}

	logger.Info(bl.ToObject().Hash().String())
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
	files := &fileTransfer{}

	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.PutPrimaryPeerKey(cfg.Peer.PrivateKey)
	local.PutContentTypes(
		new(blob.Blob).Type(),
		new(blob.Chunk).Type(),
	)
	files.local = local

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
		files.listener = lis
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
	files.resolver = res

	logger = logger.With(
		log.String("peer.privateKey", local.GetPrimaryPeerKey().String()),
		log.String("peer.publicKey", local.GetPrimaryPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", local.GetAddresses()),
	)

	logger.Info("ready")

	// construct object store
	db, err := sql.Open("sqlite3", "file_transfer.db")
	if err != nil {
		logger.Fatal("error opening sql file", log.Error(err))
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		logger.Fatal("error starting sql store", log.Error(err))
	}
	files.objectstore = str

	// construct manager
	man := objectmanager.New(
		ctx,
		net,
		res,
		str,
	)
	files.objectmanager = man

	return files, nil

}
