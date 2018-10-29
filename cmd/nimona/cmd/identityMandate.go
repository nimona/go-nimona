package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"nimona.io/go/base58"
	"nimona.io/go/primitives"
)

var (
	identityMandateResources []string
	identityMandateActions   []string
	identityMandateEffect    string
)

// identityMandateCmd represents the identityMandate command
var identityMandateCmd = &cobra.Command{
	Use:   "mandate",
	Short: "Create a mandate for a peer, given an identity private key and peer public key",
	Long:  "",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(identityMandateResources) == 0 || len(identityMandateActions) == 0 {
			return errors.New("missing policy resources or actions")
		}

		identityKeyBlock, err := primitives.BlockFromBase58(args[0])
		if err != nil {
			return err
		}

		peerKeyBlock, err := primitives.BlockFromBase58(args[1])
		if err != nil {
			return err
		}

		identityKey := &primitives.Key{}
		identityKey.FromBlock(identityKeyBlock)

		peerKey := &primitives.Key{}
		peerKey.FromBlock(peerKeyBlock)

		mandate, err := primitives.NewMandate(
			identityKey,
			peerKey,
			primitives.MandatePolicy{
				Resources: identityMandateResources,
				Actions:   identityMandateActions,
				Effect:    identityMandateEffect,
			},
		)
		if err != nil {
			return err
		}

		b, err := primitives.Marshal(mandate.Block())
		if err != nil {
			return err
		}

		cmd.Println("mandate:", base58.Encode(b))
		cmd.Println("")
		return nil
	},
}

func init() {
	identityCmd.AddCommand(identityMandateCmd)

	identityMandateCmd.PersistentFlags().StringVar(
		&identityMandateEffect,
		"effect",
		"allow",
		"policy effect",
	)

	identityMandateCmd.PersistentFlags().StringSliceVar(
		&identityMandateResources,
		"resources",
		[]string{},
		"policy resources",
	)

	identityMandateCmd.PersistentFlags().StringSliceVar(
		&identityMandateActions,
		"actions",
		[]string{
			"allow",
		},
		"policy actions",
	)
}
