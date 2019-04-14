package cmd

import (
	"fmt"
	"os"

	"github.com/cayleygraph/cayley"
	cayleyGraph "github.com/cayleygraph/cayley/graph"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/cayleygraph/cayley/graph/kv/bolt" // required for cayley

	"nimona.io/internal/api"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/hyperspace"
	"nimona.io/pkg/middleware/handshake"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object/aggregate"
	"nimona.io/pkg/object/dag"
	"nimona.io/pkg/object/exchange"
)

var (
	daemonPeerKey     string
	daemonIdentityKey string

	daemonPort             int
	daemonAPIPort          int
	daemonAnnounceHostname string
	daemonEnableMetrics    bool
	daemonToken            string

	daemonRelayAddresses []string

	daemonBootstrapAddresses = []string{
		// "tcps:andromeda.nimona.io:21013",
		// "tcps:borealis.nimona.io:21013",
		"tcps:cassiopeia.nimona.io:21013",
		// "tcps:draco.nimona.io:21013",
		// "tcps:eridanus.nimona.io:21013",
		// "tcps:fornax.nimona.io:21013",
		// "tcps:gemini.nimona.io:21013",
		// "tcps:hydra.nimona.io:21013",
		// "tcps:indus.nimona.io:21013",
		// "tcps:lacerta.nimona.io:21013",
		// "tcps:mensa.nimona.io:21013",
		// "tcps:norma.nimona.io:21013",
		// "tcps:orion.nimona.io:21013",
		// "tcps:pyxis.nimona.io:21013",
		// "tcps:stats.nimona.io:21013",
	}
)

// daemonStartCmd represents the daemon command
var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a peer as a daemon",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		bootstrapAddresses := viper.GetStringSlice("daemon.bootstraps")
		relayAddresses := viper.GetStringSlice("daemon.relays")

		if config.Daemon.PeerKey == nil {
			return errors.New("daemon not configured, please run 'daemon init'")
		}

		if err := os.MkdirAll(config.Daemon.ObjectPath, 0777); err != nil {
			return errors.Wrap(err, "could not create objects dir")
		}

		if len(bootstrapAddresses) > 0 {
			cmd.Println("Adding bootstrap nodes")
			for _, v := range bootstrapAddresses {
				cmd.Println("  *", v)
			}
		} else {
			cmd.Println("No bootstrap nodes provided")
		}

		dis := discovery.NewDiscoverer()

		k := config.Daemon.PeerKey

		li, err := net.NewLocalInfo(
			viper.GetString("daemon.announce_hostname"), k)
		if err != nil {
			return err
		}

		n, err := net.New(dis, li)
		if err != nil {
			return err
		}

		tcp := net.NewTCPTransport(li, relayAddresses)
		hsm := handshake.New(li, dis)

		n.AddMiddleware(hsm.Handle())
		n.AddTransport("tcps", tcp)

		ik := config.Daemon.IdentityKey
		if ik != nil {
			if config.Daemon.Mandate == nil {
				return errors.New("missing mandate for identity")
			}
			if err := li.AttachMandate(config.Daemon.Mandate); err != nil {
				return errors.Wrap(err, "could not attach mandate to network")
			}
		}

		err = cayleyGraph.InitQuadStore("bolt", config.Daemon.ObjectPath, nil)
		if err != nil {
			return errors.Wrap(err, "could not init quad store")
		}

		gs, err := cayley.NewGraph("bolt", config.Daemon.ObjectPath, nil)
		if err != nil {
			return errors.Wrap(err, "could not init graph store")
		}

		dpr := graph.NewCayley(gs)

		bind := fmt.Sprintf("0.0.0.0:%d", viper.GetInt("daemon.port"))
		x, err := exchange.New(k, n, dpr, dis, li, bind)
		if err != nil {
			return err
		}

		dag, err := dag.New(dpr, x, nil)
		if err != nil {
			return err
		}

		agg, err := aggregate.New(dpr, x, dag)
		if err != nil {
			return err
		}

		hsr, err := hyperspace.NewDiscoverer(k, n, x, li, bootstrapAddresses)
		if err != nil {
			return err
		}

		// _ = telemetry.NewTelemetry(x, k, "tcps:stats.nimona.io:21013")

		if err := dis.AddProvider(hsr); err != nil {
			return err
		}

		cmd.Println("Started daemon")
		cmd.Println("* Peer private key hash:\n  *", k.HashBase58())
		cmd.Println("* Peer public key hash:\n  *", k.GetPublicKey().HashBase58())
		if ik != nil {
			cmd.Println("* Identity private key hash:\n  *", ik.HashBase58())
			cmd.Println("* Identity public key hash:\n  *", ik.GetPublicKey().HashBase58())
		}
		peerAddresses := li.GetPeerInfo().Addresses
		cmd.Println("* Peer addresses:")
		if len(peerAddresses) > 0 {
			for _, addr := range peerAddresses {
				cmd.Println("  *", addr)
			}
		} else {
			cmd.Println("  * No addresses available")
		}

		apiServer := api.New(
			k,
			n,
			x,
			li,
			dpr,
			dag,
			agg,
			Version,
			Commit,
			Date,
			viper.GetString("daemon.token"),
		)

		apiPort := viper.GetInt("daemon.api_port")
		cmd.Printf("* HTTP API address:\n  * http://localhost:%d\n", apiPort)
		return apiServer.Serve(fmt.Sprintf("0.0.0.0:%d", apiPort))
	},
}

