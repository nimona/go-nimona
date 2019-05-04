package cmd

import (
	"github.com/spf13/cobra"

	"nimona.io/internal/errors"
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

		pko := config.Daemon.PeerKey.PublicKey.ToObject()
		sig, err := crypto.NewSignature(
			identityKey,
			crypto.AlgorithmObjectHash,
			pko,
		)
		if err != nil {
			return errors.Wrap(
				errors.New("could not sign peer key"),
				err,
			)
		}

		config.Daemon.IdentityKey = identityKey
		config.Daemon.PeerKey.PublicKey.Signature = sig

		if err := config.Update(cfgFile); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	identityCmd.AddCommand(identityInitCmd)
}
