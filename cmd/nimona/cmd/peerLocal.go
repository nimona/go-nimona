package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
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
		m := map[string]interface{}{}
		if err := json.Unmarshal(body, &m); err != nil {
			return err
		}

		o := &object.Object{}
		if err := o.FromMap(m); err != nil {
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
		cmd.Println("  id:", peer.Fingerprint())
		cmd.Println("  addresses:", peer.Addresses)
		cmd.Println("")
		return nil
	},
}

func init() {
	peerCmd.AddCommand(peerLocalCmd)
}