func init() {
	daemonCmd.AddCommand(daemonStartCmd)

	daemonStartCmd.PersistentFlags().IntVarP(
		&daemonPort,
		"port",
		"p",
		0,
		"peer port",
	)
	_ = viper.BindPFlag(
		"daemon.port",
		daemonStartCmd.PersistentFlags().Lookup("port"),
	)

	daemonStartCmd.PersistentFlags().StringVarP(
		&daemonToken,
		"token",
		"t",
		daemonToken,
		"daemon token",
	)
	_ = viper.BindPFlag(
		"daemon.token",
		daemonStartCmd.PersistentFlags().Lookup("token"),
	)

	daemonStartCmd.PersistentFlags().StringVar(
		&daemonAnnounceHostname,
		"announce-hostname",
		"",
		"set and announce local dns address",
	)
	_ = viper.BindPFlag(
		"daemon.announce_hostname",
		daemonStartCmd.PersistentFlags().Lookup("announce-hostname"),
	)

	daemonStartCmd.PersistentFlags().IntVar(
		&daemonAPIPort,
		"api-port",
		8030,
		"api port",
	)
	_ = viper.BindPFlag(
		"daemon.api_port",
		daemonStartCmd.PersistentFlags().Lookup("api-port"),
	)

	daemonStartCmd.PersistentFlags().BoolVarP(
		&daemonEnableMetrics,
		"metrics",
		"m",
		false,
		"enable sending anonymous metrics",
	)
	_ = viper.BindPFlag(
		"daemon.metrics",
		daemonStartCmd.PersistentFlags().Lookup("metrics"),
	)

	daemonStartCmd.PersistentFlags().StringSliceVar(
		&daemonBootstrapAddresses,
		"bootstraps",
		daemonBootstrapAddresses,
		"bootstrap addresses",
	)
	_ = viper.BindPFlag(
		"daemon.bootstraps",
		daemonStartCmd.PersistentFlags().Lookup("bootstraps"),
	)

	daemonStartCmd.PersistentFlags().StringSliceVar(
		&daemonRelayAddresses,
		"relays",
		daemonRelayAddresses,
		"relay addresses",
	)
	_ = viper.BindPFlag(
		"daemon.relays",
		daemonStartCmd.PersistentFlags().Lookup("relays"),
	)
}
