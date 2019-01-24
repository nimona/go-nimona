package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"nimona.io/pkg/object"
	"nimona.io/pkg/net/peer"
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
	peerCmd.AddCommand(peerGetCmd)
}
