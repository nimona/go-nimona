package cmd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/spf13/cobra"
	"nimona.io/go/primitives"
)

// identityCreateCmd represents the identityCreate command
var identityCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new identity",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		signingKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return err
		}

		identityKey, err := primitives.NewKey(signingKey)
		if err != nil {
			return err
		}

		cmd.Println("identity:")
		cmd.Println("  public key:", identityKey.GetPublicKey().Thumbprint())
		cmd.Println("  private key:", identityKey.Thumbprint())
		cmd.Println("")
		return nil
	},
}

func init() {
	identityCmd.AddCommand(identityCreateCmd)
}
