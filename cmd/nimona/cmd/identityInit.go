package cmd

import (
	"github.com/spf13/cobra"

	"nimona.io/pkg/crypto"
)

// identityInitCmd represents the identityInit command
var identityInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Init a new identity",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		identityKey, err := crypto.GenerateKey()
		if err != nil {
			return err
		}

		mandate, err := crypto.NewMandate(
			identityKey,
			config.Daemon.PeerKey,
			"default mandate",
			[]string{"*"},
			[]string{"*"},
			"allow",
		)
		if err != nil {
			return err
		}

		cmd.Println("identity:")
		cmd.Println("  private key:", identityKey.HashBase58())
		cmd.Println("  public key:", identityKey.GetPublicKey().HashBase58())
		cmd.Println("")

		config.Daemon.IdentityKey = identityKey
		config.Daemon.Mandate = mandate

		if err := config.Update(cfgFile); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	identityCmd.AddCommand(identityInitCmd)
}
