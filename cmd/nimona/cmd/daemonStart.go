package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"nimona.io/go/api"
	"nimona.io/go/crypto"
	"nimona.io/go/dht"
	"nimona.io/go/net"
	"nimona.io/go/storage"
	"nimona.io/go/telemetry"
)

var (
	daemonConfigPath     string
	daemonPort           int
	daemonAPIPort        int
	daemonEnableRelaying bool
	daemonEnableMetrics  bool
	apiToken             string

	relayAddresses []string

	bootstrapAddresses = []string{
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
		if daemonConfigPath == "" {
			usr, _ := user.Current()
			daemonConfigPath = path.Join(usr.HomeDir, ".nimona")
		}

		if err := os.MkdirAll(daemonConfigPath, 0777); err != nil {
			return errors.Wrap(err, "could not create config dir")
		}

		// addressBook, err := peers.NewAddressBook(daemonConfigPath)
		// if err != nil {
		// 	return errors.Wrap(err, "could not load key")
		// }

		k, err := crypto.LoadKey(filepath.Join(daemonConfigPath, "key.cbor"))
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

		n, err := net.NewNetwork(k, announceHostname, relayAddresses)
		if err != nil {
			return err
		}

		storagePath := path.Join(daemonConfigPath, "storage")
		dpr := storage.NewDiskStorage(storagePath)
		x, err := net.NewExchange(k, n, dpr, fmt.Sprintf("0.0.0.0:%d", daemonPort))
		dht, _ := dht.NewDHT(k, n, x, bootstrapAddresses)
		telemetry.NewTelemetry(x, k, "tcps:stats.nimona.io:21013")

		if err := n.Resolver().AddProvider(dht); err != nil {
			return err
		}

		netAddress := fmt.Sprintf("0.0.0.0:%d", daemonAPIPort)
		apiAddress := fmt.Sprintf("http://localhost:%d", daemonAPIPort)

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
		cmd.Println("* HTTP API address:\n  *", apiAddress)

		a := api.New(k, n, x, dht, dpr, Version, Commit, Date, apiToken)

		go func() {
			if err := a.Serve(netAddress); err != nil {
				log.Fatal("Server stoping, error:", err)
			}
		}()

		api.Wait()

		return errors.Wrap(err, "http server stopped")
	},
}

func init() {
	daemon.AddCommand(daemonStartCmd)

	daemonStartCmd.PersistentFlags().StringVar(
		&daemonConfigPath,
		"config-path",
		"",
		"daemon config path",
	)

	daemonStartCmd.PersistentFlags().IntVar(
		&daemonPort,
		"port",
		0,
		"peer port",
	)

	daemonStartCmd.PersistentFlags().IntVar(
		&daemonAPIPort,
		"api-port",
		8030,
		"api port",
	)

	daemonStartCmd.PersistentFlags().BoolVar(
		&daemonEnableMetrics,
		"metrics",
		false,
		"enable sending anonymous metrics",
	)

	daemonStartCmd.PersistentFlags().StringSliceVar(
		&bootstrapAddresses,
		"bootstraps",
		bootstrapAddresses,
		"bootstrap addresses",
	)

	daemonStartCmd.PersistentFlags().StringSliceVar(
		&relayAddresses,
		"relay-addresses",
		relayAddresses,
		"relay addresses",
	)

	daemonStartCmd.PersistentFlags().StringVar(
		&apiToken,
		"api-token",
		apiToken,
		"api token",
	)
}
