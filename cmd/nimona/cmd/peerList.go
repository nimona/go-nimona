package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"nimona.io/go/codec"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
)

// peerListCmd represents the peerList command
var peerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known peer info",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := restClient.R().Get("/peers")
		if err != nil {
			return err
		}

		body := resp.Body()
		blocks := []*primitives.Block{}
		if err := codec.Unmarshal(body, &blocks); err != nil {
			return err
		}

		if returnRaw {
			bs, err := json.MarshalIndent(blocks, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(bs))
			return nil
		}

		for _, block := range blocks {
			peer := &peers.PeerInfo{}
			peer.FromBlock(block)

			cmd.Println("peer:")
			cmd.Println("  id:", peer.Thumbprint())
			cmd.Println("  addresses:", peer.Addresses)
			cmd.Println("")
		}

		return nil
	},
}

func init() {
	peerCmd.AddCommand(peerListCmd)
}
