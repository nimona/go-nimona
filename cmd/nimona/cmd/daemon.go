package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"nimona.io/go/api"
	"nimona.io/go/dht"
	"nimona.io/go/net"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
	"nimona.io/go/storage"
	"nimona.io/go/telemetry"
)

var (
	daemonConfigPath     string
	daemonPort           int
	daemonAPIPort        int
	daemonEnableRelaying bool
	daemonEnableMetrics  bool
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
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

		addressBook, err := peers.NewAddressBook(daemonConfigPath)
		if err != nil {
			return errors.Wrap(err, "could not load key")
		}

		statsBootstrapPeer := &peers.PeerInfo{}
		for _, bootstrapPeer := range bootstrapPeerInfos {
			peerInfoBlock, err := primitives.BlockFromBase58(bootstrapPeer.key)
			if err != nil {
				return errors.Wrap(err, "could not unpack bootstrap node")
			}
			peerInfo := &peers.PeerInfo{}
			peerInfo.FromBlock(peerInfoBlock)
			if err := addressBook.PutPeerInfo(peerInfo); err != nil {
				return errors.Wrap(err, "could not put bootstrap peer")
			}
			if bootstrapPeer.alias == "stats.nimona.io" {
				statsBootstrapPeer = peerInfo
			}
			if daemonEnableRelaying {
				addressBook.AddLocalPeerRelay(peerInfo.Thumbprint())
			}
			addressBook.SetAlias(peerInfo.Signature.Key, bootstrapPeer.alias)
		}

		storagePath := path.Join(daemonConfigPath, "storage")

		dpr := storage.NewDiskStorage(storagePath)
		n, _ := net.NewExchange(addressBook, dpr, fmt.Sprintf("0.0.0.0:%d", daemonPort))
		dht, _ := dht.NewDHT(n, addressBook)
		telemetry.NewTelemetry(n, addressBook.GetLocalPeerKey(),
			statsBootstrapPeer.Signature.Key)

		n.RegisterDiscoverer(dht)

		api := api.New(addressBook, dht, n, dpr)
		err = api.Serve(fmt.Sprintf("0.0.0.0:%d", daemonAPIPort))
		return errors.Wrap(err, "http server stopped")
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)

	daemonCmd.PersistentFlags().IntVar(
		&daemonPort,
		"port",
		0,
		"peer port",
	)

	daemonCmd.PersistentFlags().IntVar(
		&daemonAPIPort,
		"api-port",
		8030,
		"api port",
	)

	daemonCmd.PersistentFlags().BoolVar(
		&daemonEnableRelaying,
		"relaying",
		false,
		"enable relaying through bootstrap peers",
	)

	daemonCmd.PersistentFlags().BoolVar(
		&daemonEnableMetrics,
		"metrics",
		false,
		"enable sending anonymous metrics",
	)
}
