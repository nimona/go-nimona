package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"nimona.io/pkg/encoding"
	"nimona.io/pkg/peers"
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

		if viper.GetBool("raw") {
			bs, err := json.MarshalIndent(ms, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(bs))
			return nil
		}

		for _, peer := range ms {
			cmd.Println("peer:")
			cmd.Println("  id:", peer.HashBase58())
			cmd.Println("  addresses:", peer.Addresses)
			cmd.Println("")
		}

		return nil
	},
}

func init() {
	peerCmd.AddCommand(peerListCmd)
}
