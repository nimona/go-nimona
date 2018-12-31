package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"nimona.io/internal/api"
	"nimona.io/internal/telemetry"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery/hyperspace"
	"nimona.io/pkg/net"
	"nimona.io/pkg/storage"
)

var (
	daemonDataDir          string
	daemonPort             int
	daemonAPIPort          int
	daemonAnnounceHostname string
	daemonEnableRelaying   bool
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
		dataDir := viper.GetString("daemon.data_dir")
		bootstrapAddresses := viper.GetStringSlice("daemon.bootstraps")
		relayAddresses := viper.GetStringSlice("daemon.relays")

		if dataDir == "" {
			usr, _ := user.Current()
			dataDir = path.Join(usr.HomeDir, ".nimona")
		}

		if err := os.MkdirAll(dataDir, 0777); err != nil {
			return errors.Wrap(err, "could not create config dir")
		}

		// addressBook, err := peer.NewAddressBook(daemonConfigPath)
		// if err != nil {
		//		return errors.Wrap(err, "could not load key")
		// }

		k, err := crypto.LoadKey(filepath.Join(dataDir, "key.cbor"))
		if err != nil {
			return errors.Wrap(err, "could not load or create peer key")
		}

		if len(bootstrapAddresses) > 0 {
			cmd.Println("Adding bootstrap nodes")
			for _, v := range bootstrapAddresses {
				cmd.Println("  *", v)
			}
		} else {
			cmd.Println("No bootstrap nodes provided")
		}

		n, err := net.New(
			k,
			viper.GetString("daemon.announce_hostname"),
			relayAddresses,
		)
		if err != nil {
			return err
		}

		storagePath := path.Join(dataDir, "storage")
		dpr := storage.NewDiskStorage(storagePath)
		bind := fmt.Sprintf("0.0.0.0:%d", viper.GetInt("daemon.port"))
		x, err := net.NewExchange(k, n, dpr, bind)
		hsr, _ := hyperspace.NewDiscoverer(k, n, x, bootstrapAddresses)
		telemetry.NewTelemetry(x, k, "tcps:stats.nimona.io:21013")

		if err := n.Discoverer().AddProvider(hsr); err != nil {
			return err
		}

		cmd.Println("Started daemon")
		cmd.Println("* Peer private key hash:\n  *", k.HashBase58())
		cmd.Println("* Peer public key hash:\n  *", k.GetPublicKey().HashBase58())
		peerAddresses := n.GetPeerInfo().Addresses
		cmd.Println("* Peer addresses:")
		if len(peerAddresses) > 0 {
			for _, addr := range peerAddresses {
				cmd.Println("  *", addr)
			}
		} else {
			cmd.Println("  * No addresses available")
		}

		apiServer := api.New(
			k, n, x, dpr,
			Version, Commit, Date,
			viper.GetString("daemon.token"),
		)

		apiPort := viper.GetInt("daemon.api_port")
		cmd.Printf("* HTTP API address:\n  * http://localhost:%d\n", apiPort)
		return apiServer.Serve(fmt.Sprintf("0.0.0.0:%d", apiPort))
	},
}

func init() {
	daemonCmd.AddCommand(daemonStartCmd)

	daemonStartCmd.PersistentFlags().StringVarP(
		&daemonDataDir,
		"data-dir",
		"d",
		"",
		"daemon data directory",
	)
	viper.BindPFlag(
		"daemon.data_dir",
		daemonStartCmd.PersistentFlags().Lookup("data-dir"),
	)

	daemonStartCmd.PersistentFlags().IntVarP(
		&daemonPort,
		"port",
		"p",
		0,
		"peer port",
	)
	viper.BindPFlag(
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
	viper.BindPFlag(
		"daemon.token",
		daemonStartCmd.PersistentFlags().Lookup("token"),
	)

	daemonStartCmd.PersistentFlags().StringVar(
		&daemonAnnounceHostname,
		"announce-hostname",
		"",
		"set and announce local dns address",
	)
	viper.BindPFlag(
		"daemon.announce_hostname",
		daemonStartCmd.PersistentFlags().Lookup("announce-hostname"),
	)

	daemonStartCmd.PersistentFlags().IntVar(
		&daemonAPIPort,
		"api-port",
		8030,
		"api port",
	)
	viper.BindPFlag(
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
	viper.BindPFlag(
		"daemon.metrics",
		daemonStartCmd.PersistentFlags().Lookup("metrics"),
	)

	daemonStartCmd.PersistentFlags().StringSliceVar(
		&daemonBootstrapAddresses,
		"bootstraps",
		daemonBootstrapAddresses,
		"bootstrap addresses",
	)
	viper.BindPFlag(
		"daemon.bootstraps",
		daemonStartCmd.PersistentFlags().Lookup("bootstraps"),
	)

	daemonStartCmd.PersistentFlags().StringSliceVar(
		&daemonRelayAddresses,
		"relays",
		daemonRelayAddresses,
		"relay addresses",
	)
	viper.BindPFlag(
		"daemon.relays",
		daemonStartCmd.PersistentFlags().Lookup("relays"),
	)
}
