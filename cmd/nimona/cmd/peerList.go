package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"nimona.io/go/encoding"
	"nimona.io/go/peers"
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
		ms := []*peers.PeerInfo{}
		if err := encoding.UnmarshalSimple(body, &ms); err != nil {
			return err
		}

		if returnRaw {
			bs, err := json.MarshalIndent(ms, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(bs))
			return nil
		}

		for _, peer := range ms {
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
