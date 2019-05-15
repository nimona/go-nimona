package cmd

import (
	"github.com/spf13/cobra"

	"nimona.io/internal/errors"
)

// peerListCmd represents the peerList command
var peerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known peer info",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("not implemented")

		// resp, err := restClient.R().Get("/peers")
		// if err != nil {
		// 	return err
		// }

		// body := resp.Body()
		// ms := []*peer.PeerInfo{}
		// if err := object.UnmarshalSimple(body, &ms); err != nil {
		// 	return err
		// }

		// if viper.GetBool("raw") {
		// 	bs, err := json.MarshalIndent(ms, "", "  ")
		// 	if err != nil {
		// 		return err
		// 	}

		// 	cmd.Println(string(bs))
		// 	return nil
		// }

		// for _, peer := range ms {
		// 	cmd.Println("peer:")
		// 	cmd.Println("  id:", peer.Fingerprint())
		// 	cmd.Println("  addresses:", peer.Addresses)
		// 	cmd.Println("")
		// }

		// return nil
	},
}

func init() {
	peerCmd.AddCommand(peerListCmd)
}
