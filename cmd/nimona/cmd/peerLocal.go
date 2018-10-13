package cmd

import (
	"encoding/json"

	"nimona.io/go/peers"

	"github.com/spf13/cobra"
	"nimona.io/go/codec"
	"nimona.io/go/primitives"
)

// peerLocalCmd represents the peerLocal command
var peerLocalCmd = &cobra.Command{
	Use:   "local",
	Short: "Get local peer information",
	Long:  "",
	Args:  cobra.MaximumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := restClient.R().Get("/local")
		if err != nil {
			return err
		}

		body := resp.Body()
		block := &primitives.Block{}
		if err := codec.Unmarshal(body, block); err != nil {
			return err
		}

		if returnRaw {
			bs, err := json.MarshalIndent(block, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(bs))
			return nil
		}

		peer := &peers.PeerInfo{}
		peer.FromBlock(block)

		cmd.Println("peer:")
		cmd.Println("  id:", peer.Thumbprint())
		cmd.Println("  addresses:", peer.Addresses)
		cmd.Println("")
		return nil
	},
}

func init() {
	peerCmd.AddCommand(peerLocalCmd)
}
