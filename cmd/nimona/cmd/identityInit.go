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

		cmd.Println("identity fingerprint:", identityKey.Fingerprint())
		cmd.Println("")

		config.Daemon.IdentityKey = identityKey

		if err := config.Update(cfgFile); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	identityCmd.AddCommand(identityInitCmd)
}
