package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"nimona.io/pkg/object"
	"nimona.io/pkg/net/peer"
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
		o, err := object.NewObjectFromBytes(body)
		if err != nil {
			return err
		}

		peer := &peer.PeerInfo{}
		if err := peer.FromObject(o); err != nil {
			return err
		}

		if returnRaw {
			bs, err := json.MarshalIndent(peer, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(bs))
			return nil
		}

		cmd.Println("peer:")
		cmd.Println("  id:", peer.HashBase58())
		cmd.Println("  addresses:", peer.Addresses)
		cmd.Println("")
		return nil
	},
}

func init() {
	peerCmd.AddCommand(peerLocalCmd)
}
