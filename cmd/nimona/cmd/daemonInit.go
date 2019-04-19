package cmd

import (
	"os"
	"path"

	cayleyGraph "github.com/cayleygraph/cayley/graph"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	_ "github.com/cayleygraph/cayley/graph/kv/bolt" // required for cayley

	"nimona.io/pkg/crypto"
)

// daemonInitCmd represents the daemon init command
var daemonInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize new peer",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(cfgFile); err == nil {
			return errors.New("daemon already initialized")
		}

		if err := os.MkdirAll(dataDir, 0777); err != nil {
			return errors.Wrap(err, "could not create config dir")
		}

		peerKey, err := crypto.GenerateKey()
		if err != nil {
			return errors.Wrap(err, "could not generate peer key")
		}

		config := &Config{
			Daemon: DaemonConfig{
				ObjectPath: path.Join(dataDir, "objects"),
				Port:       21013,
				PeerKey:    peerKey,
				// IdentityKey: identityKey,
				BootstrapAddresses: []string{
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
				},
			},
		}

		if err := config.Update(cfgFile); err != nil {
			return err
		}

		err = cayleyGraph.InitQuadStore("bolt", config.Daemon.ObjectPath, nil)
		if err != nil {
			return errors.Wrap(err, "could not init quad store")
		}

		return nil
	},
}

func init() {
	daemonCmd.AddCommand(daemonInitCmd)
}
