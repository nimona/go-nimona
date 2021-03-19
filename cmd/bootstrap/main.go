package main

import (
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pyroscope-io/pyroscope/pkg/agent/profiler"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace/provider"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/version"
)

// nolint: lll
type config struct {
	Peer struct {
		PrivateKey      *crypto.PrivateKey `envconfig:"PRIVATE_KEY"`
		BindAddress     string             `envconfig:"BIND_ADDRESS" default:"0.0.0.0:0"`
		AnnounceAddress string             `envconfig:"ANNOUNCE_ADDRESS"`
		Bootstraps      []peer.Shorthand   `envconfig:"BOOTSTRAPS"`
	} `envconfig:"PEER"`
	Metrics struct {
		PyroscopeServerName string `envconfig:"PYROSCOPE_SERVER_NAME"`
		PyroscopeServerURL  string `envconfig:"PYROSCOPE_SERVER_URL"`
		BindAddress         string `envconfig:"BIND_ADDRESS" default:"0.0.0.0:0"`
	} `envconfig:"METRICS"`
}

func main() {
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

	if cfg.Metrics.PyroscopeServerURL != "" {
		applicationName := "bootstrap"
		if cfg.Metrics.PyroscopeServerName != "" {
			applicationName = cfg.Metrics.PyroscopeServerName
		}
		// nolint: errcheck
		profiler.Start(profiler.Config{
			ApplicationName: applicationName,
			ServerAddress:   cfg.Metrics.PyroscopeServerURL,
		})
	}

	if cfg.Peer.PrivateKey == nil {
		logger.Fatal("missing peer private key")
	}

	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.PutPrimaryPeerKey(cfg.Peer.PrivateKey)

	// construct new network
	net := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

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

	// add announce address
	if cfg.Peer.AnnounceAddress != "" {
		local.PutAddresses("tcps:" + cfg.Peer.AnnounceAddress)
	}

	// convert shorthands into connection infos
	bootstrapProviders := []*peer.ConnectionInfo{}
	for _, s := range cfg.Peer.Bootstraps {
		bootstrapPeer, err := s.ConnectionInfo()
		if err != nil {
			logger.Fatal("error parsing bootstrap peer", log.Error(err))
		}
		bootstrapProviders = append(bootstrapProviders, bootstrapPeer)
	}

	// construct new hyperspace provider
	_, err = provider.New(
		ctx,
		net,
		bootstrapProviders,
	)
	if err != nil {
		logger.Fatal("error while constructing provider", log.Error(err))
	}

	logger = logger.With(
		log.String("peer.publicKey", local.GetPrimaryPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", local.GetAddresses()),
	)

	logger.Info("bootstrap node ready")

	go func() {
		promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "build_info",
				Help: "Build info",
				ConstLabels: prometheus.Labels{
					"commit":     version.Commit,
					"build_date": version.Date,
					"version":    version.Version,
					"goversion": strings.Replace(
						runtime.Version(),
						"go", "v", 1,
					),
				},
			},
			func() float64 { return 1 },
		)
		logger.Info(
			"serving metrics",
			log.String("address", cfg.Metrics.BindAddress),
		)
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(cfg.Metrics.BindAddress, nil)
		if err != nil {
			logger.Warn("error serving metrics", log.Error(err))
			return
		}
	}()

	// register for termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// and wait for one
	<-sigs

	// finally terminate everything
	logger.Info("shutting down")
	lis.Close() // nolint: errcheck
}
