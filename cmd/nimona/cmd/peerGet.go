package cmd

import (
	"encoding/json"

	"nimona.io/go/peers"

	"github.com/spf13/cobra"
	"nimona.io/go/codec"
	"nimona.io/go/primitives"
)

// peerGetCmd represents the peerGet command
var peerGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a peer's info based on their ID",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := restClient.R().Get("/peers/" + args[0])
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
	peerCmd.AddCommand(peerGetCmd)
}
