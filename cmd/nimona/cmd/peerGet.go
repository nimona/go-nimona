package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
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
		o, err := object.FromBytes(body)
		if err != nil {
			return err
		}

		peer := &peer.PeerInfo{}
		if err := peer.FromObject(o); err != nil {
			return err
		}

		if viper.GetBool("raw") {
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
